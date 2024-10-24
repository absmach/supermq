// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v5.27.3
// source: common/v1/common.proto

package v1

import (
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

type RetrieveEntitiesReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Ids []string `protobuf:"bytes,1,rep,name=ids,proto3" json:"ids,omitempty"`
}

func (x *RetrieveEntitiesReq) Reset() {
	*x = RetrieveEntitiesReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_common_v1_common_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RetrieveEntitiesReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RetrieveEntitiesReq) ProtoMessage() {}

func (x *RetrieveEntitiesReq) ProtoReflect() protoreflect.Message {
	mi := &file_common_v1_common_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RetrieveEntitiesReq.ProtoReflect.Descriptor instead.
func (*RetrieveEntitiesReq) Descriptor() ([]byte, []int) {
	return file_common_v1_common_proto_rawDescGZIP(), []int{0}
}

func (x *RetrieveEntitiesReq) GetIds() []string {
	if x != nil {
		return x.Ids
	}
	return nil
}

type RetrieveEntitiesRes struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Total    uint64         `protobuf:"varint,1,opt,name=total,proto3" json:"total,omitempty"`
	Limit    uint64         `protobuf:"varint,2,opt,name=limit,proto3" json:"limit,omitempty"`
	Offset   uint64         `protobuf:"varint,3,opt,name=offset,proto3" json:"offset,omitempty"`
	Entities []*EntityBasic `protobuf:"bytes,4,rep,name=entities,proto3" json:"entities,omitempty"`
}

func (x *RetrieveEntitiesRes) Reset() {
	*x = RetrieveEntitiesRes{}
	if protoimpl.UnsafeEnabled {
		mi := &file_common_v1_common_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RetrieveEntitiesRes) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RetrieveEntitiesRes) ProtoMessage() {}

func (x *RetrieveEntitiesRes) ProtoReflect() protoreflect.Message {
	mi := &file_common_v1_common_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RetrieveEntitiesRes.ProtoReflect.Descriptor instead.
func (*RetrieveEntitiesRes) Descriptor() ([]byte, []int) {
	return file_common_v1_common_proto_rawDescGZIP(), []int{1}
}

func (x *RetrieveEntitiesRes) GetTotal() uint64 {
	if x != nil {
		return x.Total
	}
	return 0
}

func (x *RetrieveEntitiesRes) GetLimit() uint64 {
	if x != nil {
		return x.Limit
	}
	return 0
}

func (x *RetrieveEntitiesRes) GetOffset() uint64 {
	if x != nil {
		return x.Offset
	}
	return 0
}

func (x *RetrieveEntitiesRes) GetEntities() []*EntityBasic {
	if x != nil {
		return x.Entities
	}
	return nil
}

type RetrieveEntityReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *RetrieveEntityReq) Reset() {
	*x = RetrieveEntityReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_common_v1_common_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RetrieveEntityReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RetrieveEntityReq) ProtoMessage() {}

func (x *RetrieveEntityReq) ProtoReflect() protoreflect.Message {
	mi := &file_common_v1_common_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RetrieveEntityReq.ProtoReflect.Descriptor instead.
func (*RetrieveEntityReq) Descriptor() ([]byte, []int) {
	return file_common_v1_common_proto_rawDescGZIP(), []int{2}
}

func (x *RetrieveEntityReq) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

type RetrieveEntityRes struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Entity *EntityBasic `protobuf:"bytes,1,opt,name=entity,proto3" json:"entity,omitempty"`
}

func (x *RetrieveEntityRes) Reset() {
	*x = RetrieveEntityRes{}
	if protoimpl.UnsafeEnabled {
		mi := &file_common_v1_common_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RetrieveEntityRes) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RetrieveEntityRes) ProtoMessage() {}

func (x *RetrieveEntityRes) ProtoReflect() protoreflect.Message {
	mi := &file_common_v1_common_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RetrieveEntityRes.ProtoReflect.Descriptor instead.
func (*RetrieveEntityRes) Descriptor() ([]byte, []int) {
	return file_common_v1_common_proto_rawDescGZIP(), []int{3}
}

func (x *RetrieveEntityRes) GetEntity() *EntityBasic {
	if x != nil {
		return x.Entity
	}
	return nil
}

type EntityBasic struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id       string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	DomainId string `protobuf:"bytes,2,opt,name=domain_id,json=domainId,proto3" json:"domain_id,omitempty"`
	Status   uint32 `protobuf:"varint,3,opt,name=status,proto3" json:"status,omitempty"`
}

func (x *EntityBasic) Reset() {
	*x = EntityBasic{}
	if protoimpl.UnsafeEnabled {
		mi := &file_common_v1_common_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EntityBasic) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EntityBasic) ProtoMessage() {}

func (x *EntityBasic) ProtoReflect() protoreflect.Message {
	mi := &file_common_v1_common_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EntityBasic.ProtoReflect.Descriptor instead.
func (*EntityBasic) Descriptor() ([]byte, []int) {
	return file_common_v1_common_proto_rawDescGZIP(), []int{4}
}

func (x *EntityBasic) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *EntityBasic) GetDomainId() string {
	if x != nil {
		return x.DomainId
	}
	return ""
}

func (x *EntityBasic) GetStatus() uint32 {
	if x != nil {
		return x.Status
	}
	return 0
}

type AddConnectionsReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Connections []*Connection `protobuf:"bytes,1,rep,name=connections,proto3" json:"connections,omitempty"`
}

func (x *AddConnectionsReq) Reset() {
	*x = AddConnectionsReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_common_v1_common_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AddConnectionsReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AddConnectionsReq) ProtoMessage() {}

func (x *AddConnectionsReq) ProtoReflect() protoreflect.Message {
	mi := &file_common_v1_common_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AddConnectionsReq.ProtoReflect.Descriptor instead.
func (*AddConnectionsReq) Descriptor() ([]byte, []int) {
	return file_common_v1_common_proto_rawDescGZIP(), []int{5}
}

func (x *AddConnectionsReq) GetConnections() []*Connection {
	if x != nil {
		return x.Connections
	}
	return nil
}

type AddConnectionsRes struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Ok bool `protobuf:"varint,1,opt,name=ok,proto3" json:"ok,omitempty"`
}

func (x *AddConnectionsRes) Reset() {
	*x = AddConnectionsRes{}
	if protoimpl.UnsafeEnabled {
		mi := &file_common_v1_common_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AddConnectionsRes) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AddConnectionsRes) ProtoMessage() {}

func (x *AddConnectionsRes) ProtoReflect() protoreflect.Message {
	mi := &file_common_v1_common_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AddConnectionsRes.ProtoReflect.Descriptor instead.
func (*AddConnectionsRes) Descriptor() ([]byte, []int) {
	return file_common_v1_common_proto_rawDescGZIP(), []int{6}
}

func (x *AddConnectionsRes) GetOk() bool {
	if x != nil {
		return x.Ok
	}
	return false
}

type RemoveConnectionsReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Connections []*Connection `protobuf:"bytes,1,rep,name=connections,proto3" json:"connections,omitempty"`
}

func (x *RemoveConnectionsReq) Reset() {
	*x = RemoveConnectionsReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_common_v1_common_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RemoveConnectionsReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RemoveConnectionsReq) ProtoMessage() {}

func (x *RemoveConnectionsReq) ProtoReflect() protoreflect.Message {
	mi := &file_common_v1_common_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RemoveConnectionsReq.ProtoReflect.Descriptor instead.
func (*RemoveConnectionsReq) Descriptor() ([]byte, []int) {
	return file_common_v1_common_proto_rawDescGZIP(), []int{7}
}

func (x *RemoveConnectionsReq) GetConnections() []*Connection {
	if x != nil {
		return x.Connections
	}
	return nil
}

type RemoveConnectionsRes struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Ok bool `protobuf:"varint,1,opt,name=ok,proto3" json:"ok,omitempty"`
}

func (x *RemoveConnectionsRes) Reset() {
	*x = RemoveConnectionsRes{}
	if protoimpl.UnsafeEnabled {
		mi := &file_common_v1_common_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RemoveConnectionsRes) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RemoveConnectionsRes) ProtoMessage() {}

func (x *RemoveConnectionsRes) ProtoReflect() protoreflect.Message {
	mi := &file_common_v1_common_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RemoveConnectionsRes.ProtoReflect.Descriptor instead.
func (*RemoveConnectionsRes) Descriptor() ([]byte, []int) {
	return file_common_v1_common_proto_rawDescGZIP(), []int{8}
}

func (x *RemoveConnectionsRes) GetOk() bool {
	if x != nil {
		return x.Ok
	}
	return false
}

type Connection struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ThingId   string `protobuf:"bytes,1,opt,name=thing_id,json=thingId,proto3" json:"thing_id,omitempty"`
	ChannelId string `protobuf:"bytes,2,opt,name=channel_id,json=channelId,proto3" json:"channel_id,omitempty"`
	DomainId  string `protobuf:"bytes,3,opt,name=domain_id,json=domainId,proto3" json:"domain_id,omitempty"`
}

func (x *Connection) Reset() {
	*x = Connection{}
	if protoimpl.UnsafeEnabled {
		mi := &file_common_v1_common_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Connection) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Connection) ProtoMessage() {}

func (x *Connection) ProtoReflect() protoreflect.Message {
	mi := &file_common_v1_common_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Connection.ProtoReflect.Descriptor instead.
func (*Connection) Descriptor() ([]byte, []int) {
	return file_common_v1_common_proto_rawDescGZIP(), []int{9}
}

func (x *Connection) GetThingId() string {
	if x != nil {
		return x.ThingId
	}
	return ""
}

func (x *Connection) GetChannelId() string {
	if x != nil {
		return x.ChannelId
	}
	return ""
}

func (x *Connection) GetDomainId() string {
	if x != nil {
		return x.DomainId
	}
	return ""
}

var File_common_v1_common_proto protoreflect.FileDescriptor

var file_common_v1_common_proto_rawDesc = []byte{
	0x0a, 0x16, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2f, 0x76, 0x31, 0x2f, 0x63, 0x6f, 0x6d, 0x6d,
	0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x09, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e,
	0x2e, 0x76, 0x31, 0x22, 0x27, 0x0a, 0x13, 0x52, 0x65, 0x74, 0x72, 0x69, 0x65, 0x76, 0x65, 0x45,
	0x6e, 0x74, 0x69, 0x74, 0x69, 0x65, 0x73, 0x52, 0x65, 0x71, 0x12, 0x10, 0x0a, 0x03, 0x69, 0x64,
	0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x09, 0x52, 0x03, 0x69, 0x64, 0x73, 0x22, 0x8d, 0x01, 0x0a,
	0x13, 0x52, 0x65, 0x74, 0x72, 0x69, 0x65, 0x76, 0x65, 0x45, 0x6e, 0x74, 0x69, 0x74, 0x69, 0x65,
	0x73, 0x52, 0x65, 0x73, 0x12, 0x14, 0x0a, 0x05, 0x74, 0x6f, 0x74, 0x61, 0x6c, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x04, 0x52, 0x05, 0x74, 0x6f, 0x74, 0x61, 0x6c, 0x12, 0x14, 0x0a, 0x05, 0x6c, 0x69,
	0x6d, 0x69, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x52, 0x05, 0x6c, 0x69, 0x6d, 0x69, 0x74,
	0x12, 0x16, 0x0a, 0x06, 0x6f, 0x66, 0x66, 0x73, 0x65, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x04,
	0x52, 0x06, 0x6f, 0x66, 0x66, 0x73, 0x65, 0x74, 0x12, 0x32, 0x0a, 0x08, 0x65, 0x6e, 0x74, 0x69,
	0x74, 0x69, 0x65, 0x73, 0x18, 0x04, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x63, 0x6f, 0x6d,
	0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x45, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x42, 0x61, 0x73,
	0x69, 0x63, 0x52, 0x08, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x69, 0x65, 0x73, 0x22, 0x23, 0x0a, 0x11,
	0x52, 0x65, 0x74, 0x72, 0x69, 0x65, 0x76, 0x65, 0x45, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x52, 0x65,
	0x71, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69,
	0x64, 0x22, 0x43, 0x0a, 0x11, 0x52, 0x65, 0x74, 0x72, 0x69, 0x65, 0x76, 0x65, 0x45, 0x6e, 0x74,
	0x69, 0x74, 0x79, 0x52, 0x65, 0x73, 0x12, 0x2e, 0x0a, 0x06, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e,
	0x76, 0x31, 0x2e, 0x45, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x42, 0x61, 0x73, 0x69, 0x63, 0x52, 0x06,
	0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x22, 0x52, 0x0a, 0x0b, 0x45, 0x6e, 0x74, 0x69, 0x74, 0x79,
	0x42, 0x61, 0x73, 0x69, 0x63, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x1b, 0x0a, 0x09, 0x64, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x5f,
	0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x64, 0x6f, 0x6d, 0x61, 0x69, 0x6e,
	0x49, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x0d, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x22, 0x4c, 0x0a, 0x11, 0x41, 0x64,
	0x64, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x52, 0x65, 0x71, 0x12,
	0x37, 0x0a, 0x0b, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x01,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x31,
	0x2e, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x0b, 0x63, 0x6f, 0x6e,
	0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x22, 0x23, 0x0a, 0x11, 0x41, 0x64, 0x64, 0x43,
	0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x52, 0x65, 0x73, 0x12, 0x0e, 0x0a,
	0x02, 0x6f, 0x6b, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x02, 0x6f, 0x6b, 0x22, 0x4f, 0x0a,
	0x14, 0x52, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f,
	0x6e, 0x73, 0x52, 0x65, 0x71, 0x12, 0x37, 0x0a, 0x0b, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74,
	0x69, 0x6f, 0x6e, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x63, 0x6f, 0x6d,
	0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f,
	0x6e, 0x52, 0x0b, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x22, 0x26,
	0x0a, 0x14, 0x52, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69,
	0x6f, 0x6e, 0x73, 0x52, 0x65, 0x73, 0x12, 0x0e, 0x0a, 0x02, 0x6f, 0x6b, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x08, 0x52, 0x02, 0x6f, 0x6b, 0x22, 0x63, 0x0a, 0x0a, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63,
	0x74, 0x69, 0x6f, 0x6e, 0x12, 0x19, 0x0a, 0x08, 0x74, 0x68, 0x69, 0x6e, 0x67, 0x5f, 0x69, 0x64,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x74, 0x68, 0x69, 0x6e, 0x67, 0x49, 0x64, 0x12,
	0x1d, 0x0a, 0x0a, 0x63, 0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x09, 0x63, 0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x49, 0x64, 0x12, 0x1b,
	0x0a, 0x09, 0x64, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x5f, 0x69, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x08, 0x64, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x49, 0x64, 0x42, 0x37, 0x5a, 0x35, 0x67,
	0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x61, 0x62, 0x73, 0x6d, 0x61, 0x63,
	0x68, 0x2f, 0x6d, 0x61, 0x67, 0x69, 0x73, 0x74, 0x72, 0x61, 0x6c, 0x61, 0x2f, 0x69, 0x6e, 0x74,
	0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x67, 0x72, 0x70, 0x63, 0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x6f,
	0x6e, 0x2f, 0x76, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_common_v1_common_proto_rawDescOnce sync.Once
	file_common_v1_common_proto_rawDescData = file_common_v1_common_proto_rawDesc
)

func file_common_v1_common_proto_rawDescGZIP() []byte {
	file_common_v1_common_proto_rawDescOnce.Do(func() {
		file_common_v1_common_proto_rawDescData = protoimpl.X.CompressGZIP(file_common_v1_common_proto_rawDescData)
	})
	return file_common_v1_common_proto_rawDescData
}

var file_common_v1_common_proto_msgTypes = make([]protoimpl.MessageInfo, 10)
var file_common_v1_common_proto_goTypes = []any{
	(*RetrieveEntitiesReq)(nil),  // 0: common.v1.RetrieveEntitiesReq
	(*RetrieveEntitiesRes)(nil),  // 1: common.v1.RetrieveEntitiesRes
	(*RetrieveEntityReq)(nil),    // 2: common.v1.RetrieveEntityReq
	(*RetrieveEntityRes)(nil),    // 3: common.v1.RetrieveEntityRes
	(*EntityBasic)(nil),          // 4: common.v1.EntityBasic
	(*AddConnectionsReq)(nil),    // 5: common.v1.AddConnectionsReq
	(*AddConnectionsRes)(nil),    // 6: common.v1.AddConnectionsRes
	(*RemoveConnectionsReq)(nil), // 7: common.v1.RemoveConnectionsReq
	(*RemoveConnectionsRes)(nil), // 8: common.v1.RemoveConnectionsRes
	(*Connection)(nil),           // 9: common.v1.Connection
}
var file_common_v1_common_proto_depIdxs = []int32{
	4, // 0: common.v1.RetrieveEntitiesRes.entities:type_name -> common.v1.EntityBasic
	4, // 1: common.v1.RetrieveEntityRes.entity:type_name -> common.v1.EntityBasic
	9, // 2: common.v1.AddConnectionsReq.connections:type_name -> common.v1.Connection
	9, // 3: common.v1.RemoveConnectionsReq.connections:type_name -> common.v1.Connection
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_common_v1_common_proto_init() }
func file_common_v1_common_proto_init() {
	if File_common_v1_common_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_common_v1_common_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*RetrieveEntitiesReq); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_common_v1_common_proto_msgTypes[1].Exporter = func(v any, i int) any {
			switch v := v.(*RetrieveEntitiesRes); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_common_v1_common_proto_msgTypes[2].Exporter = func(v any, i int) any {
			switch v := v.(*RetrieveEntityReq); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_common_v1_common_proto_msgTypes[3].Exporter = func(v any, i int) any {
			switch v := v.(*RetrieveEntityRes); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_common_v1_common_proto_msgTypes[4].Exporter = func(v any, i int) any {
			switch v := v.(*EntityBasic); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_common_v1_common_proto_msgTypes[5].Exporter = func(v any, i int) any {
			switch v := v.(*AddConnectionsReq); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_common_v1_common_proto_msgTypes[6].Exporter = func(v any, i int) any {
			switch v := v.(*AddConnectionsRes); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_common_v1_common_proto_msgTypes[7].Exporter = func(v any, i int) any {
			switch v := v.(*RemoveConnectionsReq); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_common_v1_common_proto_msgTypes[8].Exporter = func(v any, i int) any {
			switch v := v.(*RemoveConnectionsRes); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_common_v1_common_proto_msgTypes[9].Exporter = func(v any, i int) any {
			switch v := v.(*Connection); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_common_v1_common_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   10,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_common_v1_common_proto_goTypes,
		DependencyIndexes: file_common_v1_common_proto_depIdxs,
		MessageInfos:      file_common_v1_common_proto_msgTypes,
	}.Build()
	File_common_v1_common_proto = out.File
	file_common_v1_common_proto_rawDesc = nil
	file_common_v1_common_proto_goTypes = nil
	file_common_v1_common_proto_depIdxs = nil
}
