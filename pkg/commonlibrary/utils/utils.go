package utils

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/rivo/uniseg"
)

// CalculateAge returns the age in years given a birthdate.
func CalculateAge(birthdate time.Time) int {
	now := time.Now()

	years := now.Year() - birthdate.Year()

	// If the birthday hasn't occurred yet this year, subtract 1
	if now.Month() < birthdate.Month() ||
		(now.Month() == birthdate.Month() && now.Day() < birthdate.Day()) {
		years--
	}

	return years
}

func ValidateHTTPURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return err
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("unsupported scheme %q", u.Scheme)
	}

	if u.Host == "" {
		return fmt.Errorf("empty host")
	}

	return nil
}

func Redacted(s string) string {
	if len(s) <= 6 {
		return "***"
	}

	return "***" + s[len(s)-6:]
}

func TypePtrToStringPtr[v any](input *v) *string {
	if input == nil {
		return nil
	}

	str := fmt.Sprintf("%v", *input) // Convert the value pointed to by input to a string

	return &str
}

func TypePtrToString[v any](input *v) string {
	if input == nil {
		return ""
	}

	return fmt.Sprintf("%v", *input)
}

func PtrToString(s *string) string {
	if s != nil {
		return *s
	}

	return ""
}

var wsRE = regexp.MustCompile(`\s+`)

// CountTextLen returns a normalized, human-length count for scoring.
// 1. trims, collapses whitespace
// 2. strips control chars
// 3. counts grapheme clusters (emoji 👍🏽 = 1)
func CountTextLen(s string) int {
	// normalize whitespace and trim
	s = strings.TrimSpace(wsRE.ReplaceAllString(s, " "))

	// drop control runes (except common whitespace)
	var b strings.Builder

	for _, r := range s {
		if unicode.IsControl(r) && r != '\n' && r != '\r' && r != '\t' {
			continue
		}

		b.WriteRune(r)
	}

	norm := b.String()
	if norm == "" {
		return 0
	}

	// Prefer grapheme clusters (best for emoji and combined glyphs)
	return uniseg.GraphemeClusterCount(norm)
}
