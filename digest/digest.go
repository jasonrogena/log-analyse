package digest

import (
	"sort"
)

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
