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
	utilstrings "github.com/nautes-labs/api-server/util/string"
	resourcev1alpha1 "github.com/nautes-labs/pkg/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createDeploymentRuntimeResource(name, repoID string) *resourcev1alpha1.DeploymentRuntime {
	return &resourcev1alpha1.DeploymentRuntime{
		ObjectMeta: v1.ObjectMeta{
			Name: name,
		},
		Spec: resourcev1alpha1.DeploymentRuntimeSpec{
			Product:     defaultProductId,
			ProjectsRef: []string{"project1"},
			Destination: "env1",
			ManifestSource: resourcev1alpha1.ManifestSource{
				CodeRepo:       repoID,
				TargetRevision: "main",
				Path:           "production",
			},
		},
	}
}

func createFakeDeploymentRuntimeNode(resource *resourcev1alpha1.DeploymentRuntime) *nodestree.Node {
	return &nodestree.Node{
		Name:    resource.Name,
		Path:    fmt.Sprintf("%s/%s/%s.yaml", localRepositaryPath, _RuntimesDir, resource.Name),
		Level:   3,
		Content: resource,
		Kind:    nodestree.DeploymentRuntime,
	}
}

func createFakeDeployRuntimeNodes(node *nodestree.Node) nodestree.Node {
	fakeSubNode := &nodestree.Node{
		Name:     nodestree.DeploymentRuntime,
		Path:     fmt.Sprintf("%s/%s", localRepositaryPath, _RuntimesDir),
		IsDir:    true,
		Level:    2,
		Content:  node,
		Children: []*nodestree.Node{node},
	}
	fakeNodes := nodestree.Node{
		Name:     defaultProjectName,
		Path:     defaultProjectName,
		IsDir:    true,
		Level:    1,
		Children: []*nodestree.Node{fakeSubNode},
	}

	return fakeNodes
}

var _ = Describe("Get deployment runtime", func() {
	var (
		resourceName = "runtime1"
		toGetProject = &Project{Id: 1222, HttpUrlToRepo: fmt.Sprintf("ssh://git@gitlab.io/nautes-labs/%s.git", resourceName)}
		repoID       = fmt.Sprintf("%s%d", _RepoPrefix, int(toGetProject.Id))
		fakeResource = createDeploymentRuntimeResource(resourceName, repoID)
		fakeNode     = createFakeDeploymentRuntimeNode(fakeResource)
		fakeNodes    = createFakeDeployRuntimeNodes(fakeNode)
	)

	It("will get successed", testUseCase.GetResourceSuccess(fakeNodes, fakeNode, func(codeRepo *MockCodeRepo, secretRepo *MockSecretrepo, resourcesUsecase *ResourcesUsecase, nodestree *nodestree.MockNodesTree, gitRepo *MockGitRepo, client *kubernetes.MockClient) {
		id, _ := utilstrings.ExtractNumber("product-", fakeResource.Spec.Product)
		codeRepo.EXPECT().GetGroup(gomock.Any(), id).Return(defaultProductGroup, nil)
		id, _ = utilstrings.ExtractNumber("repo-", fakeResource.Spec.ManifestSource.CodeRepo)
		codeRepo.EXPECT().GetCodeRepo(gomock.Any(), id).Return(defautlProject, nil)

		biz := NewDeploymentRuntimeUsecase(logger, codeRepo, nodestree, resourcesUsecase)
		result, err := biz.GetDeploymentRuntime(context.Background(), resourceName, defaultGroupName)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(result).Should(Equal(fakeResource))
	}))

	It("will fail when resource is not found", testUseCase.GetResourceFail(func(codeRepo *MockCodeRepo, secretRepo *MockSecretrepo, resourceUseCase *ResourcesUsecase, nodestree *nodestree.MockNodesTree, gitRepo *MockGitRepo, client *kubernetes.MockClient) {
		biz := NewDeploymentRuntimeUsecase(logger, codeRepo, nodestree, resourceUseCase)
		_, err := biz.GetDeploymentRuntime(context.Background(), resourceName, defaultGroupName)
		Expect(err).Should(HaveOccurred())
	}))
})

var _ = Describe("List deployment runtimes", func() {
	var (
		resourceName = "runtime1"
		toGetProject = &Project{Id: 1222, HttpUrlToRepo: fmt.Sprintf("ssh://git@gitlab.io/nautes-labs/%s.git", resourceName)}
		repoID       = fmt.Sprintf("%s%d", _RepoPrefix, int(toGetProject.Id))
		fakeResource = createDeploymentRuntimeResource(resourceName, repoID)
		fakeNode     = createFakeDeploymentRuntimeNode(fakeResource)
		fakeNodes    = createFakeDeployRuntimeNodes(fakeNode)
	)

	It("will successfully", testUseCase.ListResourceSuccess(fakeNodes, func(codeRepo *MockCodeRepo, secretRepo *MockSecretrepo, resourceUseCase *ResourcesUsecase, nodestree *nodestree.MockNodesTree, gitRepo *MockGitRepo, client *kubernetes.MockClient) {
		id, _ := utilstrings.ExtractNumber("product-", fakeResource.Spec.Product)
		codeRepo.EXPECT().GetGroup(gomock.Any(), id).Return(defaultProductGroup, nil)
		id, _ = utilstrings.ExtractNumber("repo-", fakeResource.Spec.ManifestSource.CodeRepo)
		codeRepo.EXPECT().GetCodeRepo(gomock.Any(), id).Return(defautlProject, nil)

		biz := NewDeploymentRuntimeUsecase(logger, codeRepo, nodestree, resourceUseCase)
		results, err := biz.ListDeploymentRuntimes(ctx, defaultGroupName)
		Expect(err).ShouldNot(HaveOccurred())
		for _, result := range results {
			Expect(result).Should(Equal(fakeResource))
		}
	}))

	It("will fail when project is not found", func() {
		codeRepo := NewMockCodeRepo(ctl)
		codeRepo.EXPECT().GetGroup(gomock.Any(), gomock.Eq(defaultGroupName)).Return(defaultProductGroup, nil).AnyTimes()
		codeRepo.EXPECT().GetCodeRepo(gomock.Any(), gomock.Eq(defaultProjectPath)).Return(nil, ErrorProjectNotFound)

		gitRepo := NewMockGitRepo(ctl)

		in := nodestree.NewMockNodesTree(ctl)
		in.EXPECT().AppendOperators(gomock.Any())

		resourcesUsecase := NewResourcesUsecase(logger, codeRepo, nil, gitRepo, in, nautesConfigs)
		biz := NewDeploymentRuntimeUsecase(logger, codeRepo, in, resourcesUsecase)
		_, err := biz.ListDeploymentRuntimes(ctx, defaultGroupName)
		Expect(err).Should(HaveOccurred())
	})

	It("does not conform to the template layout", testUseCase.ListResourceNotMatch(fakeNodes, func(codeRepo *MockCodeRepo, secretRepo *MockSecretrepo, resourceUseCase *ResourcesUsecase, nodestree *nodestree.MockNodesTree, gitRepo *MockGitRepo, client *kubernetes.MockClient) {
		biz := NewDeploymentRuntimeUsecase(logger, codeRepo, nodestree, resourceUseCase)
		_, err := biz.ListDeploymentRuntimes(ctx, defaultGroupName)
		Expect(err).Should(HaveOccurred())
	}))
})

var _ = Describe("Save deployment runtime", func() {
	var (
		resourceName          = "runtime1"
		toGetProject          = &Project{Id: 1222, HttpUrlToRepo: fmt.Sprintf("ssh://git@gitlab.io/nautes-labs/%s.git", resourceName)}
		repoID                = fmt.Sprintf("%s%d", _RepoPrefix, int(toGetProject.Id))
		fakeResource          = createDeploymentRuntimeResource(resourceName, repoID)
		fakeNode              = createFakeDeploymentRuntimeNode(fakeResource)
		fakeNodes             = createFakeDeployRuntimeNodes(fakeNode)
		pid                   = fmt.Sprintf("%s/%s", defaultGroupName, fakeResource.Spec.ManifestSource.CodeRepo)
		project               = &Project{Id: 1222}
		deploymentRuntimeData = &DeploymentRuntimeData{
			Name: fakeResource.Name,
			Spec: fakeResource.Spec,
		}
		bizOptions = &BizOptions{
			ResouceName: resourceName,
			ProductName: defaultGroupName,
		}
	)

	It("failed to get product info", testUseCase.GetProductFail(func(codeRepo *MockCodeRepo, secretRepo *MockSecretrepo, resourceUseCase *ResourcesUsecase, nodestree *nodestree.MockNodesTree, gitRepo *MockGitRepo, client *kubernetes.MockClient) {
		biz := NewDeploymentRuntimeUsecase(logger, codeRepo, nodestree, resourceUseCase)
		err := biz.SaveDeploymentRuntime(context.Background(), bizOptions, deploymentRuntimeData)
		Expect(err).Should(HaveOccurred())
	}))

	It("failed to get default project info", testUseCase.GetDefaultProjectFail(func(codeRepo *MockCodeRepo, secretRepo *MockSecretrepo, resourceUseCase *ResourcesUsecase, nodestree *nodestree.MockNodesTree, gitRepo *MockGitRepo, client *kubernetes.MockClient) {
		codeRepo.EXPECT().GetCodeRepo(gomock.Any(), pid).Return(project, nil)
		biz := NewDeploymentRuntimeUsecase(logger, codeRepo, nodestree, resourceUseCase)
		err := biz.SaveDeploymentRuntime(context.Background(), bizOptions, deploymentRuntimeData)
		Expect(err).Should(HaveOccurred())
	}))

	It("will created successfully", testUseCase.CreateResourceSuccess(fakeNodes, func(codeRepo *MockCodeRepo, secretRepo *MockSecretrepo, resourceUseCase *ResourcesUsecase, nodestree *nodestree.MockNodesTree, gitRepo *MockGitRepo, client *kubernetes.MockClient) {
		codeRepo.EXPECT().GetCodeRepo(gomock.Any(), pid).Return(project, nil)
		biz := NewDeploymentRuntimeUsecase(logger, codeRepo, nodestree, resourceUseCase)
		err := biz.SaveDeploymentRuntime(context.Background(), bizOptions, deploymentRuntimeData)
		Expect(err).ShouldNot(HaveOccurred())
	}))

	It("will updated successfully", testUseCase.UpdateResoureSuccess(fakeNodes, fakeNode, func(codeRepo *MockCodeRepo, secretRepo *MockSecretrepo, resourceUseCase *ResourcesUsecase, nodestree *nodestree.MockNodesTree, gitRepo *MockGitRepo, client *kubernetes.MockClient) {
		codeRepo.EXPECT().GetCodeRepo(gomock.Any(), pid).Return(project, nil)

		biz := NewDeploymentRuntimeUsecase(logger, codeRepo, nodestree, resourceUseCase)
		err := biz.SaveDeploymentRuntime(context.Background(), bizOptions, deploymentRuntimeData)
		Expect(err).ShouldNot(HaveOccurred())
	}))

	It("auto merge conflict, updated successfully", testUseCase.UpdateResourceAndAutoMerge(fakeNodes, fakeNode, func(codeRepo *MockCodeRepo, secretRepo *MockSecretrepo, resourceUseCase *ResourcesUsecase, nodestree *nodestree.MockNodesTree, gitRepo *MockGitRepo, client *kubernetes.MockClient) {
		codeRepo.EXPECT().GetCodeRepo(gomock.Any(), pid).Return(project, nil)

		biz := NewDeploymentRuntimeUsecase(logger, codeRepo, nodestree, resourceUseCase)
		err := biz.SaveDeploymentRuntime(context.Background(), bizOptions, deploymentRuntimeData)
		Expect(err).ShouldNot(HaveOccurred())
	}))

	It("failed to save config", testUseCase.SaveConfigFail(fakeNodes, fakeNode, func(codeRepo *MockCodeRepo, secretRepo *MockSecretrepo, resourceUseCase *ResourcesUsecase, nodestree *nodestree.MockNodesTree, gitRepo *MockGitRepo, client *kubernetes.MockClient) {
		codeRepo.EXPECT().GetCodeRepo(gomock.Any(), pid).Return(project, nil)

		biz := NewDeploymentRuntimeUsecase(logger, codeRepo, nodestree, resourceUseCase)
		err := biz.SaveDeploymentRuntime(context.Background(), bizOptions, deploymentRuntimeData)
		Expect(err).Should(HaveOccurred())
	}))

	It("failed to push code retry three times", testUseCase.CreateResourceAndAutoRetry(fakeNodes, func(codeRepo *MockCodeRepo, secretRepo *MockSecretrepo, resourceUseCase *ResourcesUsecase, nodestree *nodestree.MockNodesTree, gitRepo *MockGitRepo, client *kubernetes.MockClient) {
		codeRepo.EXPECT().GetCodeRepo(gomock.Any(), pid).Return(project, nil)

		biz := NewDeploymentRuntimeUsecase(logger, codeRepo, nodestree, resourceUseCase)
		err := biz.SaveDeploymentRuntime(context.Background(), bizOptions, deploymentRuntimeData)
		Expect(err).Should(HaveOccurred())
	}))
	It("auto merge code push successed and retry three times when the remote code changes", func() {
		codeRepo := NewMockCodeRepo(ctl)
		codeRepo.EXPECT().GetGroup(gomock.Any(), gomock.Eq(defaultGroupName)).Return(defaultProductGroup, nil).AnyTimes()
		first := codeRepo.EXPECT().GetCodeRepo(gomock.Any(), pid).Return(project, nil)
		codeRepo.EXPECT().GetCodeRepo(gomock.Any(), gomock.Eq(defaultProjectPath)).Return(defautlProject, nil).After(first)
		codeRepo.EXPECT().GetCurrentUser(gomock.Any()).Return(_GitUser, _GitEmail, nil)

		gitRepo := NewMockGitRepo(ctl)
		gitRepo.EXPECT().Clone(gomock.Any(), cloneRepositoryParam).Return(localRepositaryPath, nil)
		firstFetch := gitRepo.EXPECT().Fetch(gomock.Any(), gomock.Any(), "origin").Return("any", nil)
		secondFetch := gitRepo.EXPECT().Fetch(gomock.Any(), gomock.Any()).Return("any", nil).After(firstFetch)
		thirdFetch := gitRepo.EXPECT().Fetch(gomock.Any(), gomock.Any(), "origin").Return("any", nil).After(secondFetch)
		fouthFetch := gitRepo.EXPECT().Fetch(gomock.Any(), gomock.Any()).Return("any", nil).After(thirdFetch)
		fifthFetch := gitRepo.EXPECT().Fetch(gomock.Any(), gomock.Any(), "origin").Return("any", nil).After(fouthFetch)
		gitRepo.EXPECT().Fetch(gomock.Any(), gomock.Any()).Return("any", nil).After(fifthFetch)
		gitRepo.EXPECT().Diff(gomock.Any(), gomock.Any(), "main", "remotes/origin/main").Return("any", nil).AnyTimes()
		gitRepo.EXPECT().Merge(gomock.Any(), localRepositaryPath).Return("successfully auto merge", nil).AnyTimes()
		gitRepo.EXPECT().Push(gomock.Any(), gomock.Any()).Return(fmt.Errorf("unable to push code")).AnyTimes()
		gitRepo.EXPECT().Commit(localRepositaryPath, gomock.Any()).AnyTimes()

		in := nodestree.NewMockNodesTree(ctl)
		in.EXPECT().AppendOperators(gomock.Any())
		resourcesUsecase := NewResourcesUsecase(logger, codeRepo, nil, gitRepo, in, nautesConfigs)
		biz := NewDeploymentRuntimeUsecase(logger, codeRepo, in, resourcesUsecase)
		in.EXPECT().Load(gomock.Eq(localRepositaryPath)).Return(emptyNodes, nil)
		in.EXPECT().Compare(gomock.Any()).Return(nil).AnyTimes()
		in.EXPECT().GetNode(gomock.Any(), gomock.Any(), gomock.Any()).Return(fakeNode)
		in.EXPECT().InsertNodes(gomock.Any(), gomock.Any()).Return(&fakeNodes, nil)

		err := biz.SaveDeploymentRuntime(context.Background(), bizOptions, deploymentRuntimeData)
		Expect(err).Should(HaveOccurred())
	})

	It("failed to auto merge conflict", testUseCase.MergeConflictFail(fakeNodes, fakeNode, func(codeRepo *MockCodeRepo, secretRepo *MockSecretrepo, resourceUseCase *ResourcesUsecase, nodestree *nodestree.MockNodesTree, gitRepo *MockGitRepo, client *kubernetes.MockClient) {
		codeRepo.EXPECT().GetCodeRepo(gomock.Any(), pid).Return(project, nil)

		biz := NewDeploymentRuntimeUsecase(logger, codeRepo, nodestree, resourceUseCase)
		err := biz.SaveDeploymentRuntime(context.Background(), bizOptions, deploymentRuntimeData)
		Expect(err).Should(HaveOccurred())
	}))

	It("modify resource but non compliant layout", testUseCase.UpdateResourceButNotConformTemplate(fakeNodes, fakeNode, func(codeRepo *MockCodeRepo, secretRepo *MockSecretrepo, resourceUseCase *ResourcesUsecase, nodestree *nodestree.MockNodesTree, gitRepo *MockGitRepo, client *kubernetes.MockClient) {
		codeRepo.EXPECT().GetCodeRepo(gomock.Any(), pid).Return(project, nil)

		biz := NewDeploymentRuntimeUsecase(logger, codeRepo, nodestree, resourceUseCase)
		err := biz.SaveDeploymentRuntime(context.Background(), bizOptions, deploymentRuntimeData)
		Expect(err).Should(HaveOccurred())
	}))

	Describe("check reference by resources", func() {
		It("incorrect product name", testUseCase.CheckReferenceButIncorrectProduct(fakeNodes, func(options nodestree.CompareOptions, nodestree *nodestree.MockNodesTree) {
			biz := NewDeploymentRuntimeUsecase(logger, nil, nodestree, nil)
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

			biz := NewDeploymentRuntimeUsecase(logger, nil, nodestree, nil)
			ok, err := biz.CheckReference(options, fakeNode, nil)
			Expect(err).Should(HaveOccurred())
			Expect(ok).To(BeTrue())
		})
		It("environment reference not found", func() {
			projectName := fakeResource.Spec.ProjectsRef[0]
			projectNodes := createProjectNodes(createProjectNode(createProjectResource(projectName)))
			fakeNodes.Children = append(fakeNodes.Children, projectNodes.Children...)

			options := nodestree.CompareOptions{
				Nodes:       fakeNodes,
				ProductName: defaultProductId,
			}
			nodestree := nodestree.NewMockNodesTree(ctl)
			nodestree.EXPECT().AppendOperators(gomock.Any())
			biz := NewDeploymentRuntimeUsecase(logger, nil, nodestree, nil)
			ok, err := biz.CheckReference(options, fakeNode, nil)
			Expect(err).Should(HaveOccurred())
			Expect(ok).To(BeTrue())
		})
		It("codeRepo reference not found", func() {
			projectName := fakeResource.Spec.ProjectsRef[0]
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

			biz := NewDeploymentRuntimeUsecase(logger, nil, nodestree, nil)
			ok, err := biz.CheckReference(options, fakeNode, nil)
			Expect(err).Should(HaveOccurred())
			Expect(ok).To(BeTrue())
		})
		It("will successed", func() {
			projectName := fakeResource.Spec.ProjectsRef[0]
			projectNodes := createProjectNodes(createProjectNode(createProjectResource(projectName)))
			env := fakeResource.Spec.Destination
			envProjects := createContainEnvironmentNodes(createEnvironmentNode(createEnvironmentResource(env)))
			fakeNodes.Children = append(fakeNodes.Children, projectNodes.Children...)
			fakeNodes.Children = append(fakeNodes.Children, envProjects.Children...)
			codeRepoNodes := createFakeCcontainingCodeRepoNodes(createFakeCodeRepoNode(createFakeCodeRepoResource(repoID)))
			fakeNodes.Children = append(fakeNodes.Children, codeRepoNodes.Children...)

			options := nodestree.CompareOptions{
				Nodes:       fakeNodes,
				ProductName: defaultProductId,
			}
			nodestree := nodestree.NewMockNodesTree(ctl)
			nodestree.EXPECT().AppendOperators(gomock.Any())

			biz := NewDeploymentRuntimeUsecase(logger, nil, nodestree, nil)
			ok, err := biz.CheckReference(options, fakeNode, nil)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(ok).To(BeTrue())
		})
	})
})

var _ = Describe("Delete deployment runtime", func() {
	var (
		resourceName = "runtime1"
		toGetProject = &Project{Id: 1222, HttpUrlToRepo: fmt.Sprintf("ssh://git@gitlab.io/nautes-labs/%s.git", resourceName)}
		repoID       = fmt.Sprintf("%s%d", _RepoPrefix, int(toGetProject.Id))
		fakeResource = createDeploymentRuntimeResource(resourceName, repoID)
		fakeNode     = createFakeDeploymentRuntimeNode(fakeResource)
		fakeNodes    = createFakeDeployRuntimeNodes(fakeNode)
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
		biz := NewDeploymentRuntimeUsecase(logger, codeRepo, nodestree, resourceUseCase)
		err := biz.DeleteDeploymentRuntime(context.Background(), bizOptions)
		Expect(err).ShouldNot(HaveOccurred())
	}))

	It("modify resource but non compliant layout standards", testUseCase.DeleteResourceErrorLayout(fakeNodes, fakeNode, func(codeRepo *MockCodeRepo, secretRepo *MockSecretrepo, resourceUseCase *ResourcesUsecase, nodestree *nodestree.MockNodesTree, gitRepo *MockGitRepo, client *kubernetes.MockClient) {
		biz := NewDeploymentRuntimeUsecase(logger, codeRepo, nodestree, resourceUseCase)
		err := biz.DeleteDeploymentRuntime(context.Background(), bizOptions)
		Expect(err).Should(HaveOccurred())
	}))
})
