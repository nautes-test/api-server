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
	"os"
	"path/filepath"

	"github.com/golang/mock/gomock"
	"github.com/nautes-labs/api-server/pkg/kubernetes"
	"github.com/nautes-labs/api-server/pkg/nodestree"
	resourcev1alpha1 "github.com/nautes-labs/pkg/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	_DefaultProjectName = "project1"
)

func createProjectPiepeLineResource(name string) *resourcev1alpha1.ProjectPipelineRuntime {
	return &resourcev1alpha1.ProjectPipelineRuntime{
		TypeMeta: v1.TypeMeta{
			Kind: nodestree.Project,
		},
		ObjectMeta: v1.ObjectMeta{
			Name: name,
		},
		Spec: resourcev1alpha1.ProjectPipelineRuntimeSpec{
			Project:        _DefaultProjectName,
			PipelineSource: "pipelineCodeRepo",
			Destination:    "env1",
			Pipelines: []resourcev1alpha1.Pipeline{
				{
					Name:   "dev",
					Label:  "test",
					Branch: "feature-xx",
					Path:   "dev",
					EventSources: []resourcev1alpha1.EventSource{
						{
							Webhook: "enabled",
						},
					},
				},
			},
			CodeSources: []string{"CodeRepo1"},
		},
	}
}

func createFakeProjectPipelineRuntimeNode(resource *resourcev1alpha1.ProjectPipelineRuntime) *nodestree.Node {
	return &nodestree.Node{
		Name:    resource.Name,
		Kind:    nodestree.ProjectPipelineRuntime,
		Path:    fmt.Sprintf("%s/%s/%s/%s.yaml", localRepositaryPath, _ProjectsDir, _DefaultProjectName, resource.Name),
		Level:   4,
		Content: resource,
	}
}

func createFakeProjectPipelineRuntimeNodes(node *nodestree.Node) nodestree.Node {
	return nodestree.Node{
		Name:  defaultProjectName,
		Path:  defaultProjectName,
		IsDir: true,
		Level: 1,
		Children: []*nodestree.Node{
			{
				Name:  _ProjectsDir,
				Path:  fmt.Sprintf("%s/%s", defaultProjectName, _ProjectsDir),
				IsDir: true,
				Level: 2,
				Children: []*nodestree.Node{
					{
						Name:  _DefaultProjectName,
						Path:  fmt.Sprintf("%s/%s/%s", defaultProjectName, _ProjectsDir, _DefaultProjectName),
						IsDir: true,
						Level: 3,
						Children: []*nodestree.Node{
							node,
						},
					},
				},
			},
		},
	}
}

var _ = Describe("Get project pipeline runtime", func() {
	var (
		resourceName = "runtime1"
		fakeResource = createProjectPiepeLineResource(resourceName)
		fakeNode     = createFakeProjectPipelineRuntimeNode(fakeResource)
		fakeNodes    = createFakeProjectPipelineRuntimeNodes(fakeNode)
	)
	It("will get success", testUseCase.GetResourceSuccess(fakeNodes, fakeNode, func(codeRepo *MockCodeRepo, secretRepo *MockSecretrepo, resourceUseCase *ResourcesUsecase, nodestree *nodestree.MockNodesTree, gitRepo *MockGitRepo, client *kubernetes.MockClient) {
		tmp, _ := fakeNode.Content.(*resourcev1alpha1.ProjectPipelineRuntime)
		tmp.Spec.PipelineSource = "repo-1"
		tmp.Spec.CodeSources[0] = "repo-2"
		codeRepo.EXPECT().GetCodeRepo(gomock.Any(), 1).Return(defautlProject, nil)
		codeRepo.EXPECT().GetCodeRepo(gomock.Any(), 2).Return(defautlProject, nil)

		biz := NewProjectPipelineRuntimeUsecase(logger, codeRepo, nodestree, resourceUseCase)
		result, err := biz.GetProjectPipelineRuntime(context.Background(), resourceName, defaultGroupName)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(result).Should(Equal(fakeResource))
	}))

	It("will fail when resource is not found", testUseCase.GetResourceFail(func(codeRepo *MockCodeRepo, secretRepo *MockSecretrepo, resourceUseCase *ResourcesUsecase, nodestree *nodestree.MockNodesTree, gitRepo *MockGitRepo, client *kubernetes.MockClient) {
		biz := NewProjectPipelineRuntimeUsecase(logger, codeRepo, nodestree, resourceUseCase)
		_, err := biz.GetProjectPipelineRuntime(context.Background(), resourceName, defaultGroupName)
		Expect(err).Should(HaveOccurred())
	}))
})

var _ = Describe("List project pipeline runtimes", func() {
	var (
		resourceName = "projectpipelineruntime1"
		fakeResource = createProjectPiepeLineResource(resourceName)
		fakeNode     = createFakeProjectPipelineRuntimeNode(fakeResource)
		fakeNodes    = createFakeProjectPipelineRuntimeNodes(fakeNode)
	)
	It("will list successfully", testUseCase.ListResourceSuccess(fakeNodes, func(codeRepo *MockCodeRepo, secretRepo *MockSecretrepo, resourceUseCase *ResourcesUsecase, nodestree *nodestree.MockNodesTree, gitRepo *MockGitRepo, client *kubernetes.MockClient) {
		tmp, _ := fakeNode.Content.(*resourcev1alpha1.ProjectPipelineRuntime)
		tmp.Spec.PipelineSource = "repo-1"
		tmp.Spec.CodeSources[0] = "repo-2"
		codeRepo.EXPECT().GetCodeRepo(gomock.Any(), 1).Return(defautlProject, nil)
		codeRepo.EXPECT().GetCodeRepo(gomock.Any(), 2).Return(defautlProject, nil)

		biz := NewProjectPipelineRuntimeUsecase(logger, codeRepo, nodestree, resourceUseCase)
		results, err := biz.ListProjectPipelineRuntimes(ctx, defaultGroupName)
		Expect(err).ShouldNot(HaveOccurred())
		for _, result := range results {
			Expect(result).Should(Equal(fakeResource))
		}
	}))

	It("does not conform to the template layout", testUseCase.ListResourceNotMatch(fakeNodes, func(codeRepo *MockCodeRepo, secretRepo *MockSecretrepo, resourceUseCase *ResourcesUsecase, nodestree *nodestree.MockNodesTree, gitRepo *MockGitRepo, client *kubernetes.MockClient) {
		biz := NewProjectPipelineRuntimeUsecase(logger, codeRepo, nodestree, resourceUseCase)
		_, err := biz.ListProjectPipelineRuntimes(ctx, defaultGroupName)
		Expect(err).Should(HaveOccurred())
	}))
})

var _ = Describe("Save project pipeline runtime", func() {
	var (
		resourceName             = "projectpipelineruntime1"
		projectForPipeline       = &Project{Name: "pipeline", Id: 1222, HttpUrlToRepo: "ssh://git@gitlab.io/nautes-labs/pipeline.git"}
		projectForPipelineRepoID = fmt.Sprintf("%s%d", _RepoPrefix, int(projectForPipeline.Id))
		projectForBase           = &Project{Name: "base", Id: 1223, HttpUrlToRepo: fmt.Sprintf("ssh://git@gitlab.io/nautes-labs/%s.git", resourceName)}
		projectForBaseRepoID     = fmt.Sprintf("%s%d", _RepoPrefix, int(projectForBase.Id))
		fakeResource             = createProjectPiepeLineResource(resourceName)
		fakeNode                 = createFakeProjectPipelineRuntimeNode(fakeResource)
		fakeNodes                = createFakeProjectPipelineRuntimeNodes(fakeNode)
		data                     = &ProjectPipelineRuntimeData{
			Name: fakeResource.Name,
			Spec: resourcev1alpha1.ProjectPipelineRuntimeSpec{
				Project:        _DefaultProjectName,
				PipelineSource: projectForPipeline.Name,
				Destination:    "env1",
				Pipelines: []resourcev1alpha1.Pipeline{
					{
						Name:   "dev",
						Label:  "test",
						Branch: "feature-xx",
						Path:   "dev",
						EventSources: []resourcev1alpha1.EventSource{
							{
								Webhook: "enabled",
							},
						},
					},
				},
				CodeSources: []string{projectForBase.Name},
			},
		}
		pipelineSourceCodeRepoPath = fmt.Sprintf("%s/%s", defaultProductGroup.Path, projectForPipeline.Name)
		codeRepoSourcePath         = fmt.Sprintf("%s/%s", defaultProductGroup.Path, projectForBase.Name)
		pipelineSouceProject       = &Project{Id: 12}
		codeRepoSouceProject       = &Project{Id: 13}
		bizOptions                 = &BizOptions{
			ResouceName: resourceName,
			ProductName: defaultGroupName,
		}
	)

	AfterEach(func() {
		data.Spec.PipelineSource = projectForPipeline.Name
		data.Spec.CodeSources[0] = projectForBase.Name
	})

	It("failed to get product info", testUseCase.GetProductFail(func(codeRepo *MockCodeRepo, secretRepo *MockSecretrepo, resourceUseCase *ResourcesUsecase, nodestree *nodestree.MockNodesTree, gitRepo *MockGitRepo, client *kubernetes.MockClient) {
		biz := NewProjectPipelineRuntimeUsecase(logger, codeRepo, nodestree, resourceUseCase)
		err := biz.SaveProjectPipelineRuntime(context.Background(), bizOptions, data)
		Expect(err).Should(HaveOccurred())
	}))

	It("failed to get default project info", testUseCase.GetDefaultProjectFail(func(codeRepo *MockCodeRepo, secretRepo *MockSecretrepo, resourceUseCase *ResourcesUsecase, nodestree *nodestree.MockNodesTree, gitRepo *MockGitRepo, client *kubernetes.MockClient) {
		codeRepo.EXPECT().GetCodeRepo(gomock.Any(), gomock.Eq(pipelineSourceCodeRepoPath)).Return(pipelineSouceProject, nil)
		codeRepo.EXPECT().GetCodeRepo(gomock.Any(), gomock.Eq(codeRepoSourcePath)).Return(codeRepoSouceProject, nil)

		biz := NewProjectPipelineRuntimeUsecase(logger, codeRepo, nodestree, resourceUseCase)
		err := biz.SaveProjectPipelineRuntime(context.Background(), bizOptions, data)
		Expect(err).Should(HaveOccurred())
	}))

	It("will created successfully", testUseCase.CreateResourceSuccess(fakeNodes, func(codeRepo *MockCodeRepo, secretRepo *MockSecretrepo, resourceUseCase *ResourcesUsecase, nodestree *nodestree.MockNodesTree, gitRepo *MockGitRepo, client *kubernetes.MockClient) {
		codeRepo.EXPECT().GetCodeRepo(gomock.Any(), gomock.Eq(pipelineSourceCodeRepoPath)).Return(pipelineSouceProject, nil)
		codeRepo.EXPECT().GetCodeRepo(gomock.Any(), gomock.Eq(codeRepoSourcePath)).Return(codeRepoSouceProject, nil)

		biz := NewProjectPipelineRuntimeUsecase(logger, codeRepo, nodestree, resourceUseCase)
		err := biz.SaveProjectPipelineRuntime(context.Background(), bizOptions, data)
		Expect(err).ShouldNot(HaveOccurred())
	}))

	It("will updated successfully", testUseCase.UpdateResoureSuccess(fakeNodes, fakeNode, func(codeRepo *MockCodeRepo, secretRepo *MockSecretrepo, resourceUseCase *ResourcesUsecase, nodestree *nodestree.MockNodesTree, gitRepo *MockGitRepo, client *kubernetes.MockClient) {
		codeRepo.EXPECT().GetCodeRepo(gomock.Any(), gomock.Eq(pipelineSourceCodeRepoPath)).Return(pipelineSouceProject, nil)
		codeRepo.EXPECT().GetCodeRepo(gomock.Any(), gomock.Eq(codeRepoSourcePath)).Return(codeRepoSouceProject, nil)

		biz := NewProjectPipelineRuntimeUsecase(logger, codeRepo, nodestree, resourceUseCase)
		err := biz.SaveProjectPipelineRuntime(context.Background(), bizOptions, data)
		Expect(err).ShouldNot(HaveOccurred())
	}))

	It("auto merge conflict, updated successfully", testUseCase.UpdateResourceAndAutoMerge(fakeNodes, fakeNode, func(codeRepo *MockCodeRepo, secretRepo *MockSecretrepo, resourceUseCase *ResourcesUsecase, nodestree *nodestree.MockNodesTree, gitRepo *MockGitRepo, client *kubernetes.MockClient) {
		codeRepo.EXPECT().GetCodeRepo(gomock.Any(), gomock.Eq(pipelineSourceCodeRepoPath)).Return(pipelineSouceProject, nil)
		codeRepo.EXPECT().GetCodeRepo(gomock.Any(), gomock.Eq(codeRepoSourcePath)).Return(codeRepoSouceProject, nil)

		biz := NewProjectPipelineRuntimeUsecase(logger, codeRepo, nodestree, resourceUseCase)
		err := biz.SaveProjectPipelineRuntime(context.Background(), bizOptions, data)
		Expect(err).ShouldNot(HaveOccurred())
	}))

	It("failed to auto merge conflict", testUseCase.MergeConflictFail(fakeNodes, fakeNode, func(codeRepo *MockCodeRepo, secretRepo *MockSecretrepo, resourceUseCase *ResourcesUsecase, nodestree *nodestree.MockNodesTree, gitRepo *MockGitRepo, client *kubernetes.MockClient) {
		codeRepo.EXPECT().GetCodeRepo(gomock.Any(), gomock.Eq(pipelineSourceCodeRepoPath)).Return(pipelineSouceProject, nil)
		codeRepo.EXPECT().GetCodeRepo(gomock.Any(), gomock.Eq(codeRepoSourcePath)).Return(codeRepoSouceProject, nil)

		biz := NewProjectPipelineRuntimeUsecase(logger, codeRepo, nodestree, resourceUseCase)
		err := biz.SaveProjectPipelineRuntime(context.Background(), bizOptions, data)
		Expect(err).Should(HaveOccurred())
	}))

	It("failed to push code retry three times", testUseCase.CreateResourceAndAutoRetry(fakeNodes, func(codeRepo *MockCodeRepo, secretRepo *MockSecretrepo, resourceUseCase *ResourcesUsecase, nodestree *nodestree.MockNodesTree, gitRepo *MockGitRepo, client *kubernetes.MockClient) {
		codeRepo.EXPECT().GetCodeRepo(gomock.Any(), gomock.Eq(pipelineSourceCodeRepoPath)).Return(pipelineSouceProject, nil)
		codeRepo.EXPECT().GetCodeRepo(gomock.Any(), gomock.Eq(codeRepoSourcePath)).Return(codeRepoSouceProject, nil)

		biz := NewProjectPipelineRuntimeUsecase(logger, codeRepo, nodestree, resourceUseCase)
		err := biz.SaveProjectPipelineRuntime(context.Background(), bizOptions, data)
		Expect(err).Should(HaveOccurred())
	}))

	It("modify resource but non compliant layout", testUseCase.UpdateResourceButNotConformTemplate(fakeNodes, fakeNode, func(codeRepo *MockCodeRepo, secretRepo *MockSecretrepo, resourceUseCase *ResourcesUsecase, nodestree *nodestree.MockNodesTree, gitRepo *MockGitRepo, client *kubernetes.MockClient) {
		codeRepo.EXPECT().GetCodeRepo(gomock.Any(), gomock.Eq(pipelineSourceCodeRepoPath)).Return(pipelineSouceProject, nil)
		codeRepo.EXPECT().GetCodeRepo(gomock.Any(), gomock.Eq(codeRepoSourcePath)).Return(codeRepoSouceProject, nil)

		biz := NewProjectPipelineRuntimeUsecase(logger, codeRepo, nodestree, resourceUseCase)
		err := biz.SaveProjectPipelineRuntime(context.Background(), bizOptions, data)
		Expect(err).Should(HaveOccurred())
	}))

	It("failed to save config", testUseCase.SaveConfigFail(fakeNodes, fakeNode, func(codeRepo *MockCodeRepo, secretRepo *MockSecretrepo, resourceUseCase *ResourcesUsecase, nodestree *nodestree.MockNodesTree, gitRepo *MockGitRepo, client *kubernetes.MockClient) {
		codeRepo.EXPECT().GetCodeRepo(gomock.Any(), gomock.Eq(pipelineSourceCodeRepoPath)).Return(pipelineSouceProject, nil)
		codeRepo.EXPECT().GetCodeRepo(gomock.Any(), gomock.Eq(codeRepoSourcePath)).Return(codeRepoSouceProject, nil)

		biz := NewProjectPipelineRuntimeUsecase(logger, codeRepo, nodestree, resourceUseCase)
		err := biz.SaveProjectPipelineRuntime(context.Background(), bizOptions, data)
		Expect(err).Should(HaveOccurred())
	}))

	Describe("check reference by resources", func() {
		It("incorrect product name", testUseCase.CheckReferenceButIncorrectProduct(fakeNodes, func(options nodestree.CompareOptions, nodestree *nodestree.MockNodesTree) {
			biz := NewProjectPipelineRuntimeUsecase(logger, nil, nodestree, nil)
			ok, err := biz.CheckReference(options, fakeNode, nil)
			Expect(err).Should(HaveOccurred())
			Expect(ok).To(BeTrue())
		}))

		It("project reference not found", func() {
			options := nodestree.CompareOptions{
				Nodes:       fakeNodes,
				ProductName: defaultProductId,
			}
			nodestree := nodestree.NewMockNodesTree(ctl)
			nodestree.EXPECT().AppendOperators(gomock.Any())

			biz := NewProjectPipelineRuntimeUsecase(logger, nil, nodestree, nil)
			ok, err := biz.CheckReference(options, fakeNode, nil)
			Expect(err).Should(HaveOccurred())
			Expect(ok).To(BeTrue())
		})

		It("repeated reference code repository", func() {
			options := nodestree.CompareOptions{
				Nodes:       fakeNodes,
				ProductName: defaultProductId,
			}
			nodestree := nodestree.NewMockNodesTree(ctl)
			nodestree.EXPECT().AppendOperators(gomock.Any())

			newResouce := fakeResource.DeepCopy()
			newResouce.Spec.PipelineSource = projectForPipelineRepoID
			newResouce.Spec.CodeSources[0] = projectForPipelineRepoID
			fakeNode.Content = newResouce

			biz := NewProjectPipelineRuntimeUsecase(logger, nil, nodestree, nil)
			ok, err := biz.CheckReference(options, fakeNode, nil)
			Expect(err).Should(HaveOccurred())
			Expect(ok).To(BeTrue())
		})

		It("environment reference not found", func() {
			projectName := fakeResource.Spec.Project
			projectNodes := createProjectNodes(createProjectNode(createProjectResource(projectName)))
			fakeNodes.Children = append(fakeNodes.Children, projectNodes.Children...)
			options := nodestree.CompareOptions{
				Nodes:       fakeNodes,
				ProductName: defaultProductId,
			}
			nodestree := nodestree.NewMockNodesTree(ctl)
			nodestree.EXPECT().AppendOperators(gomock.Any())

			biz := NewProjectPipelineRuntimeUsecase(logger, nil, nodestree, nil)
			ok, err := biz.CheckReference(options, fakeNode, nil)
			Expect(err).Should(HaveOccurred())
			Expect(ok).To(BeTrue())
		})

		It("pipeline source reference not found", func() {
			projectName := fakeResource.Spec.Project
			projectNodes := createProjectNodes(createProjectNode(createProjectResource(projectName)))
			env := fakeResource.Spec.Destination
			envProjects := createContainEnvironmentNodes(createEnvironmentNode(createEnvironmentResource(env)))
			fakeNodes.Children = append(fakeNodes.Children, projectNodes.Children...)
			fakeNodes.Children = append(fakeNodes.Children, envProjects.Children...)

			options := nodestree.CompareOptions{
				Nodes:       fakeNodes,
				ProductName: defaultProductId,
			}
			nodestree := nodestree.NewMockNodesTree(ctl)
			nodestree.EXPECT().AppendOperators(gomock.Any())

			biz := NewProjectPipelineRuntimeUsecase(logger, nil, nodestree, nil)
			ok, err := biz.CheckReference(options, fakeNode, nil)
			Expect(err).Should(HaveOccurred())
			Expect(ok).To(BeTrue())
		})

		It("coderepo sources reference not found", func() {
			projectName := fakeResource.Spec.Project
			projectNodes := createProjectNodes(createProjectNode(createProjectResource(projectName)))
			env := fakeResource.Spec.Destination
			envProjects := createContainEnvironmentNodes(createEnvironmentNode(createEnvironmentResource(env)))
			fakeNodes.Children = append(fakeNodes.Children, projectNodes.Children...)
			fakeNodes.Children = append(fakeNodes.Children, envProjects.Children...)
			codeRepoNodes := createFakeCcontainingCodeRepoNodes(createFakeCodeRepoNode(createFakeCodeRepoResource(projectForPipelineRepoID)))
			fakeNodes.Children = append(fakeNodes.Children, codeRepoNodes.Children...)

			options := nodestree.CompareOptions{
				Nodes:       fakeNodes,
				ProductName: defaultProductId,
			}
			nodestree := nodestree.NewMockNodesTree(ctl)
			nodestree.EXPECT().AppendOperators(gomock.Any())

			newResouce := fakeResource.DeepCopy()
			newResouce.Spec.PipelineSource = projectForPipelineRepoID
			fakeNode.Content = newResouce

			biz := NewProjectPipelineRuntimeUsecase(logger, nil, nodestree, nil)
			ok, err := biz.CheckReference(options, fakeNode, nil)
			Expect(err).Should(HaveOccurred())
			Expect(ok).To(BeTrue())
		})

		It("will successed", func() {
			projectName := fakeResource.Spec.Project
			projectNodes := createProjectNodes(createProjectNode(createProjectResource(projectName)))
			env := fakeResource.Spec.Destination
			envProjects := createContainEnvironmentNodes(createEnvironmentNode(createEnvironmentResource(env)))
			fakeNodes.Children = append(fakeNodes.Children, projectNodes.Children...)
			fakeNodes.Children = append(fakeNodes.Children, envProjects.Children...)
			codeRepoNodes := createFakeCcontainingCodeRepoNodes(createFakeCodeRepoNode(createFakeCodeRepoResource(projectForPipelineRepoID)))
			fakeNodes.Children = append(fakeNodes.Children, codeRepoNodes.Children...)
			codeRepoNodes = createFakeCcontainingCodeRepoNodes(createFakeCodeRepoNode(createFakeCodeRepoResource(projectForBaseRepoID)))
			fakeNodes.Children = append(fakeNodes.Children, codeRepoNodes.Children...)

			options := nodestree.CompareOptions{
				Nodes:       fakeNodes,
				ProductName: defaultProductId,
			}
			nodestree := nodestree.NewMockNodesTree(ctl)
			nodestree.EXPECT().AppendOperators(gomock.Any())

			newResouce := fakeResource.DeepCopy()
			newResouce.Spec.PipelineSource = projectForPipelineRepoID
			newResouce.Spec.CodeSources[0] = projectForBaseRepoID
			fakeNode.Content = newResouce

			biz := NewProjectPipelineRuntimeUsecase(logger, nil, nodestree, nil)
			ok, err := biz.CheckReference(options, fakeNode, nil)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(ok).To(BeTrue())
		})
	})

})

var _ = Describe("Delete project pipeline runtime", func() {
	var (
		resourceName = "projectpipelineruntime1"
		fakeResource = createProjectPiepeLineResource(resourceName)
		fakeNode     = createFakeProjectPipelineRuntimeNode(fakeResource)
		fakeNodes    = createFakeProjectPipelineRuntimeNodes(fakeNode)
		bizOptions   = &BizOptions{
			ResouceName: resourceName,
			ProductName: defaultGroupName,
		}
	)

	BeforeEach(func() {
		err := os.MkdirAll(filepath.Dir(fakeNode.Path), 0644)
		Expect(err).ShouldNot(HaveOccurred())
		_, err = os.Create(fakeNode.Path)
		Expect(err).ShouldNot(HaveOccurred())
	})

	It("will deleted successfully", testUseCase.DeleteResourceSuccess(fakeNodes, fakeNode, func(codeRepo *MockCodeRepo, secretRepo *MockSecretrepo, resourceUseCase *ResourcesUsecase, nodestree *nodestree.MockNodesTree, gitRepo *MockGitRepo, client *kubernetes.MockClient) {
		biz := NewProjectPipelineRuntimeUsecase(logger, codeRepo, nodestree, resourceUseCase)
		err := biz.DeleteProjectPipelineRuntime(context.Background(), bizOptions)
		Expect(err).ShouldNot(HaveOccurred())
	}))

	It("modify resource but non compliant layout standards", testUseCase.DeleteResourceErrorLayout(fakeNodes, fakeNode, func(codeRepo *MockCodeRepo, secretRepo *MockSecretrepo, resourceUseCase *ResourcesUsecase, nodestree *nodestree.MockNodesTree, gitRepo *MockGitRepo, client *kubernetes.MockClient) {
		biz := NewProjectPipelineRuntimeUsecase(logger, codeRepo, nodestree, resourceUseCase)
		err := biz.DeleteProjectPipelineRuntime(context.Background(), bizOptions)
		Expect(err).Should(HaveOccurred())
	}))
})
