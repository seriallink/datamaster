package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/jaswdr/faker"
	"github.com/seriallink/null"
)

type Delivery struct {
	DeliveryID            uuid.UUID          `gorm:"column:delivery_id;primaryKey"`
	PurchaseID            uuid.UUID          `gorm:"column:purchase_id"`
	CourierID             uuid.UUID          `gorm:"column:courier_id"`
	DeliveryStatus        DeliveryStatusEnum `gorm:"column:delivery_status"`
	DeliveryEstimatedDate time.Time          `gorm:"column:delivery_estimated_date"`
	DeliveryDate          null.Time          `gorm:"column:delivery_date"`
	CreatedAt             time.Time          `gorm:"column:created_at"`
	UpdatedAt             time.Time          `gorm:"column:updated_at"`
}

func (Delivery) TableName() string {
	return "dm_core.delivery"
}

func FakeDelivery(f faker.Faker) Delivery {

	delDate := null.Time{}
	estDate := time.Now().AddDate(0, 0, f.IntBetween(1, 10)) // 1–5 days ahead

	status := RandomEnum[DeliveryStatusEnum](f, DeliveryStatusEnum("").Values())

	// Set delivery_date only when status is DeliveryDelivered or DeliveryFailed
	if status == DeliveryDelivered || status == DeliveryFailed {
		delDate = null.TimeFrom(estDate.Add(time.Duration(f.IntBetween(-2, 2)) * time.Hour))
	}

	return Delivery{
		DeliveryID:            uuid.New(),
		PurchaseID:            uuid.New(), // Deve ser relacionado a uma Purchase real
		CourierID:             uuid.New(), // Deve ser relacionado a um Courier real
		DeliveryStatus:        status,
		DeliveryEstimatedDate: estDate,
		DeliveryDate:          delDate,
	}

}
