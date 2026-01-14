package leet

import "sort"

//以数组 intervals 表示若干个区间的集合，其中单个区间为 intervals[i] = [starti, endi] 。请你合并所有重叠的区间，并返回 一个不重叠的区间数组，该数组需恰好覆盖输入中的所有区间 。

//示例 1：
//
//输入：intervals = [[1,3],[2,6],[8,10],[15,18]]
//输出：[[1,6],[8,10],[15,18]]
//解释：区间 [1,3] 和 [2,6] 重叠, 将它们合并为 [1,6].
//示例 2：
//
//输入：intervals = [[1,4],[4,5]]
//输出：[[1,5]]
//解释：区间 [1,4] 和 [4,5] 可被视为重叠区间。
//示例 3：
//
//输入：intervals = [[4,7],[1,4]]
//输出：[[1,7]]
//解释：区间 [1,4] 和 [4,7] 可被视为重叠区间。
//
//
//提示：
//
//1 <= intervals.length <= 104
//intervals[i].length == 2
//0 <= starti <= endi <= 104

type Node struct {
	start int
	end   int
	next  *Node
	pre   *Node
}

var head *Node

func merge(intervals [][]int) [][]int {
	if len(intervals) == 0 {
		return nil
	}
	sort.Slice(intervals, func(i, j int) bool {
		return intervals[i][0] < intervals[j][0]
	})

	// 2. 构建链表
	// 使用 dummy 节点（哨兵节点）可以避免处理头节点为空的 edge case
	dummy := &Node{}
	tail := dummy // tail 永远指向链表的最后一个有效节点

	for _, interval := range intervals {
		start, end := interval[0], interval[1]

		// 这里的逻辑极大简化：
		// 情况 A: 链表是空的（tail == dummy）或者 当前区间和 tail 不重叠
		if tail == dummy || start > tail.end {
			// 直接在尾部追加新节点
			newNode := &Node{
				start: start,
				end:   end,
			}
			tail.next = newNode
			tail = newNode
		} else {
			// 情况 B: 发生重叠 (start <= tail.end)
			// 因为输入有序，start 一定 >= tail.start，所以只需要更新 tail.end
			if end > tail.end {
				tail.end = end
			}
		}
	}

	// 3. 将链表转换回结果数组
	var result [][]int
	cur := dummy.next
	for cur != nil {
		result = append(result, []int{cur.start, cur.end})
		cur = cur.next
	}

	return result
}

func overlap(node0 *Node, node1 *Node) bool {
	if node0 == head || node1 == head {
		return false
	}
	if node0.start > node1.end || node1.start > node0.end {
		return false
	}
	return true
}
