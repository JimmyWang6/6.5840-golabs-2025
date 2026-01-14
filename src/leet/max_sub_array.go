package leet

// 给你一个整数数组 nums ，请你找出一个具有最大和的连续子数组（子数组最少包含一个元素），返回其最大和。
//
// 子数组是数组中的一个连续部分。
//
// 示例 1：
//
// 输入：nums = [-2,1,-3,4,-1,2,1,-5,4]
// 输出：6
// 解释：连续子数组 [4,-1,2,1] 的和最大，为 6 。
// 示例 2：
//
// 输入：nums = [1]
// 输出：1
// 示例 3：
//
// 输入：nums = [5,4,-1,7,8]
// 输出：23
//
// 提示：
//
// 1 <= nums.length <= 105
// -104 <= nums[i] <= 104
func maxSubArray(nums []int) int {
	if len(nums) == 0 {
		return 0
	}
	length := len(nums)
	dp := make([]int, length)
	dp[0] = nums[0]
	maxValue := nums[0]
	for i := 1; i < length; i++ {
		dp[i] = max(dp[i-1]+nums[i], nums[i])
		if dp[i] > maxValue {
			maxValue = dp[i]
		}
	}
	return maxValue
}
