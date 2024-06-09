// Sophrosyne
//   Copyright (C) 2024  Mads R. Havmand
//
// This program is free software: you can redistribute it and/or modify
//   it under the terms of the GNU Affero General Public License as published by
//   the Free Software Foundation, either version 3 of the License, or
//   (at your option) any later version.
//
//   This program is distributed in the hope that it will be useful,
//   but WITHOUT ANY WARRANTY; without even the implied warranty of
//   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//   GNU Affero General Public License for more details.
//
//   You should have received a copy of the GNU Affero General Public License
//   along with this program.  If not, see <http://www.gnu.org/licenses/>.

// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v5.26.1
// source: sophrosyne/v0/checkProvider.proto

package v0

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
	CheckProviderService_Check_FullMethodName = "/sophrosyne.v0.CheckProviderService/Check"
)

// CheckProviderServiceClient is the client API for CheckProviderService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type CheckProviderServiceClient interface {
	Check(ctx context.Context, in *CheckProviderRequest, opts ...grpc.CallOption) (*CheckProviderResponse, error)
}

type checkProviderServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewCheckProviderServiceClient(cc grpc.ClientConnInterface) CheckProviderServiceClient {
	return &checkProviderServiceClient{cc}
}

func (c *checkProviderServiceClient) Check(ctx context.Context, in *CheckProviderRequest, opts ...grpc.CallOption) (*CheckProviderResponse, error) {
	out := new(CheckProviderResponse)
	err := c.cc.Invoke(ctx, CheckProviderService_Check_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// CheckProviderServiceServer is the server API for CheckProviderService service.
// All implementations must embed UnimplementedCheckProviderServiceServer
// for forward compatibility
type CheckProviderServiceServer interface {
	Check(context.Context, *CheckProviderRequest) (*CheckProviderResponse, error)
	mustEmbedUnimplementedCheckProviderServiceServer()
}

// UnimplementedCheckProviderServiceServer must be embedded to have forward compatible implementations.
type UnimplementedCheckProviderServiceServer struct {
}

func (UnimplementedCheckProviderServiceServer) Check(context.Context, *CheckProviderRequest) (*CheckProviderResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Check not implemented")
}
func (UnimplementedCheckProviderServiceServer) mustEmbedUnimplementedCheckProviderServiceServer() {}

// UnsafeCheckProviderServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to CheckProviderServiceServer will
// result in compilation errors.
type UnsafeCheckProviderServiceServer interface {
	mustEmbedUnimplementedCheckProviderServiceServer()
}

func RegisterCheckProviderServiceServer(s grpc.ServiceRegistrar, srv CheckProviderServiceServer) {
	s.RegisterService(&CheckProviderService_ServiceDesc, srv)
}

func _CheckProviderService_Check_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CheckProviderRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CheckProviderServiceServer).Check(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: CheckProviderService_Check_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CheckProviderServiceServer).Check(ctx, req.(*CheckProviderRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// CheckProviderService_ServiceDesc is the grpc.ServiceDesc for CheckProviderService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var CheckProviderService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "sophrosyne.v0.CheckProviderService",
	HandlerType: (*CheckProviderServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Check",
			Handler:    _CheckProviderService_Check_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "sophrosyne/v0/checkProvider.proto",
}