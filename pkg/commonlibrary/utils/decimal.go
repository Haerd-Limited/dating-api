package utils

import (
	"fmt"
	"strconv"

	"github.com/aarondl/sqlboiler/v4/types"
	"github.com/ericlagergren/decimal"
)

// FloatPtrToNullDecimal converts *float64 to types.NullDecimal (nil => NULL).
func FloatPtrToNullDecimal(f *float64) types.NullDecimal {
	if f == nil {
		return types.NullDecimal{}
	}

	d := new(decimal.Big).SetFloat64(*f)

	return types.NullDecimal{Big: d}
}

// NullDecimalToFloatPtr converts types.NullDecimal to *float64 (NULL/zero => nil).
func NullDecimalToFloatPtr(d types.NullDecimal) (*float64, error) {
	if d.IsZero() {
		return nil, nil
	}

	secs, err := strconv.ParseFloat(d.String(), 64)
	if err != nil {
		return nil, fmt.Errorf("parse media seconds: %w", err)
	}

	return &secs, nil
}
