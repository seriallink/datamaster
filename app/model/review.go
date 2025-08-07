package model

import (
	"time"
)

func init() {
	Register(&Review{})
}

// Review represents a user-submitted review of a beer in the raw dataset.
//
// This struct captures both the subjective evaluation and metadata related to
// when and by whom the review was submitted.
//
// Fields:
//   - ReviewID: Unique identifier for the review (primary key).
//   - BeerID: Foreign key referencing the reviewed beer.
//   - ProfileID: Identifier of the user who submitted the review.
//   - ReviewOverall: Overall rating given to the beer.
//   - ReviewAroma: Rating for the aroma of the beer.
//   - ReviewAppearance: Rating for the appearance of the beer.
//   - ReviewPalate: Rating for the palate/mouthfeel of the beer.
//   - ReviewTaste: Rating for the taste of the beer.
//   - ReviewTime: Timestamp indicating when the review was created.
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
