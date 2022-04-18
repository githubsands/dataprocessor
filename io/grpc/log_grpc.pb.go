// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.19.4
// source: log.proto

//go:build ignore

package grpc

import (
	context "context"

	grpc "google.golang.org/grpc"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// ProcessorClient is the client API for Processor service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ProcessorClient interface {
	// Sends a greeting
	ReceiveLog(ctx context.Context, opts ...grpc.CallOption) (Processor_ReceiveLogClient, error)
}

type processorClient struct {
	cc grpc.ClientConnInterface
}

func NewProcessorClient(cc grpc.ClientConnInterface) ProcessorClient {
	return &processorClient{cc}
}

func (c *processorClient) ReceiveLog(ctx context.Context, opts ...grpc.CallOption) (Processor_ReceiveLogClient, error) {
	stream, err := c.cc.NewStream(ctx, &Processor_ServiceDesc.Streams[0], "/Processor/ReceiveLog", opts...)
	if err != nil {
		return nil, err
	}
	x := &processorReceiveLogClient{stream}
	return x, nil
}

type Processor_ReceiveLogClient interface {
	Send(*LogRequest) error
	Recv() (*LogReply, error)
	grpc.ClientStream
}

type processorReceiveLogClient struct {
	grpc.ClientStream
}

func (x *processorReceiveLogClient) Send(m *LogRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *processorReceiveLogClient) Recv() (*LogReply, error) {
	m := new(LogReply)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// ProcessorServer is the server API for Processor service.
// All implementations must embed UnimplementedProcessorServer
// for forward compatibility
type ProcessorServer interface {
	ReceiveLog(Processor_ReceiveLogServer) error
	mustEmbedUnimplementedProcessorServer()
}

func RegisterProcessorServer(s grpc.ServiceRegistrar, srv ProcessorServer) {
	s.RegisterService(&Processor_ServiceDesc, srv)
}

func _Processor_ReceiveLog_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(ProcessorServer).ReceiveLog(&processorReceiveLogServer{stream})
}

type Processor_ReceiveLogServer interface {
	Send(*LogReply) error
	Recv() (*LogRequest, error)
	grpc.ServerStream
}

type processorReceiveLogServer struct {
	grpc.ServerStream
}

func (x *processorReceiveLogServer) Send(m *LogReply) error {
	return x.ServerStream.SendMsg(m)
}

func (x *processorReceiveLogServer) Recv() (*LogRequest, error) {
	m := new(LogRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Processor_ServiceDesc is the grpc.ServiceDesc for Processor service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Processor_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "Processor",
	HandlerType: (*ProcessorServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "ReceiveLog",
			Handler:       _Processor_ReceiveLog_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "log.proto",
}