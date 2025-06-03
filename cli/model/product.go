package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/icrowley/fake"
	"github.com/jaswdr/faker"
)

type Product struct {
	ProductID    uuid.UUID `gorm:"column:product_id;primaryKey"`
	CategoryID   uuid.UUID `gorm:"column:category_id"`
	ProductName  string    `gorm:"column:product_name"`
	CurrentStock float64   `gorm:"column:current_stock"`
	UnitPrice    float64   `gorm:"column:unit_price"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
}

func (Product) TableName() string {
	return "dm_core.product"
}

func FakeProduct(f faker.Faker) Product {
	return Product{
		ProductID:    uuid.New(),
		CategoryID:   uuid.New(), // TODO: substituir por uuid de categoria existente
		ProductName:  fake.ProductName(),
		CurrentStock: f.RandomFloat(2, 0, 1000),
		UnitPrice:    f.RandomFloat(2, 10, 500),
	}
}
