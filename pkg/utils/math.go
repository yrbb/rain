package utils

import "math"

func Round(value float64) float64 {
	return math.Floor(value + 0.5)
}

func Floor(value float64) float64 {
	return math.Floor(value)
}

func Ceil(value float64) float64 {
	return math.Ceil(value)
}

func Max(nums ...float64) float64 {
	if len(nums) < 2 {
		panic("nums: the nums length is less than 2")
	}

	max := nums[0]

	for i := 1; i < len(nums); i++ {
		max = math.Max(max, nums[i])
	}

	return max
}

func Min(nums ...float64) float64 {
	if len(nums) < 2 {
		panic("nums: the nums length is less than 2")
	}

	min := nums[0]

	for i := 1; i < len(nums); i++ {
		min = math.Min(min, nums[i])
	}

	return min
}
