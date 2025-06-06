package model

import (
	"time"

	"github.com/jaswdr/faker"
)

type Review struct {
	ReviewID         int64     `gorm:"column:review_id;primaryKey"`
	BeerID           int64     `gorm:"column:beer_id"`
	ProfileID        string    `gorm:"column:profile_id"`
	ReviewOverall    float64   `gorm:"column:review_overall"`
	ReviewAroma      float64   `gorm:"column:review_aroma"`
	ReviewAppearance float64   `gorm:"column:review_appearance"`
	ReviewPalate     float64   `gorm:"column:review_palate"`
	ReviewTaste      float64   `gorm:"column:review_taste"`
	ReviewTime       time.Time `gorm:"column:review_time"`
}

func (Review) TableName() string {
	return "dm_core.review"
}

func FakeReview(f faker.Faker, beerID int64, profileID string) Review {
	return Review{
		ReviewID:         f.Int64Between(1_000_000_000, 9_000_000_000),
		BeerID:           beerID,
		ProfileID:        profileID,
		ReviewOverall:    f.Float64(1, 1, 5),
		ReviewAroma:      f.Float64(1, 1, 5),
		ReviewAppearance: f.Float64(1, 1, 5),
		ReviewPalate:     f.Float64(1, 1, 5),
		ReviewTaste:      f.Float64(1, 1, 5),
		ReviewTime:       f.Time().TimeBetween(time.Now().AddDate(-3, 0, 0), time.Now()),
	}
}
