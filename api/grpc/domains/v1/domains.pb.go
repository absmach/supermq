// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.0
// 	protoc        v5.29.0
// source: domains/v1/domains.proto

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

type DeleteUserRes struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Deleted       bool                   `protobuf:"varint,1,opt,name=deleted,proto3" json:"deleted,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *DeleteUserRes) Reset() {
	*x = DeleteUserRes{}
	mi := &file_domains_v1_domains_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *DeleteUserRes) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeleteUserRes) ProtoMessage() {}

func (x *DeleteUserRes) ProtoReflect() protoreflect.Message {
	mi := &file_domains_v1_domains_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeleteUserRes.ProtoReflect.Descriptor instead.
func (*DeleteUserRes) Descriptor() ([]byte, []int) {
	return file_domains_v1_domains_proto_rawDescGZIP(), []int{0}
}

func (x *DeleteUserRes) GetDeleted() bool {
	if x != nil {
		return x.Deleted
	}
	return false
}

type DeleteUserReq struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Id            string                 `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *DeleteUserReq) Reset() {
	*x = DeleteUserReq{}
	mi := &file_domains_v1_domains_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *DeleteUserReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeleteUserReq) ProtoMessage() {}

func (x *DeleteUserReq) ProtoReflect() protoreflect.Message {
	mi := &file_domains_v1_domains_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeleteUserReq.ProtoReflect.Descriptor instead.
func (*DeleteUserReq) Descriptor() ([]byte, []int) {
	return file_domains_v1_domains_proto_rawDescGZIP(), []int{1}
}

func (x *DeleteUserReq) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

var File_domains_v1_domains_proto protoreflect.FileDescriptor

var file_domains_v1_domains_proto_rawDesc = []byte{
	0x0a, 0x18, 0x64, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x73, 0x2f, 0x76, 0x31, 0x2f, 0x64, 0x6f, 0x6d,
	0x61, 0x69, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0a, 0x64, 0x6f, 0x6d, 0x61,
	0x69, 0x6e, 0x73, 0x2e, 0x76, 0x31, 0x22, 0x29, 0x0a, 0x0d, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65,
	0x55, 0x73, 0x65, 0x72, 0x52, 0x65, 0x73, 0x12, 0x18, 0x0a, 0x07, 0x64, 0x65, 0x6c, 0x65, 0x74,
	0x65, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x64, 0x65, 0x6c, 0x65, 0x74, 0x65,
	0x64, 0x22, 0x1f, 0x0a, 0x0d, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x55, 0x73, 0x65, 0x72, 0x52,
	0x65, 0x71, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02,
	0x69, 0x64, 0x32, 0x61, 0x0a, 0x0e, 0x44, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x73, 0x53, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x12, 0x4f, 0x0a, 0x15, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x55, 0x73,
	0x65, 0x72, 0x46, 0x72, 0x6f, 0x6d, 0x44, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x73, 0x12, 0x19, 0x2e,
	0x64, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x44, 0x65, 0x6c, 0x65, 0x74,
	0x65, 0x55, 0x73, 0x65, 0x72, 0x52, 0x65, 0x71, 0x1a, 0x19, 0x2e, 0x64, 0x6f, 0x6d, 0x61, 0x69,
	0x6e, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x55, 0x73, 0x65, 0x72,
	0x52, 0x65, 0x73, 0x22, 0x00, 0x42, 0x30, 0x5a, 0x2e, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e,
	0x63, 0x6f, 0x6d, 0x2f, 0x61, 0x62, 0x73, 0x6d, 0x61, 0x63, 0x68, 0x2f, 0x73, 0x75, 0x70, 0x65,
	0x72, 0x6d, 0x71, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x67, 0x72, 0x70, 0x63, 0x2f, 0x64, 0x6f, 0x6d,
	0x61, 0x69, 0x6e, 0x73, 0x2f, 0x76, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_domains_v1_domains_proto_rawDescOnce sync.Once
	file_domains_v1_domains_proto_rawDescData = file_domains_v1_domains_proto_rawDesc
)

func file_domains_v1_domains_proto_rawDescGZIP() []byte {
	file_domains_v1_domains_proto_rawDescOnce.Do(func() {
		file_domains_v1_domains_proto_rawDescData = protoimpl.X.CompressGZIP(file_domains_v1_domains_proto_rawDescData)
	})
	return file_domains_v1_domains_proto_rawDescData
}

var file_domains_v1_domains_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_domains_v1_domains_proto_goTypes = []any{
	(*DeleteUserRes)(nil),        // 0: domains.v1.DeleteUserRes
	(*DeleteUserReq)(nil),        // 1: domains.v1.DeleteUserReq
	(*v1.RetrieveEntityReq)(nil), // 2: common.v1.RetrieveEntityReq
	(*v1.RetrieveEntityRes)(nil), // 3: common.v1.RetrieveEntityRes
}
var file_domains_v1_domains_proto_depIdxs = []int32{
	1, // 0: domains.v1.DomainsService.DeleteUserFromDomains:input_type -> domains.v1.DeleteUserReq
	2, // 1: domains.v1.DomainsService.RetrieveEntity:input_type -> common.v1.RetrieveEntityReq
	0, // 2: domains.v1.DomainsService.DeleteUserFromDomains:output_type -> domains.v1.DeleteUserRes
	3, // 3: domains.v1.DomainsService.RetrieveEntity:output_type -> common.v1.RetrieveEntityRes
	2, // [2:4] is the sub-list for method output_type
	0, // [0:2] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_domains_v1_domains_proto_init() }
func file_domains_v1_domains_proto_init() {
	if File_domains_v1_domains_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_domains_v1_domains_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_domains_v1_domains_proto_goTypes,
		DependencyIndexes: file_domains_v1_domains_proto_depIdxs,
		MessageInfos:      file_domains_v1_domains_proto_msgTypes,
	}.Build()
	File_domains_v1_domains_proto = out.File
	file_domains_v1_domains_proto_rawDesc = nil
	file_domains_v1_domains_proto_goTypes = nil
	file_domains_v1_domains_proto_depIdxs = nil
}
