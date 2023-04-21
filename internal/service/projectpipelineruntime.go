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

	projectpipelineruntimev1 "github.com/nautes-labs/api-server/api/projectpipelineruntime/v1"
	"github.com/nautes-labs/api-server/internal/biz"
	resourcev1alpha1 "github.com/nautes-labs/pkg/api/v1alpha1"
)

type ProjectPipelineRuntimeService struct {
	projectpipelineruntimev1.UnimplementedProjectPipelineRuntimeServer
	projectPipelineRuntime *biz.ProjectPipelineRuntimeUsecase
}

func NewProjectPipelineRuntimeService(projectPipelineRuntime *biz.ProjectPipelineRuntimeUsecase) *ProjectPipelineRuntimeService {
	return &ProjectPipelineRuntimeService{projectPipelineRuntime: projectPipelineRuntime}
}

func (s *ProjectPipelineRuntimeService) CovertCodeRepoValueToReply(projectPipelineRuntime *resourcev1alpha1.ProjectPipelineRuntime, productName string) *projectpipelineruntimev1.GetReply {
	var pipelines []*projectpipelineruntimev1.Pipeline
	var eventSources []*projectpipelineruntimev1.EventSource
	for _, pipeline := range projectPipelineRuntime.Spec.Pipelines {
		for _, eventSource := range pipeline.EventSources {
			eventSources = append(eventSources, &projectpipelineruntimev1.EventSource{
				Webhook: eventSource.Webhook,
				Calendar: &projectpipelineruntimev1.CalendarEventSource{
					Schedule:       eventSource.Calendar.Schedule,
					Interval:       eventSource.Calendar.Interval,
					ExclusionDates: eventSource.Calendar.ExclusionDates,
					Timezone:       eventSource.Calendar.Timezone,
				},
			})
		}
		pipelines = append(pipelines, &projectpipelineruntimev1.Pipeline{
			Name:   pipeline.Name,
			Branch: pipeline.Branch,
			Path:   pipeline.Path,
		})
	}
	return &projectpipelineruntimev1.GetReply{
		Name:           projectPipelineRuntime.Name,
		Project:        projectPipelineRuntime.Spec.Project,
		PipelineSource: projectPipelineRuntime.Spec.PipelineSource,
		CodeSources:    projectPipelineRuntime.Spec.CodeSources,
		Destination:    projectPipelineRuntime.Spec.Destination,
		Pipelines:      pipelines,
	}
}

func (s *ProjectPipelineRuntimeService) GetProjectPipelineRuntime(ctx context.Context, req *projectpipelineruntimev1.GetRequest) (*projectpipelineruntimev1.GetReply, error) {
	runtime, err := s.projectPipelineRuntime.GetProjectPipelineRuntime(ctx, req.ProjectPipelineRuntimeName, req.ProductName)
	if err != nil {
		return nil, err
	}

	return s.CovertCodeRepoValueToReply(runtime, req.ProductName), nil
}

func (s *ProjectPipelineRuntimeService) ListProjectPipelineRuntimes(ctx context.Context, req *projectpipelineruntimev1.ListsRequest) (*projectpipelineruntimev1.ListsReply, error) {
	runtimes, err := s.projectPipelineRuntime.ListProjectPipelineRuntimes(ctx, req.ProductName)
	if err != nil {
		return nil, err
	}

	var items []*projectpipelineruntimev1.GetReply
	for _, runtime := range runtimes {
		items = append(items, s.CovertCodeRepoValueToReply(runtime, req.ProductName))
	}

	return &projectpipelineruntimev1.ListsReply{
		Items: items,
	}, nil
}

func (s *ProjectPipelineRuntimeService) SaveProjectPipelineRuntime(ctx context.Context, req *projectpipelineruntimev1.SaveRequest) (*projectpipelineruntimev1.SaveReply, error) {
	data := &biz.ProjectPipelineRuntimeData{
		Name: req.ProjectPipelineRuntimeName,
		Spec: resourcev1alpha1.ProjectPipelineRuntimeSpec{
			Project:        req.Body.Project,
			PipelineSource: req.Body.PipelineSource,
			CodeSources:    req.Body.CodeSources,
			Destination:    req.Body.Destination,
			Pipelines:      s.getResourcePipelines(req.Body.Pipelines),
		},
	}

	options := &biz.BizOptions{
		ResouceName:       req.ProjectPipelineRuntimeName,
		ProductName:       req.ProductName,
		InsecureSkipCheck: req.InsecureSkipCheck,
	}
	err := s.projectPipelineRuntime.SaveProjectPipelineRuntime(ctx, options, data)
	if err != nil {
		return nil, err
	}

	return &projectpipelineruntimev1.SaveReply{
		Msg: fmt.Sprintf("Successfully saved %s configuration", req.ProjectPipelineRuntimeName),
	}, nil
}

func (s *ProjectPipelineRuntimeService) getResourcePipelines(pipelines []*projectpipelineruntimev1.Pipeline) []resourcev1alpha1.Pipeline {
	resourcePipelines := []resourcev1alpha1.Pipeline{}
	for _, pipeline := range pipelines {
		resourcePipeline := resourcev1alpha1.Pipeline{
			Name:   pipeline.Name,
			Branch: pipeline.Branch,
			Path:   pipeline.Path,
		}

		for _, e := range pipeline.EventSources {
			resourcePipeline.EventSources = append(resourcePipeline.EventSources, resourcev1alpha1.EventSource{
				Webhook: e.Webhook,
			})
		}

		resourcePipelines = append(resourcePipelines, resourcePipeline)

	}

	return resourcePipelines
}

func (s *ProjectPipelineRuntimeService) DeleteProjectPipelineRuntime(ctx context.Context, req *projectpipelineruntimev1.DeleteRequest) (*projectpipelineruntimev1.DeleteReply, error) {
	options := &biz.BizOptions{
		ResouceName:       req.ProjectPipelineRuntimeName,
		ProductName:       req.ProductName,
		InsecureSkipCheck: req.InsecureSkipCheck,
	}
	err := s.projectPipelineRuntime.DeleteProjectPipelineRuntime(ctx, options)
	if err != nil {
		return nil, err
	}

	return &projectpipelineruntimev1.DeleteReply{
		Msg: fmt.Sprintf("Successfully deleted %s configuration", req.ProjectPipelineRuntimeName),
	}, nil
}
