// Package dateformat handles date format inference and conversion for csvmap.
package dateformat

import (
	"time"
)

// CommonFormats is a list of common date formats to try, ordered by likelihood.
var CommonFormats = []FormatInfo{
	// ISO formats
	{Layout: "2006-01-02", Description: "ISO date (YYYY-MM-DD)"},
	{Layout: "2006-01-02 15:04:05", Description: "ISO datetime"},
	{Layout: "2006-01-02T15:04:05", Description: "ISO datetime with T"},
	{Layout: "2006-01-02T15:04:05Z07:00", Description: "ISO datetime with timezone"},

	// US formats
	{Layout: "01/02/2006", Description: "US date (MM/DD/YYYY)"},
	{Layout: "1/2/2006", Description: "US date short (M/D/YYYY)"},
	{Layout: "01-02-2006", Description: "US date with dashes"},
	{Layout: "01/02/06", Description: "US date short year (MM/DD/YY)"},
	{Layout: "1/2/06", Description: "US date short (M/D/YY)"},

	// European formats
	{Layout: "02/01/2006", Description: "European date (DD/MM/YYYY)"},
	{Layout: "2/1/2006", Description: "European date short (D/M/YYYY)"},
	{Layout: "02-01-2006", Description: "European date with dashes"},
	{Layout: "02.01.2006", Description: "European date with dots"},

	// Text formats
	{Layout: "Jan 2, 2006", Description: "Month Day, Year (Jan 2, 2006)"},
	{Layout: "January 2, 2006", Description: "Full month (January 2, 2006)"},
	{Layout: "2 Jan 2006", Description: "Day Month Year (2 Jan 2006)"},
	{Layout: "2 January 2006", Description: "Day Full Month Year"},
	{Layout: "Jan 02, 2006", Description: "Month Day, Year zero-padded"},
	{Layout: "02 Jan 2006", Description: "Day Month Year zero-padded"},

	// Other common formats
	{Layout: "20060102", Description: "Compact date (YYYYMMDD)"},
	{Layout: "2006/01/02", Description: "Year/Month/Day with slashes"},
	{Layout: "Mon, 02 Jan 2006", Description: "RFC 822 style"},
	{Layout: "Monday, January 2, 2006", Description: "Full weekday and month"},
}

// FormatInfo contains a date format layout and its description.
type FormatInfo struct {
	Layout      string
	Description string
}

// InferResult represents the result of format inference.
type InferResult struct {
	Format      FormatInfo
	Confidence  float64 // 0.0 to 1.0
	SamplesParsed int
	SamplesTotal  int
	Ambiguous   bool
}

// InferFormat attempts to detect the date format from a sample of values.
// It returns the best matching format and confidence level.
func InferFormat(values []string) *InferResult {
	if len(values) == 0 {
		return nil
	}

	// Filter out empty values
	var nonEmpty []string
	for _, v := range values {
		if v != "" {
			nonEmpty = append(nonEmpty, v)
		}
	}

	if len(nonEmpty) == 0 {
		return nil
	}

	// Sample up to 100 values for inference
	sample := nonEmpty
	if len(sample) > 100 {
		sample = sample[:100]
	}

	var bestFormat FormatInfo
	bestCount := 0
	var matchingFormats []FormatInfo

	for _, format := range CommonFormats {
		count := countParseable(sample, format.Layout)
		if count > bestCount {
			bestCount = count
			bestFormat = format
			matchingFormats = []FormatInfo{format}
		} else if count == bestCount && count > 0 {
			matchingFormats = append(matchingFormats, format)
		}
	}

	if bestCount == 0 {
		return nil
	}

	result := &InferResult{
		Format:        bestFormat,
		Confidence:    float64(bestCount) / float64(len(sample)),
		SamplesParsed: bestCount,
		SamplesTotal:  len(sample),
		Ambiguous:     len(matchingFormats) > 1,
	}

	return result
}

// ValidateFormat checks if a format string can parse all non-empty values.
func ValidateFormat(values []string, layout string) (int, int) {
	parsed := 0
	total := 0

	for _, v := range values {
		if v == "" {
			continue
		}
		total++
		if _, err := time.Parse(layout, v); err == nil {
			parsed++
		}
	}

	return parsed, total
}

// countParseable counts how many values can be parsed with the given format.
func countParseable(values []string, layout string) int {
	count := 0
	for _, v := range values {
		if _, err := time.Parse(layout, v); err == nil {
			count++
		}
	}
	return count
}

// GetFormatCheatsheet returns a list of common output formats for display.
func GetFormatCheatsheet() []FormatInfo {
	return []FormatInfo{
		{Layout: "2006-01-02", Description: "ISO date: 2006-01-02"},
		{Layout: "01/02/2006", Description: "US date: 01/02/2006"},
		{Layout: "02/01/2006", Description: "European: 02/01/2006"},
		{Layout: "Jan 2, 2006", Description: "Text: Jan 2, 2006"},
		{Layout: "2 Jan 2006", Description: "Text: 2 Jan 2006"},
		{Layout: "January 2, 2006", Description: "Full: January 2, 2006"},
		{Layout: "2006-01-02 15:04:05", Description: "With time: 2006-01-02 15:04:05"},
		{Layout: "Monday, January 2, 2006", Description: "Full weekday: Monday, January 2, 2006"},
		{Layout: "20060102", Description: "Compact: 20060102"},
	}
}
