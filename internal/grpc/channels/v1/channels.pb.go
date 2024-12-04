// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.2
// 	protoc        v5.28.3
// source: channels/v1/channels.proto

package v1

import (
	v1 "github.com/absmach/supermq/internal/grpc/common/v1"
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

type RemoveClientConnectionsReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ClientId string `protobuf:"bytes,1,opt,name=client_id,json=clientId,proto3" json:"client_id,omitempty"`
}

func (x *RemoveClientConnectionsReq) Reset() {
	*x = RemoveClientConnectionsReq{}
	mi := &file_channels_v1_channels_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RemoveClientConnectionsReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RemoveClientConnectionsReq) ProtoMessage() {}

func (x *RemoveClientConnectionsReq) ProtoReflect() protoreflect.Message {
	mi := &file_channels_v1_channels_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RemoveClientConnectionsReq.ProtoReflect.Descriptor instead.
func (*RemoveClientConnectionsReq) Descriptor() ([]byte, []int) {
	return file_channels_v1_channels_proto_rawDescGZIP(), []int{0}
}

func (x *RemoveClientConnectionsReq) GetClientId() string {
	if x != nil {
		return x.ClientId
	}
	return ""
}

type RemoveClientConnectionsRes struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *RemoveClientConnectionsRes) Reset() {
	*x = RemoveClientConnectionsRes{}
	mi := &file_channels_v1_channels_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RemoveClientConnectionsRes) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RemoveClientConnectionsRes) ProtoMessage() {}

func (x *RemoveClientConnectionsRes) ProtoReflect() protoreflect.Message {
	mi := &file_channels_v1_channels_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RemoveClientConnectionsRes.ProtoReflect.Descriptor instead.
func (*RemoveClientConnectionsRes) Descriptor() ([]byte, []int) {
	return file_channels_v1_channels_proto_rawDescGZIP(), []int{1}
}

type UnsetParentGroupFromChannelsReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ParentGroupId string `protobuf:"bytes,1,opt,name=parent_group_id,json=parentGroupId,proto3" json:"parent_group_id,omitempty"`
}

func (x *UnsetParentGroupFromChannelsReq) Reset() {
	*x = UnsetParentGroupFromChannelsReq{}
	mi := &file_channels_v1_channels_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UnsetParentGroupFromChannelsReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UnsetParentGroupFromChannelsReq) ProtoMessage() {}

func (x *UnsetParentGroupFromChannelsReq) ProtoReflect() protoreflect.Message {
	mi := &file_channels_v1_channels_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UnsetParentGroupFromChannelsReq.ProtoReflect.Descriptor instead.
func (*UnsetParentGroupFromChannelsReq) Descriptor() ([]byte, []int) {
	return file_channels_v1_channels_proto_rawDescGZIP(), []int{2}
}

func (x *UnsetParentGroupFromChannelsReq) GetParentGroupId() string {
	if x != nil {
		return x.ParentGroupId
	}
	return ""
}

type UnsetParentGroupFromChannelsRes struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *UnsetParentGroupFromChannelsRes) Reset() {
	*x = UnsetParentGroupFromChannelsRes{}
	mi := &file_channels_v1_channels_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UnsetParentGroupFromChannelsRes) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UnsetParentGroupFromChannelsRes) ProtoMessage() {}

func (x *UnsetParentGroupFromChannelsRes) ProtoReflect() protoreflect.Message {
	mi := &file_channels_v1_channels_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UnsetParentGroupFromChannelsRes.ProtoReflect.Descriptor instead.
func (*UnsetParentGroupFromChannelsRes) Descriptor() ([]byte, []int) {
	return file_channels_v1_channels_proto_rawDescGZIP(), []int{3}
}

type AuthzReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	DomainId   string `protobuf:"bytes,1,opt,name=domain_id,json=domainId,proto3" json:"domain_id,omitempty"`
	ClientId   string `protobuf:"bytes,2,opt,name=client_id,json=clientId,proto3" json:"client_id,omitempty"`
	ClientType string `protobuf:"bytes,3,opt,name=client_type,json=clientType,proto3" json:"client_type,omitempty"`
	ChannelId  string `protobuf:"bytes,4,opt,name=channel_id,json=channelId,proto3" json:"channel_id,omitempty"`
	Type       uint32 `protobuf:"varint,5,opt,name=type,proto3" json:"type,omitempty"`
}

func (x *AuthzReq) Reset() {
	*x = AuthzReq{}
	mi := &file_channels_v1_channels_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *AuthzReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AuthzReq) ProtoMessage() {}

func (x *AuthzReq) ProtoReflect() protoreflect.Message {
	mi := &file_channels_v1_channels_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AuthzReq.ProtoReflect.Descriptor instead.
func (*AuthzReq) Descriptor() ([]byte, []int) {
	return file_channels_v1_channels_proto_rawDescGZIP(), []int{4}
}

func (x *AuthzReq) GetDomainId() string {
	if x != nil {
		return x.DomainId
	}
	return ""
}

func (x *AuthzReq) GetClientId() string {
	if x != nil {
		return x.ClientId
	}
	return ""
}

func (x *AuthzReq) GetClientType() string {
	if x != nil {
		return x.ClientType
	}
	return ""
}

func (x *AuthzReq) GetChannelId() string {
	if x != nil {
		return x.ChannelId
	}
	return ""
}

func (x *AuthzReq) GetType() uint32 {
	if x != nil {
		return x.Type
	}
	return 0
}

type AuthzRes struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Authorized bool `protobuf:"varint,1,opt,name=authorized,proto3" json:"authorized,omitempty"`
}

func (x *AuthzRes) Reset() {
	*x = AuthzRes{}
	mi := &file_channels_v1_channels_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *AuthzRes) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AuthzRes) ProtoMessage() {}

func (x *AuthzRes) ProtoReflect() protoreflect.Message {
	mi := &file_channels_v1_channels_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AuthzRes.ProtoReflect.Descriptor instead.
func (*AuthzRes) Descriptor() ([]byte, []int) {
	return file_channels_v1_channels_proto_rawDescGZIP(), []int{5}
}

func (x *AuthzRes) GetAuthorized() bool {
	if x != nil {
		return x.Authorized
	}
	return false
}

var File_channels_v1_channels_proto protoreflect.FileDescriptor

var file_channels_v1_channels_proto_rawDesc = []byte{
	0x0a, 0x1a, 0x63, 0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x73, 0x2f, 0x76, 0x31, 0x2f, 0x63, 0x68,
	0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0b, 0x63, 0x68,
	0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x73, 0x2e, 0x76, 0x31, 0x1a, 0x16, 0x63, 0x6f, 0x6d, 0x6d, 0x6f,
	0x6e, 0x2f, 0x76, 0x31, 0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x22, 0x39, 0x0a, 0x1a, 0x52, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x43, 0x6c, 0x69, 0x65, 0x6e,
	0x74, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x52, 0x65, 0x71, 0x12,
	0x1b, 0x0a, 0x09, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x08, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x49, 0x64, 0x22, 0x1c, 0x0a, 0x1a,
	0x52, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x43, 0x6f, 0x6e, 0x6e,
	0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x52, 0x65, 0x73, 0x22, 0x49, 0x0a, 0x1f, 0x55, 0x6e,
	0x73, 0x65, 0x74, 0x50, 0x61, 0x72, 0x65, 0x6e, 0x74, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x46, 0x72,
	0x6f, 0x6d, 0x43, 0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x73, 0x52, 0x65, 0x71, 0x12, 0x26, 0x0a,
	0x0f, 0x70, 0x61, 0x72, 0x65, 0x6e, 0x74, 0x5f, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x5f, 0x69, 0x64,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x70, 0x61, 0x72, 0x65, 0x6e, 0x74, 0x47, 0x72,
	0x6f, 0x75, 0x70, 0x49, 0x64, 0x22, 0x21, 0x0a, 0x1f, 0x55, 0x6e, 0x73, 0x65, 0x74, 0x50, 0x61,
	0x72, 0x65, 0x6e, 0x74, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x46, 0x72, 0x6f, 0x6d, 0x43, 0x68, 0x61,
	0x6e, 0x6e, 0x65, 0x6c, 0x73, 0x52, 0x65, 0x73, 0x22, 0x98, 0x01, 0x0a, 0x08, 0x41, 0x75, 0x74,
	0x68, 0x7a, 0x52, 0x65, 0x71, 0x12, 0x1b, 0x0a, 0x09, 0x64, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x5f,
	0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x64, 0x6f, 0x6d, 0x61, 0x69, 0x6e,
	0x49, 0x64, 0x12, 0x1b, 0x0a, 0x09, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x5f, 0x69, 0x64, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x49, 0x64, 0x12,
	0x1f, 0x0a, 0x0b, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65,
	0x12, 0x1d, 0x0a, 0x0a, 0x63, 0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x5f, 0x69, 0x64, 0x18, 0x04,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x63, 0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x49, 0x64, 0x12,
	0x12, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x04, 0x74,
	0x79, 0x70, 0x65, 0x22, 0x2a, 0x0a, 0x08, 0x41, 0x75, 0x74, 0x68, 0x7a, 0x52, 0x65, 0x73, 0x12,
	0x1e, 0x0a, 0x0a, 0x61, 0x75, 0x74, 0x68, 0x6f, 0x72, 0x69, 0x7a, 0x65, 0x64, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x08, 0x52, 0x0a, 0x61, 0x75, 0x74, 0x68, 0x6f, 0x72, 0x69, 0x7a, 0x65, 0x64, 0x32,
	0x8b, 0x03, 0x0a, 0x0f, 0x43, 0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x73, 0x53, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x12, 0x3b, 0x0a, 0x09, 0x41, 0x75, 0x74, 0x68, 0x6f, 0x72, 0x69, 0x7a, 0x65,
	0x12, 0x15, 0x2e, 0x63, 0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x41,
	0x75, 0x74, 0x68, 0x7a, 0x52, 0x65, 0x71, 0x1a, 0x15, 0x2e, 0x63, 0x68, 0x61, 0x6e, 0x6e, 0x65,
	0x6c, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x41, 0x75, 0x74, 0x68, 0x7a, 0x52, 0x65, 0x73, 0x22, 0x00,
	0x12, 0x6d, 0x0a, 0x17, 0x52, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74,
	0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x27, 0x2e, 0x63, 0x68,
	0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65, 0x6d, 0x6f, 0x76, 0x65,
	0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e,
	0x73, 0x52, 0x65, 0x71, 0x1a, 0x27, 0x2e, 0x63, 0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x73, 0x2e,
	0x76, 0x31, 0x2e, 0x52, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x43,
	0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x52, 0x65, 0x73, 0x22, 0x00, 0x12,
	0x7c, 0x0a, 0x1c, 0x55, 0x6e, 0x73, 0x65, 0x74, 0x50, 0x61, 0x72, 0x65, 0x6e, 0x74, 0x47, 0x72,
	0x6f, 0x75, 0x70, 0x46, 0x72, 0x6f, 0x6d, 0x43, 0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x73, 0x12,
	0x2c, 0x2e, 0x63, 0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x55, 0x6e,
	0x73, 0x65, 0x74, 0x50, 0x61, 0x72, 0x65, 0x6e, 0x74, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x46, 0x72,
	0x6f, 0x6d, 0x43, 0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x73, 0x52, 0x65, 0x71, 0x1a, 0x2c, 0x2e,
	0x63, 0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x55, 0x6e, 0x73, 0x65,
	0x74, 0x50, 0x61, 0x72, 0x65, 0x6e, 0x74, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x46, 0x72, 0x6f, 0x6d,
	0x43, 0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x73, 0x52, 0x65, 0x73, 0x22, 0x00, 0x12, 0x4e, 0x0a,
	0x0e, 0x52, 0x65, 0x74, 0x72, 0x69, 0x65, 0x76, 0x65, 0x45, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x12,
	0x1c, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65, 0x74, 0x72,
	0x69, 0x65, 0x76, 0x65, 0x45, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x52, 0x65, 0x71, 0x1a, 0x1c, 0x2e,
	0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65, 0x74, 0x72, 0x69, 0x65,
	0x76, 0x65, 0x45, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x52, 0x65, 0x73, 0x22, 0x00, 0x42, 0x36, 0x5a,
	0x34, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x61, 0x62, 0x73, 0x6d,
	0x61, 0x63, 0x68, 0x2f, 0x73, 0x75, 0x70, 0x65, 0x72, 0x6d, 0x71, 0x2f, 0x69, 0x6e, 0x74, 0x65,
	0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x67, 0x72, 0x70, 0x63, 0x2f, 0x63, 0x68, 0x61, 0x6e, 0x6e, 0x65,
	0x6c, 0x73, 0x2f, 0x76, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_channels_v1_channels_proto_rawDescOnce sync.Once
	file_channels_v1_channels_proto_rawDescData = file_channels_v1_channels_proto_rawDesc
)

func file_channels_v1_channels_proto_rawDescGZIP() []byte {
	file_channels_v1_channels_proto_rawDescOnce.Do(func() {
		file_channels_v1_channels_proto_rawDescData = protoimpl.X.CompressGZIP(file_channels_v1_channels_proto_rawDescData)
	})
	return file_channels_v1_channels_proto_rawDescData
}

var file_channels_v1_channels_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_channels_v1_channels_proto_goTypes = []any{
	(*RemoveClientConnectionsReq)(nil),      // 0: channels.v1.RemoveClientConnectionsReq
	(*RemoveClientConnectionsRes)(nil),      // 1: channels.v1.RemoveClientConnectionsRes
	(*UnsetParentGroupFromChannelsReq)(nil), // 2: channels.v1.UnsetParentGroupFromChannelsReq
	(*UnsetParentGroupFromChannelsRes)(nil), // 3: channels.v1.UnsetParentGroupFromChannelsRes
	(*AuthzReq)(nil),                        // 4: channels.v1.AuthzReq
	(*AuthzRes)(nil),                        // 5: channels.v1.AuthzRes
	(*v1.RetrieveEntityReq)(nil),            // 6: common.v1.RetrieveEntityReq
	(*v1.RetrieveEntityRes)(nil),            // 7: common.v1.RetrieveEntityRes
}
var file_channels_v1_channels_proto_depIdxs = []int32{
	4, // 0: channels.v1.ChannelsService.Authorize:input_type -> channels.v1.AuthzReq
	0, // 1: channels.v1.ChannelsService.RemoveClientConnections:input_type -> channels.v1.RemoveClientConnectionsReq
	2, // 2: channels.v1.ChannelsService.UnsetParentGroupFromChannels:input_type -> channels.v1.UnsetParentGroupFromChannelsReq
	6, // 3: channels.v1.ChannelsService.RetrieveEntity:input_type -> common.v1.RetrieveEntityReq
	5, // 4: channels.v1.ChannelsService.Authorize:output_type -> channels.v1.AuthzRes
	1, // 5: channels.v1.ChannelsService.RemoveClientConnections:output_type -> channels.v1.RemoveClientConnectionsRes
	3, // 6: channels.v1.ChannelsService.UnsetParentGroupFromChannels:output_type -> channels.v1.UnsetParentGroupFromChannelsRes
	7, // 7: channels.v1.ChannelsService.RetrieveEntity:output_type -> common.v1.RetrieveEntityRes
	4, // [4:8] is the sub-list for method output_type
	0, // [0:4] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_channels_v1_channels_proto_init() }
func file_channels_v1_channels_proto_init() {
	if File_channels_v1_channels_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_channels_v1_channels_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_channels_v1_channels_proto_goTypes,
		DependencyIndexes: file_channels_v1_channels_proto_depIdxs,
		MessageInfos:      file_channels_v1_channels_proto_msgTypes,
	}.Build()
	File_channels_v1_channels_proto = out.File
	file_channels_v1_channels_proto_rawDesc = nil
	file_channels_v1_channels_proto_goTypes = nil
	file_channels_v1_channels_proto_depIdxs = nil
}
