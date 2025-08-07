package bronze

import (
	"strconv"
)

func init() {
	Register(&Review{})
}

// Review represents the schema of the review dataset in the bronze layer.
//
// This model is stored in Parquet format and captures detailed beer review data,
// including various rating aspects and relationships to beer, brewery, and user profile.
//
// Fields:
//   - ReviewID: Unique identifier for the review.
//   - BreweryID: ID of the associated brewery.
//   - BeerID: ID of the reviewed beer.
//   - ProfileID: ID of the user who submitted the review.
//   - ReviewOverall: Overall rating given by the user.
//   - ReviewAroma: Aroma score given by the user.
//   - ReviewAppearance: Appearance score given by the user.
//   - ReviewPalate: Palate score given by the user.
//   - ReviewTaste: Taste score given by the user.
//   - ReviewTime: Timestamp (Unix) of when the review was submitted.
//   - Operation: CDC operation type (insert, update, delete).
type Review struct {
	ReviewID         int64   `json:"review_id"         parquet:"review_id"`
	BreweryID        int64   `json:"brewery_id"        parquet:"brewery_id"`
	BeerID           int64   `json:"beer_id"           parquet:"beer_id"`
	ProfileID        int64   `json:"profile_id"        parquet:"profile_id"`
	ReviewOverall    float64 `json:"review_overall"    parquet:"review_overall,optional"`
	ReviewAroma      float64 `json:"review_aroma"      parquet:"review_aroma,optional"`
	ReviewAppearance float64 `json:"review_appearance" parquet:"review_appearance,optional"`
	ReviewPalate     float64 `json:"review_palate"     parquet:"review_palate,optional"`
	ReviewTaste      float64 `json:"review_taste"      parquet:"review_taste,optional"`
	ReviewTime       int64   `json:"review_time"       parquet:"review_time,optional"`
	Operation        string  `json:"operation"         parquet:"operation,dict"`
}

func (m *Review) TableName() string {
	return "dm_bronze.review"
}

func (m *Review) FromCSV(line []string, idx map[string]int) (any, error) {

	get := func(key string) string {
		i, ok := idx[key]
		if !ok || i >= len(line) {
			return ""
		}
		return line[i]
	}

	reviewID, _ := strconv.ParseInt(get("review_id"), 10, 64)
	breweryID, _ := strconv.ParseInt(get("brewery_id"), 10, 64)
	beerID, _ := strconv.ParseInt(get("beer_id"), 10, 64)
	profileID, _ := strconv.ParseInt(get("profile_id"), 10, 64)
	reviewTime, _ := strconv.ParseInt(get("review_time"), 10, 64)

	reviewOverall, _ := strconv.ParseFloat(get("review_overall"), 64)
	reviewAroma, _ := strconv.ParseFloat(get("review_aroma"), 64)
	reviewAppearance, _ := strconv.ParseFloat(get("review_appearance"), 64)
	reviewPalate, _ := strconv.ParseFloat(get("review_palate"), 64)
	reviewTaste, _ := strconv.ParseFloat(get("review_taste"), 64)

	return &Review{
		ReviewID:         reviewID,
		BreweryID:        breweryID,
		BeerID:           beerID,
		ProfileID:        profileID,
		ReviewTime:       reviewTime,
		ReviewOverall:    reviewOverall,
		ReviewAroma:      reviewAroma,
		ReviewAppearance: reviewAppearance,
		ReviewPalate:     reviewPalate,
		ReviewTaste:      reviewTaste,
		Operation:        "insert",
	}, nil

}
