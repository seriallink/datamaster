package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/jaswdr/faker"
)

type PurchaseItem struct {
	PurchaseID uuid.UUID `gorm:"column:purchase_id;primaryKey"`
	ProductID  uuid.UUID `gorm:"column:product_id;primaryKey"`
	Quantity   float64   `gorm:"column:quantity"`
	UnitPrice  float64   `gorm:"column:unit_price"`
	TotalPrice float64   `gorm:"column:total_price"`
	CreatedAt  time.Time `gorm:"column:created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at"`
}

func (PurchaseItem) TableName() string {
	return "dm_core.purchase_item"
}

func FakePurchaseItem(f faker.Faker, productID, purchaseID uuid.UUID) PurchaseItem {
	qty := f.Float64(0, 1, 10)
	unit := f.Float64(2, 10, 500)
	return PurchaseItem{
		PurchaseID: purchaseID,
		ProductID:  productID,
		Quantity:   qty,
		UnitPrice:  unit,
		TotalPrice: qty * unit,
	}
}
