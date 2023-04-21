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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/nautes-labs/api-server/pkg/nodestree"
	"sigs.k8s.io/kustomize/api/types"
	yaml "sigs.k8s.io/yaml"
)

func (r *ClusterRegistration) Remove() error {
	config, err := NewClusterFileConfig(r.ClusterTemplateRepoLocalPath)
	if err != nil {
		return err
	}

	nodes, err := r.DeleteClusterByType(config)
	if err != nil {
		return err
	}

	err = r.WriteFileToTenantConfigRepository(nodes)
	if err != nil {
		return err
	}

	err = r.DeleteCluster()
	if err != nil {
		return err
	}

	err = r.DeleteFileIfAppSetHasNoElements()
	if err != nil {
		return err
	}

	return nil
}

func (r *ClusterRegistration) DeleteClusterByType(config *Config) (*nodestree.Node, error) {
	var nodes nodestree.Node
	var ignorePath, ignoreFile []string
	var err error

	switch r.Usage {
	case _HostClusterType:
		ignorePath, ignoreFile = config.GetRemoveHostClusterConfig()
		nodes, err = r.LoadTemplateNodesTree(ignorePath, ignoreFile)
		if err != nil {
			return nil, err
		}
		err = r.DeleteHostCluster(&nodes)
		if err != nil {
			return nil, err
		}
	case _PhysicalRuntime:
		ignorePath, ignoreFile = config.GetRemovePhysicalRuntimeConfig()
		nodes, err = r.LoadTemplateNodesTree(ignorePath, ignoreFile)
		if err != nil {
			return nil, err
		}
		err = r.DeleteRuntime(&nodes)
		if err != nil {
			return nil, err
		}
	case _VirtualRuntime:
		ignorePath, ignoreFile = config.GetRemoveVirtualRuntimeConfig()
		nodes, err = r.LoadTemplateNodesTree(ignorePath, ignoreFile)
		if err != nil {
			return nil, err
		}
		err = r.DeleteRuntime(&nodes)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("unknown cluster usage")
	}

	return &nodes, nil
}

func (r *ClusterRegistration) DeleteHostCluster(nodes *nodestree.Node) error {
	dir := fmt.Sprintf("%s/%s", GetHostClustesrDir(r.TenantConfigRepoLocalPath), r.Cluster.Name)
	vclustersDir := fmt.Sprintf("%s/vclusters", dir)
	exist := isDirExist(vclustersDir)
	if exist {
		return fmt.Errorf("unable to delete cluster %s because the host cluster is referenced by other virtual cluster", r.Cluster.Name)
	}

	err := DeleteSpecifyDir(dir)
	if err != nil {
		return err
	}

	err = r.DeleteClusterToKustomization()
	if err != nil {
		return err
	}

	err = r.GetAndDeleteHostClusterNames()
	if err != nil {
		return err
	}

	err = r.Execute(nodes)
	if err != nil {
		return err
	}

	r.ReplaceFilePath(nodes)

	return nil
}

func (r *ClusterRegistration) DeleteClusterToKustomization() (err error) {
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
		r.ClusterResouceFiles = RemoveStringFromArray(kustomization.Resources, filename)
	} else {
		r.ClusterResouceFiles = append(r.ClusterResouceFiles, filename)
	}

	return nil
}

func (r *ClusterRegistration) GetAndDeleteHostClusterNames() error {
	hostClusterAppsetFilePath := fmt.Sprintf("%s/host-cluster-appset.yaml", GetTenantProductionDir(r.TenantConfigRepoLocalPath))
	clusterNames, err := GetHostClusterNames(hostClusterAppsetFilePath)
	if err != nil {
		return err
	}

	r.HostClusterNames = RemoveStringFromArray(clusterNames, r.HostCluster.Name)

	return nil
}

func (r *ClusterRegistration) GetAndDeleteVclusterNames() error {
	vclusterAppsetFilePath := fmt.Sprintf("%s/%s/production/vcluster-appset.yaml", GetHostClustesrDir(r.TenantConfigRepoLocalPath), r.Vcluster.HostCluster)
	clusterNames, err := GetVclusterNames(vclusterAppsetFilePath)
	if err != nil {
		return err
	}

	r.VclusterNames = RemoveStringFromArray(clusterNames, r.Vcluster.Name)

	return nil
}

func (r *ClusterRegistration) DeleteRuntime(nodes *nodestree.Node) error {
	runtimeDir := fmt.Sprintf("%s/%s", GetRuntimesDir(r.TenantConfigRepoLocalPath), r.Runtime.Name)
	err := DeleteSpecifyDir(runtimeDir)
	if err != nil {
		return err
	}

	r.ReplaceFilePath(nodes)

	if r.Usage == _VirtualRuntime {
		vclustersDir := fmt.Sprintf("%s/%s", GetVclustersDir(r.TenantConfigRepoLocalPath, r.Vcluster.HostCluster.Name), r.Cluster.Name)
		err := DeleteSpecifyDir(vclustersDir)
		if err != nil {
			return err
		}

		r.OverlayTemplateDirectoryPlaceholder(nodes, _HostClusterDirectoreyPlaceholder, r.Vcluster.HostCluster.Name)

		err = r.GetAndDeleteVclusterNames()
		if err != nil {
			return err
		}
	}

	err = r.DeleteClusterToKustomization()
	if err != nil {
		return err
	}

	err = r.Execute(nodes)
	if err != nil {
		return err
	}

	return nil
}

func (r *ClusterRegistration) DeleteFileIfAppSetHasNoElements() error {
	if len(r.HostClusterNames) == 0 {
		err := r.DeleteHostClusterAppSet()
		if err != nil {
			return err
		}
	}

	if len(r.VclusterNames) == 0 {
		err := r.DeleteVclusterAppSet()
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *ClusterRegistration) DeleteHostClusterAppSet() error {
	if r.Usage != _HostClusterType {
		return nil
	}

	hostClusterAppSetPath := fmt.Sprintf("%s/host-cluster-appset.yaml", GetTenantProductionDir(r.TenantConfigRepoLocalPath))
	err := DeleteSpecifyFile(hostClusterAppSetPath)
	if err != nil {
		return err
	}

	return nil
}

func (r *ClusterRegistration) DeleteRuntimeAppSet() error {
	if r.Usage == _HostClusterType {
		return nil
	}

	tenantProductionDir := GetTenantProductionDir(r.TenantConfigRepoLocalPath)
	err := filepath.Walk(tenantProductionDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && IsValidRuntimeAppSetFilename(info.Name()) {
			err := DeleteSpecifyFile(path)
			if err != nil {
				return fmt.Errorf("failed to delete file %s", info.Name())
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func IsValidRuntimeAppSetFilename(filename string) bool {
	pattern := "^runtime(-[a-zA-Z0-9]+)*-appset\\.yaml$"
	match, err := regexp.MatchString(pattern, filename)
	if err != nil {
		return false
	}
	return match
}

func (r *ClusterRegistration) DeleteVclusterAppSet() error {
	if r.Usage != _VirtualRuntime {
		return nil
	}

	vclusterAppSetPath := fmt.Sprintf("%s/%s/production/vcluster-appset.yaml", GetHostClustesrDir(r.TenantConfigRepoLocalPath), r.Vcluster.HostCluster.Name)
	err := DeleteSpecifyFile(vclusterAppSetPath)
	if err != nil {
		return err
	}

	return nil
}

func DeleteSpecifyDir(dir string) error {
	fileInfo, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}

	if !fileInfo.IsDir() {
		return fmt.Errorf("'%s' is not a directory", dir)
	}

	return os.RemoveAll(dir)
}

func DeleteSpecifyFile(filename string) error {
	fileInfo, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}

	if fileInfo.IsDir() {
		return fmt.Errorf("'%s' is a directory", filename)
	}

	return os.Remove(filename)
}

func (r *ClusterRegistration) DeleteCluster() error {
	filename := fmt.Sprintf("%s/%s.yaml", GetClustersDir(r.TenantConfigRepoLocalPath), r.Cluster.Name)
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}

	return os.Remove(filename)
}

func isDirExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func RemoveStringFromArray(arr []string, target string) []string {
	for i := 0; i < len(arr); i++ {
		if arr[i] == target {
			arr = append(arr[:i], arr[i+1:]...)
			i--
		}
	}
	return arr
}
