package model

import (
	"time"

	"github.com/seriallink/datamaster/cli/enum"

	"github.com/google/uuid"
	"github.com/jaswdr/faker"
)

type Tracking struct {
	TrackingID     uuid.UUID               `gorm:"column:tracking_id;primaryKey"`
	DeliveryID     uuid.UUID               `gorm:"column:delivery_id"`
	Notes          string                  `gorm:"column:notes"`
	TrackingStatus enum.TrackingStatusEnum `gorm:"column:tracking_status"`
	EventTimestamp time.Time               `gorm:"column:event_timestamp"`
	CreatedAt      time.Time               `gorm:"column:created_at"`
	UpdatedAt      time.Time               `gorm:"column:updated_at"`
}

func (Tracking) TableName() string {
	return "dm_core.tracking"
}

func FakeTracking(f faker.Faker, deliveryID uuid.UUID) Tracking {
	return Tracking{
		TrackingID:     uuid.New(),
		DeliveryID:     deliveryID,
		Notes:          f.Lorem().Sentence(8),
		TrackingStatus: enum.RandomEnum[enum.TrackingStatusEnum](f, enum.TrackingStatusEnum("").Values()),
		EventTimestamp: time.Now(),
	}
}
