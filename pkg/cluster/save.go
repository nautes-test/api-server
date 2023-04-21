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

package cluster

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/nautes-labs/api-server/pkg/nodestree"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"
)

func (r *ClusterRegistration) Save() error {
	config, err := NewClusterFileConfig(r.ClusterTemplateRepoLocalPath)
	if err != nil {
		return err
	}

	nodes, err := r.SaveClusterByType(config)
	if err != nil {
		return err
	}

	err = r.WriteFileToTenantConfigRepository(nodes)
	if err != nil {
		return err
	}

	err = r.WriteCluster()
	if err != nil {
		return err
	}

	return nil
}

func (r *ClusterRegistration) SaveClusterByType(config *Config) (*nodestree.Node, error) {
	var nodes nodestree.Node
	var ignorePath, ignoreFile []string
	var err error

	switch r.Usage {
	case _HostClusterType:
		ignorePath, ignoreFile = config.GetSaveHostClusterConfig()
		nodes, err = r.LoadTemplateNodesTree(ignorePath, ignoreFile)
		if err != nil {
			return nil, err
		}
		err = r.SaveHostCluster(&nodes)
		if err != nil {
			return nil, err
		}
	case _PhysicalRuntime:
		ignorePath, ignoreFile = config.GetSavePhysicalRuntimeConfig()
		nodes, err = r.LoadTemplateNodesTree(ignorePath, ignoreFile)
		if err != nil {
			return nil, err
		}
		err = r.SaveRuntime(&nodes)
		if err != nil {
			return nil, err
		}
	case _VirtualRuntime:
		ignorePath, ignoreFile = config.GetSaveVirtualRuntimeConfig()
		nodes, err = r.LoadTemplateNodesTree(ignorePath, ignoreFile)
		if err != nil {
			return nil, err
		}
		err = r.SaveRuntime(&nodes)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("unknown cluster usage")
	}

	return &nodes, nil
}

func (r *ClusterRegistration) SaveClusterToKustomization() (err error) {
	kustomizationFilePath := fmt.Sprintf("%s/nautes/overlays/production/clusters/kustomization.yaml", r.TenantConfigRepoLocalPath)
	if _, err := os.Stat(kustomizationFilePath); os.IsNotExist(err) {
		return err
	}

	bytes, err := os.ReadFile(kustomizationFilePath)
	if err != nil {
		return
	}
	var kustomization types.Kustomization
	err = yaml.Unmarshal(bytes, &kustomization)
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("%s.yaml", r.Cluster.Name)
	if len(kustomization.Resources) > 0 {
		r.ClusterResouceFiles = AddIfNotExists(kustomization.Resources, filename)
	} else {
		r.ClusterResouceFiles = append(r.ClusterResouceFiles, filename)
	}

	return nil
}

func (r *ClusterRegistration) LoadTemplateNodesTree(ignorePath, ignoreFile []string) (nodes nodestree.Node, err error) {
	fileOptions := &nodestree.FileOptions{
		IgnorePath:       ignorePath,
		IgnoreFile:       ignoreFile,
		ExclusionsSuffix: []string{".txt", ".md"},
		ContentType:      nodestree.StringContentType,
	}
	client := nodestree.NewNodestree(fileOptions, nil, nil)
	nodes, err = client.Load(r.ClusterTemplateRepoLocalPath)
	if err != nil {
		return
	}

	return
}

func (r *ClusterRegistration) SaveHostCluster(nodes *nodestree.Node) error {
	err := r.SaveClusterToKustomization()
	if err != nil {
		return err
	}

	err = r.GetAndMergeHostClusterNames()
	if err != nil {
		return err
	}

	err = r.Execute(nodes)
	if err != nil {
		return err
	}

	r.OverlayTemplateDirectoryPlaceholder(nodes, _HostClusterDirectoreyPlaceholder, r.HostCluster.Name)
	r.ReplaceFilePath(nodes)

	return nil
}

func (r *ClusterRegistration) CheckHostClusterDirExists() error {
	if r.Usage == _VirtualRuntime && r.Vcluster != nil {
		hostClusterDir := fmt.Sprintf("%s/host-clusters/%s", r.TenantConfigRepoLocalPath, r.Vcluster.HostCluster.Name)
		_, err := os.Stat(hostClusterDir)
		if err != nil {
			return fmt.Errorf("the specified host cluster for this virtual cluster does not exist")
		}
	}

	return nil
}

func (r *ClusterRegistration) SaveRuntime(nodes *nodestree.Node) error {
	err := r.CheckHostClusterDirExists()
	if err != nil {
		return err
	}

	err = r.SaveClusterToKustomization()
	if err != nil {
		return err
	}

	err = r.GetAndMergeVclusterNames()
	if err != nil {
		return err
	}

	err = r.Execute(nodes)
	if err != nil {
		return err
	}

	if r.Usage == _VirtualRuntime {
		r.OverlayTemplateDirectoryPlaceholder(nodes, _VclusterDirectoryDirectoreyPlaceholder, r.Vcluster.Name)
		r.OverlayTemplateDirectoryPlaceholder(nodes, _HostClusterDirectoreyPlaceholder, r.Vcluster.HostCluster.Name)
	}

	r.OverlayTemplateDirectoryPlaceholder(nodes, _RuntimeDirectoryDirectoreyPlaceholder, r.Runtime.Name)

	r.ReplaceFilePath(nodes)

	return nil
}

func (r *ClusterRegistration) GetAndMergeHostClusterNames() error {
	if r.HostCluster == nil || r.HostCluster.Name == "" {
		return nil
	}

	path := GetHostClusterAppsetFilePath(r.TenantConfigRepoLocalPath)
	clusterNames, err := GetHostClusterNames(path)
	if err != nil {
		return err
	}
	r.HostClusterNames = AddIfNotExists(clusterNames, r.HostCluster.Name)

	return nil
}

func (r *ClusterRegistration) GetAndMergeVclusterNames() error {
	if r.Vcluster == nil {
		return nil
	}

	vclusterAppsetFilePath := fmt.Sprintf("%s/host-clusters/%s/production/vcluster-appset.yaml", r.TenantConfigRepoLocalPath, r.Vcluster.HostCluster.Name)
	clusterNames, err := GetVclusterNames(vclusterAppsetFilePath)
	if err != nil {
		return err
	}

	r.VclusterNames = AddIfNotExists(clusterNames, r.Vcluster.Name)

	return nil
}

func (r *ClusterRegistration) WriteFileToTenantConfigRepository(nodes *nodestree.Node) error {
	for _, node := range nodes.Children {
		if node.IsDir {
			err := r.WriteFileToTenantConfigRepository(node)
			if err != nil {
				return err
			}
		} else {
			content, ok := node.Content.(string)
			if ok {
				err := WriteConfigFile(node.Path, content)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (r *ClusterRegistration) WriteCluster() error {
	bytes, err := yaml.Marshal(r.Cluster)
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("%s/%s.yaml", GetClustersDir(r.TenantConfigRepoLocalPath), r.Cluster.Name)
	err = ioutil.WriteFile(filename, bytes, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (r *ClusterRegistration) OverlayTemplateDirectoryPlaceholder(nodes *nodestree.Node, placeholder string, replaceVaule string) {
	for _, node := range nodes.Children {
		if ok := strings.Contains(node.Path, placeholder); ok {
			if node.IsDir {
				node.Name = replaceVaule
			}
			node.Path = ReplaceFilePath(node.Path, placeholder, replaceVaule)
		}

		if node.IsDir {
			r.OverlayTemplateDirectoryPlaceholder(node, placeholder, replaceVaule)
		}
	}
}

func (r *ClusterRegistration) ReplaceFilePath(nodes *nodestree.Node) {
	var oldDir, newDir string
	oldDir = r.ClusterTemplateRepoLocalPath
	newDir = r.TenantConfigRepoLocalPath

	for _, node := range nodes.Children {
		node.Path = ReplaceFilePath(node.Path, oldDir, newDir)
		if node.IsDir {
			r.ReplaceFilePath(node)
		}
	}
}

func (r *ClusterRegistration) Execute(nodes *nodestree.Node) error {
	for _, node := range nodes.Children {
		if node.IsDir {
			err := r.Execute(node)
			if err != nil {
				return err
			}
		}

		if content, ok := node.Content.(string); ok {
			t, err := template.New(node.Name).Funcs(template.FuncMap{
				"split": strings.Split,
			}).Parse(content)
			if err != nil {
				return err
			}

			buf := new(bytes.Buffer)
			err = t.Execute(buf, r)
			if err != nil {
				return err
			}
			node.Content = buf.String()
		}
	}

	return nil
}

func WriteConfigFile(filePath, content string) error {
	dir := filepath.Dir(filePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return err
		}
	}

	err := ioutil.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return err
	}

	return nil
}

func ReplaceFilePath(filePath, oldDir, newDir string) (newPath string) {
	return strings.Replace(filePath, oldDir, newDir, 1)
}
