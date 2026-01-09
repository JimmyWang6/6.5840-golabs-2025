package leet

import (
	"strconv"
	"strings"
)

// 给你一个仅由数字（0 - 9）组成的字符串 num 。
//
// 请你找出能够使用 num 中数字形成的 最大回文 整数，并以字符串形式返回。该整数不含 前导零 。
//
// 注意：
//
// 你 无需 使用 num 中的所有数字，但你必须使用 至少 一个数字。
// 数字可以重新排序。
//
// 示例 1：
//
// 输入：num = "444947137"
// 输出："7449447"
// 解释：
// 从 "444947137" 中选用数字 "4449477"，可以形成回文整数 "7449447" 。
// 可以证明 "7449447" 是能够形成的最大回文整数。
// 示例 2：
//
// 输入：num = "00009"
// 输出："9"
// 解释：
// 可以证明 "9" 能够形成的最大回文整数。
// 注意返回的整数不应含前导零。
func largestPalindromic(num string) string {
	// 1. 统计频率
	count := make([]int, 10)
	for _, c := range num {
		count[c-'0']++
	}
	var leftBuilder strings.Builder

	for i := 9; i >= 0; i-- {
		if i == 0 && leftBuilder.Len() == 0 {
			continue
		}

		for count[i] > 1 {
			leftBuilder.WriteString(strconv.Itoa(i))
			count[i] -= 2
		}
	}

	leftPart := leftBuilder.String()

	mid := ""
	for i := 9; i >= 0; i-- {
		if count[i] > 0 {
			mid = strconv.Itoa(i)
			break
		}
	}
	if leftPart == "" && mid == "" {
		return "0"
	}

	return leftPart + mid + reverseString(leftPart)
}

// 辅助函数：翻转字符串
func reverseString(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}
