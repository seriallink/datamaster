package model

import (
	"time"

	"github.com/jaswdr/faker"
)

func init() {
	Register(&Review{})
}

type Review struct {
	ReviewID         int64     `json:"review_id"         gorm:"column:review_id;primaryKey"`
	BeerID           int64     `json:"beer_id"           gorm:"column:beer_id"`
	ProfileID        string    `json:"profile_id"        gorm:"column:profile_id"`
	ReviewOverall    float64   `json:"review_overall"    gorm:"column:review_overall"`
	ReviewAroma      float64   `json:"review_aroma"      gorm:"column:review_aroma"`
	ReviewAppearance float64   `json:"review_appearance" gorm:"column:review_appearance"`
	ReviewPalate     float64   `json:"review_palate"     gorm:"column:review_palate"`
	ReviewTaste      float64   `json:"review_taste"      gorm:"column:review_taste"`
	ReviewTime       time.Time `json:"review_time"       gorm:"column:review_time"`
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
