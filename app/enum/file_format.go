package enum

type FileFormatEnum string

const (
	FileFormatCsv     FileFormatEnum = "csv"
	FileFormatJson    FileFormatEnum = "json"
	FileFormatParquet FileFormatEnum = "parquet"
)

func (enum FileFormatEnum) Values() []Enum {
	return []Enum{
		FileFormatCsv,
		FileFormatJson,
		FileFormatParquet,
	}
}

func (enum FileFormatEnum) String() string {
	return string(enum)
}

func (enum FileFormatEnum) IsEmpty() bool {
	return isEmptyEnum(enum)
}

func (enum FileFormatEnum) IsValid() bool {
	return isValidEnum(enum, enum)
}
