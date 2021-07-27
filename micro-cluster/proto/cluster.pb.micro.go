// Code generated by protoc-gen-micro. DO NOT EDIT.
// source: cluster.proto

package cluster

import (
	fmt "fmt"
	proto "google.golang.org/protobuf/proto"
	math "math"
)

import (
	context "context"
	api "github.com/asim/go-micro/v3/api"
	client "github.com/asim/go-micro/v3/client"
	server "github.com/asim/go-micro/v3/server"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Reference imports to suppress errors if they are not otherwise used.
var _ api.Endpoint
var _ context.Context
var _ client.Option
var _ server.Option

// Api Endpoints for ClusterService service

func NewClusterServiceEndpoints() []*api.Endpoint {
	return []*api.Endpoint{}
}

// Client API for ClusterService service

type ClusterService interface {
	CreateCluster(ctx context.Context, in *ClusterCreateReqDTO, opts ...client.CallOption) (*ClusterCreateRespDTO, error)
	QueryCluster(ctx context.Context, in *ClusterQueryReqDTO, opts ...client.CallOption) (*ClusterQueryRespDTO, error)
	DeleteCluster(ctx context.Context, in *ClusterDeleteReqDTO, opts ...client.CallOption) (*ClusterDeleteRespDTO, error)
	DetailCluster(ctx context.Context, in *ClusterDetailReqDTO, opts ...client.CallOption) (*ClusterDetailRespDTO, error)
	ImportData(ctx context.Context, in *DataImportRequest, opts ...client.CallOption) (*DataImportResponse, error)
	ExportData(ctx context.Context, in *DataExportRequest, opts ...client.CallOption) (*DataExportResponse, error)
	DescribeDataTransport(ctx context.Context, in *DataTransportQueryRequest, opts ...client.CallOption) (*DataTransportQueryResponse, error)
}

type clusterService struct {
	c    client.Client
	name string
}

func NewClusterService(name string, c client.Client) ClusterService {
	return &clusterService{
		c:    c,
		name: name,
	}
}

func (c *clusterService) CreateCluster(ctx context.Context, in *ClusterCreateReqDTO, opts ...client.CallOption) (*ClusterCreateRespDTO, error) {
	req := c.c.NewRequest(c.name, "ClusterService.CreateCluster", in)
	out := new(ClusterCreateRespDTO)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clusterService) QueryCluster(ctx context.Context, in *ClusterQueryReqDTO, opts ...client.CallOption) (*ClusterQueryRespDTO, error) {
	req := c.c.NewRequest(c.name, "ClusterService.QueryCluster", in)
	out := new(ClusterQueryRespDTO)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clusterService) DeleteCluster(ctx context.Context, in *ClusterDeleteReqDTO, opts ...client.CallOption) (*ClusterDeleteRespDTO, error) {
	req := c.c.NewRequest(c.name, "ClusterService.DeleteCluster", in)
	out := new(ClusterDeleteRespDTO)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clusterService) DetailCluster(ctx context.Context, in *ClusterDetailReqDTO, opts ...client.CallOption) (*ClusterDetailRespDTO, error) {
	req := c.c.NewRequest(c.name, "ClusterService.DetailCluster", in)
	out := new(ClusterDetailRespDTO)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clusterService) ImportData(ctx context.Context, in *DataImportRequest, opts ...client.CallOption) (*DataImportResponse, error) {
	req := c.c.NewRequest(c.name, "ClusterService.ImportData", in)
	out := new(DataImportResponse)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clusterService) ExportData(ctx context.Context, in *DataExportRequest, opts ...client.CallOption) (*DataExportResponse, error) {
	req := c.c.NewRequest(c.name, "ClusterService.ExportData", in)
	out := new(DataExportResponse)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clusterService) DescribeDataTransport(ctx context.Context, in *DataTransportQueryRequest, opts ...client.CallOption) (*DataTransportQueryResponse, error) {
	req := c.c.NewRequest(c.name, "ClusterService.DescribeDataTransport", in)
	out := new(DataTransportQueryResponse)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for ClusterService service

type ClusterServiceHandler interface {
	CreateCluster(context.Context, *ClusterCreateReqDTO, *ClusterCreateRespDTO) error
	QueryCluster(context.Context, *ClusterQueryReqDTO, *ClusterQueryRespDTO) error
	DeleteCluster(context.Context, *ClusterDeleteReqDTO, *ClusterDeleteRespDTO) error
	DetailCluster(context.Context, *ClusterDetailReqDTO, *ClusterDetailRespDTO) error
	ImportData(context.Context, *DataImportRequest, *DataImportResponse) error
	ExportData(context.Context, *DataExportRequest, *DataExportResponse) error
	DescribeDataTransport(context.Context, *DataTransportQueryRequest, *DataTransportQueryResponse) error
}

func RegisterClusterServiceHandler(s server.Server, hdlr ClusterServiceHandler, opts ...server.HandlerOption) error {
	type clusterService interface {
		CreateCluster(ctx context.Context, in *ClusterCreateReqDTO, out *ClusterCreateRespDTO) error
		QueryCluster(ctx context.Context, in *ClusterQueryReqDTO, out *ClusterQueryRespDTO) error
		DeleteCluster(ctx context.Context, in *ClusterDeleteReqDTO, out *ClusterDeleteRespDTO) error
		DetailCluster(ctx context.Context, in *ClusterDetailReqDTO, out *ClusterDetailRespDTO) error
		ImportData(ctx context.Context, in *DataImportRequest, out *DataImportResponse) error
		ExportData(ctx context.Context, in *DataExportRequest, out *DataExportResponse) error
		DescribeDataTransport(ctx context.Context, in *DataTransportQueryRequest, out *DataTransportQueryResponse) error
	}
	type ClusterService struct {
		clusterService
	}
	h := &clusterServiceHandler{hdlr}
	return s.Handle(s.NewHandler(&ClusterService{h}, opts...))
}

type clusterServiceHandler struct {
	ClusterServiceHandler
}

func (h *clusterServiceHandler) CreateCluster(ctx context.Context, in *ClusterCreateReqDTO, out *ClusterCreateRespDTO) error {
	return h.ClusterServiceHandler.CreateCluster(ctx, in, out)
}

func (h *clusterServiceHandler) QueryCluster(ctx context.Context, in *ClusterQueryReqDTO, out *ClusterQueryRespDTO) error {
	return h.ClusterServiceHandler.QueryCluster(ctx, in, out)
}

func (h *clusterServiceHandler) DeleteCluster(ctx context.Context, in *ClusterDeleteReqDTO, out *ClusterDeleteRespDTO) error {
	return h.ClusterServiceHandler.DeleteCluster(ctx, in, out)
}

func (h *clusterServiceHandler) DetailCluster(ctx context.Context, in *ClusterDetailReqDTO, out *ClusterDetailRespDTO) error {
	return h.ClusterServiceHandler.DetailCluster(ctx, in, out)
}

func (h *clusterServiceHandler) ImportData(ctx context.Context, in *DataImportRequest, out *DataImportResponse) error {
	return h.ClusterServiceHandler.ImportData(ctx, in, out)
}

func (h *clusterServiceHandler) ExportData(ctx context.Context, in *DataExportRequest, out *DataExportResponse) error {
	return h.ClusterServiceHandler.ExportData(ctx, in, out)
}

func (h *clusterServiceHandler) DescribeDataTransport(ctx context.Context, in *DataTransportQueryRequest, out *DataTransportQueryResponse) error {
	return h.ClusterServiceHandler.DescribeDataTransport(ctx, in, out)
}
