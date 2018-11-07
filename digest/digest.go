package digest

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/jasonrogena/log-analyse/config"
	"github.com/jasonrogena/log-analyse/types"
)

type UrlPathDigester struct {
	tree *Tree
}

type DigestPayload struct {
	Field *types.Field
	Cnf   *config.Config
}

func (digester UrlPathDigester) Absorb(someData interface{}) (err error) {
	config, ok := someData.(*config.Config)
	if ok {
		fmt.Printf("Tree has the %d root nodes\n", len(digester.tree.rootNodes))
		err = digester.tree.generalizeTree(config)
	}
	return
}

func getField(someData interface{}) (*types.Field, bool) {
	field, ok := someData.(*types.Field)
	if ok {
		if field.FieldType.Name != "request" {
			ok = false
		}
	}
	return field, ok
}

func GetDigestPayload(someData interface{}) (DigestPayload, bool) {
	payload, ok := someData.(DigestPayload)
	if ok {
		if payload.Field == nil || payload.Field.FieldType.Name != "request" {
			ok = false
		}
	}
	return payload, ok
}

func GetUriParts(request string, uriRegex string) ([]string, error) {
	urlRegex, regexErr := regexp.Compile(uriRegex)
	if regexErr == nil && len(request) > 0 {
		//GET /api/v1/forms/197928.json?a=dfds HTTP/1.1
		reqPartsArr := strings.Split(request, " ")
		if len(reqPartsArr) == 3 {
			reqPathWithArgs := CleanUri(reqPartsArr[1])
			if urlRegex.MatchString(reqPathWithArgs) {
				reqPathParts := strings.Split(reqPathWithArgs, "?")
				cleanReqPath := reqPathParts[0]
				cleanReqPathParts := strings.Split(cleanReqPath, "/")
				return cleanReqPathParts, nil
			} else {
				return nil, errors.New("Request uri did not match regex '" + uriRegex + "'")
			}
		} else {
			return nil, errors.New("Request string doesn't have three parts")
		}
	} else if regexErr != nil {
		return nil, regexErr
	} else {
		return nil, errors.New("Request is an empty string")
	}
}

func (digester UrlPathDigester) Digest(someData interface{}) error {
	payload, payloadOk := GetDigestPayload(someData)
	if payloadOk {
		field := payload.Field
		uriParts, uriErr := GetUriParts(field.ValueString, payload.Cnf.Digest.UriRegex)
		if uriErr == nil {
			addPathErr := digester.tree.addRequestPath(nil, digester.tree.rootNodes, uriParts, -1, field)
			if addPathErr != nil {
				return addPathErr
			}
		} else {
			return uriErr
		}
	}

	return nil
}

func CleanUri(uri string) string {
	re := regexp.MustCompile(`//+`)
	uri = re.ReplaceAllString(uri, "/")
	return uri
}

func (tree *Tree) addRequestPath(parentNode *TreeNode, nodes map[string]*TreeNode, requestPathParts []string, lastVisitedPathIndex int, field *types.Field) error {
	if lastVisitedPathIndex < len(requestPathParts) {
		nodeFound := false
		var foundNode *TreeNode
		pathIndex := lastVisitedPathIndex + 1
		for _, curNode := range nodes {
			if curNode.value == GENERIC_VALUE {
				for curCombinedIndex := range curNode.combinedValues {
					if curNode.combinedValues[curCombinedIndex] == requestPathParts[pathIndex] {
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

		if !nodeFound { // no node in current level has the value
			newNodeUuid, uuidErr := genUUID()
			if uuidErr != nil {
				return uuidErr
			}
			newTreeNode := TreeNode{level: pathIndex, uuid: newNodeUuid, value: requestPathParts[pathIndex], parent: parentNode}
			foundNode = &newTreeNode

			if parentNode == nil { // newTreeNode is a root node
				tree.addNodeToIndex(foundNode)
				nodes[foundNode.uuid] = foundNode
			} else {
				parentNode.addChild(tree, foundNode)
			}
		}

		if foundNode != nil {
			if pathIndex == len(requestPathParts)-1 { // leaf node for this request path
				foundNode.payload = append(foundNode.payload, field)
			} else {
				addReqErr := tree.addRequestPath(foundNode, foundNode.children, requestPathParts, pathIndex, field)
				if addReqErr != nil {
					return addReqErr
				}
			}
		}
	}

	return nil
}

func (digester UrlPathDigester) IsDigestable(someData interface{}) bool {
	_, ok := getField(someData)
	return ok
}

func InitUrlPathDigester(rbfsLayerCap int) (digester UrlPathDigester) {
	rootNodes := make(map[string]*TreeNode)
	index := make(map[int]map[string]*TreeNode)
	tree := Tree{rootNodes: rootNodes, inOrderNodeIndex: index, rbfsLayerCap: rbfsLayerCap}
	digester = UrlPathDigester{tree: &tree}
	return
}

func (tree *Tree) generalizeTree(config *config.Config) error {
	// Use reverse breadth first search to travers the tree, starting from it's leaf nodes
	var treeLayers []int

	for cl := range tree.inOrderNodeIndex {
		treeLayers = append(treeLayers, cl)
	}
	sort.Ints(treeLayers)
	fmt.Printf("Going through %d tree layers\n", len(treeLayers))

	// perform the traversal, starting from the bottom most (bottom most leafs) layer
	for curIndex := len(treeLayers) - 1; curIndex >= 0; curIndex-- {
		curLayer := treeLayers[curIndex]
		for curUUID := range tree.inOrderNodeIndex[curLayer] { //TODO: Does range work with growing map
			tree.inOrderNodeIndex[curLayer][curUUID].generalizeTreeNode(tree)
		}
	}

	// write to the database
	for curIndex := len(treeLayers) - 1; curIndex >= 0; curIndex-- {
		curLayer := treeLayers[curIndex]
		for curUUID := range tree.inOrderNodeIndex[curLayer] {
			if len(tree.inOrderNodeIndex[curLayer][curUUID].payload) > 0 {
				curNode := tree.inOrderNodeIndex[curLayer][curUUID]
				nodePath, noPermutations := curNode.reconstructPath("", 1)
				if config.Digest.MinPathPermutations <= noPermutations {
					fmt.Printf("About to store %q as a generalized path (%d permutations)\n", nodePath, noPermutations)
				}
			}
		}
	}

	return nil
}

func (tree *Tree) printOrderedIndexTree() {
	var treeLayers []int

	for cl := range tree.inOrderNodeIndex {
		treeLayers = append(treeLayers, cl)
	}
	sort.Ints(treeLayers)

	for curIndex := len(treeLayers) - 1; curIndex >= 0; curIndex-- {
		curLayer := treeLayers[curIndex]
		fmt.Printf("Layer %d has these many nodes %d \n", curLayer, len(tree.inOrderNodeIndex[curLayer]))
		for curUUID := range tree.inOrderNodeIndex[curLayer] { //TODO: Does range work with growing map
			fmt.Printf("%s (%v) ", tree.inOrderNodeIndex[curLayer][curUUID].value, tree.inOrderNodeIndex[curLayer][curUUID].combinedValues)
			if curLayer == 4 {
				fmt.Printf("_%s_ ", tree.inOrderNodeIndex[curLayer][curUUID].parent.value)
			}
		}
		fmt.Printf("\n")
	}
}

func (node *TreeNode) reconstructPath(curPath string, noPermutations int) (string, int) {
	if len(curPath) > 0 {
		curPath = "/" + curPath
	}
	curPath = node.value + curPath
	noPermutations = noPermutations * node.getNoValuePermutations()

	if node.parent != nil {
		curPath, noPermutations = node.parent.reconstructPath(curPath, noPermutations)
	}

	return curPath, noPermutations
}

func (node *TreeNode) generalizeTreeNode(tree *Tree) error {
	// Check each of it's children, combine those that have equivalent child trees
	if node.children != nil {
		var childUUIDS []string
		for curUUID := range node.children {
			childUUIDS = append(childUUIDS, curUUID)
		}

		for x, _ := range childUUIDS {
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
