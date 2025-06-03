package model

import (
	"time"

	"github.com/seriallink/datamaster/cli/enum"

	"github.com/google/uuid"
	"github.com/jaswdr/faker"
)

type Purchase struct {
	PurchaseID     uuid.UUID               `gorm:"column:purchase_id;primaryKey"`
	CustomerID     uuid.UUID               `gorm:"column:customer_id"`
	PurchaseStatus enum.PurchaseStatusEnum `gorm:"column:purchase_status"`
	TotalValue     float64                 `gorm:"column:total_value"`
	PurchaseDate   time.Time               `gorm:"column:purchase_date"`
	CreatedAt      time.Time               `gorm:"column:created_at"`
	UpdatedAt      time.Time               `gorm:"column:updated_at"`
}

func (Purchase) TableName() string {
	return "dm_core.purchase"
}

func FakePurchase(f faker.Faker) Purchase {
	return Purchase{
		PurchaseID:     uuid.New(),
		CustomerID:     uuid.New(), // Deve ser vinculado a Customer real
		PurchaseStatus: enum.RandomEnum[enum.PurchaseStatusEnum](f, enum.PurchaseStatusEnum("").Values()),
		TotalValue:     f.RandomFloat(2, 50, 1500),
		PurchaseDate:   time.Now().AddDate(0, 0, -f.IntBetween(0, 30)), // até 30 dias atrás
	}
}
