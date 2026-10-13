package store

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/cng1985/erpgo/backend/internal/config"
	"github.com/cng1985/erpgo/backend/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Open(cfg config.Config) (*gorm.DB, error) {
	if cfg.DatabasePath != ":memory:" {
		if err := os.MkdirAll(filepath.Dir(cfg.DatabasePath), 0o755); err != nil {
			return nil, fmt.Errorf("create database directory: %w", err)
		}
	}
	db, err := gorm.Open(sqlite.Open(cfg.DatabasePath+"?_busy_timeout=5000&_foreign_keys=on"), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}
	if err := db.Exec("PRAGMA journal_mode=WAL").Error; err != nil {
		return nil, fmt.Errorf("enable sqlite WAL: %w", err)
	}
	return db, nil
}

func Seed(db *gorm.DB) error {
	var count int64
	if err := db.Model(&model.Customer{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	customers := []model.Customer{
		{Name: "深圳市华信科技有限公司", Contact: "林经理", Phone: "13800001001", Address: "深圳市南山区科技园"},
		{Name: "杭州迅达电子有限公司", Contact: "周女士", Phone: "13800001002", Address: "杭州市滨江区"},
		{Name: "苏州联创智能科技有限公司", Contact: "陈先生", Phone: "13800001003", Address: "苏州市工业园区"},
		{Name: "南京博远信息技术有限公司", Contact: "赵经理", Phone: "13800001004", Address: "南京市雨花台区"},
		{Name: "上海天创实业有限公司", Contact: "王女士", Phone: "13800001005", Address: "上海市闵行区"},
	}
	products := []model.Product{
		{SKU: "SP0001", Name: "A4 复印纸 70g", Category: "办公耗材", Unit: "箱", PriceCents: 24800, CostCents: 18800, Stock: 18, ReorderLevel: 20},
		{SKU: "SP0002", Name: "无线键盘套装", Category: "电脑配件", Unit: "套", PriceCents: 15900, CostCents: 11200, Stock: 86, ReorderLevel: 20},
		{SKU: "SP0003", Name: "激光打印机墨粉", Category: "办公耗材", Unit: "支", PriceCents: 32800, CostCents: 24500, Stock: 6, ReorderLevel: 10},
		{SKU: "SP0004", Name: "千兆交换机 8 口", Category: "网络设备", Unit: "台", PriceCents: 38900, CostCents: 29200, Stock: 4, ReorderLevel: 8},
		{SKU: "SP0005", Name: "USB-C 扩展坞", Category: "电脑配件", Unit: "个", PriceCents: 21900, CostCents: 15800, Stock: 42, ReorderLevel: 12},
		{SKU: "SP0006", Name: "标签打印纸", Category: "办公耗材", Unit: "卷", PriceCents: 3200, CostCents: 1800, Stock: 120, ReorderLevel: 30},
	}
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&customers).Error; err != nil {
			return err
		}
		if err := tx.Create(&products).Error; err != nil {
			return err
		}
		today := time.Now()
		amounts := []int64{856000, 632000, 250000, 200000, 160000, 140000, 130000, 120000, 110000, 100000, 90000, 76000}
		for i, amount := range amounts {
			order := model.SalesOrder{Number: fmt.Sprintf("SO%s%03d", today.Format("20060102"), i+1), CustomerID: uint(i%len(customers)) + 1, Status: "pending", TotalCents: amount, CreatedAt: today.Add(-time.Duration(i) * time.Minute)}
			order.Items = []model.OrderItem{{ProductID: uint(i%len(products)) + 1, Quantity: 1, UnitPriceCents: amount, SubtotalCents: amount}}
			if err := tx.Create(&order).Error; err != nil {
				return err
			}
		}
		pastAmounts := []int64{1800000, 2500000, 2200000, 3200000, 2800000, 3600000}
		for i, amount := range pastAmounts {
			created := today.AddDate(0, 0, -(i + 1)).Add(12 * time.Hour)
			order := model.SalesOrder{Number: fmt.Sprintf("SO%sP%02d", created.Format("20060102"), i+1), CustomerID: uint(i%len(customers)) + 1, Status: "completed", TotalCents: amount, CreatedAt: created}
			order.Items = []model.OrderItem{{ProductID: uint(i%len(products)) + 1, Quantity: 1, UnitPriceCents: amount, SubtotalCents: amount}}
			if err := tx.Create(&order).Error; err != nil {
				return err
			}
		}
		shippedAt := today.AddDate(0, 0, -8)
		shipped := model.SalesOrder{Number: "SO" + shippedAt.Format("20060102") + "900", CustomerID: 1, Status: "shipped", TotalCents: 1680000, CreatedAt: shippedAt}
		shipped.Items = []model.OrderItem{{ProductID: 2, Quantity: 1, UnitPriceCents: shipped.TotalCents, SubtotalCents: shipped.TotalCents}}
		if err := tx.Create(&shipped).Error; err != nil {
			return err
		}
		return nil
	})
}
