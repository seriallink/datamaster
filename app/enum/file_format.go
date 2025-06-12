package enum

type FileFormatEnum string

const (
	FileFormatCsv  FileFormatEnum = "csv"
	FileFormatJson FileFormatEnum = "json"
)

func (enum FileFormatEnum) Values() []Enum {
	return []Enum{
		FileFormatCsv,
		FileFormatJson,
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
