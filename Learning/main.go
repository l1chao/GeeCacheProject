package main

func twoSum(nums []int, target int) []int {
	for i := 0; i < len(nums); i++ {
		pair_val := target - nums[i]
		for j := i + 1; j < len(nums); j++ {
			if pair_val == nums[j] {
				return []int{i, j}
			}
		}
	}
	return nil
}
