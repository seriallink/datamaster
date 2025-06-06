package bronze

import "github.com/google/uuid"

func init() {
	Register(&Review{})
}

type Review struct {
	ReviewID       string `json:"review_id"       parquet:"review_id,optional,snappy"`
	ProductID      string `json:"product_id"      parquet:"product_id,optional,snappy"`
	CustomerID     string `json:"customer_id"     parquet:"customer_id,optional,snappy"`
	ReviewText     string `json:"review_text"     parquet:"review_text,optional,snappy"`
	Rating         string `json:"rating"          parquet:"rating,optional,snappy"`
	ReviewDate     string `json:"review_date"     parquet:"review_date,optional,snappy"`
	SentimentLabel string `json:"sentiment_label" parquet:"sentiment_label,optional,snappy,dict"`
	SentimentScore string `json:"sentiment_score" parquet:"sentiment_score,optional,snappy"`
	CreatedAt      string `json:"created_at"      parquet:"created_at,optional,snappy"`
	UpdatedAt      string `json:"updated_at"      parquet:"updated_at,optional,snappy"`
	Operation      string `json:"operation"       parquet:"operation,optional,snappy,dict"`
}

func (m *Review) TableName() string {
	return "dm_bronze.review"
}

func (m *Review) TableId() uuid.UUID {
	return uuid.MustParse("7426e8d6-b84a-420c-adcc-43c61dbd111c")
}
