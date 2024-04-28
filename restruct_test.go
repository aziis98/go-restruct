package restruct_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aziis98/lfr/lang/runtime/restruct"
	"gotest.tools/assert"
)

type Foo struct {
	First  int
	Second string
}

type SlugString string

func NewSlugString(s string) SlugString {
	return SlugString(strings.ReplaceAll(strings.TrimSpace(s), " ", "-"))
}

type Bar struct {
	FirstField  int
	SecondField SlugString
}

func TestBasicFeatures(t *testing.T) {
	foo1 := Foo{
		First:  1,
		Second: "  this is my foo    ",
	}

	bar1, err := restruct.Convert[Bar](foo1,
		restruct.MustFunc[SlugString, string](NewSlugString),
		restruct.StructFromStruct[Bar, Foo]{
			"FirstField":  "First",
			"SecondField": "Second",
		},
	)
	assert.NilError(t, err)
	assert.DeepEqual(t,
		Bar{
			FirstField:  1,
			SecondField: SlugString("this-is-my-foo"),
		},
		bar1,
	)
}

//
// Tree conversion
//

// Fist tree structure

type Node interface{}

type TreeLeaf int

type TreeBranch struct {
	Left  Node
	Right Node
}

// Second tree structure

type TreeNode struct {
	Value int
	Left  *TreeNode
	Right *TreeNode
}

func TestTreeConversion(t *testing.T) {
	sourceTree := TreeBranch{
		Left: TreeBranch{
			Left:  TreeLeaf(1),
			Right: TreeLeaf(2),
		},
		Right: TreeBranch{
			Left: TreeLeaf(3),
			Right: TreeBranch{
				Left:  TreeLeaf(4),
				Right: TreeLeaf(5),
			},
		},
	}

	expectedTreeNode := &TreeNode{
		Left: &TreeNode{
			Left: &TreeNode{
				Value: 1,
			},
			Right: &TreeNode{
				Value: 2,
			},
		},
		Right: &TreeNode{
			Left: &TreeNode{
				Value: 3,
			},
			Right: &TreeNode{
				Left: &TreeNode{
					Value: 4,
				},
				Right: &TreeNode{
					Value: 5,
				},
			},
		},
	}

	t.Run("Generic Recursive Conversion", func(t *testing.T) {
		actualTreeNode, err := restruct.Convert[*TreeNode](sourceTree,
			restruct.RecursiveFunc[*TreeNode, Node](func(cnv restruct.Converter, n Node) (*TreeNode, error) {
				switch n := n.(type) {
				case TreeLeaf:
					value := int(n)
					return &TreeNode{Value: value}, nil
				case TreeBranch:
					left, err := restruct.ConvertWith[*TreeNode](cnv, n.Left)
					if err != nil {
						return nil, err
					}
					right, err := restruct.ConvertWith[*TreeNode](cnv, n.Right)
					if err != nil {
						return nil, err
					}

					return &TreeNode{Left: left, Right: right}, nil
				default:
					panic(fmt.Sprintf("unknown node type: %T", n))
				}
			}),
		)
		assert.NilError(t, err)
		assert.DeepEqual(t, expectedTreeNode, actualTreeNode)
	})

	t.Run("Specialized Recursive Conversion", func(t *testing.T) {
		actualTreeNode, err := restruct.Convert[*TreeNode](sourceTree,
			restruct.Func[*TreeNode, TreeLeaf](func(n TreeLeaf) (*TreeNode, error) {
				value := int(n)
				return &TreeNode{Value: value}, nil
			}),
			restruct.RecursiveFunc[*TreeNode, TreeBranch](func(cnv restruct.Converter, n TreeBranch) (*TreeNode, error) {
				left, err := restruct.ConvertWith[*TreeNode](cnv, n.Left)
				if err != nil {
					return nil, err
				}
				right, err := restruct.ConvertWith[*TreeNode](cnv, n.Right)
				if err != nil {
					return nil, err
				}

				return &TreeNode{Left: left, Right: right}, nil
			}),
		)
		assert.NilError(t, err)
		assert.DeepEqual(t, expectedTreeNode, actualTreeNode)
	})

	t.Run("StructFromStruct Recursive Conversion", func(t *testing.T) {
		actualTreeNode, err := restruct.Convert[*TreeNode](sourceTree,
			restruct.StructFromStruct[*TreeNode, TreeBranch]{
				"Left":  "Left",
				"Right": "Right",
			},
			restruct.Func[*TreeNode, TreeLeaf](func(n TreeLeaf) (*TreeNode, error) {
				value := int(n)
				return &TreeNode{Value: value}, nil
			}),
		)
		assert.NilError(t, err)
		assert.DeepEqual(t, expectedTreeNode, actualTreeNode)
	})
}
