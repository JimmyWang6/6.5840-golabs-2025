package leet

//给定两个字符串 s 和 p，找到 s 中所有 p 的 异位词 的子串，返回这些子串的起始索引。不考虑答案输出的顺序。
//
//示例 1:
//
//输入: s = "cbaebabacd", p = "abc"
//输出: [0,6]
//解释:
//起始索引等于 0 的子串是 "cba", 它是 "abc" 的异位词。
//起始索引等于 6 的子串是 "bac", 它是 "abc" 的异位词。
//示例 2:
//
//输入: s = "abab", p = "ab"
//输出: [0,1,2]
//解释:
//起始索引等于 0 的子串是 "ab", 它是 "ab" 的异位词。
//起始索引等于 1 的子串是 "ba", 它是 "ab" 的异位词。
//起始索引等于 2 的子串是 "ab", 它是 "ab" 的异位词。

func findAnagrams(s string, p string) []int {
	n, m := len(s), len(p)
	if n < m {
		return nil
	}

	// 1. 使用数组代替 map，索引 0-25 对应 a-z
	// pCount 记录目标 p 的字符分布
	// sCount 记录当前窗口的字符分布
	var pCount, sCount [26]int

	// 2. 初始化：统计 p 的字符，以及 s 中前 m 个字符（第一个窗口）
	for i := 0; i < m; i++ {
		pCount[p[i]-'a']++
		sCount[s[i]-'a']++
	}

	var result []int

	// 3. 检查第一个窗口
	// Go 语言的神奇之处：数组可以直接比较！
	if sCount == pCount {
		result = append(result, 0)
	}

	// 4. 开始滑动窗口
	for i := m; i < n; i++ {
		// 移入右边的新字符
		sCount[s[i]-'a']++
		// 移出左边的旧字符
		sCount[s[i-m]-'a']--

		// 再次比较两个数组是否相同
		if sCount == pCount {
			result = append(result, i-m+1) // 记录起始索引
		}
	}

	return result
}
