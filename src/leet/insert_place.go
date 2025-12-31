package leet

//35. 搜索插入位置
//尝试过
//简单
//相关标签
//premium lock icon
//相关企业
//给定一个排序数组和一个目标值，在数组中找到目标值，并返回其索引。如果目标值不存在于数组中，返回它将会被按顺序插入的位置。
//
//请必须使用时间复杂度为 O(log n) 的算法。
//
//
//
//示例 1:
//
//输入: nums = [1,3,5,6], target = 5
//输出: 2
//示例 2:
//
//输入: nums = [1,3,5,6], target = 2
//输出: 1
//示例 3:
//
//输入: nums = [1,3,5,6], target = 7
//输出: 4

func searchInsert(nums []int, target int) int {
	if len(nums) == 0 {
		return 0
	}
	if len(nums) == 1 {
		if nums[0] >= target {
			return 0
		} else {
			return 1
		}
	}
	for i := 1; i < len(nums); i++ {
		if nums[i] > target {
			return i - 1
		}
		if nums[i] == target || (nums[i] > target && nums[i] > nums[i-1]) {
			return i
		}
	}
	return len(nums)
}
