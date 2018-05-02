package digest

import (
	"strings"
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

func getField(someData interface{}) (*types.Field, bool) {
	field, ok := someData.(*types.Field)
	if ok {
		if field.FieldType != "request" {
			ok = false
		}
	}
	return field, ok
}

func (digester UrlPathDigester) Digest(someData interface{}) error {
	field, fieldOk := getField(types.Field)
	if fieldOk && len(field.ValueString) > 0 {
		//GET /api/v1/forms/197928.json?a=dfds HTTP/1.1
		reqPartsArr := strings.Split(field.ValueString, " ")
		if len(reqPartsArr) == 3 {
			reqPathWithArgs := reqPartsArr[2]
			reqPathParts := strings.Split(reqPathWithArgs, "?")
			cleanReqPath := reqPathParts[0]
			cleanReqPathParts := strings.Split(cleanReqPath, "/")
			digester.tree.addRequestPath(nil, &digester.tree.rootNodes, cleanReqPathParts, -1)
		}
	}
}

func (tree *Tree)addRequestPath(parentNode *TreeNode, nodes *map[string]*TreeNode, requestPathParts []string, lastVisitedPathIndex int) error {
	if lastVisitedPathIndex < len(requestPathParts) {
		nodeFound := false
		var foundNode *TreeNode
		pathIndex := lastVisitedPathIndex + 1
		for _, curNode := range nodes {
			if curNode.value == GENERIC_VALUE {
				for curCombinedValue := range curNode.combinedValues {
					if curCombinedValue == requestPathParts[pathIndex] {
						nodeFound = true
						break
					}
				}
			} else if curNode.value == requestPathParts[pathIndex] {
				nodeFound = true
			}

			if nodeFound {
				foundNode = curNode
				break
			}
		}

		if !nodeFound {// no node in current level has the value
			newNodeUuid, uuidErr := genUUID()
			if uuidErr != nil {
				return uuidErr
			}
			newTreeNode := TreeNode{level:pathIndex, uuid: newNodeUuid, value: requestPathParts[pathIndex], parent:parentNode}
			foundNode = &newTreeNode
			
			if parentNode == nil {// newTreeNode is a root node
				tree.addNodeToIndex(foundNode)
				nodes[foundNode.uuid] = foundNode
			} else {
				parentNode.addChild(tree, foundNode)
			}			
		}

		if foundNode != nil {
			addReqErr := tree.addRequestPath(foundNode, &foundNode.children, requestPathParts, pathIndex)
			if addReqErr != nil {
				return addReqErr
			}
		}
	}

	return nil
}

func (digester UrlPathDigester) IsDigestable(someData interface{}) bool {
	field, ok := getField(types.Field)
	return ok
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
