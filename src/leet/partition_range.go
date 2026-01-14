package leet

//给你一个字符串 s 。我们要把这个字符串划分为尽可能多的片段，同一字母最多出现在一个片段中。例如，字符串 "ababcc" 能够被分为 ["abab", "cc"]，但类似 ["aba", "bcc"] 或 ["ab", "ab", "cc"] 的划分是非法的。
//
//注意，划分结果需要满足：将所有划分结果按顺序连接，得到的字符串仍然是 s 。
//
//返回一个表示每个字符串片段的长度的列表。

var result = make([]int, 0)

func partitionLabels(s string) []int {
	var flag = make([]int, 26)
	for i := 0; i < len(s); i++ {
		flag[s[i]-'a'] = i
	}
	lp, rp, end := 0, 0, 0
	for {
		if lp >= len(s) || rp >= len(s) {
			break
		}
		index := s[rp] - 'a'
		last := flag[index]
		if last > end {
			end = last
		}
		if end == rp {
			result = append(result, rp-lp+1)
			lp = end + 1
			rp = lp
		} else {
			rp = rp + 1
		}
	}
	return result
}
