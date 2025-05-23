package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/jaswdr/faker"
)

type Courier struct {
	CourierID           uuid.UUID     `gorm:"column:courier_id;primaryKey"`
	CourierName         string        `gorm:"column:courier_name"`
	ContactName         string        `gorm:"column:contact_name"`
	ContactEmail        string        `gorm:"column:contact_email"`
	ContactPhone        string        `gorm:"column:contact_phone"`
	CostPerKM           float64       `gorm:"column:cost_per_km"`
	AverageDeliveryTime time.Duration `gorm:"column:average_delivery_time"`
	ReliabilityScore    int16         `gorm:"column:reliability_score"`
	CreatedAt           time.Time     `gorm:"column:created_at"`
	UpdatedAt           time.Time     `gorm:"column:updated_at"`
}

func (Courier) TableName() string {
	return "dm_core.courier"
}

func FakeCourier(f faker.Faker) Courier {
	return Courier{
		CourierID:           uuid.New(),
		CourierName:         f.Company().Name(),
		ContactName:         f.Person().Name(),
		ContactEmail:        f.Internet().Email(),
		ContactPhone:        f.Phone().Number(),
		CostPerKM:           f.RandomFloat(2, 2, 10),
		AverageDeliveryTime: time.Duration(f.IntBetween(30, 240)) * time.Minute,
		ReliabilityScore:    int16(f.IntBetween(70, 100)),
	}
}
