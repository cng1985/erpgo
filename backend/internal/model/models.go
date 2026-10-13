package model

import (
	"time"

	"gorm.io/gorm"
)

type Customer struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Name      string         `json:"name" gorm:"size:120;not null;index"`
	Contact   string         `json:"contact" gorm:"size:80"`
	Phone     string         `json:"phone" gorm:"size:40"`
	Address   string         `json:"address" gorm:"size:240"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type Product struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	SKU          string         `json:"sku" gorm:"size:64;not null;uniqueIndex"`
	Name         string         `json:"name" gorm:"size:160;not null;index"`
	Category     string         `json:"category" gorm:"size:80"`
	Unit         string         `json:"unit" gorm:"size:20;default:件"`
	PriceCents   int64          `json:"price_cents" gorm:"not null"`
	CostCents    int64          `json:"cost_cents" gorm:"not null"`
	Stock        int            `json:"stock" gorm:"not null;default:0"`
	ReorderLevel int            `json:"reorder_level" gorm:"not null;default:0"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

type SalesOrder struct {
	ID         uint        `json:"id" gorm:"primaryKey"`
	Number     string      `json:"number" gorm:"size:40;not null;uniqueIndex"`
	CustomerID uint        `json:"customer_id" gorm:"not null;index"`
	Customer   Customer    `json:"customer"`
	Status     string      `json:"status" gorm:"size:20;not null;index"`
	TotalCents int64       `json:"total_cents" gorm:"not null"`
	Note       string      `json:"note" gorm:"size:500"`
	Items      []OrderItem `json:"items" gorm:"constraint:OnDelete:CASCADE"`
	CreatedAt  time.Time   `json:"created_at" gorm:"index"`
	UpdatedAt  time.Time   `json:"updated_at"`
}

type OrderItem struct {
	ID             uint    `json:"id" gorm:"primaryKey"`
	SalesOrderID   uint    `json:"sales_order_id" gorm:"not null;index"`
	ProductID      uint    `json:"product_id" gorm:"not null;index"`
	Product        Product `json:"product"`
	Quantity       int     `json:"quantity" gorm:"not null"`
	UnitPriceCents int64   `json:"unit_price_cents" gorm:"not null"`
	SubtotalCents  int64   `json:"subtotal_cents" gorm:"not null"`
}

type PurchaseReceipt struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	Number        string    `json:"number" gorm:"size:40;not null;uniqueIndex"`
	ProductID     uint      `json:"product_id" gorm:"not null;index"`
	Product       Product   `json:"product"`
	Quantity      int       `json:"quantity" gorm:"not null"`
	UnitCostCents int64     `json:"unit_cost_cents" gorm:"not null"`
	Supplier      string    `json:"supplier" gorm:"size:160"`
	Note          string    `json:"note" gorm:"size:500"`
	CreatedAt     time.Time `json:"created_at" gorm:"index"`
}

type StockMovement struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	ProductID uint      `json:"product_id" gorm:"not null;index"`
	Product   Product   `json:"product"`
	Kind      string    `json:"kind" gorm:"size:24;not null"`
	Quantity  int       `json:"quantity" gorm:"not null"`
	Reference string    `json:"reference" gorm:"size:64"`
	Note      string    `json:"note" gorm:"size:500"`
	CreatedAt time.Time `json:"created_at" gorm:"index"`
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&Customer{}, &Product{}, &SalesOrder{}, &OrderItem{}, &PurchaseReceipt{}, &StockMovement{})
}
