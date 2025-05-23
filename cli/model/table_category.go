package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/jaswdr/faker"
)

type Category struct {
	CategoryID          uuid.UUID `gorm:"column:category_id;primaryKey"`
	CategoryName        string    `gorm:"column:category_name"`
	CategoryDescription string    `gorm:"column:category_description"`
	CreatedAt           time.Time `gorm:"column:created_at"`
	UpdatedAt           time.Time `gorm:"column:updated_at"`
}

func (Category) TableName() string {
	return "dm_core.category"
}

func FakeCategory(f faker.Faker) Category {
	return Category{
		CategoryID:          uuid.New(),
		CategoryName:        f.Lorem().Word(),
		CategoryDescription: f.Lorem().Sentence(10),
	}
}
