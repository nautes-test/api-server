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
	"strconv"

	"github.com/golang/mock/gomock"
	commonv1 "github.com/nautes-labs/api-server/api/common/v1"
	"github.com/nautes-labs/api-server/pkg/kubernetes"
	"github.com/nautes-labs/api-server/pkg/nodestree"
	utilstrings "github.com/nautes-labs/api-server/util/string"
	resourcev1alpha1 "github.com/nautes-labs/pkg/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	_DefaultProjectResourceName = "project1"
)

func createFakeCodeRepoResource(name string) *resourcev1alpha1.CodeRepo {
	return &resourcev1alpha1.CodeRepo{
		ObjectMeta: v1.ObjectMeta{
			Name: name,
		},
		TypeMeta: v1.TypeMeta{
			Kind: nodestree.CodeRepo,
		},
		Spec: resourcev1alpha1.CodeRepoSpec{
			Product:           defaultProductId,
			RepoName:          name,
			Project:           _DefaultProjectResourceName,
			DeploymentRuntime: true,
			Webhook: &resourcev1alpha1.Webhook{
				Events:    []string{"push_events"},
				Isolation: "true",
			},
		},
	}
}

func createFakeCodeRepoNode(resource *resourcev1alpha1.CodeRepo) *nodestree.Node {
	return &nodestree.Node{
		Name:    resource.Name,
		Path:    fmt.Sprintf("%s/%s/%s/%s.yaml", localRepositaryPath, _CodeReposSubDir, resource.Name, resource.Name),
		Level:   4,
		Content: resource,
		Kind:    nodestree.CodeRepo,
	}
}

func createFakeCcontainingCodeRepoNodes(node *nodestree.Node) nodestree.Node {
	return nodestree.Node{
		Name:  defaultProjectName,
		Path:  defaultProjectName,
		IsDir: true,
		Level: 1,
		Children: []*nodestree.Node{
			{
				Name:  _CodeReposSubDir,
				Path:  fmt.Sprintf("%v/%v", defaultProjectName, _CodeReposSubDir),
				IsDir: true,
				Level: 2,
				Children: []*nodestree.Node{
					{
						Name:  node.Name,
						Path:  fmt.Sprintf("%s/%s/%s", localRepositaryPath, _CodeReposSubDir, node.Name),
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

var _ = Describe("Get codeRepo", func() {
	var (
		resourceName      = "toGetCodeRepo"
		fakeResource      = createFakeCodeRepoResource(resourceName)
		fakeNode          = createFakeCodeRepoNode(fakeResource)
		fakeNodes         = createFakeCcontainingCodeRepoNodes(fakeNode)
		project           = &Project{Id: 1222, HttpUrlToRepo: fmt.Sprintf("ssh://git@gitlab.io/nautes-labs/%s.git", resourceName)}
		toGetCodeRepoPath = fmt.Sprintf("%s/%s", defaultProductGroup.Path, resourceName)
	)

	It("will get successfully", testUseCase.GetResourceSuccess(fakeNodes, fakeNode, func(codeRepo *MockCodeRepo, secretRepo *MockSecretrepo, resourcesUsecase *ResourcesUsecase, nodestree *nodestree.MockNodesTree, gitRepo *MockGitRepo, client *kubernetes.MockClient) {
		codeRepo.EXPECT().GetCodeRepo(gomock.Any(), gomock.Eq(toGetCodeRepoPath)).Return(project, nil)
		id, _ := utilstrings.ExtractNumber("product-", fakeResource.Spec.Product)
		codeRepo.EXPECT().GetGroup(gomock.Any(), id).Return(defaultProductGroup, nil)

		biz := NewCodeRepoUsecase(logger, codeRepo, secretRepo, nodestree, nautesConfigs, resourcesUsecase, nil)
		item, _, err := biz.GetCodeRepo(context.Background(), resourceName, defaultGroupName)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(item).To(Equal(fakeResource))
	}))

	It("will fail when resource is not found", testUseCase.GetResourceFail(func(codeRepo *MockCodeRepo, secretRepo *MockSecretrepo, resourceUseCase *ResourcesUsecase, nodestree *nodestree.MockNodesTree, gitRepo *MockGitRepo, client *kubernetes.MockClient) {
		codeRepo.EXPECT().GetCodeRepo(gomock.Any(), gomock.Eq(toGetCodeRepoPath)).Return(project, nil)

		biz := NewCodeRepoUsecase(logger, codeRepo, secretRepo, nodestree, nautesConfigs, resourceUseCase, nil)
		_, _, err := biz.GetCodeRepo(context.Background(), resourceName, defaultGroupName)
		Expect(err).Should(HaveOccurred())
	}))
})

var _ = Describe("List coderepos", func() {
	var (
		resourceName       = "toGetCodeRepo"
		fakeResource       = createFakeCodeRepoResource(resourceName)
		fakeNode           = createFakeCodeRepoNode(fakeResource)
		fakeNodes          = createFakeCcontainingCodeRepoNodes(fakeNode)
		codeRepoAndProject = &CodeRepoAndProject{
			CodeRepo: fakeResource,
			Project:  defautlProject,
		}
	)

	It("will list successfully", testUseCase.ListResourceSuccess(fakeNodes, func(codeRepo *MockCodeRepo, secretRepo *MockSecretrepo, resourceUseCase *ResourcesUsecase, nodestree *nodestree.MockNodesTree, gitRepo *MockGitRepo, client *kubernetes.MockClient) {
		gid, _ := utilstrings.ExtractNumber("product-", fakeResource.Spec.Product)
		pid := fmt.Sprintf("%s/%s", defaultGroupName, fakeResource.Spec.RepoName)
		codeRepo.EXPECT().GetGroup(gomock.Any(), gid).Return(defaultProductGroup, nil)
		codeRepo.EXPECT().GetCodeRepo(gomock.Any(), pid).Return(defautlProject, nil)

		biz := NewCodeRepoUsecase(logger, codeRepo, secretRepo, nodestree, nautesConfigs, resourceUseCase, nil)
		results, err := biz.ListCodeRepos(ctx, defaultGroupName)
		Expect(err).ShouldNot(HaveOccurred())
		for _, result := range results {
			Expect(result).Should(Equal(codeRepoAndProject))
		}
	}))

	It("does not conform to the template layout", testUseCase.ListResourceNotMatch(fakeNodes, func(codeRepo *MockCodeRepo, secretRepo *MockSecretrepo, resourceUseCase *ResourcesUsecase, nodestree *nodestree.MockNodesTree, gitRepo *MockGitRepo, client *kubernetes.MockClient) {
		biz := NewCodeRepoUsecase(logger, codeRepo, secretRepo, nodestree, nautesConfigs, resourceUseCase, nil)
		_, err := biz.ListCodeRepos(ctx, defaultGroupName)
		Expect(err).Should(HaveOccurred())
	}))
})

var _ = Describe("Save codeRepo", func() {
	var (
		resourceName           = "toSaveCodeRepo"
		fakeResource           = createFakeCodeRepoResource(resourceName)
		fakeNode               = createFakeCodeRepoNode(fakeResource)
		fakeNodes              = createFakeCcontainingCodeRepoNodes(fakeNode)
		toSaveProject          = &Project{Id: 1222, HttpUrlToRepo: "https://gitlab.com/nautes-labs/test.git"}
		toSaveProjectDeployKey = &ProjectDeployKey{
			ID:  2013,
			Key: "FingerprintData",
		}
		extendKVs         = make(map[string]string)
		toGetCodeRepoPath = fmt.Sprintf("%s/%s", defaultProductGroup.Path, resourceName)
		repoName          = fmt.Sprintf("repo-%d", toSaveProject.Id)
		data              = &CodeRepoData{
			Name: resourceName,
			Spec: resourcev1alpha1.CodeRepoSpec{
				CodeRepoProvider:  "provider",
				Product:           "product",
				Project:           _DefaultProjectResourceName,
				RepoName:          repoName,
				DeploymentRuntime: true,
				Webhook: &resourcev1alpha1.Webhook{
					Events: []string{"push_events"},
				},
			},
		}
		options = &GitCodeRepoOptions{
			Gitlab: &GitlabCodeRepoOptions{
				Name:        repoName,
				Path:        repoName,
				NamespaceID: defaultProductGroup.Id,
			},
		}
		bizOptions = &BizOptions{
			ResouceName: resourceName,
			ProductName: defaultGroupName,
		}
		listDeployKeys = []*ProjectDeployKey{
			{
				ID:  2013,
				Key: "Key1",
			},
			{
				ID:  2014,
				Key: "Key2",
			},
		}
	)

	BeforeEach(func() {
		extendKVs["fingerprint"] = toSaveProjectDeployKey.Key
		extendKVs["id"] = strconv.Itoa(toSaveProjectDeployKey.ID)
	})

	It("failed to get product info", testUseCase.GetProductFail(func(codeRepo *MockCodeRepo, secretRepo *MockSecretrepo, resourceUseCase *ResourcesUsecase, nodestree *nodestree.MockNodesTree, gitRepo *MockGitRepo, client *kubernetes.MockClient) {
		biz := NewCodeRepoUsecase(logger, codeRepo, secretRepo, nodestree, nautesConfigs, resourceUseCase, nil)
		err := biz.SaveCodeRepo(context.Background(), bizOptions, data, options)
		Expect(err).Should(HaveOccurred())
	}))

	It("failed to get default project info", testUseCase.GetDefaultProjectFail(func(codeRepo *MockCodeRepo, secretRepo *MockSecretrepo, resourceUseCase *ResourcesUsecase, nodestree *nodestree.MockNodesTree, gitRepo *MockGitRepo, client *kubernetes.MockClient) {
		codeRepo.EXPECT().GetCodeRepo(gomock.Any(), gomock.Eq(toGetCodeRepoPath)).Return(toSaveProject, nil)
		codeRepo.EXPECT().UpdateCodeRepo(gomock.Any(), gomock.Eq(int(toSaveProject.Id)), options).Return(toSaveProject, nil)

		biz := NewCodeRepoUsecase(logger, codeRepo, secretRepo, nodestree, nautesConfigs, resourceUseCase, nil)
		err := biz.SaveCodeRepo(context.Background(), bizOptions, data, options)
		Expect(err).Should(HaveOccurred())
	}))

	It("will created successfully", testUseCase.CreateResourceSuccess(fakeNodes, func(codeRepo *MockCodeRepo, secretRepo *MockSecretrepo, resourceUseCase *ResourcesUsecase, nodestree *nodestree.MockNodesTree, gitRepo *MockGitRepo, client *kubernetes.MockClient) {
		codeRepo.EXPECT().GetCodeRepo(gomock.Any(), gomock.Eq(toGetCodeRepoPath)).Return(nil, ErrorProjectNotFound)
		codeRepo.EXPECT().CreateCodeRepo(gomock.Any(), gomock.Eq(int(defaultProductGroup.Id)), options).Return(toSaveProject, nil)
		codeRepo.EXPECT().SaveDeployKey(gomock.Any(), gomock.Any(), gomock.Any()).Return(toSaveProjectDeployKey, nil)
		codeRepo.EXPECT().ListDeployKeys(gomock.Any(), int(toSaveProject.Id), gomock.Any()).Return(listDeployKeys, nil)
		codeRepo.EXPECT().DeleteDeployKey(gomock.Any(), int(toSaveProject.Id), int(2014)).Return(nil)

		secretRepo.EXPECT().GetDeployKey(gomock.Any(), gomock.Any()).Return(nil, commonv1.ErrorSecretNotFound("secret data is not found"))
		secretRepo.EXPECT().SaveDeployKey(gomock.Any(), gomock.Eq(int(toSaveProject.Id)), gomock.Any(), extendKVs).Return(nil)
		client.EXPECT().List(context.Background(), gomock.Any(), gomock.Any()).Return(nil)

		biz := NewCodeRepoUsecase(logger, codeRepo, secretRepo, nodestree, nautesConfigs, resourceUseCase, client)
		err := biz.SaveCodeRepo(context.Background(), bizOptions, data, options)
		Expect(err).ShouldNot(HaveOccurred())
	}))

	It("will updated successfully", testUseCase.UpdateResoureSuccess(fakeNodes, fakeNode, func(codeRepo *MockCodeRepo, secretRepo *MockSecretrepo, resourceUseCase *ResourcesUsecase, nodestree *nodestree.MockNodesTree, gitRepo *MockGitRepo, client *kubernetes.MockClient) {
		codeRepo.EXPECT().GetCodeRepo(gomock.Any(), gomock.Eq(toGetCodeRepoPath)).Return(toSaveProject, nil)
		codeRepo.EXPECT().UpdateCodeRepo(gomock.Any(), gomock.Eq(int(toSaveProject.Id)), options).Return(toSaveProject, nil)
		codeRepo.EXPECT().SaveDeployKey(gomock.Any(), gomock.Any(), gomock.Any()).Return(toSaveProjectDeployKey, nil)
		codeRepo.EXPECT().ListDeployKeys(gomock.Any(), int(toSaveProject.Id), gomock.Any()).Return(listDeployKeys, nil)
		codeRepo.EXPECT().DeleteDeployKey(gomock.Any(), int(toSaveProject.Id), int(2014)).Return(nil)

		secretRepo.EXPECT().GetDeployKey(gomock.Any(), gomock.Any()).Return(nil, commonv1.ErrorSecretNotFound("secret data is not found"))
		secretRepo.EXPECT().SaveDeployKey(gomock.Any(), gomock.Eq(int(toSaveProject.Id)), gomock.Any(), extendKVs).Return(nil)
		client.EXPECT().List(context.Background(), gomock.Any(), gomock.Any()).Return(nil)

		biz := NewCodeRepoUsecase(logger, codeRepo, secretRepo, nodestree, nautesConfigs, resourceUseCase, client)
		err := biz.SaveCodeRepo(context.Background(), bizOptions, data, options)
		Expect(err).ShouldNot(HaveOccurred())
	}))

	It("auto merge conflict, updated successfully", testUseCase.UpdateResourceAndAutoMerge(fakeNodes, fakeNode, func(codeRepo *MockCodeRepo, secretRepo *MockSecretrepo, resourceUseCase *ResourcesUsecase, nodestree *nodestree.MockNodesTree, gitRepo *MockGitRepo, client *kubernetes.MockClient) {
		codeRepo.EXPECT().GetCodeRepo(gomock.Any(), gomock.Eq(toGetCodeRepoPath)).Return(toSaveProject, nil)
		codeRepo.EXPECT().UpdateCodeRepo(gomock.Any(), gomock.Eq(int(toSaveProject.Id)), options).Return(toSaveProject, nil)
		codeRepo.EXPECT().SaveDeployKey(gomock.Any(), gomock.Any(), gomock.Any()).Return(toSaveProjectDeployKey, nil)
		codeRepo.EXPECT().ListDeployKeys(gomock.Any(), int(toSaveProject.Id), gomock.Any()).Return(listDeployKeys, nil)
		codeRepo.EXPECT().DeleteDeployKey(gomock.Any(), int(toSaveProject.Id), int(2014)).Return(nil)

		secretRepo.EXPECT().GetDeployKey(gomock.Any(), gomock.Any()).Return(nil, commonv1.ErrorSecretNotFound("secret data is not found"))
		secretRepo.EXPECT().SaveDeployKey(gomock.Any(), gomock.Eq(int(toSaveProject.Id)), gomock.Any(), extendKVs).Return(nil)
		client.EXPECT().List(context.Background(), gomock.Any(), gomock.Any()).Return(nil)

		biz := NewCodeRepoUsecase(logger, codeRepo, secretRepo, nodestree, nautesConfigs, resourceUseCase, client)
		err := biz.SaveCodeRepo(context.Background(), bizOptions, data, options)
		Expect(err).ShouldNot(HaveOccurred())
	}))

	It("failed to auto merge conflict", testUseCase.MergeConflictFail(fakeNodes, fakeNode, func(codeRepo *MockCodeRepo, secretRepo *MockSecretrepo, resourceUseCase *ResourcesUsecase, nodestree *nodestree.MockNodesTree, gitRepo *MockGitRepo, client *kubernetes.MockClient) {
		codeRepo.EXPECT().GetCodeRepo(gomock.Any(), gomock.Eq(toGetCodeRepoPath)).Return(nil, ErrorProjectNotFound)
		codeRepo.EXPECT().CreateCodeRepo(gomock.Any(), gomock.Eq(int(defaultProductGroup.Id)), options).Return(toSaveProject, nil)
		client.EXPECT().List(context.Background(), gomock.Any(), gomock.Any()).Return(nil)

		biz := NewCodeRepoUsecase(logger, codeRepo, secretRepo, nodestree, nautesConfigs, resourceUseCase, client)
		err := biz.SaveCodeRepo(context.Background(), bizOptions, data, options)
		Expect(err).Should(HaveOccurred())
	}))

	It("failed to push code retry three times", testUseCase.CreateResourceAndAutoRetry(fakeNodes, func(codeRepo *MockCodeRepo, secretRepo *MockSecretrepo, resourceUseCase *ResourcesUsecase, nodestree *nodestree.MockNodesTree, gitRepo *MockGitRepo, client *kubernetes.MockClient) {
		codeRepo.EXPECT().GetCodeRepo(gomock.Any(), gomock.Eq(toGetCodeRepoPath)).Return(nil, ErrorProjectNotFound)
		codeRepo.EXPECT().CreateCodeRepo(gomock.Any(), gomock.Eq(int(defaultProductGroup.Id)), options).Return(toSaveProject, nil)

		client.EXPECT().List(context.Background(), gomock.Any(), gomock.Any()).Return(nil)

		biz := NewCodeRepoUsecase(logger, codeRepo, secretRepo, nodestree, nautesConfigs, resourceUseCase, client)
		err := biz.SaveCodeRepo(context.Background(), bizOptions, data, options)
		Expect(err).Should(HaveOccurred())
	}))

	It("modify resource but non compliant layout", testUseCase.UpdateResourceButNotConformTemplate(fakeNodes, fakeNode, func(codeRepo *MockCodeRepo, secretRepo *MockSecretrepo, resourceUseCase *ResourcesUsecase, nodestree *nodestree.MockNodesTree, gitRepo *MockGitRepo, client *kubernetes.MockClient) {
		codeRepo.EXPECT().GetCodeRepo(gomock.Any(), gomock.Eq(toGetCodeRepoPath)).Return(toSaveProject, nil)
		codeRepo.EXPECT().UpdateCodeRepo(gomock.Any(), gomock.Eq(int(toSaveProject.Id)), options).Return(toSaveProject, nil)
		client.EXPECT().List(context.Background(), gomock.Any(), gomock.Any()).Return(nil)

		biz := NewCodeRepoUsecase(logger, codeRepo, secretRepo, nodestree, nautesConfigs, resourceUseCase, client)
		err := biz.SaveCodeRepo(context.Background(), bizOptions, data, options)
		Expect(err).Should(HaveOccurred())
	}))

	It("failed to save config", testUseCase.SaveConfigFail(fakeNodes, fakeNode, func(codeRepo *MockCodeRepo, secretRepo *MockSecretrepo, resourceUseCase *ResourcesUsecase, nodestree *nodestree.MockNodesTree, gitRepo *MockGitRepo, client *kubernetes.MockClient) {
		codeRepo.EXPECT().GetCodeRepo(gomock.Any(), gomock.Eq(toGetCodeRepoPath)).Return(toSaveProject, nil)
		codeRepo.EXPECT().UpdateCodeRepo(gomock.Any(), gomock.Eq(int(toSaveProject.Id)), options).Return(toSaveProject, nil)
		client.EXPECT().List(context.Background(), gomock.Any(), gomock.Any()).Return(nil)

		biz := NewCodeRepoUsecase(logger, codeRepo, secretRepo, nodestree, nautesConfigs, resourceUseCase, client)
		err := biz.SaveCodeRepo(context.Background(), bizOptions, data, options)
		Expect(err).Should(HaveOccurred())
	}))

	Describe("check reference by resources", func() {
		It("incorrect product name", testUseCase.CheckReferenceButIncorrectProduct(fakeNodes, func(options nodestree.CompareOptions, nodestree *nodestree.MockNodesTree) {
			biz := NewCodeRepoUsecase(logger, nil, nil, nodestree, nautesConfigs, nil, nil)
			ok, err := biz.CheckReference(options, fakeNode, nil)
			Expect(err).Should(HaveOccurred())
			Expect(ok).To(BeTrue())
		}))

		It("webhook matching failed ", func() {
			options := nodestree.CompareOptions{
				Nodes:       fakeNodes,
				ProductName: defaultProductId,
			}
			nodestree := nodestree.NewMockNodesTree(ctl)
			nodestree.EXPECT().AppendOperators(gomock.Any())
			newResouce := fakeResource.DeepCopy()
			newResouce.Spec.Webhook.Events = append(newResouce.Spec.Webhook.Events, "errorWebhook")
			node := createFakeCodeRepoNode(newResouce)

			biz := NewCodeRepoUsecase(logger, nil, nil, nodestree, nautesConfigs, nil, nil)
			ok, err := biz.CheckReference(options, node, nil)
			Expect(err).Should(HaveOccurred())
			Expect(ok).To(BeTrue())
		})

		It("project reference not found", func() {
			options := nodestree.CompareOptions{
				Nodes:       fakeNodes,
				ProductName: defaultProductId,
			}
			nodestree := nodestree.NewMockNodesTree(ctl)
			nodestree.EXPECT().AppendOperators(gomock.Any())

			biz := NewCodeRepoUsecase(logger, nil, nil, nodestree, nautesConfigs, nil, nil)
			ok, err := biz.CheckReference(options, fakeNode, nil)
			Expect(err).Should(HaveOccurred())
			Expect(ok).To(BeTrue())
		})

		It("codeRepo provider reference not found", func() {
			options := nodestree.CompareOptions{
				Nodes:       fakeNodes,
				ProductName: defaultProductId,
			}
			nodestree := nodestree.NewMockNodesTree(ctl)
			nodestree.EXPECT().AppendOperators(gomock.Any())

			objKey := client.ObjectKey{
				Namespace: nautesConfigs.Nautes.Namespace,
				Name:      fakeResource.Spec.CodeRepoProvider,
			}

			client := kubernetes.NewMockClient(ctl)
			client.EXPECT().Get(gomock.Any(), objKey, &resourcev1alpha1.CodeRepoProvider{}).Return(ErrorResourceNoFound)

			projectName := fakeResource.Spec.Project
			projectNodes := createProjectNodes(createProjectNode(createProjectResource(projectName)))
			options.Nodes.Children = append(options.Nodes.Children, projectNodes.Children...)

			biz := NewCodeRepoUsecase(logger, nil, nil, nodestree, nautesConfigs, nil, nil)
			ok, err := biz.CheckReference(options, fakeNode, client)
			Expect(err).Should(HaveOccurred())
			Expect(ok).To(BeTrue())
		})

		It("will successed", func() {
			options := nodestree.CompareOptions{
				Nodes:       fakeNodes,
				ProductName: defaultProductId,
			}
			nodestree := nodestree.NewMockNodesTree(ctl)
			nodestree.EXPECT().AppendOperators(gomock.Any())

			objKey := client.ObjectKey{
				Namespace: nautesConfigs.Nautes.Namespace,
				Name:      fakeResource.Spec.CodeRepoProvider,
			}

			client := kubernetes.NewMockClient(ctl)
			client.EXPECT().Get(gomock.Any(), objKey, &resourcev1alpha1.CodeRepoProvider{}).Return(nil)

			projectName := fakeResource.Spec.Project
			projectNodes := createProjectNodes(createProjectNode(createProjectResource(projectName)))
			options.Nodes.Children = append(options.Nodes.Children, projectNodes.Children...)

			biz := NewCodeRepoUsecase(logger, nil, nil, nodestree, nautesConfigs, nil, nil)
			ok, err := biz.CheckReference(options, fakeNode, client)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(ok).To(BeTrue())
		})
	})
})

var _ = Describe("Delete codeRepo", func() {
	var (
		resourceName      = "toDeleteCodeRepo"
		fakeResource      = createFakeCodeRepoResource(resourceName)
		fakeNode          = createFakeCodeRepoNode(fakeResource)
		fakeNodes         = createFakeCcontainingCodeRepoNodes(fakeNode)
		deletedProject    = &Project{Id: 1222}
		toGetCodeRepoPath = fmt.Sprintf("%s/%s", defaultProductGroup.Path, resourceName)
		bizOptions        = &BizOptions{
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
		codeRepo.EXPECT().GetCodeRepo(gomock.Any(), gomock.Eq(toGetCodeRepoPath)).Return(deletedProject, nil)
		codeRepo.EXPECT().DeleteCodeRepo(gomock.Any(), gomock.Eq(int(deletedProject.Id))).Return(nil)
		secretRepo.EXPECT().DeleteSecret(gomock.Any(), gomock.Eq(int(deletedProject.Id))).Return(nil)

		biz := NewCodeRepoUsecase(logger, codeRepo, secretRepo, nodestree, nautesConfigs, resourceUseCase, nil)
		err := biz.DeleteCodeRepo(context.Background(), bizOptions)
		Expect(err).ShouldNot(HaveOccurred())
	}))

	It("modify resource but non compliant layout standards", testUseCase.DeleteResourceErrorLayout(fakeNodes, fakeNode, func(codeRepo *MockCodeRepo, secretRepo *MockSecretrepo, resourceUseCase *ResourcesUsecase, nodestree *nodestree.MockNodesTree, gitRepo *MockGitRepo, client *kubernetes.MockClient) {
		codeRepo.EXPECT().GetCodeRepo(gomock.Any(), gomock.Eq(toGetCodeRepoPath)).Return(deletedProject, nil)

		biz := NewCodeRepoUsecase(logger, codeRepo, secretRepo, nodestree, nautesConfigs, resourceUseCase, nil)
		err := biz.DeleteCodeRepo(context.Background(), bizOptions)
		Expect(err).Should(HaveOccurred())
	}))
})
