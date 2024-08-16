package jsonextractor

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ExtractJSONData extracts the JSON substring from the input string.
func ExtractJSONData(input string) (string, error) {
	startIndex := strings.Index(input, "{")
	endIndex := strings.LastIndex(input, "}")

	if startIndex == -1 || endIndex == -1 || startIndex >= endIndex {
		return "", fmt.Errorf("JSON data not found or invalid format")
	}

	jsonStr := input[startIndex : endIndex+1]
	return jsonStr, nil
}

// UnmarshalJSONData unmarshals the JSON data into the provided struct.
func UnmarshalJSONData[T any](input string, result *T) error {
	jsonStr, err := ExtractJSONData(input)
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(jsonStr), result)
	if err != nil {
		return err
	}

	return nil
}
