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

	"github.com/go-kratos/kratos/v2/log"
	commonv1 "github.com/nautes-labs/api-server/api/common/v1"
	projectpipelineruntimev1 "github.com/nautes-labs/api-server/api/projectpipelineruntime/v1"
	"github.com/nautes-labs/api-server/pkg/nodestree"
	resourcev1alpha1 "github.com/nautes-labs/pkg/api/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	_PipelineRuntimeKind = "PipelineRuntime"
)

type ProjectPipelineRuntimeUsecase struct {
	log              *log.Helper
	codeRepo         CodeRepo
	nodestree        nodestree.NodesTree
	resourcesUsecase *ResourcesUsecase
}

type ProjectPipelineRuntimeData struct {
	Name string
	Spec resourcev1alpha1.ProjectPipelineRuntimeSpec
}

func NewProjectPipelineRuntimeUsecase(logger log.Logger, codeRepo CodeRepo, nodestree nodestree.NodesTree, resourcesUsecase *ResourcesUsecase) *ProjectPipelineRuntimeUsecase {
	runtime := &ProjectPipelineRuntimeUsecase{log: log.NewHelper(log.With(logger)), codeRepo: codeRepo, nodestree: nodestree, resourcesUsecase: resourcesUsecase}
	nodestree.AppendOperators(runtime)
	return runtime
}

func (p *ProjectPipelineRuntimeUsecase) convertCodeRepoToRepoName(ctx context.Context, runtime *resourcev1alpha1.ProjectPipelineRuntime) error {
	if runtime.Spec.PipelineSource == "" {
		return fmt.Errorf("the pipelineSource field value of projectPipelineRuntime %s should not be empty", runtime.Name)
	}
	if runtime.Spec.PipelineSource != "" {
		repoName, err := p.resourcesUsecase.convertCodeRepoToRepoName(ctx, runtime.Spec.PipelineSource)
		if err != nil {
			return err
		}
		runtime.Spec.PipelineSource = repoName
	}

	for idx, codeRepo := range runtime.Spec.CodeSources {
		repoName, err := p.resourcesUsecase.convertCodeRepoToRepoName(ctx, codeRepo)
		if err != nil {
			return err
		}
		runtime.Spec.CodeSources[idx] = repoName
	}

	return nil
}

func (p *ProjectPipelineRuntimeUsecase) GetProjectPipelineRuntime(ctx context.Context, projectPipelineName, productName string) (*resourcev1alpha1.ProjectPipelineRuntime, error) {
	resourceNode, err := p.resourcesUsecase.Get(ctx, nodestree.ProjectPipelineRuntime, productName, p, func(nodes nodestree.Node) (string, error) {
		return projectPipelineName, nil
	})
	if err != nil {
		return nil, err
	}

	runtime, ok := resourceNode.Content.(*resourcev1alpha1.ProjectPipelineRuntime)
	if !ok {
		return nil, fmt.Errorf("the resource %s type is inconsistent", projectPipelineName)
	}

	err = p.convertCodeRepoToRepoName(ctx, runtime)
	if err != nil {
		return nil, err
	}

	return runtime, nil
}

func (p *ProjectPipelineRuntimeUsecase) ListProjectPipelineRuntimes(ctx context.Context, productName string) ([]*resourcev1alpha1.ProjectPipelineRuntime, error) {
	var runtimes []*resourcev1alpha1.ProjectPipelineRuntime

	resourceNodes, err := p.resourcesUsecase.List(ctx, productName, p)
	if err != nil {
		return nil, err
	}

	nodes := nodestree.ListsResourceNodes(*resourceNodes, nodestree.ProjectPipelineRuntime)
	for _, node := range nodes {
		if node.Kind == nodestree.ProjectPipelineRuntime && !node.IsDir {
			runtime, ok := node.Content.(*resourcev1alpha1.ProjectPipelineRuntime)
			if ok {
				err = p.convertCodeRepoToRepoName(ctx, runtime)
				if err != nil {
					return nil, err
				}
				runtimes = append(runtimes, runtime)
			}
		}
	}

	return runtimes, nil
}

func (p *ProjectPipelineRuntimeUsecase) SaveProjectPipelineRuntime(ctx context.Context, options *BizOptions, data *ProjectPipelineRuntimeData) error {
	project, err := p.resourcesUsecase.GetCodeRepo(ctx, options.ProductName, data.Spec.PipelineSource)
	if err != nil {
		if ok := commonv1.IsProjectNotFound(err); ok {
			return projectpipelineruntimev1.ErrorPipelineResourceNotFound("failed to get repository please check pipeline source %s or product %s valid", data.Spec.PipelineSource, options.ProductName)
		} else {
			return err
		}
	}

	data.Spec.PipelineSource = SpliceCodeRepoResourceName(int(project.Id))

	if len(data.Spec.CodeSources) > 0 {
		for i, source := range data.Spec.CodeSources {
			project, err := p.resourcesUsecase.GetCodeRepo(ctx, options.ProductName, source)
			if err != nil {
				return fmt.Errorf("failed to get repository please check codeRepo source or product name, err: %w", err)
			}

			data.Spec.CodeSources[i] = SpliceCodeRepoResourceName(int(project.Id))
		}
	}

	resourceOptions := &resourceOptions{
		resourceKind:      nodestree.ProjectPipelineRuntime,
		productName:       options.ProductName,
		insecureSkipCheck: options.InsecureSkipCheck,
		operator:          p,
	}
	err = p.resourcesUsecase.Save(ctx, resourceOptions, data)
	if err != nil {
		return err
	}

	return nil
}

func (p *ProjectPipelineRuntimeUsecase) DeleteProjectPipelineRuntime(ctx context.Context, options *BizOptions) error {
	resourceOptions := &resourceOptions{
		resourceKind:      nodestree.ProjectPipelineRuntime,
		productName:       options.ProductName,
		insecureSkipCheck: options.InsecureSkipCheck,
		operator:          p,
	}
	err := p.resourcesUsecase.Delete(ctx, resourceOptions, func(nodes nodestree.Node) (string, error) {
		return options.ResouceName, nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (p *ProjectPipelineRuntimeUsecase) CreateNode(path string, data interface{}) (*nodestree.Node, error) {
	var resourceNode *nodestree.Node

	val, ok := data.(*ProjectPipelineRuntimeData)
	if !ok {
		return nil, fmt.Errorf("failed to save project when create specify node path: %s", path)
	}

	runtime := &resourcev1alpha1.ProjectPipelineRuntime{
		TypeMeta: v1.TypeMeta{
			APIVersion: resourcev1alpha1.GroupVersion.String(),
			Kind:       nodestree.ProjectPipelineRuntime,
		},
		ObjectMeta: v1.ObjectMeta{
			Name: val.Name,
		},
		Spec: val.Spec,
	}

	storageResourceDirectory := fmt.Sprintf("%s/%s", path, _ProjectsDir)
	resourceParentDir := fmt.Sprintf("%s/%s", storageResourceDirectory, val.Spec.Project)
	resourceFile := fmt.Sprintf("%s/%s.yaml", resourceParentDir, val.Name)
	resourceNode = &nodestree.Node{
		Name:    val.Name,
		Path:    resourceFile,
		Content: runtime,
		Kind:    nodestree.ProjectPipelineRuntime,
		Level:   4,
	}

	return resourceNode, nil
}
func (p *ProjectPipelineRuntimeUsecase) UpdateNode(node *nodestree.Node, data interface{}) (*nodestree.Node, error) {
	val, ok := data.(*ProjectPipelineRuntimeData)
	if !ok {
		return nil, fmt.Errorf("failed to get project data when update %s node", node.Name)
	}

	runtime, ok := node.Content.(*resourcev1alpha1.ProjectPipelineRuntime)
	if !ok {
		return nil, fmt.Errorf("failed to get project insatnce when update %s node", node.Name)
	}

	if val.Spec.Project != runtime.Spec.Project {
		return nil, fmt.Errorf("existing pipeline runtime is not allow modifying the project field")
	}

	runtime.Spec = val.Spec
	node.Content = runtime

	return node, nil
}

func (p *ProjectPipelineRuntimeUsecase) CheckReference(options nodestree.CompareOptions, node *nodestree.Node, k8sClient client.Client) (bool, error) {
	if node.Kind != nodestree.ProjectPipelineRuntime {
		return false, nil
	}

	projectPipelineRuntime, ok := node.Content.(*resourcev1alpha1.ProjectPipelineRuntime)
	if !ok {
		return true, fmt.Errorf("node %s resource type error", node.Name)
	}

	ok, err := p.isRepeatRepositories(projectPipelineRuntime)
	if ok {
		return true, err
	}

	ok, err = p.isRepeatPipelinePath(projectPipelineRuntime)
	if ok {
		return true, err
	}

	projectName := projectPipelineRuntime.Spec.Project
	ok = nodestree.IsResourceExist(options, projectName, nodestree.Project)
	if !ok {
		return true, fmt.Errorf(_ResourceDoesNotExistOrUnavailable, _ProjectKind, projectName, _PipelineRuntimeKind,
			projectPipelineRuntime.Name, _ProjectsDir+projectPipelineRuntime.Spec.Project)
	}

	targetEnvironment := projectPipelineRuntime.Spec.Destination
	ok = nodestree.IsResourceExist(options, targetEnvironment, nodestree.Enviroment)
	if !ok {
		return true, fmt.Errorf(_ResourceDoesNotExistOrUnavailable, _EnvironmentKind, targetEnvironment, _PipelineRuntimeKind,
			projectPipelineRuntime.Name, projectPipelineRuntime.Spec.Project)
	}

	codeRepoName := projectPipelineRuntime.Spec.PipelineSource
	ok = nodestree.IsResourceExist(options, codeRepoName, nodestree.CodeRepo)
	if !ok {
		return true, fmt.Errorf(_ResourceDoesNotExistOrUnavailable, _CodeRepoKind, codeRepoName, _PipelineRuntimeKind,
			projectPipelineRuntime.Name, projectPipelineRuntime.Spec.Project)
	}

	if len(projectPipelineRuntime.Spec.CodeSources) > 0 {
		codeSources := projectPipelineRuntime.Spec.CodeSources
		for _, source := range codeSources {
			ok = nodestree.IsResourceExist(options, source, nodestree.CodeRepo)
			if !ok {
				return true, fmt.Errorf(_ResourceDoesNotExistOrUnavailable, _CodeRepoKind, codeRepoName, _PipelineRuntimeKind,
					projectPipelineRuntime.Name, projectPipelineRuntime.Spec.Project)
			}
		}
	}

	ok, err = p.compare(options.Nodes)
	if ok && err != nil {
		return true, err
	}

	return true, nil
}

func (p *ProjectPipelineRuntimeUsecase) isRepeatRepositories(runtime *resourcev1alpha1.ProjectPipelineRuntime) (bool, error) {
	codeRepos := runtime.Spec.CodeSources
	if len(codeRepos) > 0 {
		for _, source := range codeRepos {
			if source == runtime.Spec.PipelineSource {
				return true, fmt.Errorf("CodeSource for ProjectPipelineRuntime %s has duplicate item, as found in the global validation", runtime.Name)
			}
		}
	}

	return false, nil
}

func (p *ProjectPipelineRuntimeUsecase) isRepeatPipelinePath(runtime *resourcev1alpha1.ProjectPipelineRuntime) (bool, error) {
	len := len(runtime.Spec.Pipelines)

	for i := 0; i < len-1; i++ {
		for j := i + 1; j < len; j++ {
			if runtime.Spec.Pipelines[i].Path == runtime.Spec.Pipelines[j].Path {
				return true, fmt.Errorf("ProjectPipelineRuntime %s uses the same code repository for both codeSource and pipelineSource under %s directory, as found in the global validation", runtime.Name, runtime.Spec.Project)
			}
		}
	}

	return false, nil
}

func (e *ProjectPipelineRuntimeUsecase) compare(nodes nodestree.Node) (bool, error) {
	resourceNodes := nodestree.ListsResourceNodes(nodes, nodestree.ProjectPipelineRuntime)
	for i := 0; i < len(resourceNodes); i++ {
		for j := i + 1; j < len(resourceNodes); j++ {
			if v1, ok := resourceNodes[i].Content.(*resourcev1alpha1.ProjectPipelineRuntime); ok {
				if v2, ok := resourceNodes[j].Content.(*resourcev1alpha1.ProjectPipelineRuntime); ok {
					ok, err := v1.Compare(v2)
					if err != nil {
						return true, err
					}
					if ok {
						n1 := resourceNodes[i].Name
						n2 := resourceNodes[j].Name
						p1 := nodestree.GetResourceValue(resourceNodes[i].Content, "Spec", "Project")
						p2 := nodestree.GetResourceValue(resourceNodes[j].Content, "Spec", "Project")
						d1 := fmt.Sprintf("%s/%s", p1, n1)
						d2 := fmt.Sprintf("%s/%s", p2, n2)
						return true, fmt.Errorf("duplicate pipeline found in verify the validity of the global template, respectively %s and %s", d1, d2)
					}
				}
			}
		}
	}

	return false, nil
}

func (p *ProjectPipelineRuntimeUsecase) CreateResource(kind string) interface{} {
	if kind != nodestree.ProjectPipelineRuntime {
		return nil
	}

	return &resourcev1alpha1.ProjectPipelineRuntime{}
}

func SpliceCodeRepoResourceName(id int) string {
	return fmt.Sprintf("%s%d", _RepoPrefix, int(id))
}
