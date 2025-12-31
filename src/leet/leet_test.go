package leet

import "testing"

func Test(t *testing.T) {
	res := generateParenthesis(3)
	for i := 0; i < len(res); i++ {
		println(res[i])
	}
}
