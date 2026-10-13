package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/cng1985/erpgo/backend/internal/config"
	"github.com/cng1985/erpgo/backend/internal/model"
	"github.com/cng1985/erpgo/backend/internal/store"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func testHandler(t *testing.T) (*gorm.DB, http.Handler) {
	t.Helper()
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatal(err)
	}
	time.Local = location
	databaseName := strings.NewReplacer("/", "-").Replace(t.Name())
	db, err := gorm.Open(sqlite.Open("file:erp-"+databaseName+"?mode=memory&cache=shared&_busy_timeout=5000&_foreign_keys=on"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := model.Migrate(db); err != nil {
		t.Fatal(err)
	}
	if err := store.Seed(db); err != nil {
		t.Fatal(err)
	}
	return db, New(db, config.Config{AllowedOrigins: []string{"http://localhost:5173"}})
}

func TestOrderCreationReservesStockAndCancellationRestoresIt(t *testing.T) {
	db, handler := testHandler(t)
	var product model.Product
	if err := db.First(&product, 1).Error; err != nil {
		t.Fatal(err)
	}
	initialStock := product.Stock
	body := []byte(`{"customer_id":1,"items":[{"product_id":1,"quantity":2}]}`)
	create := httptest.NewRecorder()
	handler.ServeHTTP(create, httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewReader(body)))
	if create.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body = %s", create.Code, create.Body.String())
	}
	var order struct {
		ID    uint  `json:"id"`
		Total int64 `json:"total_cents"`
	}
	if err := json.Unmarshal(create.Body.Bytes(), &order); err != nil {
		t.Fatal(err)
	}
	if order.Total != product.PriceCents*2 {
		t.Fatalf("total = %d, want %d", order.Total, product.PriceCents*2)
	}
	if err := db.First(&product, product.ID).Error; err != nil {
		t.Fatal(err)
	}
	if product.Stock != initialStock-2 {
		t.Fatalf("reserved stock = %d, want %d", product.Stock, initialStock-2)
	}
	patchBody := []byte(`{"status":"cancelled"}`)
	cancel := httptest.NewRecorder()
	handler.ServeHTTP(cancel, httptest.NewRequest(http.MethodPatch, "/api/orders/"+uintString(order.ID)+"/status", bytes.NewReader(patchBody)))
	if cancel.Code != http.StatusOK {
		t.Fatalf("cancel status = %d, body = %s", cancel.Code, cancel.Body.String())
	}
	if err := db.First(&product, product.ID).Error; err != nil {
		t.Fatal(err)
	}
	if product.Stock != initialStock {
		t.Fatalf("restored stock = %d, want %d", product.Stock, initialStock)
	}
}

func TestHealthAndDashboardReturnSeededData(t *testing.T) {
	_, handler := testHandler(t)
	health := httptest.NewRecorder()
	handler.ServeHTTP(health, httptest.NewRequest(http.MethodGet, "/api/health", nil))
	if health.Code != http.StatusOK || !bytes.Contains(health.Body.Bytes(), []byte(`"status":"ok"`)) {
		t.Fatalf("health response: %d %s", health.Code, health.Body.String())
	}
	dashboard := httptest.NewRecorder()
	handler.ServeHTTP(dashboard, httptest.NewRequest(http.MethodGet, "/api/dashboard", nil))
	if dashboard.Code != http.StatusOK {
		t.Fatalf("dashboard status = %d, body = %s", dashboard.Code, dashboard.Body.String())
	}
	var data map[string]any
	if err := json.Unmarshal(dashboard.Body.Bytes(), &data); err != nil {
		t.Fatal(err)
	}
	if data["today_sales_cents"] != float64(2864000) {
		t.Fatalf("today sales = %v", data["today_sales_cents"])
	}
	if data["pending_orders"] != float64(12) {
		t.Fatalf("pending orders = %v", data["pending_orders"])
	}
}

func uintString(value uint) string {
	if value == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for value > 0 {
		i--
		buf[i] = byte('0' + value%10)
		value /= 10
	}
	return string(buf[i:])
}
