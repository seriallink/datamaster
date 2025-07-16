package core

import (
	"context"
	"fmt"
	"strings"

	"github.com/seriallink/datamaster/app/misc"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/comprehend"
	"github.com/aws/aws-sdk-go-v2/service/comprehend/types"
)

// MaskPIIData scans and masks personally identifiable information (PII) fields in the provided dataset.
//
// It uses a sample of the data to detect potential PII fields and then masks their values across the entire dataset.
//
// Parameters:
//   - cfg: aws.Config - AWS configuration used for service clients.
//   - ctx: context.Context - Context for the AWS operations.
//   - data: []map[string]any - The dataset where PII fields may be masked.
//
// Returns:
//   - []map[string]any: The dataset with masked PII values (if any).
//   - error: An error if PII detection or masking fails.
func MaskPIIData(cfg aws.Config, ctx context.Context, data []map[string]any) ([]map[string]any, error) {

	if len(data) == 0 {
		return data, nil
	}

	sample := SampleData(data, 10)

	piiFields, err := DetectPIIFields(cfg, ctx, sample)
	if err != nil {
		return nil, fmt.Errorf("error detecting PII: %w", err)
	}

	if len(piiFields) == 0 {
		return data, nil
	}

	for i := range data {
		for _, field := range piiFields {
			if val, ok := data[i][field]; ok {
				data[i][field] = MaskValue(val)
			}
		}
	}

	return data, nil

}

// DetectPIIFields analyzes sample data to identify fields that likely contain personally identifiable information (PII).
//
// It uses AWS Comprehend to evaluate the combined text of each field across multiple records. A field is considered PII
// if at least 50% of its non-empty values are detected as PII with an average confidence score of 0.7 or higher.
//
// Parameters:
//   - cfg: aws.Config - AWS configuration used to initialize the Comprehend client.
//   - ctx: context.Context - Context for the Comprehend API call.
//   - sample: []map[string]any - Sample data used to evaluate potential PII fields.
//
// Returns:
//   - []string: A list of field names detected as containing PII.
//   - error: An error if the detection fails.
func DetectPIIFields(cfg aws.Config, ctx context.Context, sample []map[string]any) ([]string, error) {

	if len(sample) == 0 {
		return nil, nil
	}

	client := comprehend.NewFromConfig(cfg)

	result := []string{}

	fieldSet := make(map[string]struct{})
	for _, row := range sample {
		for field := range row {
			fieldSet[field] = struct{}{}
		}
	}

	for field := range fieldSet {

		var combined strings.Builder
		nonEmptyCount := 0

		for _, row := range sample {
			if val, ok := row[field]; ok && val != nil {
				str := fmt.Sprintf("%v", val)
				if strings.TrimSpace(str) != "" {
					combined.WriteString(str + "\n")
					nonEmptyCount++
				}
			}
		}

		if nonEmptyCount == 0 {
			continue
		}

		input := &comprehend.DetectPiiEntitiesInput{
			Text:         aws.String(combined.String()),
			LanguageCode: types.LanguageCodeEn,
		}

		output, err := client.DetectPiiEntities(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("detect PII for field %s failed: %w", field, err)
		}

		var scoreSum float32
		matchCount := 0

		for _, entity := range output.Entities {
			if entity.Score != nil {
				scoreSum += *entity.Score
				matchCount++
			}
		}

		if matchCount == 0 {
			continue
		}

		avgScore := scoreSum / float32(matchCount)
		proportion := float32(matchCount) / float32(nonEmptyCount)

		if proportion >= 0.5 && avgScore >= 0.7 {
			result = append(result, field)
		}

	}

	return result, nil

}

// MaskValue returns a masked version of the input value for PII obfuscation.
//
// The function converts the input to a string and applies masking as follows:
//   - If the string is empty, returns an empty string.
//   - If the string has 4 or fewer characters, replaces all characters with asterisks.
//   - If the string has more than 4 characters, keeps the first and last characters and replaces the middle with asterisks.
//
// Parameters:
//   - val: any - The value to be masked.
//
// Returns:
//   - any: The masked string, or the original value if it could not be converted.
func MaskValue(val any) any {

	str := misc.AnyToString(val)

	if len(str) == 0 {
		return ""
	}

	if len(str) <= 4 {
		return strings.Repeat("*", len(str))
	}

	return str[:1] + strings.Repeat("*", len(str)-2) + str[len(str)-1:]

}

// SampleData returns a subset of the input data containing at most `n` elements.
//
// If the length of the input slice is less than `n`, it returns the entire slice.
// Otherwise, it returns the first `n` elements.
//
// Parameters:
//   - data: []map[string]any - The input dataset.
//   - n: int - The maximum number of elements to return.
//
// Returns:
//   - []map[string]any: A subslice of the original dataset with at most `n` entries.
func SampleData(data []map[string]any, n int) []map[string]any {
	if len(data) < n {
		return data
	}
	return data[:n]
}
