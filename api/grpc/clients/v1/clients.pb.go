// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.0
// 	protoc        v5.29.0
// source: clients/v1/clients.proto

package v1

import (
	v1 "github.com/absmach/supermq/api/grpc/common/v1"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type AuthnReq struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	ClientId      string                 `protobuf:"bytes,1,opt,name=client_id,json=clientId,proto3" json:"client_id,omitempty"`
	ClientSecret  string                 `protobuf:"bytes,2,opt,name=client_secret,json=clientSecret,proto3" json:"client_secret,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *AuthnReq) Reset() {
	*x = AuthnReq{}
	mi := &file_clients_v1_clients_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *AuthnReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AuthnReq) ProtoMessage() {}

func (x *AuthnReq) ProtoReflect() protoreflect.Message {
	mi := &file_clients_v1_clients_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AuthnReq.ProtoReflect.Descriptor instead.
func (*AuthnReq) Descriptor() ([]byte, []int) {
	return file_clients_v1_clients_proto_rawDescGZIP(), []int{0}
}

func (x *AuthnReq) GetClientId() string {
	if x != nil {
		return x.ClientId
	}
	return ""
}

func (x *AuthnReq) GetClientSecret() string {
	if x != nil {
		return x.ClientSecret
	}
	return ""
}

type AuthnRes struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Authenticated bool                   `protobuf:"varint,1,opt,name=authenticated,proto3" json:"authenticated,omitempty"`
	Id            string                 `protobuf:"bytes,2,opt,name=id,proto3" json:"id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *AuthnRes) Reset() {
	*x = AuthnRes{}
	mi := &file_clients_v1_clients_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *AuthnRes) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AuthnRes) ProtoMessage() {}

func (x *AuthnRes) ProtoReflect() protoreflect.Message {
	mi := &file_clients_v1_clients_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AuthnRes.ProtoReflect.Descriptor instead.
func (*AuthnRes) Descriptor() ([]byte, []int) {
	return file_clients_v1_clients_proto_rawDescGZIP(), []int{1}
}

func (x *AuthnRes) GetAuthenticated() bool {
	if x != nil {
		return x.Authenticated
	}
	return false
}

func (x *AuthnRes) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

type RemoveChannelConnectionsReq struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	ChannelId     string                 `protobuf:"bytes,1,opt,name=channel_id,json=channelId,proto3" json:"channel_id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *RemoveChannelConnectionsReq) Reset() {
	*x = RemoveChannelConnectionsReq{}
	mi := &file_clients_v1_clients_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RemoveChannelConnectionsReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RemoveChannelConnectionsReq) ProtoMessage() {}

func (x *RemoveChannelConnectionsReq) ProtoReflect() protoreflect.Message {
	mi := &file_clients_v1_clients_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RemoveChannelConnectionsReq.ProtoReflect.Descriptor instead.
func (*RemoveChannelConnectionsReq) Descriptor() ([]byte, []int) {
	return file_clients_v1_clients_proto_rawDescGZIP(), []int{2}
}

func (x *RemoveChannelConnectionsReq) GetChannelId() string {
	if x != nil {
		return x.ChannelId
	}
	return ""
}

type RemoveChannelConnectionsRes struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *RemoveChannelConnectionsRes) Reset() {
	*x = RemoveChannelConnectionsRes{}
	mi := &file_clients_v1_clients_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RemoveChannelConnectionsRes) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RemoveChannelConnectionsRes) ProtoMessage() {}

func (x *RemoveChannelConnectionsRes) ProtoReflect() protoreflect.Message {
	mi := &file_clients_v1_clients_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RemoveChannelConnectionsRes.ProtoReflect.Descriptor instead.
func (*RemoveChannelConnectionsRes) Descriptor() ([]byte, []int) {
	return file_clients_v1_clients_proto_rawDescGZIP(), []int{3}
}

type UnsetParentGroupFromClientReq struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	ParentGroupId string                 `protobuf:"bytes,1,opt,name=parent_group_id,json=parentGroupId,proto3" json:"parent_group_id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *UnsetParentGroupFromClientReq) Reset() {
	*x = UnsetParentGroupFromClientReq{}
	mi := &file_clients_v1_clients_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UnsetParentGroupFromClientReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UnsetParentGroupFromClientReq) ProtoMessage() {}

func (x *UnsetParentGroupFromClientReq) ProtoReflect() protoreflect.Message {
	mi := &file_clients_v1_clients_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UnsetParentGroupFromClientReq.ProtoReflect.Descriptor instead.
func (*UnsetParentGroupFromClientReq) Descriptor() ([]byte, []int) {
	return file_clients_v1_clients_proto_rawDescGZIP(), []int{4}
}

func (x *UnsetParentGroupFromClientReq) GetParentGroupId() string {
	if x != nil {
		return x.ParentGroupId
	}
	return ""
}

type UnsetParentGroupFromClientRes struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *UnsetParentGroupFromClientRes) Reset() {
	*x = UnsetParentGroupFromClientRes{}
	mi := &file_clients_v1_clients_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UnsetParentGroupFromClientRes) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UnsetParentGroupFromClientRes) ProtoMessage() {}

func (x *UnsetParentGroupFromClientRes) ProtoReflect() protoreflect.Message {
	mi := &file_clients_v1_clients_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UnsetParentGroupFromClientRes.ProtoReflect.Descriptor instead.
func (*UnsetParentGroupFromClientRes) Descriptor() ([]byte, []int) {
	return file_clients_v1_clients_proto_rawDescGZIP(), []int{5}
}

var File_clients_v1_clients_proto protoreflect.FileDescriptor

var file_clients_v1_clients_proto_rawDesc = []byte{
	0x0a, 0x18, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x73, 0x2f, 0x76, 0x31, 0x2f, 0x63, 0x6c, 0x69,
	0x65, 0x6e, 0x74, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0a, 0x63, 0x6c, 0x69, 0x65,
	0x6e, 0x74, 0x73, 0x2e, 0x76, 0x31, 0x1a, 0x16, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2f, 0x76,
	0x31, 0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x4c,
	0x0a, 0x08, 0x41, 0x75, 0x74, 0x68, 0x6e, 0x52, 0x65, 0x71, 0x12, 0x1b, 0x0a, 0x09, 0x63, 0x6c,
	0x69, 0x65, 0x6e, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x63,
	0x6c, 0x69, 0x65, 0x6e, 0x74, 0x49, 0x64, 0x12, 0x23, 0x0a, 0x0d, 0x63, 0x6c, 0x69, 0x65, 0x6e,
	0x74, 0x5f, 0x73, 0x65, 0x63, 0x72, 0x65, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c,
	0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x53, 0x65, 0x63, 0x72, 0x65, 0x74, 0x22, 0x40, 0x0a, 0x08,
	0x41, 0x75, 0x74, 0x68, 0x6e, 0x52, 0x65, 0x73, 0x12, 0x24, 0x0a, 0x0d, 0x61, 0x75, 0x74, 0x68,
	0x65, 0x6e, 0x74, 0x69, 0x63, 0x61, 0x74, 0x65, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52,
	0x0d, 0x61, 0x75, 0x74, 0x68, 0x65, 0x6e, 0x74, 0x69, 0x63, 0x61, 0x74, 0x65, 0x64, 0x12, 0x0e,
	0x0a, 0x02, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x22, 0x3c,
	0x0a, 0x1b, 0x52, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x43, 0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x43,
	0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x52, 0x65, 0x71, 0x12, 0x1d, 0x0a,
	0x0a, 0x63, 0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x09, 0x63, 0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x49, 0x64, 0x22, 0x1d, 0x0a, 0x1b,
	0x52, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x43, 0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x43, 0x6f, 0x6e,
	0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x52, 0x65, 0x73, 0x22, 0x47, 0x0a, 0x1d, 0x55,
	0x6e, 0x73, 0x65, 0x74, 0x50, 0x61, 0x72, 0x65, 0x6e, 0x74, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x46,
	0x72, 0x6f, 0x6d, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x71, 0x12, 0x26, 0x0a, 0x0f,
	0x70, 0x61, 0x72, 0x65, 0x6e, 0x74, 0x5f, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x5f, 0x69, 0x64, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x70, 0x61, 0x72, 0x65, 0x6e, 0x74, 0x47, 0x72, 0x6f,
	0x75, 0x70, 0x49, 0x64, 0x22, 0x1f, 0x0a, 0x1d, 0x55, 0x6e, 0x73, 0x65, 0x74, 0x50, 0x61, 0x72,
	0x65, 0x6e, 0x74, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x46, 0x72, 0x6f, 0x6d, 0x43, 0x6c, 0x69, 0x65,
	0x6e, 0x74, 0x52, 0x65, 0x73, 0x32, 0x83, 0x05, 0x0a, 0x0e, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74,
	0x73, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x3c, 0x0a, 0x0c, 0x41, 0x75, 0x74, 0x68,
	0x65, 0x6e, 0x74, 0x69, 0x63, 0x61, 0x74, 0x65, 0x12, 0x14, 0x2e, 0x63, 0x6c, 0x69, 0x65, 0x6e,
	0x74, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x41, 0x75, 0x74, 0x68, 0x6e, 0x52, 0x65, 0x71, 0x1a, 0x14,
	0x2e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x41, 0x75, 0x74, 0x68,
	0x6e, 0x52, 0x65, 0x73, 0x22, 0x00, 0x12, 0x4e, 0x0a, 0x0e, 0x52, 0x65, 0x74, 0x72, 0x69, 0x65,
	0x76, 0x65, 0x45, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x12, 0x1c, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f,
	0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65, 0x74, 0x72, 0x69, 0x65, 0x76, 0x65, 0x45, 0x6e, 0x74,
	0x69, 0x74, 0x79, 0x52, 0x65, 0x71, 0x1a, 0x1c, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e,
	0x76, 0x31, 0x2e, 0x52, 0x65, 0x74, 0x72, 0x69, 0x65, 0x76, 0x65, 0x45, 0x6e, 0x74, 0x69, 0x74,
	0x79, 0x52, 0x65, 0x73, 0x22, 0x00, 0x12, 0x54, 0x0a, 0x10, 0x52, 0x65, 0x74, 0x72, 0x69, 0x65,
	0x76, 0x65, 0x45, 0x6e, 0x74, 0x69, 0x74, 0x69, 0x65, 0x73, 0x12, 0x1e, 0x2e, 0x63, 0x6f, 0x6d,
	0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65, 0x74, 0x72, 0x69, 0x65, 0x76, 0x65, 0x45,
	0x6e, 0x74, 0x69, 0x74, 0x69, 0x65, 0x73, 0x52, 0x65, 0x71, 0x1a, 0x1e, 0x2e, 0x63, 0x6f, 0x6d,
	0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65, 0x74, 0x72, 0x69, 0x65, 0x76, 0x65, 0x45,
	0x6e, 0x74, 0x69, 0x74, 0x69, 0x65, 0x73, 0x52, 0x65, 0x73, 0x22, 0x00, 0x12, 0x4e, 0x0a, 0x0e,
	0x41, 0x64, 0x64, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x1c,
	0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x41, 0x64, 0x64, 0x43, 0x6f,
	0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x52, 0x65, 0x71, 0x1a, 0x1c, 0x2e, 0x63,
	0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x41, 0x64, 0x64, 0x43, 0x6f, 0x6e, 0x6e,
	0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x52, 0x65, 0x73, 0x22, 0x00, 0x12, 0x57, 0x0a, 0x11,
	0x52, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e,
	0x73, 0x12, 0x1f, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65,
	0x6d, 0x6f, 0x76, 0x65, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x52,
	0x65, 0x71, 0x1a, 0x1f, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x52,
	0x65, 0x6d, 0x6f, 0x76, 0x65, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73,
	0x52, 0x65, 0x73, 0x22, 0x00, 0x12, 0x6e, 0x0a, 0x18, 0x52, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x43,
	0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e,
	0x73, 0x12, 0x27, 0x2e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x52,
	0x65, 0x6d, 0x6f, 0x76, 0x65, 0x43, 0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x43, 0x6f, 0x6e, 0x6e,
	0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x52, 0x65, 0x71, 0x1a, 0x27, 0x2e, 0x63, 0x6c, 0x69,
	0x65, 0x6e, 0x74, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x43, 0x68,
	0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73,
	0x52, 0x65, 0x73, 0x22, 0x00, 0x12, 0x74, 0x0a, 0x1a, 0x55, 0x6e, 0x73, 0x65, 0x74, 0x50, 0x61,
	0x72, 0x65, 0x6e, 0x74, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x46, 0x72, 0x6f, 0x6d, 0x43, 0x6c, 0x69,
	0x65, 0x6e, 0x74, 0x12, 0x29, 0x2e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x76, 0x31,
	0x2e, 0x55, 0x6e, 0x73, 0x65, 0x74, 0x50, 0x61, 0x72, 0x65, 0x6e, 0x74, 0x47, 0x72, 0x6f, 0x75,
	0x70, 0x46, 0x72, 0x6f, 0x6d, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x71, 0x1a, 0x29,
	0x2e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x55, 0x6e, 0x73, 0x65,
	0x74, 0x50, 0x61, 0x72, 0x65, 0x6e, 0x74, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x46, 0x72, 0x6f, 0x6d,
	0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x73, 0x22, 0x00, 0x42, 0x30, 0x5a, 0x2e, 0x67,
	0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x61, 0x62, 0x73, 0x6d, 0x61, 0x63,
	0x68, 0x2f, 0x73, 0x75, 0x70, 0x65, 0x72, 0x6d, 0x71, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x67, 0x72,
	0x70, 0x63, 0x2f, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x73, 0x2f, 0x76, 0x31, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_clients_v1_clients_proto_rawDescOnce sync.Once
	file_clients_v1_clients_proto_rawDescData = file_clients_v1_clients_proto_rawDesc
)

func file_clients_v1_clients_proto_rawDescGZIP() []byte {
	file_clients_v1_clients_proto_rawDescOnce.Do(func() {
		file_clients_v1_clients_proto_rawDescData = protoimpl.X.CompressGZIP(file_clients_v1_clients_proto_rawDescData)
	})
	return file_clients_v1_clients_proto_rawDescData
}

var file_clients_v1_clients_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_clients_v1_clients_proto_goTypes = []any{
	(*AuthnReq)(nil),                      // 0: clients.v1.AuthnReq
	(*AuthnRes)(nil),                      // 1: clients.v1.AuthnRes
	(*RemoveChannelConnectionsReq)(nil),   // 2: clients.v1.RemoveChannelConnectionsReq
	(*RemoveChannelConnectionsRes)(nil),   // 3: clients.v1.RemoveChannelConnectionsRes
	(*UnsetParentGroupFromClientReq)(nil), // 4: clients.v1.UnsetParentGroupFromClientReq
	(*UnsetParentGroupFromClientRes)(nil), // 5: clients.v1.UnsetParentGroupFromClientRes
	(*v1.RetrieveEntityReq)(nil),          // 6: common.v1.RetrieveEntityReq
	(*v1.RetrieveEntitiesReq)(nil),        // 7: common.v1.RetrieveEntitiesReq
	(*v1.AddConnectionsReq)(nil),          // 8: common.v1.AddConnectionsReq
	(*v1.RemoveConnectionsReq)(nil),       // 9: common.v1.RemoveConnectionsReq
	(*v1.RetrieveEntityRes)(nil),          // 10: common.v1.RetrieveEntityRes
	(*v1.RetrieveEntitiesRes)(nil),        // 11: common.v1.RetrieveEntitiesRes
	(*v1.AddConnectionsRes)(nil),          // 12: common.v1.AddConnectionsRes
	(*v1.RemoveConnectionsRes)(nil),       // 13: common.v1.RemoveConnectionsRes
}
var file_clients_v1_clients_proto_depIdxs = []int32{
	0,  // 0: clients.v1.ClientsService.Authenticate:input_type -> clients.v1.AuthnReq
	6,  // 1: clients.v1.ClientsService.RetrieveEntity:input_type -> common.v1.RetrieveEntityReq
	7,  // 2: clients.v1.ClientsService.RetrieveEntities:input_type -> common.v1.RetrieveEntitiesReq
	8,  // 3: clients.v1.ClientsService.AddConnections:input_type -> common.v1.AddConnectionsReq
	9,  // 4: clients.v1.ClientsService.RemoveConnections:input_type -> common.v1.RemoveConnectionsReq
	2,  // 5: clients.v1.ClientsService.RemoveChannelConnections:input_type -> clients.v1.RemoveChannelConnectionsReq
	4,  // 6: clients.v1.ClientsService.UnsetParentGroupFromClient:input_type -> clients.v1.UnsetParentGroupFromClientReq
	1,  // 7: clients.v1.ClientsService.Authenticate:output_type -> clients.v1.AuthnRes
	10, // 8: clients.v1.ClientsService.RetrieveEntity:output_type -> common.v1.RetrieveEntityRes
	11, // 9: clients.v1.ClientsService.RetrieveEntities:output_type -> common.v1.RetrieveEntitiesRes
	12, // 10: clients.v1.ClientsService.AddConnections:output_type -> common.v1.AddConnectionsRes
	13, // 11: clients.v1.ClientsService.RemoveConnections:output_type -> common.v1.RemoveConnectionsRes
	3,  // 12: clients.v1.ClientsService.RemoveChannelConnections:output_type -> clients.v1.RemoveChannelConnectionsRes
	5,  // 13: clients.v1.ClientsService.UnsetParentGroupFromClient:output_type -> clients.v1.UnsetParentGroupFromClientRes
	7,  // [7:14] is the sub-list for method output_type
	0,  // [0:7] is the sub-list for method input_type
	0,  // [0:0] is the sub-list for extension type_name
	0,  // [0:0] is the sub-list for extension extendee
	0,  // [0:0] is the sub-list for field type_name
}

func init() { file_clients_v1_clients_proto_init() }
func file_clients_v1_clients_proto_init() {
	if File_clients_v1_clients_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_clients_v1_clients_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_clients_v1_clients_proto_goTypes,
		DependencyIndexes: file_clients_v1_clients_proto_depIdxs,
		MessageInfos:      file_clients_v1_clients_proto_msgTypes,
	}.Build()
	File_clients_v1_clients_proto = out.File
	file_clients_v1_clients_proto_rawDesc = nil
	file_clients_v1_clients_proto_goTypes = nil
	file_clients_v1_clients_proto_depIdxs = nil
}
