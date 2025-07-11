package main

import "fmt"

type TreeNode struct {
	Val   int
	Left  *TreeNode
	Right *TreeNode
}

func invertTree(root *TreeNode) *TreeNode {
	queue := []*TreeNode{root}

	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]

		node.Left, node.Right = node.Right, node.Left

		if node.Left != nil {
			queue = append(queue, node.Left)
		}

		if node.Right != nil {
			queue = append(queue, node.Right)
		}
	}
	return root
}

func arrToTree(arr []interface{}) *TreeNode {
	if len(arr) == 0 || arr[0] == nil {
		return nil
	}

	root := &TreeNode{Val: arr[0].(int)}
	queue := []*TreeNode{root}

	i := 1
	for i < len(arr) {
		node := queue[0]
		queue = queue[1:]

		if i < len(arr) && arr[i] != nil {
			node.Left = &TreeNode{Val: arr[i].(int)}
			queue = append(queue, node.Left)
		}
		i++

		if i < len(arr) && arr[i] != nil {
			node.Right = &TreeNode{Val: arr[i].(int)}
			queue = append(queue, node.Right)
		}
		i++
	}

	return root
}

func treeToArr(root *TreeNode) []interface{} {
	res := []interface{}{}
	if root == nil {
		return res
	}

	queue := []*TreeNode{root}
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]

		if node != nil {
			res = append(res, node.Val)
			queue = append(queue, node.Left)
			queue = append(queue, node.Right)
		} else {
			res = append(res, nil) // q3 追加nil
		}
	}

	// del尾端多餘的 nil
	for len(res) > 0 && res[len(res)-1] == nil {
		res = res[:len(res)-1]
	}

	return res
}

func main() {
	// ex1
	q1 := []interface{}{5, 3, 8, 1, 7, 2, 6}
	tree := arrToTree(q1)
	ansTree := invertTree(tree)
	fmt.Printf("q1:%v ANS: %v\n", q1, treeToArr(ansTree))

	// ex2
	q2 := []interface{}{6, 8, 9}
	tree2 := arrToTree(q2)
	ansTree2 := invertTree(tree2)
	fmt.Printf("q2: %v ANS: %v\n", q2, treeToArr(ansTree2))

	// // // ex3
	q3 := []interface{}{5, 3, 8, 1, 7, 2, 6, 100, 3, -1}
	tree3 := arrToTree(q3)
	ansTree3 := invertTree(tree3)
	fmt.Printf("q3: %v ANS: %v\n", q3, treeToArr(ansTree3))

	// // // ex4
	// q4 := []int{}
	// ans4 := []int{}
	// fmt.Printf("q1: %v ANS: %v\n", q4, ans4)
}
