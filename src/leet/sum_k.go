package leet

func subarraySum(nums []int, k int) int {
	if len(nums) == 0 {
		return 0
	}
	store := make(map[int]int)
	total := 0
	result := 0
	store[0] = 1
	for i := 0; i < len(nums); i++ {
		total += nums[i]
		target := total - k
		value, ok := store[target]
		if ok {
			result += value
		}
		store[total]++
	}
	return result
}
