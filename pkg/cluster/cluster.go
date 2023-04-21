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
	"fmt"
	"io/ioutil"
	"os"

	utilstrings "github.com/nautes-labs/api-server/util/string"
	v1 "k8s.io/api/core/v1"
	yaml "sigs.k8s.io/yaml"
)

const (
	_HostClusterDirectoreyPlaceholder                    = "_HOST_CLUSTER_"
	_RuntimeDirectoryDirectoreyPlaceholder               = "_RUNTIME_"
	_VclusterDirectoryDirectoreyPlaceholder              = "_VCLUSTER_"
	_HostClusterType                        ClusterUsage = "HostCluster"
	_PhysicalRuntime                        ClusterUsage = "PhysicalRuntime"
	_VirtualRuntime                         ClusterUsage = "VirtualRuntime"
)

func NewClusterRegistration() ClusterRegistrationOperator {
	return &ClusterRegistration{}
}

func (r *ClusterRegistration) InitializeDependencies(param *ClusterRegistrationParam) error {
	var hostCluster *HostCluster
	var runtime *Runtime
	var vcluster *Vcluster
	var hostClusterNames []string
	var err error

	if ok := IsHostCluser(param.Cluster); ok {
		hostCluster = &HostCluster{
			Name:          param.Cluster.Name,
			ApiServer:     param.Cluster.Spec.ApiServer,
			ArgocdProject: param.Configs.Nautes.TenantName,
		}
	}

	if ok := IsHostCluser(param.Cluster); !ok {
		runtime = &Runtime{
			Name:          fmt.Sprintf("%s-runtime", param.Cluster.Name),
			Type:          getRuntimeType(param.Cluster),
			MountPath:     param.Cluster.Name,
			ApiServer:     param.Cluster.Spec.ApiServer,
			ArgocdProject: fmt.Sprintf("%s-runtime-project", param.Cluster.Name),
		}
	}

	if ok := IsVirtualRuntime(param.Cluster); ok {
		vcluster, err = ConstructVcluster(param)
		if err != nil {
			return err
		}

		// Virtual cluster argocd port uses the traefik of the host cluster
		httpsNodePort, err := r.GetTraefikNodePortToHostCluster(param.TenantConfigRepoLocalPath, param.Cluster.Spec.HostCluster)
		if err != nil {
			return fmt.Errorf("failed to get host cluster %s tarefik https NodePort, please check if the host cluster exists", param.Cluster.Spec.HostCluster)
		}
		// Automatically obtain argocd domain name
		if param.ArgocdHost == "" {
			hostCluster, err := getHostCluster(param.TenantConfigRepoLocalPath, param.Cluster.Spec.HostCluster, param.Configs.Nautes.TenantName)
			if err != nil {
				return fmt.Errorf("argocd host not filled in and automatic acquisition of host cluster IP failed, err: %v", err)
			}
			ip, err := utilstrings.ParseUrl(hostCluster.ApiServer)
			if err != nil {
				return fmt.Errorf("argocd host not filled in and automatic parse host cluster IP failed, err: %v", err)
			}
			param.ArgocdHost = fmt.Sprintf("argocd.%s.%s.nip.io", param.Cluster.Name, ip)
		}
		argocdURL := fmt.Sprintf("https://%s:%d", param.ArgocdHost, httpsNodePort)
		runtime.Argocd = &Argocd{
			Host: param.ArgocdHost,
			URL:  argocdURL,
		}
	}

	if IsPhysicalRuntime(param.Cluster) {
		if param.ArgocdHost == "" {
			ip, err := utilstrings.ParseUrl(param.Cluster.Spec.ApiServer)
			if err != nil {
				return fmt.Errorf("argocd host not filled in and automatic parse host cluster IP failed, err: %v", err)
			}
			param.ArgocdHost = fmt.Sprintf("argocd.%s.%s.nip.io", param.Cluster.Name, ip)
		}

		if param.Traefik != nil {
			argocdURL := fmt.Sprintf("https://%s:%s", param.ArgocdHost, param.Traefik.HttpsNodePort)
			runtime.Argocd = &Argocd{
				Host: param.ArgocdHost,
				URL:  argocdURL,
			}
		}
	}

	usage := _PhysicalRuntime
	if hostCluster != nil {
		usage = _HostClusterType
	}
	if vcluster != nil {
		usage = _VirtualRuntime
	}

	*r = ClusterRegistration{
		Cluster:                      param.Cluster,
		ClusterTemplateRepoLocalPath: param.ClusterTemplateRepoLocalPath,
		TenantConfigRepoLocalPath:    param.TenantConfigRepoLocalPath,
		RepoURL:                      param.RepoURL,
		CaBundle:                     param.CaBundle,
		Usage:                        usage,
		HostClusterNames:             hostClusterNames,
		HostCluster:                  hostCluster,
		Vcluster:                     vcluster,
		Runtime:                      runtime,
		Traefik:                      param.Traefik,
		NautesConfigs:                param.Configs.Nautes,
		SecretConfigs:                param.Configs.Secret,
		OauthConfigs:                 param.Configs.OAuth,
		GitConfigs:                   param.Configs.Git,
	}

	return nil
}

func ConstructVcluster(param *ClusterRegistrationParam) (*Vcluster, error) {
	vcluster := &Vcluster{
		Name:      param.Cluster.Name,
		ApiServer: param.Cluster.Spec.ApiServer,
		Namespace: param.Cluster.Name,
	}
	// NodePort of vcluster
	if param.Vcluster != nil && param.Vcluster.HttpsNodePort != "" {
		vcluster.HttpsNodePort = param.Vcluster.HttpsNodePort
	} else {
		if vcluster.ApiServer == "" {
			return nil, fmt.Errorf("the apiserver of vcluster %s is not empty", vcluster.Name)
		}

		port, err := utilstrings.ExtractPortFromURL(vcluster.ApiServer)
		if err != nil {
			return nil, fmt.Errorf("failed to automatically obtain vcluster host, err: %w", err)
		}
		vcluster.HttpsNodePort = port
	}

	// Get hostcluster information from the tenant configuration library
	hostCluster, err := getHostCluster(param.TenantConfigRepoLocalPath, param.Cluster.Spec.HostCluster, param.Configs.Nautes.TenantName)
	if err != nil {
		return nil, err
	}
	vcluster.HostCluster = hostCluster
	// Set host tls
	hostClusterIP, err := utilstrings.ParseUrl(hostCluster.ApiServer)
	if err != nil {
		return nil, err
	}
	vcluster.TlsSan = hostClusterIP

	return vcluster, nil
}

func (r *ClusterRegistration) GetArgocdURL() (string, error) {
	filename := fmt.Sprintf("%s/%s-runtime/argocd/overlays/production/patch-argocd-cm.yaml", GetRuntimesDir(r.TenantConfigRepoLocalPath), r.Cluster.Name)
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return "", err
	}

	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}
	var cm v1.ConfigMap
	if err := yaml.Unmarshal(bytes, &cm); err != nil {
		return "", err
	}

	return cm.Data["url"], nil
}

func (r *ClusterRegistration) GetTraefikNodePortToHostCluster(tenantLocalPath, hostCluster string) (int, error) {
	traefikFileName := fmt.Sprintf("%s/%s/production/traefik-app.yaml", GetHostClustesrDir(tenantLocalPath), hostCluster)
	app, err := parseArgocdApplication(traefikFileName)
	if err != nil {
		return 0, err
	}

	httpsNodePort, err := getTraefikHttpsNodePort(app)
	if err != nil {
		return 0, err
	}

	return httpsNodePort, nil
}

func getHostCluster(tenantConfigRepoLocalPath, hostClusterName, tenantName string) (*HostCluster, error) {
	clusterFileName := fmt.Sprintf("%s/%s.yaml", GetClustersDir(tenantConfigRepoLocalPath), hostClusterName)
	if _, err := os.Stat(clusterFileName); os.IsNotExist(err) {
		return nil, err
	}

	cluster, err := parseCluster(clusterFileName)
	if err != nil {
		return nil, err
	}

	return &HostCluster{
		Name:          cluster.Name,
		ApiServer:     cluster.Spec.ApiServer,
		ArgocdProject: tenantName,
	}, nil
}
