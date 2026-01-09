package leet

type TreeNode struct {
	Val   int
	Left  *TreeNode
	Right *TreeNode
}

func invertTree(root *TreeNode) *TreeNode {
	if root == nil {
		return nil
	}
	cur := root.Left
	root.Left = invertTree(root.Right)
	root.Right = invertTree(cur)
	return root
}
