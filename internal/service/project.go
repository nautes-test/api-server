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

package service

import (
	"context"
	"fmt"

	projectv1 "github.com/nautes-labs/api-server/api/project/v1"
	"github.com/nautes-labs/api-server/internal/biz"
	resourcev1alpha1 "github.com/nautes-labs/pkg/api/v1alpha1"
)

type ProjectService struct {
	projectv1.UnimplementedProjectServer
	project *biz.ProjectUsecase
}

func NewProjectService(project *biz.ProjectUsecase) *ProjectService {
	return &ProjectService{
		project: project,
	}
}

func (s *ProjectService) CovertCodeRepoValueToReply(project *resourcev1alpha1.Project) *projectv1.GetReply {
	return &projectv1.GetReply{
		Product:  project.Spec.Product,
		Name:     project.Name,
		Language: project.Spec.Language,
	}
}

func (s *ProjectService) GetProject(ctx context.Context, req *projectv1.GetRequest) (*projectv1.GetReply, error) {
	project, err := s.project.GetProject(ctx, req.ProjectName, req.ProductName)
	if err != nil {
		return nil, err
	}

	return s.CovertCodeRepoValueToReply(project), nil
}

func (s *ProjectService) ListProjects(ctx context.Context, req *projectv1.ListsRequest) (*projectv1.ListsReply, error) {
	projects, err := s.project.ListProjects(ctx, req.ProductName)
	if err != nil {
		return nil, err
	}

	var items []*projectv1.GetReply
	for _, project := range projects {
		items = append(items, s.CovertCodeRepoValueToReply(project))
	}

	return &projectv1.ListsReply{Items: items}, nil
}

func (s *ProjectService) SaveProject(ctx context.Context, req *projectv1.SaveRequest) (*projectv1.SaveReply, error) {
	project := &biz.ProjectData{
		ProjectName: req.ProjectName,
		Language:    req.Body.Language,
	}
	options := &biz.BizOptions{
		ResouceName:       req.ProjectName,
		ProductName:       req.ProductName,
		InsecureSkipCheck: req.InsecureSkipCheck,
	}
	err := s.project.SaveProject(ctx, options, project)
	if err != nil {
		return nil, err
	}

	return &projectv1.SaveReply{
		Msg: fmt.Sprintf("Successfully saved %v configuration", project.ProjectName),
	}, nil
}

func (s *ProjectService) DeleteProject(ctx context.Context, req *projectv1.DeleteRequest) (*projectv1.DeleteReply, error) {
	options := &biz.BizOptions{
		ResouceName:       req.ProjectName,
		ProductName:       req.ProductName,
		InsecureSkipCheck: req.InsecureSkipCheck,
	}
	err := s.project.DeleteProject(ctx, options)
	if err != nil {
		return nil, err
	}

	return &projectv1.DeleteReply{
		Msg: fmt.Sprintf("Successfully deleted %v configuration", req.ProjectName),
	}, nil
}
