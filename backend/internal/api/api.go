package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cng1985/erpgo/backend/internal/config"
	"github.com/cng1985/erpgo/backend/internal/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type API struct {
	db       *gorm.DB
	location *time.Location
}

func New(db *gorm.DB, cfg config.Config) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery(), cors(cfg.AllowedOrigins))
	location := time.Local
	if cfg.TimeZone != "" {
		if configured, err := time.LoadLocation(cfg.TimeZone); err == nil {
			location = configured
		}
	}
	a := &API{db: db, location: location}
	r.GET("/api/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	v1 := r.Group("/api")
	v1.GET("/dashboard", a.dashboard)
	v1.GET("/customers", a.listCustomers)
	v1.POST("/customers", a.createCustomer)
	v1.GET("/products", a.listProducts)
	v1.POST("/products", a.createProduct)
	v1.GET("/orders", a.listOrders)
	v1.POST("/orders", a.createOrder)
	v1.PATCH("/orders/:id/status", a.updateOrderStatus)
	v1.GET("/inventory", a.listInventory)
	v1.POST("/inventory/adjust", a.adjustInventory)
	v1.GET("/purchases", a.listPurchases)
	v1.POST("/purchases/receive", a.receivePurchase)
	return r
}

func cors(origins []string) gin.HandlerFunc {
	allowed := map[string]bool{}
	for _, origin := range origins {
		if origin != "" {
			allowed[origin] = true
		}
	}
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if allowed[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func (a *API) dashboard(c *gin.Context) {
	now := time.Now().In(a.location)
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	var today int64
	var pending, lowStock int64
	var receivable int64
	if err := a.db.Model(&model.SalesOrder{}).Where("created_at >= ? AND created_at < ? AND status <> ?", start, start.AddDate(0, 0, 1), "cancelled").Select("COALESCE(SUM(total_cents),0)").Scan(&today).Error; err != nil {
		fail(c, err)
		return
	}
	if err := a.db.Model(&model.SalesOrder{}).Where("status = ?", "pending").Count(&pending).Error; err != nil {
		fail(c, err)
		return
	}
	if err := a.db.Model(&model.Product{}).Where("stock <= reorder_level").Count(&lowStock).Error; err != nil {
		fail(c, err)
		return
	}
	if err := a.db.Model(&model.SalesOrder{}).Where("status = ?", "shipped").Select("COALESCE(SUM(total_cents),0)").Scan(&receivable).Error; err != nil {
		fail(c, err)
		return
	}
	trendStart := start.AddDate(0, 0, -6)
	_, offset := now.Zone()
	dateModifier := fmt.Sprintf("%+d seconds", offset)
	dateExpression := fmt.Sprintf("date(created_at, '%s')", dateModifier)
	var rows []struct {
		Date   string `json:"date"`
		Amount int64  `json:"amount"`
	}
	if err := a.db.Model(&model.SalesOrder{}).Select(dateExpression+" as date, COALESCE(SUM(total_cents),0) as amount").Where("created_at >= ? AND created_at < ? AND status <> ?", trendStart, start.AddDate(0, 0, 1), "cancelled").Group(dateExpression).Scan(&rows).Error; err != nil {
		fail(c, err)
		return
	}
	values := map[string]int64{}
	for _, row := range rows {
		values[row.Date] = row.Amount
	}
	trend := make([]gin.H, 0, 7)
	for i := 0; i < 7; i++ {
		day := trendStart.AddDate(0, 0, i)
		key := day.Format("2006-01-02")
		trend = append(trend, gin.H{"date": key, "amount": values[key]})
	}
	var orders []model.SalesOrder
	if err := a.db.Preload("Customer").Preload("Items.Product").Order("created_at DESC").Limit(5).Find(&orders).Error; err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"today_sales_cents": today, "pending_orders": pending, "low_stock_products": lowStock, "receivable_cents": receivable, "sales_trend": trend, "recent_orders": orders})
}

func (a *API) listCustomers(c *gin.Context) {
	var items []model.Customer
	q := a.db.Model(&model.Customer{})
	if s := strings.TrimSpace(c.Query("q")); s != "" {
		like := "%" + s + "%"
		q = q.Where("name LIKE ? OR contact LIKE ? OR phone LIKE ?", like, like, like)
	}
	var total int64
	q.Count(&total)
	if err := q.Order("id DESC").Scopes(pageScope(c)).Find(&items).Error; err != nil {
		fail(c, err)
		return
	}
	list(c, items, total)
}

func (a *API) createCustomer(c *gin.Context) {
	var in struct {
		Name    string `json:"name"`
		Contact string `json:"contact"`
		Phone   string `json:"phone"`
		Address string `json:"address"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		bad(c, err)
		return
	}
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		badMessage(c, "客户名称不能为空")
		return
	}
	item := model.Customer{Name: in.Name, Contact: strings.TrimSpace(in.Contact), Phone: strings.TrimSpace(in.Phone), Address: strings.TrimSpace(in.Address)}
	if err := a.db.Create(&item).Error; err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusCreated, item)
}

func (a *API) listProducts(c *gin.Context) {
	var items []model.Product
	q := a.db.Model(&model.Product{})
	if s := strings.TrimSpace(c.Query("q")); s != "" {
		like := "%" + s + "%"
		q = q.Where("name LIKE ? OR sku LIKE ? OR category LIKE ?", like, like, like)
	}
	if c.Query("low_stock") == "true" {
		q = q.Where("stock <= reorder_level")
	}
	var total int64
	q.Count(&total)
	if err := q.Order("id DESC").Scopes(pageScope(c)).Find(&items).Error; err != nil {
		fail(c, err)
		return
	}
	list(c, items, total)
}

func (a *API) createProduct(c *gin.Context) {
	var in struct {
		SKU          string `json:"sku"`
		Name         string `json:"name"`
		Category     string `json:"category"`
		Unit         string `json:"unit"`
		PriceCents   int64  `json:"price_cents"`
		CostCents    int64  `json:"cost_cents"`
		Stock        int    `json:"stock"`
		ReorderLevel int    `json:"reorder_level"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		bad(c, err)
		return
	}
	in.SKU = strings.TrimSpace(in.SKU)
	in.Name = strings.TrimSpace(in.Name)
	if in.SKU == "" || in.Name == "" {
		badMessage(c, "商品编码和名称不能为空")
		return
	}
	if in.PriceCents < 0 || in.CostCents < 0 || in.Stock < 0 || in.ReorderLevel < 0 {
		badMessage(c, "价格和库存不能为负数")
		return
	}
	if in.Unit == "" {
		in.Unit = "件"
	}
	item := model.Product{SKU: in.SKU, Name: in.Name, Category: in.Category, Unit: in.Unit, PriceCents: in.PriceCents, CostCents: in.CostCents, Stock: in.Stock, ReorderLevel: in.ReorderLevel}
	if err := a.db.Create(&item).Error; err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusCreated, item)
}

func (a *API) listOrders(c *gin.Context) {
	var items []model.SalesOrder
	q := a.db.Model(&model.SalesOrder{}).Preload("Customer").Preload("Items.Product")
	if s := strings.TrimSpace(c.Query("q")); s != "" {
		q = q.Joins("LEFT JOIN customers ON customers.id = sales_orders.customer_id").Where("sales_orders.number LIKE ? OR customers.name LIKE ?", "%"+s+"%", "%"+s+"%")
	}
	if s := c.Query("status"); s != "" {
		q = q.Where("sales_orders.status = ?", s)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		fail(c, err)
		return
	}
	if err := q.Order("sales_orders.created_at DESC").Scopes(pageScope(c)).Find(&items).Error; err != nil {
		fail(c, err)
		return
	}
	list(c, items, total)
}

func (a *API) createOrder(c *gin.Context) {
	var in struct {
		CustomerID uint   `json:"customer_id"`
		Note       string `json:"note"`
		Items      []struct {
			ProductID uint `json:"product_id"`
			Quantity  int  `json:"quantity"`
		} `json:"items"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		bad(c, err)
		return
	}
	if in.CustomerID == 0 || len(in.Items) == 0 {
		badMessage(c, "请选择客户并添加至少一项商品")
		return
	}
	order := model.SalesOrder{Number: "SO" + time.Now().Format("20060102150405.000000"), CustomerID: in.CustomerID, Status: "pending", Note: strings.TrimSpace(in.Note)}
	err := a.db.Transaction(func(tx *gorm.DB) error {
		var customer model.Customer
		if err := tx.First(&customer, in.CustomerID).Error; err != nil {
			return err
		}
		for _, line := range in.Items {
			if line.ProductID == 0 || line.Quantity < 1 {
				return fmt.Errorf("商品数量必须大于 0")
			}
			var product model.Product
			if err := tx.First(&product, line.ProductID).Error; err != nil {
				return err
			}
			if product.Stock < line.Quantity {
				return fmt.Errorf("商品 %s 库存不足", product.Name)
			}
			subtotal := product.PriceCents * int64(line.Quantity)
			order.TotalCents += subtotal
			order.Items = append(order.Items, model.OrderItem{ProductID: product.ID, Quantity: line.Quantity, UnitPriceCents: product.PriceCents, SubtotalCents: subtotal})
		}
		if err := tx.Create(&order).Error; err != nil {
			return err
		}
		for _, line := range order.Items {
			res := tx.Model(&model.Product{}).Where("id = ? AND stock >= ?", line.ProductID, line.Quantity).UpdateColumn("stock", gorm.Expr("stock - ?", line.Quantity))
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected != 1 {
				return fmt.Errorf("库存已变化，请刷新后重试")
			}
			if err := tx.Create(&model.StockMovement{ProductID: line.ProductID, Kind: "sale_reserved", Quantity: -line.Quantity, Reference: order.Number, Note: "销售订单预占"}).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		badMessage(c, err.Error())
		return
	}
	a.db.Preload("Customer").Preload("Items.Product").First(&order, order.ID)
	c.JSON(http.StatusCreated, order)
}

func (a *API) updateOrderStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id < 1 {
		badMessage(c, "订单编号无效")
		return
	}
	var in struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		bad(c, err)
		return
	}
	err = a.db.Transaction(func(tx *gorm.DB) error {
		var order model.SalesOrder
		if err := tx.Preload("Items").First(&order, id).Error; err != nil {
			return err
		}
		valid := order.Status == "pending" && (in.Status == "shipped" || in.Status == "cancelled") || order.Status == "shipped" && in.Status == "completed"
		if !valid {
			return fmt.Errorf("订单状态不能从 %s 变更为 %s", order.Status, in.Status)
		}
		if in.Status == "cancelled" {
			for _, line := range order.Items {
				if err := tx.Model(&model.Product{}).Where("id = ?", line.ProductID).UpdateColumn("stock", gorm.Expr("stock + ?", line.Quantity)).Error; err != nil {
					return err
				}
				if err := tx.Create(&model.StockMovement{ProductID: line.ProductID, Kind: "sale_cancelled", Quantity: line.Quantity, Reference: order.Number, Note: "取消订单返还库存"}).Error; err != nil {
					return err
				}
			}
		}
		return tx.Model(&order).Update("status", in.Status).Error
	})
	if err != nil {
		badMessage(c, err.Error())
		return
	}
	var order model.SalesOrder
	if err := a.db.Preload("Customer").Preload("Items.Product").First(&order, id).Error; err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, order)
}

func (a *API) listInventory(c *gin.Context) {
	var items []model.Product
	q := a.db.Model(&model.Product{}).Select("*")
	if c.Query("low_stock") == "true" {
		q = q.Where("stock <= reorder_level")
	}
	if s := strings.TrimSpace(c.Query("q")); s != "" {
		like := "%" + s + "%"
		q = q.Where("name LIKE ? OR sku LIKE ?", like, like)
	}
	var total int64
	q.Count(&total)
	if err := q.Order("CASE WHEN stock <= reorder_level THEN 0 ELSE 1 END, stock ASC").Scopes(pageScope(c)).Find(&items).Error; err != nil {
		fail(c, err)
		return
	}
	list(c, items, total)
}

func (a *API) adjustInventory(c *gin.Context) {
	var in struct {
		ProductID uint   `json:"product_id"`
		Quantity  int    `json:"quantity"`
		Note      string `json:"note"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		bad(c, err)
		return
	}
	if in.ProductID == 0 || in.Quantity == 0 {
		badMessage(c, "请选择商品并填写调整数量")
		return
	}
	err := a.db.Transaction(func(tx *gorm.DB) error {
		var p model.Product
		if err := tx.First(&p, in.ProductID).Error; err != nil {
			return err
		}
		if p.Stock+in.Quantity < 0 {
			return fmt.Errorf("调整后库存不能小于 0")
		}
		if err := tx.Model(&p).UpdateColumn("stock", gorm.Expr("stock + ?", in.Quantity)).Error; err != nil {
			return err
		}
		kind := "stock_in"
		if in.Quantity < 0 {
			kind = "stock_out"
		}
		return tx.Create(&model.StockMovement{ProductID: p.ID, Kind: kind, Quantity: in.Quantity, Note: strings.TrimSpace(in.Note)}).Error
	})
	if err != nil {
		badMessage(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (a *API) listPurchases(c *gin.Context) {
	var items []model.PurchaseReceipt
	q := a.db.Model(&model.PurchaseReceipt{}).Preload("Product")
	var total int64
	q.Count(&total)
	if err := q.Order("created_at DESC").Scopes(pageScope(c)).Find(&items).Error; err != nil {
		fail(c, err)
		return
	}
	list(c, items, total)
}

func (a *API) receivePurchase(c *gin.Context) {
	var in struct {
		ProductID     uint   `json:"product_id"`
		Quantity      int    `json:"quantity"`
		UnitCostCents int64  `json:"unit_cost_cents"`
		Supplier      string `json:"supplier"`
		Note          string `json:"note"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		bad(c, err)
		return
	}
	if in.ProductID == 0 || in.Quantity < 1 || in.UnitCostCents < 0 {
		badMessage(c, "商品、正数入库数量和有效成本为必填项")
		return
	}
	receipt := model.PurchaseReceipt{Number: "PO" + time.Now().Format("20060102150405.000000"), ProductID: in.ProductID, Quantity: in.Quantity, UnitCostCents: in.UnitCostCents, Supplier: strings.TrimSpace(in.Supplier), Note: strings.TrimSpace(in.Note)}
	err := a.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&receipt).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.Product{}).Where("id = ?", in.ProductID).UpdateColumn("stock", gorm.Expr("stock + ?", in.Quantity)).Error; err != nil {
			return err
		}
		return tx.Create(&model.StockMovement{ProductID: in.ProductID, Kind: "purchase", Quantity: in.Quantity, Reference: receipt.Number, Note: "采购入库"}).Error
	})
	if err != nil {
		fail(c, err)
		return
	}
	a.db.Preload("Product").First(&receipt, receipt.ID)
	c.JSON(http.StatusCreated, receipt)
}

func pageScope(c *gin.Context) func(*gorm.DB) *gorm.DB {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	return func(db *gorm.DB) *gorm.DB { return db.Offset((page - 1) * size).Limit(size) }
}
func list(c *gin.Context, items any, total int64) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": total, "page": page, "page_size": size})
}
func fail(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	if err == gorm.ErrRecordNotFound {
		status = http.StatusNotFound
	}
	c.JSON(status, gin.H{"error": err.Error()})
}
func bad(c *gin.Context, err error) { badMessage(c, err.Error()) }
func badMessage(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, gin.H{"error": message})
}
