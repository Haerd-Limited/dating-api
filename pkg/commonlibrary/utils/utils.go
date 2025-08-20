package utils

import "fmt"

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
