// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package rpc

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// MediaApiClient is the client API for MediaApi service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type MediaApiClient interface {
	PrepareSession(ctx context.Context, in *CreateParam, opts ...grpc.CallOption) (*Session, error)
	UpdateSession(ctx context.Context, in *UpdateParam, opts ...grpc.CallOption) (*Status, error)
	StartSession(ctx context.Context, in *StartParam, opts ...grpc.CallOption) (*Status, error)
	StopSession(ctx context.Context, in *StopParam, opts ...grpc.CallOption) (*Status, error)
	ExecuteAction(ctx context.Context, in *Action, opts ...grpc.CallOption) (*ActionResult, error)
	ExecuteActionWithNotify(ctx context.Context, in *Action, opts ...grpc.CallOption) (MediaApi_ExecuteActionWithNotifyClient, error)
	SystemChannel(ctx context.Context, opts ...grpc.CallOption) (MediaApi_SystemChannelClient, error)
}

type mediaApiClient struct {
	cc grpc.ClientConnInterface
}

func NewMediaApiClient(cc grpc.ClientConnInterface) MediaApiClient {
	return &mediaApiClient{cc}
}

func (c *mediaApiClient) PrepareSession(ctx context.Context, in *CreateParam, opts ...grpc.CallOption) (*Session, error) {
	out := new(Session)
	err := c.cc.Invoke(ctx, "/rpc.MediaApi/PrepareSession", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *mediaApiClient) UpdateSession(ctx context.Context, in *UpdateParam, opts ...grpc.CallOption) (*Status, error) {
	out := new(Status)
	err := c.cc.Invoke(ctx, "/rpc.MediaApi/UpdateSession", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *mediaApiClient) StartSession(ctx context.Context, in *StartParam, opts ...grpc.CallOption) (*Status, error) {
	out := new(Status)
	err := c.cc.Invoke(ctx, "/rpc.MediaApi/StartSession", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *mediaApiClient) StopSession(ctx context.Context, in *StopParam, opts ...grpc.CallOption) (*Status, error) {
	out := new(Status)
	err := c.cc.Invoke(ctx, "/rpc.MediaApi/StopSession", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *mediaApiClient) ExecuteAction(ctx context.Context, in *Action, opts ...grpc.CallOption) (*ActionResult, error) {
	out := new(ActionResult)
	err := c.cc.Invoke(ctx, "/rpc.MediaApi/ExecuteAction", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *mediaApiClient) ExecuteActionWithNotify(ctx context.Context, in *Action, opts ...grpc.CallOption) (MediaApi_ExecuteActionWithNotifyClient, error) {
	stream, err := c.cc.NewStream(ctx, &MediaApi_ServiceDesc.Streams[0], "/rpc.MediaApi/ExecuteActionWithNotify", opts...)
	if err != nil {
		return nil, err
	}
	x := &mediaApiExecuteActionWithNotifyClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type MediaApi_ExecuteActionWithNotifyClient interface {
	Recv() (*ActionEvent, error)
	grpc.ClientStream
}

type mediaApiExecuteActionWithNotifyClient struct {
	grpc.ClientStream
}

func (x *mediaApiExecuteActionWithNotifyClient) Recv() (*ActionEvent, error) {
	m := new(ActionEvent)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *mediaApiClient) SystemChannel(ctx context.Context, opts ...grpc.CallOption) (MediaApi_SystemChannelClient, error) {
	stream, err := c.cc.NewStream(ctx, &MediaApi_ServiceDesc.Streams[1], "/rpc.MediaApi/SystemChannel", opts...)
	if err != nil {
		return nil, err
	}
	x := &mediaApiSystemChannelClient{stream}
	return x, nil
}

type MediaApi_SystemChannelClient interface {
	Send(*SystemEvent) error
	Recv() (*SystemEvent, error)
	grpc.ClientStream
}

type mediaApiSystemChannelClient struct {
	grpc.ClientStream
}

func (x *mediaApiSystemChannelClient) Send(m *SystemEvent) error {
	return x.ClientStream.SendMsg(m)
}

func (x *mediaApiSystemChannelClient) Recv() (*SystemEvent, error) {
	m := new(SystemEvent)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// MediaApiServer is the server API for MediaApi service.
// All implementations must embed UnimplementedMediaApiServer
// for forward compatibility
type MediaApiServer interface {
	PrepareSession(context.Context, *CreateParam) (*Session, error)
	UpdateSession(context.Context, *UpdateParam) (*Status, error)
	StartSession(context.Context, *StartParam) (*Status, error)
	StopSession(context.Context, *StopParam) (*Status, error)
	ExecuteAction(context.Context, *Action) (*ActionResult, error)
	ExecuteActionWithNotify(*Action, MediaApi_ExecuteActionWithNotifyServer) error
	SystemChannel(MediaApi_SystemChannelServer) error
	mustEmbedUnimplementedMediaApiServer()
}

// UnimplementedMediaApiServer must be embedded to have forward compatible implementations.
type UnimplementedMediaApiServer struct {
}

func (UnimplementedMediaApiServer) PrepareSession(context.Context, *CreateParam) (*Session, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PrepareSession not implemented")
}
func (UnimplementedMediaApiServer) UpdateSession(context.Context, *UpdateParam) (*Status, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateSession not implemented")
}
func (UnimplementedMediaApiServer) StartSession(context.Context, *StartParam) (*Status, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StartSession not implemented")
}
func (UnimplementedMediaApiServer) StopSession(context.Context, *StopParam) (*Status, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StopSession not implemented")
}
func (UnimplementedMediaApiServer) ExecuteAction(context.Context, *Action) (*ActionResult, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ExecuteAction not implemented")
}
func (UnimplementedMediaApiServer) ExecuteActionWithNotify(*Action, MediaApi_ExecuteActionWithNotifyServer) error {
	return status.Errorf(codes.Unimplemented, "method ExecuteActionWithNotify not implemented")
}
func (UnimplementedMediaApiServer) SystemChannel(MediaApi_SystemChannelServer) error {
	return status.Errorf(codes.Unimplemented, "method SystemChannel not implemented")
}
func (UnimplementedMediaApiServer) mustEmbedUnimplementedMediaApiServer() {}

// UnsafeMediaApiServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to MediaApiServer will
// result in compilation errors.
type UnsafeMediaApiServer interface {
	mustEmbedUnimplementedMediaApiServer()
}

func RegisterMediaApiServer(s grpc.ServiceRegistrar, srv MediaApiServer) {
	s.RegisterService(&MediaApi_ServiceDesc, srv)
}

func _MediaApi_PrepareSession_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateParam)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MediaApiServer).PrepareSession(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rpc.MediaApi/PrepareSession",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MediaApiServer).PrepareSession(ctx, req.(*CreateParam))
	}
	return interceptor(ctx, in, info, handler)
}

func _MediaApi_UpdateSession_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateParam)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MediaApiServer).UpdateSession(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rpc.MediaApi/UpdateSession",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MediaApiServer).UpdateSession(ctx, req.(*UpdateParam))
	}
	return interceptor(ctx, in, info, handler)
}

func _MediaApi_StartSession_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StartParam)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MediaApiServer).StartSession(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rpc.MediaApi/StartSession",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MediaApiServer).StartSession(ctx, req.(*StartParam))
	}
	return interceptor(ctx, in, info, handler)
}

func _MediaApi_StopSession_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StopParam)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MediaApiServer).StopSession(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rpc.MediaApi/StopSession",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MediaApiServer).StopSession(ctx, req.(*StopParam))
	}
	return interceptor(ctx, in, info, handler)
}

func _MediaApi_ExecuteAction_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Action)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MediaApiServer).ExecuteAction(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rpc.MediaApi/ExecuteAction",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MediaApiServer).ExecuteAction(ctx, req.(*Action))
	}
	return interceptor(ctx, in, info, handler)
}

func _MediaApi_ExecuteActionWithNotify_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(Action)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(MediaApiServer).ExecuteActionWithNotify(m, &mediaApiExecuteActionWithNotifyServer{stream})
}

type MediaApi_ExecuteActionWithNotifyServer interface {
	Send(*ActionEvent) error
	grpc.ServerStream
}

type mediaApiExecuteActionWithNotifyServer struct {
	grpc.ServerStream
}

func (x *mediaApiExecuteActionWithNotifyServer) Send(m *ActionEvent) error {
	return x.ServerStream.SendMsg(m)
}

func _MediaApi_SystemChannel_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(MediaApiServer).SystemChannel(&mediaApiSystemChannelServer{stream})
}

type MediaApi_SystemChannelServer interface {
	Send(*SystemEvent) error
	Recv() (*SystemEvent, error)
	grpc.ServerStream
}

type mediaApiSystemChannelServer struct {
	grpc.ServerStream
}

func (x *mediaApiSystemChannelServer) Send(m *SystemEvent) error {
	return x.ServerStream.SendMsg(m)
}

func (x *mediaApiSystemChannelServer) Recv() (*SystemEvent, error) {
	m := new(SystemEvent)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// MediaApi_ServiceDesc is the grpc.ServiceDesc for MediaApi service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var MediaApi_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "rpc.MediaApi",
	HandlerType: (*MediaApiServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "PrepareSession",
			Handler:    _MediaApi_PrepareSession_Handler,
		},
		{
			MethodName: "UpdateSession",
			Handler:    _MediaApi_UpdateSession_Handler,
		},
		{
			MethodName: "StartSession",
			Handler:    _MediaApi_StartSession_Handler,
		},
		{
			MethodName: "StopSession",
			Handler:    _MediaApi_StopSession_Handler,
		},
		{
			MethodName: "ExecuteAction",
			Handler:    _MediaApi_ExecuteAction_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "ExecuteActionWithNotify",
			Handler:       _MediaApi_ExecuteActionWithNotify_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "SystemChannel",
			Handler:       _MediaApi_SystemChannel_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "msapi.proto",
}
