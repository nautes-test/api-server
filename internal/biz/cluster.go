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

package biz

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/go-kratos/kratos/v2/log"
	cluster "github.com/nautes-labs/api-server/pkg/cluster"
	utilstrings "github.com/nautes-labs/api-server/util/string"
	resourcev1alpha1 "github.com/nautes-labs/pkg/api/v1alpha1"
	nautesconfigs "github.com/nautes-labs/pkg/pkg/nautesconfigs"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/kops/pkg/kubeconfig"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

const (
	defaultClusterTemplateURL = "https://github.com/nautes-labs/cluster-templates.git"
	SecretPath                = "default"
	SecretEngine              = "pki"
)

type ClusterUsecase struct {
	log              *log.Helper
	secretRepo       Secretrepo
	codeRepo         CodeRepo
	resourcesUsecase *ResourcesUsecase
	configs          *nautesconfigs.Config
	client           client.Client
	cluster          cluster.ClusterRegistrationOperator
	dex              DexRepo
}

type ClusterData struct {
	ClusterName string
	ApiServer   string
	ClusterType string
	Usage       string
	HostCluster string
}

func NewClusterUsecase(logger log.Logger, codeRepo CodeRepo, secretRepo Secretrepo, resourcesUsecase *ResourcesUsecase, configs *nautesconfigs.Config, client client.Client, cluster cluster.ClusterRegistrationOperator, dex DexRepo) *ClusterUsecase {
	return &ClusterUsecase{log: log.NewHelper(log.With(logger)), codeRepo: codeRepo, secretRepo: secretRepo, resourcesUsecase: resourcesUsecase, configs: configs, client: client, cluster: cluster, dex: dex}
}

func (c *ClusterUsecase) CloneRepository(ctx context.Context, url string) (string, error) {
	path, err := c.resourcesUsecase.CloneCodeRepo(ctx, url)
	if err != nil {
		return "", err
	}

	return path, nil
}

func (c *ClusterUsecase) SaveKubeconfig(ctx context.Context, id, server, config string) error {
	config, err := c.ConvertKubeconfig(config, server)
	if err != nil {
		return err
	}
	err = c.secretRepo.SaveClusterConfig(ctx, id, config)
	if err != nil {
		return err
	}

	return nil
}

func (r *ClusterUsecase) ConvertKubeconfig(config, server string) (string, error) {
	kubeconfig := &kubeconfig.KubectlConfig{}
	jsonData, err := yaml.YAMLToJSONStrict([]byte(config))
	if err != nil {
		return "", err
	}

	err = json.Unmarshal([]byte(jsonData), kubeconfig)
	if err != nil {
		return "", err
	}

	if len(kubeconfig.Clusters) < 1 {
		return "", fmt.Errorf("invalid kubeconfig file: must have at least one cluster")
	}

	if len(kubeconfig.Users) < 1 {
		return "", fmt.Errorf("invalid kubeconfig file: must have at least one user")
	}

	kubeconfig.Clusters[0].Cluster.Server = server

	bytes, err := yaml.Marshal(kubeconfig)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

func (c *ClusterUsecase) GetCacert(ctx context.Context) (string, error) {
	secretOptions := &SecretOptions{
		SecretPath:   SecretPath,
		SecretEngine: SecretEngine,
		SecretKey:    "cacert",
	}
	cacert, err := c.secretRepo.GetSecret(ctx, secretOptions)
	if err != nil {
		return "", err
	}

	return cacert, nil
}

func (c *ClusterUsecase) GetTenantRepository(ctx context.Context) (*Project, error) {
	codeRepos := &resourcev1alpha1.CodeRepoList{}
	labelSelector := labels.SelectorFromSet(map[string]string{"coderepo.resource.nautes.io/tenant-management": c.configs.Nautes.TenantName})
	err := c.client.List(context.Background(), codeRepos, &client.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		return nil, err
	}
	if len(codeRepos.Items) == 0 {
		return nil, fmt.Errorf("tenant repository is not found")
	}

	pid, _ := utilstrings.ExtractNumber("repo-", codeRepos.Items[0].Name)
	project, err := c.codeRepo.GetCodeRepo(ctx, pid)
	if err != nil {
		return nil, err
	}

	return project, nil
}

func (c *ClusterUsecase) SaveCluster(ctx context.Context, param *cluster.ClusterRegistrationParam, kubeconfig string) error {
	if ok := cluster.IsVirtualRuntime(param.Cluster); !ok {
		err := c.SaveKubeconfig(ctx, param.Cluster.Name, param.Cluster.Spec.ApiServer, kubeconfig)
		if err != nil {
			c.log.Debugf("failed to  saved kubeconfig to secre repo, cluster name: %s", param.Cluster.Name)
			return err
		}
		c.log.Infof("successfully saved kubeconfig to secre repo, cluster name: %s", param.Cluster.Name)
	}

	cacert, err := c.GetCacert(ctx)
	if err != nil {
		c.log.Debugf("failed to get cacert to secre repo, cluster name: %s", param.Cluster.Name)
		return err
	}

	url := GetClusterTemplateHttpsURL(c.configs)
	clusterTemplateLocalPath, err := c.CloneRepository(ctx, url)
	if err != nil {
		c.log.Debugf("failed to clone cluster template repository, cluster name: %s, url: %s", param.Cluster.Name, url)
		return err
	}
	defer cleanCodeRepo(clusterTemplateLocalPath)

	project, err := c.GetTenantRepository(ctx)
	if err != nil {
		c.log.Debugf("failed to get tenant repository, cluster name: %s", param.Cluster.Name)
		return err
	}
	tenantRepositoryLocalPath, err := c.CloneRepository(ctx, project.HttpUrlToRepo)
	if err != nil {
		c.log.Debugf("failed to clone tenant repository, cluster name: %s, url: %s", param.Cluster.Name, url)
		return err
	}
	defer cleanCodeRepo(tenantRepositoryLocalPath)

	param.ClusterTemplateRepoLocalPath = clusterTemplateLocalPath
	param.CaBundle = base64.StdEncoding.EncodeToString([]byte(cacert))
	param.TenantConfigRepoLocalPath = tenantRepositoryLocalPath
	param.RepoURL = project.SshUrlToRepo
	param.Configs = c.configs
	err = c.cluster.InitializeDependencies(param)
	if err != nil {
		return err
	}
	err = c.cluster.Save()
	if err != nil {
		c.log.Debugf("failed to save cluster, cluster name: %s", param.Cluster.Name)
		return err
	}

	err = c.resourcesUsecase.SaveConfig(ctx, tenantRepositoryLocalPath)
	if err != nil {
		c.log.Debugf("failed to save config to git, cluster name: %s", param.Cluster.Name)
		return err
	}

	c.log.Infof("successfully register cluster, cluster name: %s", param.Cluster.Name)

	err = c.SaveDexConfig(param, tenantRepositoryLocalPath)
	if err != nil {
		return err
	}

	return nil
}

func (c *ClusterUsecase) DeleteCluster(ctx context.Context, clusterName string) error {
	url := GetClusterTemplateHttpsURL(c.configs)
	clusterTemplateLocalPath, err := c.CloneRepository(ctx, url)
	if err != nil {
		c.log.Debugf("failed to clone cluster template repository, cluster name: %s, url: %s", clusterName, url)
		return err
	}
	defer cleanCodeRepo(clusterTemplateLocalPath)

	project, err := c.GetTenantRepository(ctx)
	if err != nil {
		c.log.Debugf("failed to get tenant repository, cluster name: %s", clusterName)
		return err
	}
	tenantRepositoryLocalPath, err := c.CloneRepository(ctx, project.HttpUrlToRepo)
	if err != nil {
		c.log.Debugf("failed to get tenant repository local path, cluster name: %s", clusterName)
		return err
	}
	defer cleanCodeRepo(tenantRepositoryLocalPath)

	clusterFilePath := fmt.Sprintf("%s/nautes/overlays/production/clusters/%s.yaml", tenantRepositoryLocalPath, clusterName)
	clusterResouce, err := GetCluster(clusterFilePath)
	if err != nil {
		c.log.Debugf("failed to get cluster cluster resource, cluster name: %s", clusterName)
		return err
	}

	param := &cluster.ClusterRegistrationParam{
		Cluster:                      clusterResouce,
		RepoURL:                      project.SshUrlToRepo,
		Configs:                      c.configs,
		ClusterTemplateRepoLocalPath: clusterTemplateLocalPath,
		TenantConfigRepoLocalPath:    tenantRepositoryLocalPath,
	}
	err = c.cluster.InitializeDependencies(param)
	if err != nil {
		return err
	}

	err = c.DeleteDexConfig(param)
	if err != nil {
		return err
	}

	err = c.cluster.Remove()
	if err != nil {
		c.log.Debugf("failed to remove cluster, cluster name: %s", clusterName)
		return err
	}

	err = c.resourcesUsecase.SaveConfig(ctx, tenantRepositoryLocalPath)
	if err != nil {
		c.log.Debugf("failed to save config to git, cluster name: %s", clusterName)
		return err
	}

	c.log.Infof("successfully remove cluster, cluster name: %s", clusterName)

	return nil
}

func (c *ClusterUsecase) SaveDexConfig(param *cluster.ClusterRegistrationParam, teantLocalPath string) error {
	if ok := cluster.IsHostCluser(param.Cluster); ok {
		return nil
	}

	// The port of the physical runtime cluster is obtained by user parameters
	// Obtained by querying the traefik port of the host cluster during virtual runtime
	var callback string
	if ok := cluster.IsPhysicalRuntime(param.Cluster); ok {
		callback = fmt.Sprintf("https://%s:%s/api/dex/callback", param.ArgocdHost, param.Traefik.HttpsNodePort)
	} else {
		httpsNodePort, err := c.cluster.GetTraefikNodePortToHostCluster(teantLocalPath, param.Cluster.Spec.HostCluster)
		if err != nil {
			return err
		}
		callback = fmt.Sprintf("https://%s:%s/api/dex/callback", param.ArgocdHost, strconv.Itoa(httpsNodePort))
	}

	err := c.dex.UpdateRedirectURIs(callback)
	if err != nil {
		return err
	}

	return nil
}

func (c *ClusterUsecase) DeleteDexConfig(param *cluster.ClusterRegistrationParam) error {
	if ok := cluster.IsHostCluser(param.Cluster); ok {
		return nil
	}

	url, err := c.cluster.GetArgocdURL()
	if err != nil {
		return err
	}

	callback := fmt.Sprintf("%s/api/dex/callback", url)
	err = c.dex.RemoveRedirectURIs(callback)
	if err != nil {
		return err
	}

	return nil
}

func GetClusterTemplateHttpsURL(configs *nautesconfigs.Config) string {
	if configs.Nautes.RuntimeTemplateSource != "" {
		return configs.Nautes.RuntimeTemplateSource
	}

	return defaultClusterTemplateURL
}

func GetCluster(filename string) (*resourcev1alpha1.Cluster, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil, err
	}

	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var cluster resourcev1alpha1.Cluster
	err = yaml.Unmarshal(content, &cluster)
	if err != nil {
		return nil, err
	}

	return &cluster, nil
}
