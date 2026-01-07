package leet

import "testing"

func Test(t *testing.T) {
	res := findAnagrams("cbaebabacd", "abc")
	length := len(res)
	for i := 0; i < length; i++ {
		println(res[i])
	}
}
