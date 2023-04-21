// Copyright 2023 Nautes Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package nodestree

import (
	"fmt"
)

type nodeInfo struct {
	Kind        string
	Count       int
	TargetCount int
}

func CheckNumberOfResources(options CompareOptions, in *nodesTree) error {
	config := in.config
	configInfos := listConfigInfos([]Config{*config})

	ok, err := IsEmptyParentNode(options, configInfos)
	if ok && err != nil {
		return err
	}

	err = checkCountResource(options, configInfos)
	if err != nil {
		return err
	}

	return nil
}

// checkReqiredResource check whether the node parent is empty
func IsEmptyParentNode(options CompareOptions, configs []configInfo) (bool, error) {
	for _, node := range options.Nodes.Children {
		for _, config := range configs {
			if !node.IsDir && node.Name != config.Category && config.Optional {
				continue
			}

			// Check whether the node has resources
			// nodes := searchSubNodes(node, config.Level)
			// for _, n := range nodes {
			// 	if len(n.Children) == 0 {
			// 		return true, fmt.Errorf("there are must be 1 resource in the %s directory", n.Name)
			// 	}
			// }
		}
	}
	return false, nil
}

// searchSubNodes search the parent node of node
func searchSubNodes(node *Node, level int) (nodes []*Node) {
	if !node.IsDir {
		return
	}

	if (node.Level + 1) == level {
		nodes = append(nodes, node)
		return
	} else {
		for _, n := range node.Children {
			nodes = append(nodes, searchSubNodes(n, level)...)
		}
	}

	return
}

func checkCountResource(options CompareOptions, configs []configInfo) error {
	mapping := generateNodeInfos(options, configs)
	for _, m := range mapping {
		if m.TargetCount > 1 && m.Count > m.TargetCount {
			return fmt.Errorf("the number of %s resources does not match, expected %d but is %d", m.Kind, m.TargetCount, m.Count)
		}
	}

	return nil
}

func generateNodeInfos(options CompareOptions, configs []configInfo) map[string]*nodeInfo {
	var mapping = make(map[string]*nodeInfo)

	for _, node := range options.Nodes.Children {
		for _, config := range configs {
			if node.IsDir {
				newOptions := CompareOptions{
					Nodes:            *node,
					ProductName:      options.ProductName,
					LocalProjectPath: options.LocalProjectPath,
				}

				newMapping := generateNodeInfos(newOptions, configs)

				for key, value := range newMapping {
					mapping[key] = value
				}
			} else {
				if config.Kind == node.Kind {
					mapping = setMapping(mapping, node, config.Count)
				}
			}
		}
	}

	return mapping
}

func setMapping(mapping map[string]*nodeInfo, node *Node, targetCount int) map[string]*nodeInfo {
	val, ok := mapping[node.Kind]
	if !ok {
		mapping[node.Kind] = &nodeInfo{
			Kind:        node.Kind,
			Count:       1,
			TargetCount: targetCount,
		}
	} else {
		val.Count += 1
	}

	return mapping
}
