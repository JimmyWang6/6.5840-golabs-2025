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
	len1 := len(s)
	len2 := len(p)
	if len1 < len2 {
		return []int{}
	}
	if len1 == len2 {
		if s == p {
			return []int{0}
		}
	}
	result := make([]int, 0)
	res := make(map[byte]int)
	for i := 0; i < len2; i++ {
		res[p[i]]++
	}
	for i := 0; i < len2; i++ {
		_, ok := res[s[i]]
		if ok {
			res[s[i]]--
		}
	}
	for lp, rp := 0, len2-1; ; {
		if allZero(res) {
			// empty map means submap
			result = append(result, lp)
		}
		_, lpOk := res[s[lp]]
		if lpOk {
			res[s[lp]]++
		}
		lp++
		rp++
		if rp == len1 {
			break
		}
		//
		_, rpOk := res[s[rp]]
		if rpOk {
			res[s[rp]]--
		}
	}
	return result
}

func allZero(value map[byte]int) bool {
	for _, v := range value {
		if v != 0 {
			return false
		}
	}
	return true
}
