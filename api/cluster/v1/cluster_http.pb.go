// Code generated by protoc-gen-go-http. DO NOT EDIT.
// versions:
// - protoc-gen-go-http v2.5.3
// - protoc             v3.6.1
// source: cluster/v1/cluster.proto

package v1

import (
	context "context"
	http "github.com/go-kratos/kratos/v2/transport/http"
	binding "github.com/go-kratos/kratos/v2/transport/http/binding"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the kratos package it is being compiled against.
var _ = new(context.Context)
var _ = binding.EncodeURL

const _ = http.SupportPackageIsVersion1

const OperationClusterDeleteCluster = "/api.cluster.v1.Cluster/DeleteCluster"
const OperationClusterSaveCluster = "/api.cluster.v1.Cluster/SaveCluster"

type ClusterHTTPServer interface {
	DeleteCluster(context.Context, *DeleteRequest) (*DeleteReply, error)
	SaveCluster(context.Context, *SaveRequest) (*SaveReply, error)
}

func RegisterClusterHTTPServer(s *http.Server, srv ClusterHTTPServer) {
	r := s.Route("/")
	r.POST("/api/v1/clusters/{clusterName}", _Cluster_SaveCluster0_HTTP_Handler(srv))
	r.DELETE("/api/v1/clusters/{clusterName}", _Cluster_DeleteCluster0_HTTP_Handler(srv))
}

func _Cluster_SaveCluster0_HTTP_Handler(srv ClusterHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in SaveRequest
		if err := ctx.Bind(&in.Body); err != nil {
			return err
		}
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		if err := ctx.BindVars(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, OperationClusterSaveCluster)
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.SaveCluster(ctx, req.(*SaveRequest))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		reply := out.(*SaveReply)
		return ctx.Result(200, reply)
	}
}

func _Cluster_DeleteCluster0_HTTP_Handler(srv ClusterHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in DeleteRequest
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		if err := ctx.BindVars(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, OperationClusterDeleteCluster)
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.DeleteCluster(ctx, req.(*DeleteRequest))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		reply := out.(*DeleteReply)
		return ctx.Result(200, reply)
	}
}

type ClusterHTTPClient interface {
	DeleteCluster(ctx context.Context, req *DeleteRequest, opts ...http.CallOption) (rsp *DeleteReply, err error)
	SaveCluster(ctx context.Context, req *SaveRequest, opts ...http.CallOption) (rsp *SaveReply, err error)
}

type ClusterHTTPClientImpl struct {
	cc *http.Client
}

func NewClusterHTTPClient(client *http.Client) ClusterHTTPClient {
	return &ClusterHTTPClientImpl{client}
}

func (c *ClusterHTTPClientImpl) DeleteCluster(ctx context.Context, in *DeleteRequest, opts ...http.CallOption) (*DeleteReply, error) {
	var out DeleteReply
	pattern := "/api/v1/clusters/{clusterName}"
	path := binding.EncodeURL(pattern, in, true)
	opts = append(opts, http.Operation(OperationClusterDeleteCluster))
	opts = append(opts, http.PathTemplate(pattern))
	err := c.cc.Invoke(ctx, "DELETE", path, nil, &out, opts...)
	if err != nil {
		return nil, err
	}
	return &out, err
}

func (c *ClusterHTTPClientImpl) SaveCluster(ctx context.Context, in *SaveRequest, opts ...http.CallOption) (*SaveReply, error) {
	var out SaveReply
	pattern := "/api/v1/clusters/{clusterName}"
	path := binding.EncodeURL(pattern, in, false)
	opts = append(opts, http.Operation(OperationClusterSaveCluster))
	opts = append(opts, http.PathTemplate(pattern))
	err := c.cc.Invoke(ctx, "POST", path, in.Body, &out, opts...)
	if err != nil {
		return nil, err
	}
	return &out, err
}
