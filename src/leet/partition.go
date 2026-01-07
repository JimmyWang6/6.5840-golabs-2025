package leet

// 给你一个字符串 s，请你将 s 分割成一些 子串，使每个子串都是 回文串 。返回 s 所有可能的分割方案。
//
// 示例 1：
//
// 输入：s = "aab"
// 输出：[["a","a","b"],["aa","b"]]
// 示例 2：
//
// 输入：s = "a"
// 输出：[["a"]]
//
// 提示：
//
// 1 <= s.length <= 16
// s 仅由小写英文字母组成
func partition(s string) [][]string {
	var dp [][]bool = make([][]bool, len(s))
	var res [][]string = make([][]string, 0)
	for i := 0; i < len(s); i++ {
		dp[i] = make([]bool, len(s))
		dp[i][i] = true
	}
	n := len(s)
	for i := n - 1; i >= 0; i-- {
		for j := i + 1; j < len(s); j++ {
			if s[i] == s[j] {
				if j-i < 2 {
					dp[i][j] = true
				} else {
					dp[i][j] = dp[i+1][j-1]
				}
			} else {
				dp[i][j] = false
			}
		}
	}

	splits := make([]string, 0)
	var dfs func(int)

	dfs = func(i int) {
		if i == len(s) {
			temp := make([]string, len(splits))
			copy(temp, splits)
			res = append(res, temp)
			return
		}
		for j := i; j < len(s); j++ {
			if dp[i][j] {
				splits = append(splits, s[i:j+1])
				dfs(j + 1)
				splits = splits[:len(splits)-1]
			}
		}
	}
	dfs(0)
	return res
}
