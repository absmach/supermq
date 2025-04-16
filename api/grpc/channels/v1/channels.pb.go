// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v6.30.2
// source: channels/v1/channels.proto

package v1

import (
	v1 "github.com/absmach/supermq/api/grpc/common/v1"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type RemoveClientConnectionsReq struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	ClientId      string                 `protobuf:"bytes,1,opt,name=client_id,json=clientId,proto3" json:"client_id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
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
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
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
	state         protoimpl.MessageState `protogen:"open.v1"`
	ParentGroupId string                 `protobuf:"bytes,1,opt,name=parent_group_id,json=parentGroupId,proto3" json:"parent_group_id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
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
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
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
	state         protoimpl.MessageState `protogen:"open.v1"`
	DomainId      string                 `protobuf:"bytes,1,opt,name=domain_id,json=domainId,proto3" json:"domain_id,omitempty"`
	ClientId      string                 `protobuf:"bytes,2,opt,name=client_id,json=clientId,proto3" json:"client_id,omitempty"`
	ClientType    string                 `protobuf:"bytes,3,opt,name=client_type,json=clientType,proto3" json:"client_type,omitempty"`
	ChannelId     string                 `protobuf:"bytes,4,opt,name=channel_id,json=channelId,proto3" json:"channel_id,omitempty"`
	Type          uint32                 `protobuf:"varint,5,opt,name=type,proto3" json:"type,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
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
	state         protoimpl.MessageState `protogen:"open.v1"`
	Authorized    bool                   `protobuf:"varint,1,opt,name=authorized,proto3" json:"authorized,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
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

const file_channels_v1_channels_proto_rawDesc = "" +
	"\n" +
	"\x1achannels/v1/channels.proto\x12\vchannels.v1\x1a\x16common/v1/common.proto\"9\n" +
	"\x1aRemoveClientConnectionsReq\x12\x1b\n" +
	"\tclient_id\x18\x01 \x01(\tR\bclientId\"\x1c\n" +
	"\x1aRemoveClientConnectionsRes\"I\n" +
	"\x1fUnsetParentGroupFromChannelsReq\x12&\n" +
	"\x0fparent_group_id\x18\x01 \x01(\tR\rparentGroupId\"!\n" +
	"\x1fUnsetParentGroupFromChannelsRes\"\x98\x01\n" +
	"\bAuthzReq\x12\x1b\n" +
	"\tdomain_id\x18\x01 \x01(\tR\bdomainId\x12\x1b\n" +
	"\tclient_id\x18\x02 \x01(\tR\bclientId\x12\x1f\n" +
	"\vclient_type\x18\x03 \x01(\tR\n" +
	"clientType\x12\x1d\n" +
	"\n" +
	"channel_id\x18\x04 \x01(\tR\tchannelId\x12\x12\n" +
	"\x04type\x18\x05 \x01(\rR\x04type\"*\n" +
	"\bAuthzRes\x12\x1e\n" +
	"\n" +
	"authorized\x18\x01 \x01(\bR\n" +
	"authorized2\xdd\x03\n" +
	"\x0fChannelsService\x12;\n" +
	"\tAuthorize\x12\x15.channels.v1.AuthzReq\x1a\x15.channels.v1.AuthzRes\"\x00\x12m\n" +
	"\x17RemoveClientConnections\x12'.channels.v1.RemoveClientConnectionsReq\x1a'.channels.v1.RemoveClientConnectionsRes\"\x00\x12|\n" +
	"\x1cUnsetParentGroupFromChannels\x12,.channels.v1.UnsetParentGroupFromChannelsReq\x1a,.channels.v1.UnsetParentGroupFromChannelsRes\"\x00\x12N\n" +
	"\x0eRetrieveEntity\x12\x1c.common.v1.RetrieveEntityReq\x1a\x1c.common.v1.RetrieveEntityRes\"\x00\x12P\n" +
	"\x0fRetrieveByRoute\x12\x1d.common.v1.RetrieveByRouteReq\x1a\x1c.common.v1.RetrieveEntityRes\"\x00B1Z/github.com/absmach/supermq/api/grpc/channels/v1b\x06proto3"

var (
	file_channels_v1_channels_proto_rawDescOnce sync.Once
	file_channels_v1_channels_proto_rawDescData []byte
)

func file_channels_v1_channels_proto_rawDescGZIP() []byte {
	file_channels_v1_channels_proto_rawDescOnce.Do(func() {
		file_channels_v1_channels_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_channels_v1_channels_proto_rawDesc), len(file_channels_v1_channels_proto_rawDesc)))
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
	(*v1.RetrieveByRouteReq)(nil),           // 7: common.v1.RetrieveByRouteReq
	(*v1.RetrieveEntityRes)(nil),            // 8: common.v1.RetrieveEntityRes
}
var file_channels_v1_channels_proto_depIdxs = []int32{
	4, // 0: channels.v1.ChannelsService.Authorize:input_type -> channels.v1.AuthzReq
	0, // 1: channels.v1.ChannelsService.RemoveClientConnections:input_type -> channels.v1.RemoveClientConnectionsReq
	2, // 2: channels.v1.ChannelsService.UnsetParentGroupFromChannels:input_type -> channels.v1.UnsetParentGroupFromChannelsReq
	6, // 3: channels.v1.ChannelsService.RetrieveEntity:input_type -> common.v1.RetrieveEntityReq
	7, // 4: channels.v1.ChannelsService.RetrieveByRoute:input_type -> common.v1.RetrieveByRouteReq
	5, // 5: channels.v1.ChannelsService.Authorize:output_type -> channels.v1.AuthzRes
	1, // 6: channels.v1.ChannelsService.RemoveClientConnections:output_type -> channels.v1.RemoveClientConnectionsRes
	3, // 7: channels.v1.ChannelsService.UnsetParentGroupFromChannels:output_type -> channels.v1.UnsetParentGroupFromChannelsRes
	8, // 8: channels.v1.ChannelsService.RetrieveEntity:output_type -> common.v1.RetrieveEntityRes
	8, // 9: channels.v1.ChannelsService.RetrieveByRoute:output_type -> common.v1.RetrieveEntityRes
	5, // [5:10] is the sub-list for method output_type
	0, // [0:5] is the sub-list for method input_type
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
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_channels_v1_channels_proto_rawDesc), len(file_channels_v1_channels_proto_rawDesc)),
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
	file_channels_v1_channels_proto_goTypes = nil
	file_channels_v1_channels_proto_depIdxs = nil
}
