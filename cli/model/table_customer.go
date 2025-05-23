package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/jaswdr/faker"
)

type Customer struct {
	CustomerID    uuid.UUID  `gorm:"column:customer_id;primaryKey"`
	CustomerName  string     `gorm:"column:customer_name"`
	CustomerEmail string     `gorm:"column:customer_email"`
	CustomerPhone string     `gorm:"column:customer_phone"`
	Gender        GenderEnum `gorm:"column:gender"`
	CreatedAt     time.Time  `gorm:"column:created_at"`
	UpdatedAt     time.Time  `gorm:"column:updated_at"`
}

func (Customer) TableName() string {
	return "dm_core.customer"
}

func FakeCustomer(f faker.Faker) Customer {
	return Customer{
		CustomerID:    uuid.New(),
		CustomerName:  f.Person().Name(),
		CustomerEmail: f.Internet().Email(),
		CustomerPhone: f.Phone().Number(),
		Gender:        RandomEnum[GenderEnum](f, GenderEnum("").Values()),
	}
}
