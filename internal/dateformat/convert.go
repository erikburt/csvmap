package dateformat

import (
	"time"
)

// Converter handles date format conversion.
type Converter struct {
	InputFormat  string
	OutputFormat string
}

// NewConverter creates a new date converter.
func NewConverter(inputFormat, outputFormat string) *Converter {
	return &Converter{
		InputFormat:  inputFormat,
		OutputFormat: outputFormat,
	}
}

// Convert converts a date string from input format to output format.
// Returns the original value if conversion fails or value is empty.
func (c *Converter) Convert(value string) string {
	if value == "" {
		return value
	}

	t, err := time.Parse(c.InputFormat, value)
	if err != nil {
		// Return original if can't parse
		return value
	}

	return t.Format(c.OutputFormat)
}

// ConvertAll converts all values in a slice.
func (c *Converter) ConvertAll(values []string) []string {
	result := make([]string, len(values))
	for i, v := range values {
		result[i] = c.Convert(v)
	}
	return result
}

// TransformFunc returns a function that can be used with Data.TransformColumn.
func (c *Converter) TransformFunc() func(string) string {
	return c.Convert
}

// Preview shows how a sample of values would be converted.
func (c *Converter) Preview(values []string, maxSamples int) []ConversionPreview {
	if maxSamples > len(values) {
		maxSamples = len(values)
	}

	var previews []ConversionPreview
	for i := 0; i < maxSamples; i++ {
		if values[i] == "" {
			continue
		}
		preview := ConversionPreview{
			Original:  values[i],
			Converted: c.Convert(values[i]),
		}
		preview.Success = preview.Original != preview.Converted || c.canParse(values[i])
		previews = append(previews, preview)
		if len(previews) >= maxSamples {
			break
		}
	}
	return previews
}

func (c *Converter) canParse(value string) bool {
	_, err := time.Parse(c.InputFormat, value)
	return err == nil
}

// ConversionPreview shows before/after for a single value.
type ConversionPreview struct {
	Original  string
	Converted string
	Success   bool
}

// ValidateLayout checks if a layout string is valid by testing with a reference time.
func ValidateLayout(layout string) bool {
	// Use Go's reference time to validate
	refTime := time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)
	formatted := refTime.Format(layout)

	// Try to parse it back
	_, err := time.Parse(layout, formatted)
	return err == nil
}
