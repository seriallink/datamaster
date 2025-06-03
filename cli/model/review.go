package model

import (
	"time"

	"github.com/seriallink/datamaster/cli/enum"

	"github.com/google/uuid"
	"github.com/jaswdr/faker"
)

type Review struct {
	ReviewID       uuid.UUID          `gorm:"column:review_id;primaryKey"`
	ProductID      uuid.UUID          `gorm:"column:product_id"`
	CustomerID     uuid.UUID          `gorm:"column:customer_id"`
	ReviewText     string             `gorm:"column:review_text"`
	Rating         int16              `gorm:"column:rating"`
	ReviewDate     time.Time          `gorm:"column:review_date"`
	SentimentLabel enum.SentimentEnum `gorm:"column:sentiment_label"`
	SentimentScore float64            `gorm:"column:sentiment_score"`
	CreatedAt      time.Time          `gorm:"column:created_at"`
	UpdatedAt      time.Time          `gorm:"column:updated_at"`
}

func (Review) TableName() string {
	return "dm_core.review"
}

func FakeReview(f faker.Faker, productID, customerID uuid.UUID) Review {
	return Review{
		ReviewID:       uuid.New(),
		ProductID:      productID,
		CustomerID:     customerID,
		ReviewText:     f.Lorem().Sentence(12),
		Rating:         int16(f.IntBetween(1, 5)),
		ReviewDate:     time.Now().AddDate(0, 0, -f.IntBetween(0, 30)),
		SentimentLabel: enum.RandomEnum[enum.SentimentEnum](f, enum.SentimentEnum("").Values()),
		SentimentScore: f.RandomFloat(2, 0, 1),
	}
}
