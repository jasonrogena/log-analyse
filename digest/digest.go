package digest

import (
	"sort"
	"github.com/jasonrogena/log-analyse/types"
)

type UrlPathDigester struct {
	tree *Tree
}

func (digester UrlPathDigester) Absorb() (err error) {
	err = digester.tree.generalizeTree()
	return
}

func (digester UrlPathDigester) Digest(logLine interface{}) error {
	// TODO: implement this
}

func (digester UrlPathDigester) IsDigestable(someData interface{}) bool {
	field, ok := someData.(types.Field)
	if ok {
		if field.FieldType == "request" {
			return true
		}
	}
	return false
}

func InitUrlPathDigester(rbfsLayerCap int) (digester UrlPathDigester) {
	rootNodes := make(map[string]*TreeNode)
	index := make(map[int]map[string]*TreeNode)
	tree := Tree{rootNodes:rootNodes, inOrderNodeIndex: index, rbfsLayerCap: rbfsLayerCap}
	digester = UrlPathDigester{tree:&tree}
	return
}

func (tree *Tree) generalizeTree() error {
	// Use reverse breadth first search to travers the tree, starting from it's leaf nodes
	var treeLayers []int

	for cl := range tree.inOrderNodeIndex {
		treeLayers = append(treeLayers, cl)
	}
	sort.Ints(treeLayers)

	// perform the traversal, starting from the bottom most (bottom most leafs) layer
	for curIndex := len(treeLayers) - 1; curIndex--; curIndex >= 0 {
		curLayer := treeLayers[curIndex]
		for curUUID := range tree.inOrderNodeIndex[curLayer] {//TODO: Does range work with growing map
			tree.inOrderNodeIndex[curLayer][curUUID].generalizeTreeNode(tree)
		}
	}

	// TODO: save data in the database

	return nil
}

func (node *TreeNode) generalizeTreeNode(tree *Tree) error {
	// Check each of it's children, combine those that have equivalent child trees
	if node.children != nil {
		var childUUIDS []string
		for curUUID := range node.children {
			childUUIDS = append(childUUIDS, curUUID)
		}

		for x,_ := range childUUIDS {
			if _, xOk := node.children[childUUIDS[x]]; xOk {
				var equivalentSiblings []string
				for y := x + 1; y < len(childUUIDS); y++ {
					if _, yOk := node.children[childUUIDS[y]]; yOk && node.children[childUUIDS[x]].hasEquivalentChildTree(node.children[childUUIDS[y]]) {
						equivalentSiblings = append(equivalentSiblings, childUUIDS[y])
					}
				}

				if len(equivalentSiblings) > 0 {
					equivalentSiblings = append(equivalentSiblings, childUUIDS[x])
					node.combineChildren(tree, equivalentSiblings)
				}
			}
		}
	}

	return nil
}
