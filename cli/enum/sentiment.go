package enum

type SentimentEnum string

const (
	SentimentPositive SentimentEnum = "positive"
	SentimentNeutral  SentimentEnum = "neutral"
	SentimentNegative SentimentEnum = "negative"
)

func (enum SentimentEnum) Values() []Enum {
	return []Enum{
		SentimentPositive,
		SentimentNeutral,
		SentimentNegative,
	}
}

func (enum SentimentEnum) String() string {
	return string(enum)
}

func (enum SentimentEnum) IsEmpty() bool {
	return isEmptyEnum(enum)
}

func (enum SentimentEnum) IsValid() bool {
	return isValidEnum(enum, enum)
}
