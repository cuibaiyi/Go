package main

import (
	"fmt"
)

type treeNode struct {
	value       string
	left, right *treeNode
}

func main() {
	//创建一颗树
	root := treeNode{"A", nil, nil}
	root.left = &treeNode{value: "B"}
	root.right = &treeNode{value: "C"}
	root.left.left = &treeNode{value: "D"}
	root.left.right = &treeNode{value: "E"}
	root.left.right.left = new(treeNode)
	root.left.right.left.value = "F"
	root.right.left = &treeNode{value: "G"}
	root.right.left.right = &treeNode{value: "H"}
	root.right.right = &treeNode{value: "I"}
	root.traverse()
}

//前序遍历
func (node *treeNode) traverse() {
	if node == nil {
		return
	}
	fmt.Print(node.value + " ")
	node.left.traverse()
	node.right.traverse()
}
// 输出：A B D E F C G H I

//中序遍历
func (node *treeNode) traverse() {
   if(node == nil){
      return
   }
   node.left.traverse()
   fmt.Print(node.value + " ")
   node.right.traverse()
}
// 输出：D B F E A G H C I

// //后序遍历
func (node *treeNode) traverse() {
   if(node == nil){
      return
   }
   node.left.traverse()
   node.right.traverse()
   fmt.Print(node.value + " ")
}
//遍历结果：D F E B H G I C A 
