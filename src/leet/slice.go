package leet

import (
	"container/heap"
)

var a []int

type Item struct {
	Val   int
	Index int
}

// 定义堆结构
type MaxHeap []Item

func (h MaxHeap) Len() int { return len(h) }
func (h MaxHeap) Less(i, j int) bool {
	return h[i].Val > h[j].Val
}
func (h MaxHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }
func (h *MaxHeap) Push(x any) {
	*h = append(*h, x.(Item))
}
func (h *MaxHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}
func (h MaxHeap) Peek() Item {
	if len(h) == 0 {
		return Item{}
	}
	return h[0]
}
func maxSlidingWindow(nums []int, k int) []int {
	result := make([]int, len(nums)-k+1)
	h := &MaxHeap{}
	heap.Init(h)
	if len(nums) == 0 {
		return []int{}
	}
	for i := 0; i < k; i++ {
		heap.Push(h, Item{Val: nums[i], Index: i})
	}
	result[0] = h.Peek().Val
	for i := 1; i < len(result); i++ {
		index := i + k - 1
		heap.Push(h, Item{Val: nums[index], Index: index})
		for h.Len() > 0 && h.Peek().Index < i {
			heap.Pop(h)
		}
		result[i] = h.Peek().Val
	}
	return result
}
