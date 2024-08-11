// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v5.26.1
// source: tao.proto

package proto

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

const (
	TaoService_ObjectAdd_FullMethodName  = "/tao.TaoService/ObjectAdd"
	TaoService_ObjectGet_FullMethodName  = "/tao.TaoService/ObjectGet"
	TaoService_AssocAdd_FullMethodName   = "/tao.TaoService/AssocAdd"
	TaoService_AssocGet_FullMethodName   = "/tao.TaoService/AssocGet"
	TaoService_AssocRange_FullMethodName = "/tao.TaoService/AssocRange"
)

// TaoServiceClient is the client API for TaoService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type TaoServiceClient interface {
	ObjectAdd(ctx context.Context, in *ObjectAddRequest, opts ...grpc.CallOption) (*GenericOkResponse, error)
	ObjectGet(ctx context.Context, in *ObjectGetRequest, opts ...grpc.CallOption) (*AssocGetResponse, error)
	AssocAdd(ctx context.Context, in *AssocAddRequest, opts ...grpc.CallOption) (*GenericOkResponse, error)
	AssocGet(ctx context.Context, in *AssocGetRequest, opts ...grpc.CallOption) (*AssocGetResponse, error)
	AssocRange(ctx context.Context, in *AssocRangeRequest, opts ...grpc.CallOption) (*AssocGetResponse, error)
}

type taoServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewTaoServiceClient(cc grpc.ClientConnInterface) TaoServiceClient {
	return &taoServiceClient{cc}
}

func (c *taoServiceClient) ObjectAdd(ctx context.Context, in *ObjectAddRequest, opts ...grpc.CallOption) (*GenericOkResponse, error) {
	out := new(GenericOkResponse)
	err := c.cc.Invoke(ctx, TaoService_ObjectAdd_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *taoServiceClient) ObjectGet(ctx context.Context, in *ObjectGetRequest, opts ...grpc.CallOption) (*AssocGetResponse, error) {
	out := new(AssocGetResponse)
	err := c.cc.Invoke(ctx, TaoService_ObjectGet_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *taoServiceClient) AssocAdd(ctx context.Context, in *AssocAddRequest, opts ...grpc.CallOption) (*GenericOkResponse, error) {
	out := new(GenericOkResponse)
	err := c.cc.Invoke(ctx, TaoService_AssocAdd_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *taoServiceClient) AssocGet(ctx context.Context, in *AssocGetRequest, opts ...grpc.CallOption) (*AssocGetResponse, error) {
	out := new(AssocGetResponse)
	err := c.cc.Invoke(ctx, TaoService_AssocGet_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *taoServiceClient) AssocRange(ctx context.Context, in *AssocRangeRequest, opts ...grpc.CallOption) (*AssocGetResponse, error) {
	out := new(AssocGetResponse)
	err := c.cc.Invoke(ctx, TaoService_AssocRange_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TaoServiceServer is the server API for TaoService service.
// All implementations must embed UnimplementedTaoServiceServer
// for forward compatibility
type TaoServiceServer interface {
	ObjectAdd(context.Context, *ObjectAddRequest) (*GenericOkResponse, error)
	ObjectGet(context.Context, *ObjectGetRequest) (*AssocGetResponse, error)
	AssocAdd(context.Context, *AssocAddRequest) (*GenericOkResponse, error)
	AssocGet(context.Context, *AssocGetRequest) (*AssocGetResponse, error)
	AssocRange(context.Context, *AssocRangeRequest) (*AssocGetResponse, error)
	mustEmbedUnimplementedTaoServiceServer()
}

// UnimplementedTaoServiceServer must be embedded to have forward compatible implementations.
type UnimplementedTaoServiceServer struct {
}

func (UnimplementedTaoServiceServer) ObjectAdd(context.Context, *ObjectAddRequest) (*GenericOkResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ObjectAdd not implemented")
}
func (UnimplementedTaoServiceServer) ObjectGet(context.Context, *ObjectGetRequest) (*AssocGetResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ObjectGet not implemented")
}
func (UnimplementedTaoServiceServer) AssocAdd(context.Context, *AssocAddRequest) (*GenericOkResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AssocAdd not implemented")
}
func (UnimplementedTaoServiceServer) AssocGet(context.Context, *AssocGetRequest) (*AssocGetResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AssocGet not implemented")
}
func (UnimplementedTaoServiceServer) AssocRange(context.Context, *AssocRangeRequest) (*AssocGetResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AssocRange not implemented")
}
func (UnimplementedTaoServiceServer) mustEmbedUnimplementedTaoServiceServer() {}

// UnsafeTaoServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to TaoServiceServer will
// result in compilation errors.
type UnsafeTaoServiceServer interface {
	mustEmbedUnimplementedTaoServiceServer()
}

func RegisterTaoServiceServer(s grpc.ServiceRegistrar, srv TaoServiceServer) {
	s.RegisterService(&TaoService_ServiceDesc, srv)
}

func _TaoService_ObjectAdd_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ObjectAddRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TaoServiceServer).ObjectAdd(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TaoService_ObjectAdd_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TaoServiceServer).ObjectAdd(ctx, req.(*ObjectAddRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TaoService_ObjectGet_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ObjectGetRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TaoServiceServer).ObjectGet(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TaoService_ObjectGet_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TaoServiceServer).ObjectGet(ctx, req.(*ObjectGetRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TaoService_AssocAdd_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AssocAddRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TaoServiceServer).AssocAdd(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TaoService_AssocAdd_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TaoServiceServer).AssocAdd(ctx, req.(*AssocAddRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TaoService_AssocGet_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AssocGetRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TaoServiceServer).AssocGet(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TaoService_AssocGet_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TaoServiceServer).AssocGet(ctx, req.(*AssocGetRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TaoService_AssocRange_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AssocRangeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TaoServiceServer).AssocRange(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TaoService_AssocRange_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TaoServiceServer).AssocRange(ctx, req.(*AssocRangeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// TaoService_ServiceDesc is the grpc.ServiceDesc for TaoService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var TaoService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "tao.TaoService",
	HandlerType: (*TaoServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ObjectAdd",
			Handler:    _TaoService_ObjectAdd_Handler,
		},
		{
			MethodName: "ObjectGet",
			Handler:    _TaoService_ObjectGet_Handler,
		},
		{
			MethodName: "AssocAdd",
			Handler:    _TaoService_AssocAdd_Handler,
		},
		{
			MethodName: "AssocGet",
			Handler:    _TaoService_AssocGet_Handler,
		},
		{
			MethodName: "AssocRange",
			Handler:    _TaoService_AssocRange_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "tao.proto",
}
