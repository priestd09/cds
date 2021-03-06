// Code generated by protoc-gen-go.
// source: workflowqueue.proto
// DO NOT EDIT!

package grpc

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import github_com_ovh_cds_sdk "github.com/ovh/cds/sdk"
import github_com_ovh_cds_sdk1 "github.com/ovh/cds/sdk"
import google_protobuf1 "github.com/golang/protobuf/ptypes/empty"

import (
	context "golang.org/x/net/context"
	grpc1 "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc1.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc1.SupportPackageIsVersion4

// Client API for WorkflowQueue service

type WorkflowQueueClient interface {
	SendLog(ctx context.Context, opts ...grpc1.CallOption) (WorkflowQueue_SendLogClient, error)
	SendResult(ctx context.Context, in *github_com_ovh_cds_sdk1.Result, opts ...grpc1.CallOption) (*google_protobuf1.Empty, error)
}

type workflowQueueClient struct {
	cc *grpc1.ClientConn
}

func NewWorkflowQueueClient(cc *grpc1.ClientConn) WorkflowQueueClient {
	return &workflowQueueClient{cc}
}

func (c *workflowQueueClient) SendLog(ctx context.Context, opts ...grpc1.CallOption) (WorkflowQueue_SendLogClient, error) {
	stream, err := grpc1.NewClientStream(ctx, &_WorkflowQueue_serviceDesc.Streams[0], c.cc, "/grpc.WorkflowQueue/SendLog", opts...)
	if err != nil {
		return nil, err
	}
	x := &workflowQueueSendLogClient{stream}
	return x, nil
}

type WorkflowQueue_SendLogClient interface {
	Send(*github_com_ovh_cds_sdk.Log) error
	CloseAndRecv() (*google_protobuf1.Empty, error)
	grpc1.ClientStream
}

type workflowQueueSendLogClient struct {
	grpc1.ClientStream
}

func (x *workflowQueueSendLogClient) Send(m *github_com_ovh_cds_sdk.Log) error {
	return x.ClientStream.SendMsg(m)
}

func (x *workflowQueueSendLogClient) CloseAndRecv() (*google_protobuf1.Empty, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(google_protobuf1.Empty)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *workflowQueueClient) SendResult(ctx context.Context, in *github_com_ovh_cds_sdk1.Result, opts ...grpc1.CallOption) (*google_protobuf1.Empty, error) {
	out := new(google_protobuf1.Empty)
	err := grpc1.Invoke(ctx, "/grpc.WorkflowQueue/SendResult", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for WorkflowQueue service

type WorkflowQueueServer interface {
	SendLog(WorkflowQueue_SendLogServer) error
	SendResult(context.Context, *github_com_ovh_cds_sdk1.Result) (*google_protobuf1.Empty, error)
}

func RegisterWorkflowQueueServer(s *grpc1.Server, srv WorkflowQueueServer) {
	s.RegisterService(&_WorkflowQueue_serviceDesc, srv)
}

func _WorkflowQueue_SendLog_Handler(srv interface{}, stream grpc1.ServerStream) error {
	return srv.(WorkflowQueueServer).SendLog(&workflowQueueSendLogServer{stream})
}

type WorkflowQueue_SendLogServer interface {
	SendAndClose(*google_protobuf1.Empty) error
	Recv() (*github_com_ovh_cds_sdk.Log, error)
	grpc1.ServerStream
}

type workflowQueueSendLogServer struct {
	grpc1.ServerStream
}

func (x *workflowQueueSendLogServer) SendAndClose(m *google_protobuf1.Empty) error {
	return x.ServerStream.SendMsg(m)
}

func (x *workflowQueueSendLogServer) Recv() (*github_com_ovh_cds_sdk.Log, error) {
	m := new(github_com_ovh_cds_sdk.Log)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _WorkflowQueue_SendResult_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc1.UnaryServerInterceptor) (interface{}, error) {
	in := new(github_com_ovh_cds_sdk1.Result)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WorkflowQueueServer).SendResult(ctx, in)
	}
	info := &grpc1.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grpc.WorkflowQueue/SendResult",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WorkflowQueueServer).SendResult(ctx, req.(*github_com_ovh_cds_sdk1.Result))
	}
	return interceptor(ctx, in, info, handler)
}

var _WorkflowQueue_serviceDesc = grpc1.ServiceDesc{
	ServiceName: "grpc.WorkflowQueue",
	HandlerType: (*WorkflowQueueServer)(nil),
	Methods: []grpc1.MethodDesc{
		{
			MethodName: "SendResult",
			Handler:    _WorkflowQueue_SendResult_Handler,
		},
	},
	Streams: []grpc1.StreamDesc{
		{
			StreamName:    "SendLog",
			Handler:       _WorkflowQueue_SendLog_Handler,
			ClientStreams: true,
		},
	},
	Metadata: "workflowqueue.proto",
}

func init() { proto.RegisterFile("workflowqueue.proto", fileDescriptor1) }

var fileDescriptor1 = []byte{
	// 196 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x74, 0xce, 0xbf, 0xca, 0xc2, 0x30,
	0x14, 0x05, 0xf0, 0xaf, 0xf0, 0xa1, 0x10, 0x70, 0x89, 0xe0, 0xd0, 0x82, 0x88, 0x2e, 0x4e, 0x37,
	0xa0, 0x6f, 0x20, 0xe8, 0xd4, 0x45, 0x1d, 0xdc, 0x9b, 0xa4, 0xa9, 0x34, 0xf5, 0xd6, 0xfc, 0x69,
	0xf1, 0x59, 0x7c, 0x59, 0x69, 0x53, 0x71, 0xea, 0x78, 0xef, 0x39, 0xfc, 0x38, 0x64, 0xde, 0xa2,
	0x29, 0x73, 0x8d, 0xed, 0xd3, 0x4b, 0x2f, 0xa1, 0x36, 0xe8, 0x90, 0xfe, 0x2b, 0x53, 0xf3, 0x78,
	0xa5, 0xee, 0xae, 0xf0, 0x19, 0x70, 0xac, 0x18, 0x36, 0x05, 0xe3, 0xc2, 0x32, 0x2b, 0x4a, 0xa6,
	0x51, 0x85, 0x5e, 0xbc, 0x19, 0x69, 0x18, 0x69, 0xbd, 0x76, 0x43, 0x29, 0x51, 0x88, 0x4a, 0x4b,
	0xd6, 0x5f, 0x99, 0xcf, 0x99, 0xac, 0x6a, 0xf7, 0x0a, 0xe1, 0xee, 0x1d, 0x91, 0xd9, 0x6d, 0x58,
	0x70, 0xee, 0x16, 0xd0, 0x03, 0x99, 0x5e, 0xe5, 0x43, 0xa4, 0xa8, 0x68, 0x02, 0x3f, 0x1f, 0xb0,
	0x29, 0x80, 0x0b, 0x0b, 0x56, 0x94, 0x90, 0xa2, 0x8a, 0x17, 0x10, 0x5c, 0xf8, 0xba, 0x70, 0xec,
	0xdc, 0xf5, 0xdf, 0x36, 0xa2, 0x27, 0x42, 0x3a, 0xe3, 0xd2, 0xcf, 0xa0, 0xcb, 0x31, 0x26, 0xe4,
	0xe3, 0x52, 0x36, 0xe9, 0x3f, 0xfb, 0x4f, 0x00, 0x00, 0x00, 0xff, 0xff, 0xaa, 0x1f, 0xda, 0xec,
	0x25, 0x01, 0x00, 0x00,
}
