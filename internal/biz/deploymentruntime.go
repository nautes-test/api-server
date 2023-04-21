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
	"fmt"
	"reflect"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/nautes-labs/api-server/pkg/nodestree"
	resourcev1alpha1 "github.com/nautes-labs/pkg/api/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	_RuntimesDir = "runtimes"
)

type DeploymentRuntimeUsecase struct {
	log              *log.Helper
	codeRepo         CodeRepo
	nodestree        nodestree.NodesTree
	resourcesUsecase *ResourcesUsecase
}

type DeploymentRuntimeData struct {
	Name string
	Spec resourcev1alpha1.DeploymentRuntimeSpec
}

func NewDeploymentRuntimeUsecase(logger log.Logger, codeRepo CodeRepo, nodestree nodestree.NodesTree, resourcesUsecase *ResourcesUsecase) *DeploymentRuntimeUsecase {
	runtime := &DeploymentRuntimeUsecase{log: log.NewHelper(log.With(logger)), codeRepo: codeRepo, nodestree: nodestree, resourcesUsecase: resourcesUsecase}
	nodestree.AppendOperators(runtime)
	return runtime
}

func (p *DeploymentRuntimeUsecase) convertCodeRepoToRepoName(ctx context.Context, runtime *resourcev1alpha1.DeploymentRuntime) error {
	if runtime.Spec.ManifestSource.CodeRepo == "" {
		return fmt.Errorf("the codeRepo field value of deploymentruntime %s should not be empty", runtime.Name)
	}

	repoName, err := p.resourcesUsecase.convertCodeRepoToRepoName(ctx, runtime.Spec.ManifestSource.CodeRepo)
	if err != nil {
		return err
	}
	runtime.Spec.ManifestSource.CodeRepo = repoName

	return nil
}

func (c *DeploymentRuntimeUsecase) convertProductToGroupName(ctx context.Context, runtime *resourcev1alpha1.DeploymentRuntime) error {
	if runtime.Spec.Product == "" {
		return fmt.Errorf("the product field value of deploymentruntime %s should not be empty", runtime.Name)
	}

	groupName, err := c.resourcesUsecase.convertProductToGroupName(ctx, runtime.Spec.Product)
	if err != nil {
		return err
	}

	runtime.Spec.Product = groupName

	return nil
}

func (d *DeploymentRuntimeUsecase) GetDeploymentRuntime(ctx context.Context, deploymentRuntimeName, productName string) (*resourcev1alpha1.DeploymentRuntime, error) {
	resourceNode, err := d.resourcesUsecase.Get(ctx, nodestree.DeploymentRuntime, productName, d, func(nodes nodestree.Node) (string, error) {
		return deploymentRuntimeName, nil
	})
	if err != nil {
		return nil, err
	}

	runtime, ok := resourceNode.Content.(*resourcev1alpha1.DeploymentRuntime)
	if !ok {
		return nil, fmt.Errorf("the resource type of %s is inconsistent", deploymentRuntimeName)
	}

	err = d.convertCodeRepoToRepoName(ctx, runtime)
	if err != nil {
		return nil, err
	}

	err = d.convertProductToGroupName(ctx, runtime)
	if err != nil {
		return nil, err
	}

	return runtime, nil
}

func (d *DeploymentRuntimeUsecase) ListDeploymentRuntimes(ctx context.Context, productName string) ([]*resourcev1alpha1.DeploymentRuntime, error) {
	var runtimes []*resourcev1alpha1.DeploymentRuntime

	resourceNodes, err := d.resourcesUsecase.List(ctx, productName, d)
	if err != nil {
		return nil, err
	}

	nodes := nodestree.ListsResourceNodes(*resourceNodes, nodestree.DeploymentRuntime)
	for _, node := range nodes {
		if node.Kind == nodestree.DeploymentRuntime && !node.IsDir {
			runtime, ok := node.Content.(*resourcev1alpha1.DeploymentRuntime)
			if ok {

				err = d.convertCodeRepoToRepoName(ctx, runtime)
				if err != nil {
					return nil, err
				}

				err = d.convertProductToGroupName(ctx, runtime)
				if err != nil {
					return nil, err
				}

				runtimes = append(runtimes, runtime)
			}
		}
	}

	return runtimes, nil
}

func (d *DeploymentRuntimeUsecase) SaveDeploymentRuntime(ctx context.Context, options *BizOptions, data *DeploymentRuntimeData) error {
	group, err := d.codeRepo.GetGroup(ctx, options.ProductName)
	if err != nil {
		return err
	}

	pid := fmt.Sprintf("%s/%s", group.Path, data.Spec.ManifestSource.CodeRepo)
	project, err := d.codeRepo.GetCodeRepo(ctx, pid)
	if err != nil {
		return fmt.Errorf("the referenced code repository %s does not exist", data.Spec.ManifestSource.CodeRepo)
	}

	data.Spec.ManifestSource.CodeRepo = fmt.Sprintf("%s%d", _RepoPrefix, int(project.Id))
	data.Spec.Product = fmt.Sprintf("%s%d", _ProductPrefix, int(group.Id))
	resourceOptions := &resourceOptions{
		resourceKind:      nodestree.DeploymentRuntime,
		productName:       options.ProductName,
		insecureSkipCheck: options.InsecureSkipCheck,
		operator:          d,
	}
	err = d.resourcesUsecase.Save(ctx, resourceOptions, data)
	if err != nil {
		return err
	}

	return nil
}

func (e *DeploymentRuntimeUsecase) CreateNode(path string, data interface{}) (*nodestree.Node, error) {
	val, ok := data.(*DeploymentRuntimeData)
	if !ok {
		return nil, fmt.Errorf("failed to create node, the path is %s", path)
	}

	resource := &resourcev1alpha1.DeploymentRuntime{
		TypeMeta: v1.TypeMeta{
			APIVersion: resourcev1alpha1.GroupVersion.String(),
			Kind:       nodestree.DeploymentRuntime,
		},
		ObjectMeta: v1.ObjectMeta{
			Name: val.Name,
		},
		Spec: val.Spec,
	}

	resourceDirectory := fmt.Sprintf("%s/%s", path, _RuntimesDir)
	resourceFile := fmt.Sprintf("%s/%s.yaml", resourceDirectory, val.Name)

	return &nodestree.Node{
		Name:    val.Name,
		Path:    resourceFile,
		Kind:    nodestree.DeploymentRuntime,
		Content: resource,
		Level:   3,
	}, nil
}

func (e *DeploymentRuntimeUsecase) UpdateNode(resourceNode *nodestree.Node, data interface{}) (*nodestree.Node, error) {
	val, ok := data.(*DeploymentRuntimeData)
	if !ok {
		return nil, fmt.Errorf("failed to update node %s", resourceNode.Name)
	}

	deployRuntime, ok := resourceNode.Content.(*resourcev1alpha1.DeploymentRuntime)
	if !ok {
		return nil, fmt.Errorf("failed to get resource instance when update node %s", val.Name)
	}

	ok = reflect.DeepEqual(deployRuntime.Spec, val.Spec)
	if ok {
		return resourceNode, nil
	}

	deployRuntime.Spec = val.Spec
	resourceNode.Content = deployRuntime

	return resourceNode, nil
}

func (d *DeploymentRuntimeUsecase) CheckReference(options nodestree.CompareOptions, node *nodestree.Node, k8sClient client.Client) (bool, error) {
	if node.Kind != nodestree.DeploymentRuntime {
		return false, nil
	}

	deploymentRuntime, ok := node.Content.(*resourcev1alpha1.DeploymentRuntime)
	if !ok {
		return true, fmt.Errorf("node %s resource type error", node.Name)
	}

	productName := deploymentRuntime.Spec.Product
	if productName != options.ProductName {
		return true, fmt.Errorf("the product name of resource %s does not match the current product name, expected %s, but now is %s ", deploymentRuntime.Name, options.ProductName, productName)
	}

	projectsRef := deploymentRuntime.Spec.ProjectsRef
	for _, projectName := range projectsRef {
		ok := nodestree.IsResourceExist(options, projectName, nodestree.Project)
		if !ok {
			return true, fmt.Errorf("the referenced project %s by the deployment runtime %s does not exist while verifying the validity of the global template", projectName, deploymentRuntime.Name)
		}
	}

	ok = nodestree.IsResourceExist(options, deploymentRuntime.Spec.Destination, nodestree.Enviroment)
	if !ok {
		return true, fmt.Errorf("the referenced environment %s by the deployment runtime %s does not exist while verifying the validity of the global template", deploymentRuntime.Spec.Destination, deploymentRuntime.Name)
	}

	codeRepoName := deploymentRuntime.Spec.ManifestSource.CodeRepo
	ok = nodestree.IsResourceExist(options, codeRepoName, nodestree.CodeRepo)
	if !ok {
		return true, fmt.Errorf("the referenced repository %s by the deployment runtime %s does not exist while verifying the validity of the global template", codeRepoName, deploymentRuntime.Name)
	}

	ok, err := d.compare(options.Nodes)
	if ok {
		return true, err
	}

	return true, nil
}

func (d *DeploymentRuntimeUsecase) compare(nodes nodestree.Node) (bool, error) {
	resourceNodes := nodestree.ListsResourceNodes(nodes, nodestree.DeploymentRuntime)
	for i := 0; i < len(resourceNodes); i++ {
		for j := i + 1; j < len(resourceNodes); j++ {
			if v1, ok := resourceNodes[i].Content.(*resourcev1alpha1.DeploymentRuntime); ok {
				if v2, ok := resourceNodes[j].Content.(*resourcev1alpha1.DeploymentRuntime); ok {
					ok, err := v1.Compare(v2)
					if err != nil {
						return false, err
					}
					if ok {
						return ok, fmt.Errorf("duplicate reference found between resource %s and resource %s", resourceNodes[i].Name, resourceNodes[j].Name)
					}
				}
			}
		}
	}

	return false, nil
}

func (d *DeploymentRuntimeUsecase) CreateResource(kind string) interface{} {
	if kind != nodestree.DeploymentRuntime {
		return nil
	}

	return &resourcev1alpha1.DeploymentRuntime{}
}

func (d *DeploymentRuntimeUsecase) DeleteDeploymentRuntime(ctx context.Context, options *BizOptions) error {
	resourceOptions := &resourceOptions{
		resourceKind:      nodestree.DeploymentRuntime,
		productName:       options.ProductName,
		insecureSkipCheck: options.InsecureSkipCheck,
		operator:          d,
	}
	err := d.resourcesUsecase.Delete(ctx, resourceOptions, func(nodes nodestree.Node) (string, error) {
		return options.ResouceName, nil
	})

	if err != nil {
		return err
	}

	return nil
}
