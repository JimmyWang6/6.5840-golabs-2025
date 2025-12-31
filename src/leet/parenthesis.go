package leet

//数字 n 代表生成括号的对数，请你设计一个函数，用于能够生成所有可能的并且 有效的 括号组合。
//
//
//
//示例 1：
//
//输入：n = 3
//输出：["((()))","(()())","(())()","()(())","()()()"]
//示例 2：
//
//输入：n = 1
//输出：["()"]

func generateParenthesis(n int) []string {
	res := make([]string, 0)
	add(&res, "", 0, 0, n)
	return res
}

func add(res *[]string, str string, lb int, rb int, target int) {
	if len(str) == target*2 {
		*res = append(*res, str)
		return
	}
	if target-lb > 0 {
		add(res, str+"(", lb+1, rb, target)
	}
	if rb < lb {
		add(res, str+")", lb, rb+1, target)
	}
}
