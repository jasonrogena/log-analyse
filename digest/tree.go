package digest

import (
	"errors"

	uuid "github.com/satori/go.uuid"
)

const GENERIC_VALUE string = "*"

type Tree struct {
	rootNodes        map[string]*TreeNode
	inOrderNodeIndex map[int]map[string]*TreeNode
	rbfsLayerCap     int
}

type TreeNode struct {
	level          int
	uuid           string
	value          string
	combinedValues []string
	children       map[string]*TreeNode
	parent         *TreeNode
	payload        []interface{}
}

func (node *TreeNode) getSiblings() []*TreeNode {
	var siblings []*TreeNode

	if node.parent != nil {
		for _, curChild := range node.parent.children {
			if curChild != node {
				siblings = append(siblings, curChild)
			}
		}
	}

	return siblings
}

func (node *TreeNode) equal(otherNode *TreeNode) bool {
	return node.hasEqualValue(otherNode) && node.hasEquivalentChildTree(otherNode)
}

func (node *TreeNode) hasEqualValue(otherNode *TreeNode) bool {
	return node.value == otherNode.value
}

func (node *TreeNode) hasEquivalentChildTree(otherNode *TreeNode) bool {
	if len(node.children) == len(otherNode.children) {
		for _, curChild := range node.children {
			childFound := false
			for _, curOtherChild := range otherNode.children {
				if curChild.hasEqualValue(curOtherChild) {
					childFound = true
					if curChild.hasEquivalentChildTree(curOtherChild) == false {
						return false
					}
				}
			}

			if childFound == false {
				return false
			}
		}
		return true
	}

	return false
}

func (node *TreeNode) addChild(tree *Tree, newChild *TreeNode) error {
	if len(newChild.uuid) == 0 {
		return errors.New("Could not add node as child to node with uuid " + node.uuid)
	}

	if _, containsKey := node.children[newChild.uuid]; containsKey {
		return errors.New("Child already exists with uuid " + newChild.uuid + " cannot add another child with the same uuid")
	}

	node.children[newChild.uuid] = newChild
	newChild.parent = node
	newChild.level = node.level + 1

	tree.addNodeToIndex(newChild)

	return nil
}

func (tree *Tree) addNodeToIndex(node *TreeNode) {
	if _, ok := tree.inOrderNodeIndex[node.level]; !ok {
		tree.inOrderNodeIndex[node.level] = make(map[string]*TreeNode)
	}
	tree.inOrderNodeIndex[node.level][node.uuid] = node
}

func (node *TreeNode) removeChild(tree *Tree, child *TreeNode) error {
	if _, ok := tree.inOrderNodeIndex[child.level]; ok {
		delete(tree.inOrderNodeIndex[child.level], child.uuid)
	}
	delete(node.children, child.uuid)
	return nil
}

func (node *TreeNode) addCombinedValue(value string) {
	found := false
	for _, curValue := range node.combinedValues {
		if curValue == value {
			found = true
			break
		}
	}

	if !found {
		node.combinedValues = append(node.combinedValues, value)
	}
}

func (node *TreeNode) phagocytose(weakNode *TreeNode, includeChildren bool) {
	// copy over the value
	if weakNode.value != GENERIC_VALUE {
		node.addCombinedValue(weakNode.value)
	}

	for _, curCombinedValue := range weakNode.combinedValues {
		node.addCombinedValue(curCombinedValue)
	}

	// copy over the payloads
	node.payload = append(node.payload, weakNode.payload...)

	if includeChildren {
		var weakNodeChildrenKeys []string
		for curKey, _ := range weakNode.children {
			weakNodeChildrenKeys = append(weakNodeChildrenKeys, curKey)
		}

		for _, curMasterChild := range node.children {
			indexToRemove := -1
			for i := 0; i < len(weakNodeChildrenKeys); i++ {
				curWeakChild := weakNode.children[weakNodeChildrenKeys[i]]
				if curMasterChild.value == curWeakChild.value {
					curMasterChild.phagocytose(curWeakChild, true)
					indexToRemove = i
					break
				}
			}

			if indexToRemove != -1 {
				if indexToRemove == len(weakNodeChildrenKeys)-1 {
					weakNodeChildrenKeys = weakNodeChildrenKeys[:indexToRemove]
				} else {
					weakNodeChildrenKeys = append(weakNodeChildrenKeys[:indexToRemove], weakNodeChildrenKeys[indexToRemove+1:]...)
				}
			}
		}
	}
}

func (node *TreeNode) combineChildren(tree *Tree, uuids []string) error {
	newChild := new(TreeNode)
	childUUID, uuidErr := genUUID()

	if uuidErr != nil {
		return uuidErr
	}

	newChild.uuid = childUUID
	newChild.value = GENERIC_VALUE

	for _, curUUID := range uuids {
		if curChild, ok := node.children[curUUID]; ok {
			includeItsChildren := true
			if len(newChild.children) == 0 && len(curChild.children) > 0 {
				includeItsChildren = false
				newChild.children = curChild.children
				for _, curNewGrandchild := range newChild.children {
					curNewGrandchild.parent = newChild
				}
			}

			newChild.phagocytose(curChild, includeItsChildren)
			node.removeChild(tree, curChild)
		}
	}

	return node.addChild(tree, newChild)
}

func genUUID() (string, error) {
	u, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	return u.String(), err
}
