package msgo

import (
	"fmt"
	"testing"
)

func TestTreeNode_Put(t *testing.T) {
	root := &treeNode{name: "", children: make([]*treeNode, 0)}
	root.Put("/user/get/:id")
	root.Put("/user/create")
	root.Put("/order/list")
	root.Put("/order/create")

	node := root.Get("/user/get/1")
	fmt.Println(node)
	node = root.Get("/user/get/2")
	fmt.Println(node)
	node = root.Get("/user/hello")
	fmt.Println(node)
	node = root.Get("/order/create")
	fmt.Println(node)
}
