package models_test

import (
	. "code.cloudfoundry.org/app-autoscaler/src/autoscaler/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"slices"
)

var _ = Describe("Adjustment Parsing", func() {
	DescribeTable("ParseAdjustment",
		func(input string, expected Adjustment, expectError bool) {
			result, err := ParseAdjustment(input)
			if expectError {
				Expect(err).To(HaveOccurred())
			} else {
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(expected))
			}
		},
		Entry("positive absolute value", "+10", Adjustment{IsScaleOut: true, IsRelative: false, Value: 10}, false),
		Entry("negative absolute value", "-10", Adjustment{IsScaleOut: false, IsRelative: false, Value: 10}, false),
		Entry("positive relative value", "+10%", Adjustment{IsScaleOut: true, IsRelative: true, Value: 10}, false),
		Entry("negative relative value", "-10%", Adjustment{IsScaleOut: false, IsRelative: true, Value: 10}, false),
		Entry("invalid value", "abc", Adjustment{}, true),
	)
})

var _ = Describe("Adjustment Comparison", func() {
	DescribeTable("CompareAdjustments",
		func(adj1, adj2 Adjustment, expected int) {
			result := CompareAdjustments(adj1, adj2)
			Expect(result).To(Equal(expected))
		},
		Entry("both positive absolute values, equal", Adjustment{IsScaleOut: true, IsRelative: false, Value: 10}, Adjustment{IsScaleOut: true, IsRelative: false, Value: 10}, 0),
		Entry("both positive absolute values, different", Adjustment{IsScaleOut: true, IsRelative: false, Value: 10}, Adjustment{IsScaleOut: true, IsRelative: false, Value: 5}, 5),
		Entry("positive vs negative absolute values", Adjustment{IsScaleOut: true, IsRelative: false, Value: 10}, Adjustment{IsScaleOut: false, IsRelative: false, Value: 10}, 1),
		Entry("both negative absolute values, equal", Adjustment{IsScaleOut: false, IsRelative: false, Value: 10}, Adjustment{IsScaleOut: false, IsRelative: false, Value: 10}, 0),
		Entry("both positive relative values, equal", Adjustment{IsScaleOut: true, IsRelative: true, Value: 10}, Adjustment{IsScaleOut: true, IsRelative: true, Value: 10}, 0),
		Entry("positive vs negative relative values", Adjustment{IsScaleOut: true, IsRelative: true, Value: 10}, Adjustment{IsScaleOut: false, IsRelative: true, Value: 10}, 1),
		Entry("absolute vs relative values", Adjustment{IsScaleOut: true, IsRelative: false, Value: 10}, Adjustment{IsScaleOut: true, IsRelative: true, Value: 10}, -1),
		Entry("relative vs absolute values", Adjustment{IsScaleOut: true, IsRelative: true, Value: 10}, Adjustment{IsScaleOut: true, IsRelative: false, Value: 10}, 1),
	)
})

var _ = Describe("Adjustment Sorting", func() {
	It("should sort adjustments using CompareAdjustments", func() {
		adjustments := []Adjustment{
			{IsScaleOut: true, IsRelative: false, Value: 5},
			{IsScaleOut: false, IsRelative: true, Value: 10},
			{IsScaleOut: true, IsRelative: true, Value: 10},
			{IsScaleOut: false, IsRelative: false, Value: 5},
		}

		expected := []Adjustment{
			{IsScaleOut: false, IsRelative: true, Value: 10},
			{IsScaleOut: false, IsRelative: false, Value: 5},
			{IsScaleOut: true, IsRelative: false, Value: 5},
			{IsScaleOut: true, IsRelative: true, Value: 10},
		}

		slices.SortStableFunc(adjustments, func(a, b Adjustment) int {
			return CompareAdjustments(a, b)
		})

		Expect(adjustments).To(Equal(expected))
	})
})
