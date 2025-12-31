package leet

// 给你一个按照非递减顺序排列的整数数组 nums，和一个目标值 target。请你找出给定目标值在数组中的开始位置和结束位置。
//
// 如果数组中不存在目标值 target，返回 [-1, -1]。
//
// 你必须设计并实现时间复杂度为 O(log n) 的算法解决此问题。
//
// 示例 1：
//
// 输入：nums = [5,7,7,8,8,10], target = 8
// 输出：[3,4]
// 示例 2：
//
// 输入：nums = [5,7,7,8,8,10], target = 6
// 输出：[-1,-1]
// 示例 3：
//
// 输入：nums = [], target = 0
// 输出：[-1,-1]
//
// 提示：
//
// 0 <= nums.length <= 105
// -109 <= nums[i] <= 109
// nums 是一个非递减数组
// -109 <= target <= 109
func searchRange(nums []int, target int) []int {
	return binarySearch(nums, target, 0, len(nums)-1)
}

func binarySearch(nums []int, target int, from int, to int) []int {
	if from > to || nums[from] > target || nums[to] < target {
		return []int{-1, -1}
	}
	if from == to {
		if nums[from] == target {
			return []int{from, from}
		} else {
			return []int{-1, -1}
		}
	}

	mid := (from + to) / 2
	if nums[mid] < target {
		return binarySearch(nums, target, mid+1, to)
	}
	if nums[mid] > target {
		return binarySearch(nums, target, from, mid)
	}
	if nums[mid] == target {
		left := binarySearch(nums, target, from, mid)
		right := binarySearch(nums, target, mid+1, to)
		ld := mid
		rd := mid

		if left[1] == -1 && left[0] == -1 {
			ld = mid
		} else if left[0] != -1 && left[1] != -1 {
			ld = left[0]
		} else {
			ld = left[1]
		}

		if right[1] == -1 && right[0] == -1 {
			rd = mid
		} else if right[0] != -1 && right[1] != -1 {
			rd = right[1]
		} else {
			rd = right[0]
		}
		return []int{ld, rd}
	}
	return []int{-1, -1}
}
