package models

import (
	"strconv"
	"strings"
)

type Adjustment struct {
	IsScaleOut bool
	IsRelative bool
	Value      int
}

func ParseAdjustment(adjustment string) (Adjustment, error) {
	isScaleOut := strings.HasPrefix(adjustment, "+")
	isRelative := strings.HasSuffix(adjustment, "%")

	if isScaleOut || strings.HasPrefix(adjustment, "-") {
		adjustment = adjustment[1:]
	}

	if isRelative {
		adjustment = adjustment[:len(adjustment)-1]
	}

	value, err := strconv.Atoi(adjustment)
	if err != nil {
		return Adjustment{}, err
	}

	return Adjustment{
		IsScaleOut: isScaleOut,
		IsRelative: isRelative,
		Value:      value,
	}, nil
}

func compareScaleOut(a, b Adjustment) int {
	if a.IsScaleOut != b.IsScaleOut {
		if a.IsScaleOut {
			return 1
		}
		return -1
	}
	return 0
}

func compareRelative(a, b Adjustment) int {
	if a.IsRelative != b.IsRelative {
		if a.IsScaleOut {
			if a.IsRelative {
				return 1
			}
			return -1
		}
		if a.IsRelative {
			return -1
		}
		return 1
	}
	return 0
}

func compareValue(a, b Adjustment) int {
	if a.Value != b.Value {
		if a.IsScaleOut {
			return a.Value - b.Value
		}
		return b.Value - a.Value
	}
	return 0
}

func CompareAdjustments(a, b Adjustment) int {
	// See [slices.SortFunc](https://pkg.go.dev/slices#SortFunc):
	// cmp(a, b) should return a negative number when a < b, a positive number when
	// a > b and zero when a == b or a and b are incomparable in the sense of
	// a strict weak ordering.

	if differsByScaleOut := compareScaleOut(a, b); differsByScaleOut != 0 {
		return differsByScaleOut
	}

	if differsByRelative := compareRelative(a, b); differsByRelative != 0 {
		return differsByRelative
	}

	if differsByValue := compareValue(a, b); differsByValue != 0 {
		return differsByValue
	}

	return 0
}
