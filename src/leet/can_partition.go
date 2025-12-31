package leet

// 给你一个 只包含正整数 的 非空 数组 nums 。请你判断是否可以将这个数组分割成两个子集，使得两个子集的元素和相等。
//
// 示例 1：
//
// 输入：nums = [1,5,11,5]
// 输出：true
// 解释：数组可以分割成 [1, 5, 5] 和 [11] 。
// 示例 2：
//
// 输入：nums = [1,2,3,5]
// 输出：false
// 解释：数组不能分割成两个元素和相等的子集。
//
// 提示：
//
// 1 <= nums.length <= 200
// 1 <= nums[i] <= 100
func canPartition(nums []int) bool {
	if len(nums) < 2 {
		return false
	}
	total := 0
	for i := 0; i < len(nums); i++ {
		total += nums[i]
	}
	if total%2 != 0 {
		return false
	}
	target := total / 2
	dp := make([]int, target+1)
	dp[0] = 0
	for i := 0; i < len(nums); i++ {
		num := nums[i]
		for j := target; j >= num; j-- {
			if dp[j] > dp[j-num]+num {
				dp[j] = dp[j]
			} else {
				dp[j] = dp[j-num] + num
			}
		}
		if dp[target] == target {
			return true
		}
	}
	return false
}
