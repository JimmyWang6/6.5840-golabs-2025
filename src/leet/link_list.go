package leet

func isPalindrome(head *ListNode) bool {
	if head == nil || head.Next == nil {
		return true
	}
	slow, fast := head, head
	for fast != nil && fast.Next != nil {
		fast = fast.Next.Next
		slow = slow.Next
	}

	if fast == nil {
		//偶数
		// slow 开始为后半部分
		cur := slow
		slow = slow.Next
		cur.Next = nil
		// 反转
		return reserve(head, slow)
	} else {
		cur := slow
		slow = slow.Next
		cur.Next = nil
		return reserve(head, slow)
	}

}

func reserve(head *ListNode, half *ListNode) bool {
	cur := half
	for {
		pre := half
		pre.Next = nil
		half = half.Next
		if half == nil {
			break
		}
		half.Next = pre
	}
	for cur != nil && head != nil {
		if cur.Val != head.Val {
			return false
		}
		cur = cur.Next
		head = head.Next
	}
	return true
}
