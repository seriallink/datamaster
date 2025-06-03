package model

import (
	"time"

	"github.com/seriallink/datamaster/app/enum"

	"github.com/google/uuid"
	"github.com/jaswdr/faker"
)

func init() {
	Register(&Customer{})
}

type Customer struct {
	CustomerID    uuid.UUID           `json:"customer_id"    gorm:"column:customer_id;primaryKey"`
	CustomerName  string              `json:"customer_name"  gorm:"column:customer_name"`
	CustomerEmail string              `json:"customer_email" gorm:"column:customer_email"`
	CustomerPhone string              `json:"customer_phone" gorm:"column:customer_phone"`
	Gender        enum.GenderTypeEnum `json:"gender"         gorm:"column:gender"`
	CreatedAt     time.Time           `json:"created_at"     gorm:"column:created_at"`
	UpdatedAt     time.Time           `json:"updated_at"     gorm:"column:updated_at"`
}

func (m *Customer) TableName() string {
	return "dm_core.customer"
}

func (m *Customer) TableId() uuid.UUID {
	return uuid.MustParse("53a8770f-a07e-4b77-8d9e-41785f0b7b25")
}

func FakeCustomer(f faker.Faker) Customer {
	return Customer{
		CustomerID:    uuid.New(),
		CustomerName:  f.Person().Name(),
		CustomerEmail: f.Internet().Email(),
		CustomerPhone: f.Phone().Number(),
		Gender:        enum.RandomEnum[enum.GenderTypeEnum](f, enum.GenderTypeEnum("").Values()),
	}
}
