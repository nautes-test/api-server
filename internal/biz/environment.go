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

	errors "github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	enviromentv1 "github.com/nautes-labs/api-server/api/environment/v1"
	"github.com/nautes-labs/api-server/pkg/nodestree"
	resourcev1alpha1 "github.com/nautes-labs/pkg/api/v1alpha1"
	nautesconfigs "github.com/nautes-labs/pkg/pkg/nautesconfigs"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	_EnvironmentKind = "Environment"
	_EnvSubDir       = "envs"
)

type EnvironmentUsecase struct {
	log              *log.Helper
	codeRepo         CodeRepo
	nodestree        nodestree.NodesTree
	config           *nautesconfigs.Config
	resourcesUsecase *ResourcesUsecase
}

type EnviromentData struct {
	Name string
	Spec resourcev1alpha1.EnvironmentSpec
}

func NewEnviromentUsecase(logger log.Logger, config *nautesconfigs.Config, codeRepo CodeRepo, nodestree nodestree.NodesTree, resourcesUsecase *ResourcesUsecase) *EnvironmentUsecase {
	env := &EnvironmentUsecase{log: log.NewHelper(log.With(logger)), config: config, codeRepo: codeRepo, nodestree: nodestree, resourcesUsecase: resourcesUsecase}
	nodestree.AppendOperators(env)
	return env
}

func (c *EnvironmentUsecase) convertProductToGroupName(ctx context.Context, env *resourcev1alpha1.Environment) error {
	if env.Spec.Product == "" {
		return fmt.Errorf("the product field value of enviroment %s should not be empty", env.Spec.Product)
	}

	groupName, err := c.resourcesUsecase.convertProductToGroupName(ctx, env.Spec.Product)
	if err != nil {
		return err
	}

	env.Spec.Product = groupName

	return nil
}

func (e *EnvironmentUsecase) GetEnvironment(ctx context.Context, enviromentName, productName string) (*resourcev1alpha1.Environment, error) {
	node, err := e.resourcesUsecase.Get(ctx, nodestree.Enviroment, productName, e, func(nodes nodestree.Node) (string, error) {
		return enviromentName, nil
	})
	if err != nil {
		return nil, err
	}

	env, err := e.nodeToResource(node)
	if err != nil {
		return nil, err
	}

	err = e.convertProductToGroupName(ctx, env)
	if err != nil {
		return nil, err
	}

	return env, nil
}

func (e *EnvironmentUsecase) ListEnvironments(ctx context.Context, productName string) ([]*resourcev1alpha1.Environment, error) {
	nodes, err := e.resourcesUsecase.List(ctx, productName, e)
	if err != nil {
		return nil, err
	}

	envs, err := e.nodesToLists(*nodes)
	if err != nil {
		return nil, err
	}

	for _, env := range envs {
		err = e.convertProductToGroupName(ctx, env)
		if err != nil {
			return nil, err
		}
	}

	return envs, nil
}

func (e *EnvironmentUsecase) nodesToLists(nodes nodestree.Node) ([]*resourcev1alpha1.Environment, error) {
	var resourcesSubDir *nodestree.Node
	var resources []*resourcev1alpha1.Environment

	for _, child := range nodes.Children {
		if child.Name == _EnvSubDir {
			resourcesSubDir = child
			break
		}
	}

	if resourcesSubDir == nil {
		return nil, ErrorNodetNotFound
	}

	for _, node := range resourcesSubDir.Children {
		env, err := e.nodeToResource(node)
		if err != nil {
			return nil, err
		}

		resources = append(resources, env)
	}

	return resources, nil
}

func (e *EnvironmentUsecase) nodeToResource(node *nodestree.Node) (*resourcev1alpha1.Environment, error) {
	env, ok := node.Content.(*resourcev1alpha1.Environment)
	if !ok {
		return nil, errors.New(503, enviromentv1.ErrorReason_ASSERT_ERROR.String(), fmt.Sprintf("%s assert error", node.Name))
	}

	return env, nil
}

func (e *EnvironmentUsecase) SaveEnvironment(ctx context.Context, options *BizOptions, data *EnviromentData) error {
	group, err := e.codeRepo.GetGroup(ctx, options.ProductName)
	if err != nil {
		return err
	}

	data.Spec.Product = fmt.Sprintf("%s%d", _ProductPrefix, int(group.Id))
	resourceOptions := &resourceOptions{
		resourceKind:      nodestree.Enviroment,
		productName:       options.ProductName,
		insecureSkipCheck: options.InsecureSkipCheck,
		operator:          e,
	}
	err = e.resourcesUsecase.Save(ctx, resourceOptions, data)
	if err != nil {
		return err
	}

	return nil
}

func (e *EnvironmentUsecase) CreateNode(path string, data interface{}) (*nodestree.Node, error) {
	val, ok := data.(*EnviromentData)
	if !ok {
		return nil, errors.New(503, enviromentv1.ErrorReason_ASSERT_ERROR.String(), fmt.Sprintf("failed to assert EnviromentData when create node, data: %v", val))
	}

	env := &resourcev1alpha1.Environment{
		TypeMeta: v1.TypeMeta{
			APIVersion: resourcev1alpha1.GroupVersion.String(),
			Kind:       nodestree.Enviroment,
		},
		ObjectMeta: v1.ObjectMeta{
			Name: val.Name,
		},
		Spec: val.Spec,
	}

	resourceDirectory := fmt.Sprintf("%s/%s", path, _EnvSubDir)
	resourceFile := fmt.Sprintf("%s/%s.yaml", resourceDirectory, val.Name)

	return &nodestree.Node{
		Name:    val.Name,
		Path:    resourceFile,
		Content: env,
		Kind:    nodestree.Enviroment,
		Level:   3,
	}, nil
}

func (e *EnvironmentUsecase) UpdateNode(resourceNode *nodestree.Node, data interface{}) (*nodestree.Node, error) {
	val, ok := data.(*EnviromentData)
	if !ok {
		return nil, errors.New(503, enviromentv1.ErrorReason_ASSERT_ERROR.String(), fmt.Sprintf("failed to assert EnviromentData when update node, data: %v", val))
	}

	env, ok := resourceNode.Content.(*resourcev1alpha1.Environment)
	if !ok {
		return nil, errors.New(503, enviromentv1.ErrorReason_ASSERT_ERROR.String(), fmt.Sprintf("failed to assert resourcev1alpha1.Environment when update node, data: %v", val))
	}

	env.Name = val.Name
	if env.Spec == val.Spec {
		return resourceNode, nil
	}

	env.Spec = val.Spec
	resourceNode.Content = env

	return resourceNode, nil
}

func (e *EnvironmentUsecase) CheckReference(options nodestree.CompareOptions, node *nodestree.Node, k8sClient client.Client) (bool, error) {
	if node.Kind != nodestree.Enviroment {
		return false, nil
	}

	env, ok := node.Content.(*resourcev1alpha1.Environment)
	if !ok {
		return true, fmt.Errorf("node %s resource type error", node.Name)
	}

	productName := env.Spec.Product
	if productName != options.ProductName {
		return true, fmt.Errorf("the product name of resource %s does not match the current product name, expected is %s, but now is %s", env.Name, options.ProductName, productName)
	}

	tenantAdminNamespace := e.config.Nautes.Namespace
	if tenantAdminNamespace == "" {
		return true, fmt.Errorf("tenant admin namspace cannot be empty")
	}

	objKey := client.ObjectKey{
		Namespace: tenantAdminNamespace,
		Name:      env.Spec.Cluster,
	}

	err := k8sClient.Get(context.TODO(), objKey, &resourcev1alpha1.Cluster{})
	if err != nil {
		return true, fmt.Errorf(_ResourceDoesNotExistOrUnavailable+"err: "+err.Error(), nodestree.Cluster, env.Spec.Cluster,
			_EnvironmentKind, env.Name, _EnvSubDir)
	}

	ok, err = e.compare(options.Nodes)
	if ok {
		return ok, err
	}

	return true, nil
}

func (e *EnvironmentUsecase) CreateResource(kind string) interface{} {
	if kind != nodestree.Enviroment {
		return nil
	}

	return &resourcev1alpha1.Environment{}
}

func (e *EnvironmentUsecase) DeleteEnvironment(ctx context.Context, options *BizOptions) error {
	resourceOptions := &resourceOptions{
		resourceKind:      nodestree.Enviroment,
		productName:       options.ProductName,
		insecureSkipCheck: options.InsecureSkipCheck,
		operator:          e,
	}
	err := e.resourcesUsecase.Delete(ctx, resourceOptions, func(nodes nodestree.Node) (string, error) {
		return options.ResouceName, nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (e *EnvironmentUsecase) compare(nodes nodestree.Node) (bool, error) {
	resourceNodes := nodestree.ListsResourceNodes(nodes, nodestree.Enviroment)
	for i := 0; i < len(resourceNodes); i++ {
		for j := i + 1; j < len(resourceNodes); j++ {
			if v1, ok := resourceNodes[i].Content.(*resourcev1alpha1.Environment); ok {
				if v2, ok := resourceNodes[j].Content.(*resourcev1alpha1.Environment); ok {
					ok, err := v1.Compare(v2)
					if err != nil {
						return false, err
					}
					if ok {
						return ok, fmt.Errorf("duplicate reference cluster resources sbetween resource %s and resource %s", resourceNodes[i].Name, resourceNodes[j].Name)
					}
				}
			}
		}
	}

	return false, nil
}
