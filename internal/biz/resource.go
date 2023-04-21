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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/nautes-labs/api-server/pkg/nodestree"
	utilstrings "github.com/nautes-labs/api-server/util/string"
	nautesconfigs "github.com/nautes-labs/pkg/pkg/nautesconfigs"
	sjson "github.com/tidwall/sjson"
	kustomize "sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"
)

const (
	KustomizationFileName                 = "kustomization.yaml"
	RretryCount           RretryCountType = "RretryCount"
	_ProductPrefix                        = "product-"
)

type RretryCountType string
type getResouceName func(nodes nodestree.Node) (string, error)
type isDeleteAllowed func(nodes nodestree.Node, resourceName string) (bool, error)

type ResourcesUsecase struct {
	log        log.Logger
	codeRepo   CodeRepo
	secretRepo Secretrepo
	gitRepo    GitRepo
	nodestree  nodestree.NodesTree
	configs    *nautesconfigs.Config
}

func NewResourcesUsecase(log log.Logger, codeRepo CodeRepo, secretRepo Secretrepo, gitRepo GitRepo, nodestree nodestree.NodesTree, configs *nautesconfigs.Config) *ResourcesUsecase {
	return &ResourcesUsecase{
		log:        log,
		codeRepo:   codeRepo,
		secretRepo: secretRepo,
		gitRepo:    gitRepo,
		nodestree:  nodestree,
		configs:    configs,
	}
}

func (r *ResourcesUsecase) Get(ctx context.Context, resourceKind, productName string, operator nodestree.NodesOperator, getResourceName getResouceName) (*nodestree.Node, error) {
	product, project, err := r.GetProductAndCodeRepo(ctx, productName)
	if err != nil {
		return nil, err
	}

	localPath, err := r.CloneCodeRepo(ctx, project.HttpUrlToRepo)
	if err != nil {
		return nil, err
	}

	defer func(path string) {
		cleanCodeRepo(path)
	}(localPath)

	nodes, err := r.nodestree.Load(localPath)
	if err != nil {
		return nil, err
	}

	options := nodestree.CompareOptions{
		Nodes:            nodes,
		ProductName:      fmt.Sprintf("%s%d", _ProductPrefix, int(product.Id)),
		LocalProjectPath: localPath,
	}

	err = r.nodestree.Compare(options)
	if err != nil {
		return nil, err
	}

	resourceName, err := getResourceName(nodes)
	if err != nil {
		return nil, err
	}

	resourceNode := r.GetNode(&nodes, resourceKind, resourceName)
	if resourceNode == nil {
		return nil, ErrorResourceNoFound
	}

	return resourceNode, nil
}

func (r *ResourcesUsecase) List(ctx context.Context, productName string, operator nodestree.NodesOperator) (*nodestree.Node, error) {
	product, project, err := r.GetProductAndCodeRepo(ctx, productName)
	if err != nil {
		return nil, err
	}

	localPath, err := r.CloneCodeRepo(ctx, project.HttpUrlToRepo)
	if err != nil {
		return nil, err
	}

	defer func(path string) {
		cleanCodeRepo(path)
	}(localPath)

	nodes, err := r.nodestree.Load(localPath)
	if err != nil {
		return nil, err
	}

	options := nodestree.CompareOptions{
		Nodes:            nodes,
		ProductName:      fmt.Sprintf("%s%d", _ProductPrefix, int(product.Id)),
		LocalProjectPath: localPath,
	}

	err = r.nodestree.Compare(options)
	if err != nil {
		return nil, err
	}

	return &nodes, nil
}

type resourceOptions struct {
	resourceKind      string
	resourceName      string
	productName       string
	insecureSkipCheck bool
	operator          nodestree.NodesOperator
}

// Save create or update config to git platform
func (r *ResourcesUsecase) Save(ctx context.Context, resourceOptions *resourceOptions, data interface{}) error {
	var resourceNode *nodestree.Node

	product, project, err := r.GetProductAndCodeRepo(ctx, resourceOptions.productName)
	if err != nil {
		r.log.Log(-1, "msg", "failed to get product and coderepo data", "err", err)
		return err
	}

	localPath, err := r.CloneCodeRepo(ctx, project.HttpUrlToRepo)
	if err != nil {
		r.log.Log(-1, "msg", "failed to clone coderepo", "url", project.HttpUrlToRepo)
		return err
	}

	defer func(path string) {
		cleanCodeRepo(path)
	}(localPath)

	nodes, err := r.nodestree.Load(localPath)
	if err != nil {
		r.log.Log(-1, "msg", "first load n failed", "err", err)
		return err
	}

	options := nodestree.CompareOptions{
		Nodes:            nodes,
		ProductName:      fmt.Sprintf("%s%d", _ProductPrefix, int(product.Id)),
		LocalProjectPath: localPath,
	}

	resourceNode = r.GetNode(&nodes, resourceOptions.resourceKind, resourceOptions.resourceName)
	if resourceNode == nil {
		resourceNode, err = resourceOptions.operator.CreateNode(localPath, data)
		if err != nil {
			r.log.Log(-1, "msg", "failed to create node", "err", err)
			return err
		}
	} else {
		resourceNode, err = resourceOptions.operator.UpdateNode(resourceNode, data)
		if err != nil {
			r.log.Log(-1, "failed to update node", "err", err)
			return err
		}
	}

	newNodes, err := r.InsertNodes(r.nodestree, &nodes, resourceNode)
	if err != nil {
		r.log.Log(-1, "msg", "failed to insert node", "err", err)
		return err
	}

	if !resourceOptions.insecureSkipCheck {
		options.Nodes = *newNodes
		err = r.nodestree.Compare(options)
		if err != nil {
			r.log.Log(-1, "msg", "recheck failed", "err", err)
			return err
		}
	}

	err = r.WriteResource(resourceNode)
	if err != nil {
		r.log.Log(-1, "msg", "failed to write resource", "err", err)
		return err
	}

	err = r.SaveDeployConfig(&nodes, localPath)
	if err != nil {
		r.log.Log(-1, "msg", "failed to saved deploy config", "err", err)
		return err
	}

	err = r.SaveConfig(ctx, localPath)
	if err != nil {
		r.log.Log(-1, "msg", "failed to git submission", "err", err)
		return err
	}

	return nil
}

func (r *ResourcesUsecase) Delete(ctx context.Context, resourceOptions *resourceOptions, getResourceName getResouceName) error {
	product, project, err := r.GetProductAndCodeRepo(ctx, resourceOptions.productName)
	if err != nil {
		return err
	}

	localPath, err := r.CloneCodeRepo(ctx, project.HttpUrlToRepo)
	if err != nil {
		return err
	}

	defer func(path string) {
		cleanCodeRepo(path)
	}(localPath)

	nodes, err := r.nodestree.Load(localPath)
	if err != nil {
		return err
	}

	options := nodestree.CompareOptions{
		Nodes:            nodes,
		ProductName:      fmt.Sprintf("%s%d", _ProductPrefix, int(product.Id)),
		LocalProjectPath: localPath,
	}

	resourceName, err := getResourceName(nodes)
	if err != nil {
		return err
	}

	resourceNode := r.GetNode(&nodes, resourceOptions.resourceKind, resourceName)
	if resourceNode == nil {
		return fmt.Errorf("the resource %s of type %s was not found", resourceName, resourceOptions.resourceKind)
	}

	newNodes, err := r.RemoveNode(&nodes, resourceNode)
	if err != nil {
		return err
	}

	if !resourceOptions.insecureSkipCheck {
		options.Nodes = *newNodes
		err = r.nodestree.Compare(options)
		if err != nil {
			return err
		}
	}

	err = deleteResource(resourceNode)
	if err != nil {
		return err
	}

	err = r.SaveDeployConfig(&nodes, localPath)
	if err != nil {
		return err
	}

	err = r.SaveConfig(ctx, localPath)
	if err != nil {
		return err
	}

	return nil
}

func (r *ResourcesUsecase) InsertNodes(nodestree nodestree.NodesTree, nodes, resource *nodestree.Node) (*nodestree.Node, error) {
	return nodestree.InsertNodes(nodes, resource)
}

// GetNode get specifial node accroding to resource kind and name
func (r *ResourcesUsecase) GetNode(nodes *nodestree.Node, kind, resourceName string) *nodestree.Node {
	return r.nodestree.GetNode(nodes, kind, resourceName)
}

// RemoveNode delete the specified node accroding to the path
func (r *ResourcesUsecase) RemoveNode(nodes, node *nodestree.Node) (*nodestree.Node, error) {
	nodes, err := r.nodestree.RemoveNode(nodes, node)
	if err != nil {
		return nil, err
	}
	return nodes, nil
}

func (r *ResourcesUsecase) GetProductAndCodeRepo(ctx context.Context, ProductName string) (*Group, *Project, error) {
	group, err := r.codeRepo.GetGroup(ctx, ProductName)
	if err != nil {
		return nil, nil, err
	}

	toGetCodeRepo := fmt.Sprintf("%v/%v", group.Path, r.configs.Git.DefaultProductName)
	project, err := r.codeRepo.GetCodeRepo(ctx, toGetCodeRepo)
	if err != nil {
		return nil, nil, err
	}

	return group, project, nil
}

// GetCodeRepoName The name of the codeRepo resource must be prefixed with repo-, eg: repo-1
func (r *ResourcesUsecase) GetCodeRepo(ctx context.Context, ProductName, codeRepoName string) (*Project, error) {
	pid := ""
	group, err := r.codeRepo.GetGroup(ctx, ProductName)
	if err != nil {
		return nil, err
	}

	pid = fmt.Sprintf("%s/%s", group.Path, codeRepoName)
	project, err := r.codeRepo.GetCodeRepo(ctx, pid)
	if err != nil {
		return nil, err
	}

	return project, nil
}

func (r *ResourcesUsecase) CloneCodeRepo(ctx context.Context, url string) (path string, err error) {
	user, email, err := r.codeRepo.GetCurrentUser(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get current user, err: %w", err)
	}

	param := &CloneRepositoryParam{
		URL:   url,
		User:  user,
		Email: email,
	}
	localCodeRepoPath, err := r.gitRepo.Clone(ctx, param)
	if err != nil {
		return "", fmt.Errorf("failed to clone repository, the repository url: %s, err: %w", url, err)
	}

	return localCodeRepoPath, nil
}

// WriteResource Write project resource content to a file
func (r *ResourcesUsecase) WriteResource(node *nodestree.Node) (err error) {
	jsonBytes, err := json.Marshal(node.Content)
	if err != nil {
		return fmt.Errorf("failed to convert resource to json data, err: %w", err)
	}

	jsonString, err := sjson.Delete(string(jsonBytes), "status")
	if err != nil {
		return fmt.Errorf("failed to delete status field of resource, err: %w", err)
	}

	yamlBytes, err := yaml.JSONToYAML([]byte(jsonString))
	if err != nil {
		return fmt.Errorf("failed to convert json to yaml data, err: %w", err)
	}

	subPath := filepath.Dir(node.Path)
	_, err = os.Stat(subPath)
	if !os.IsExist(err) {
		err := os.MkdirAll(subPath, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to write resource directory, err: %w", err)
		}
	}

	err = os.WriteFile(node.Path, yamlBytes, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to write resource file, err: %w", err)
	}

	return
}

func (r *ResourcesUsecase) SaveDeployConfig(nodes *nodestree.Node, path string) error {
	var deployDirectory = fmt.Sprintf("%s/%s", path, r.configs.Deploy.ArgoCD.Kustomize.DefaultPath.DefaultProject)
	var kustomizationFilePath = fmt.Sprintf("%s/%s", deployDirectory, KustomizationFileName)
	var kustomization = &kustomize.Kustomization{
		TypeMeta: kustomize.TypeMeta{
			APIVersion: kustomize.KustomizationVersion,
			Kind:       kustomize.KustomizationKind,
		},
		Resources: []string{},
	}

	addKustomizeResources(nodes, kustomization, path)

	bytes, err := yaml.Marshal(kustomization)
	if err != nil {
		return err
	}

	err = writeKustomize(kustomizationFilePath, bytes)
	if err != nil {
		return err
	}

	return nil
}

func addKustomizeResources(nodes *nodestree.Node, kustomization *kustomize.Kustomization, path string) {
	if nodes != nil {
		for _, v := range nodes.Children {
			if !v.IsDir {
				relativePath := strings.ReplaceAll(v.Path, path, "..")
				kustomization.Resources = append(kustomization.Resources, relativePath)
			} else if v.IsDir && len(v.Children) > 0 {
				addKustomizeResources(v, kustomization, path)
			}
		}
	}
}

func writeKustomize(path string, bytes []byte) error {
	_, err := os.Stat(filepath.Dir(path))
	if err != nil {
		err = os.MkdirAll(filepath.Dir(path), os.ModePerm)
		if err != nil {
			return err
		}
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	if err != nil {
		return err
	}

	defer file.Close()

	_, err = file.Write(bytes)
	if err != nil {
		return err
	}

	return nil
}

// SaveConfig Save project resource config to git platform
// If automatic merge fails will retry three times
func (r *ResourcesUsecase) SaveConfig(ctx context.Context, path string) error {
	count := getCount(ctx)
	if count == nil {
		ctx = withCount(ctx, 1)
	}

	_, err := r.gitRepo.Fetch(ctx, path, "origin")
	if err != nil {
		return err
	}

	data, err := r.gitRepo.Diff(ctx, path, "main", "remotes/origin/main")
	if err != nil {
		return err
	}

	if data == "" {
		err = r.gitRepo.SaveConfig(ctx, path)
		if err != nil {
			return err
		}
	} else {
		err = r.retryAutoMerge(ctx, path)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *ResourcesUsecase) convertCodeRepoToRepoName(ctx context.Context, codeRepoName string) (string, error) {
	id, err := utilstrings.ExtractNumber("repo-", codeRepoName)
	if err != nil {
		return "", err
	}

	project, err := r.codeRepo.GetCodeRepo(ctx, id)
	if err != nil {
		return "", err
	}

	return project.Name, nil
}

func (r *ResourcesUsecase) convertProductToGroupName(ctx context.Context, productName string) (string, error) {
	id, err := utilstrings.ExtractNumber("product-", productName)
	if err != nil {
		return "", err
	}

	group, err := r.codeRepo.GetGroup(ctx, id)
	if err != nil {
		return "", err
	}

	return group.Name, nil
}

func (r *ResourcesUsecase) retryAutoMerge(ctx context.Context, path string) error {
	_, err := r.gitRepo.Fetch(ctx, path)
	if err != nil {
		return fmt.Errorf("when the save configuration cannot be fetch remote branch, err: %v", err)
	}

	err = r.gitRepo.Commit(path, "api: saved configuration")
	if err != nil {
		return err
	}

	_, err = r.gitRepo.Merge(ctx, path)
	if err != nil {
		return fmt.Errorf("when the save configuration cannot be merge automatically, manual approval may be required, err: %v", err)
	}

	err = r.gitRepo.Push(ctx, path)
	if err != nil {
		ok, count, err := isMergeExceededTimes(ctx, 3)
		if err != nil {
			return err
		}

		if !ok {
			count += 1
			ctx = withCount(ctx, count)
			time.Sleep(3 * time.Second)
			return r.SaveConfig(ctx, path)
		}

		err = fmt.Errorf("failed to save config, err: %v", err)
		return err
	}

	return nil
}

func isMergeExceededTimes(ctx context.Context, exceed int) (bool, int, error) {
	count := getCount(ctx)
	val, ok := count.(int)
	if !ok {
		return false, 0, fmt.Errorf("count type is not int")
	}

	if val == exceed {
		return true, val, nil
	}

	return false, val, nil
}

func cleanCodeRepo(filename string) error {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil
	}

	err := os.RemoveAll(filename)
	return err
}

func withCount(ctx context.Context, val interface{}) context.Context {
	return context.WithValue(ctx, RretryCount, val)
}

func getCount(ctx context.Context) interface{} {
	count := ctx.Value(RretryCount)
	return count
}

func deleteResource(node *nodestree.Node) (err error) {
	fileinfos, err := ioutil.ReadDir(filepath.Dir(node.Path))
	if err != nil {
		return err
	}

	if len(fileinfos) == 1 {
		err = os.RemoveAll(filepath.Dir(node.Path))
		if err != nil {
			return err
		}
	} else {
		err = os.Remove(node.Path)
		if err != nil {
			return
		}
	}

	return
}
