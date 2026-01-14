package leet

import "testing"

func TestSpiralOrder(t *testing.T) {
	// 直接在函数参数中初始化
	array := [][]int{
		{1, 2, 3},
		{4, 5, 6},
		{7, 8, 9},
	}
	res := spiralOrder(array)
	t.Log(res)
}
