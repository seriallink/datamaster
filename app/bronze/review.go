package bronze

import (
	"strconv"
)

func init() {
	Register(&Review{})
}

type Review struct {
	ReviewID         int64   `json:"review_id"         parquet:"review_id,optional,snappy"`
	BeerID           int64   `json:"beer_id"           parquet:"beer_id,optional,snappy"`
	ProfileID        int64   `json:"profile_id"        parquet:"profile_id,optional,snappy"`
	ReviewOverall    float64 `json:"review_overall"    parquet:"review_overall,optional,snappy"`
	ReviewAroma      float64 `json:"review_aroma"      parquet:"review_aroma,optional,snappy"`
	ReviewAppearance float64 `json:"review_appearance" parquet:"review_appearance,optional,snappy"`
	ReviewPalate     float64 `json:"review_palate"     parquet:"review_palate,optional,snappy"`
	ReviewTaste      float64 `json:"review_taste"      parquet:"review_taste,optional,snappy"`
	ReviewTime       int64   `json:"review_time"       parquet:"review_time,optional,snappy"` // Unix timestamp
	Operation        string  `json:"operation"         parquet:"operation,optional,snappy,dict"`
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
	beerID, _ := strconv.ParseInt(get("beer_id"), 10, 64)
	profileID, _ := strconv.ParseInt(get("profile_id"), 10, 64)
	reviewTime, _ := strconv.ParseInt(get("review_time"), 10, 64)

	reviewOverall, _ := strconv.ParseFloat(get("review_overall"), 64)
	reviewAroma, _ := strconv.ParseFloat(get("review_aroma"), 64)
	reviewAppearance, _ := strconv.ParseFloat(get("review_appearance"), 64)
	reviewPalate, _ := strconv.ParseFloat(get("review_palate"), 64)
	reviewTaste, _ := strconv.ParseFloat(get("review_taste"), 64)

	return Review{
		ReviewID:         reviewID,
		BeerID:           beerID,
		ProfileID:        profileID,
		ReviewTime:       reviewTime,
		ReviewOverall:    reviewOverall,
		ReviewAroma:      reviewAroma,
		ReviewAppearance: reviewAppearance,
		ReviewPalate:     reviewPalate,
		ReviewTaste:      reviewTaste,
	}, nil

}
