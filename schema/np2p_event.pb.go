// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.32.0
// 	protoc        v3.19.1
// source: np2p_event.proto

package schema

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

type Tag struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Tag [][]byte `protobuf:"bytes,1,rep,name=tag,proto3" json:"tag,omitempty"`
}

func (x *Tag) Reset() {
	*x = Tag{}
	if protoimpl.UnsafeEnabled {
		mi := &file_np2p_event_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Tag) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Tag) ProtoMessage() {}

func (x *Tag) ProtoReflect() protoreflect.Message {
	mi := &file_np2p_event_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Tag.ProtoReflect.Descriptor instead.
func (*Tag) Descriptor() ([]byte, []int) {
	return file_np2p_event_proto_rawDescGZIP(), []int{0}
}

func (x *Tag) GetTag() [][]byte {
	if x != nil {
		return x.Tag
	}
	return nil
}

type Np2PEventPB struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id        []byte `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Pubkey    []byte `protobuf:"bytes,2,opt,name=pubkey,proto3" json:"pubkey,omitempty"`
	CreatedAt uint64 `protobuf:"varint,3,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	Kind      uint32 `protobuf:"varint,4,opt,name=kind,proto3" json:"kind,omitempty"`
	Tags      []*Tag `protobuf:"bytes,5,rep,name=tags,proto3" json:"tags,omitempty"`
	Content   string `protobuf:"bytes,6,opt,name=content,proto3" json:"content,omitempty"`
	Sig       []byte `protobuf:"bytes,7,opt,name=sig,proto3,oneof" json:"sig,omitempty"`
}

func (x *Np2PEventPB) Reset() {
	*x = Np2PEventPB{}
	if protoimpl.UnsafeEnabled {
		mi := &file_np2p_event_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Np2PEventPB) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Np2PEventPB) ProtoMessage() {}

func (x *Np2PEventPB) ProtoReflect() protoreflect.Message {
	mi := &file_np2p_event_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Np2PEventPB.ProtoReflect.Descriptor instead.
func (*Np2PEventPB) Descriptor() ([]byte, []int) {
	return file_np2p_event_proto_rawDescGZIP(), []int{1}
}

func (x *Np2PEventPB) GetId() []byte {
	if x != nil {
		return x.Id
	}
	return nil
}

func (x *Np2PEventPB) GetPubkey() []byte {
	if x != nil {
		return x.Pubkey
	}
	return nil
}

func (x *Np2PEventPB) GetCreatedAt() uint64 {
	if x != nil {
		return x.CreatedAt
	}
	return 0
}

func (x *Np2PEventPB) GetKind() uint32 {
	if x != nil {
		return x.Kind
	}
	return 0
}

func (x *Np2PEventPB) GetTags() []*Tag {
	if x != nil {
		return x.Tags
	}
	return nil
}

func (x *Np2PEventPB) GetContent() string {
	if x != nil {
		return x.Content
	}
	return ""
}

func (x *Np2PEventPB) GetSig() []byte {
	if x != nil {
		return x.Sig
	}
	return nil
}

var File_np2p_event_proto protoreflect.FileDescriptor

var file_np2p_event_proto_rawDesc = []byte{
	0x0a, 0x10, 0x6e, 0x70, 0x32, 0x70, 0x5f, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x12, 0x06, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x22, 0x17, 0x0a, 0x03, 0x54, 0x61,
	0x67, 0x12, 0x10, 0x0a, 0x03, 0x74, 0x61, 0x67, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0c, 0x52, 0x03,
	0x74, 0x61, 0x67, 0x22, 0xc2, 0x01, 0x0a, 0x0b, 0x4e, 0x70, 0x32, 0x70, 0x45, 0x76, 0x65, 0x6e,
	0x74, 0x50, 0x42, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52,
	0x02, 0x69, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x70, 0x75, 0x62, 0x6b, 0x65, 0x79, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0c, 0x52, 0x06, 0x70, 0x75, 0x62, 0x6b, 0x65, 0x79, 0x12, 0x1d, 0x0a, 0x0a, 0x63,
	0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x5f, 0x61, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x04, 0x52,
	0x09, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x41, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x6b, 0x69,
	0x6e, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x04, 0x6b, 0x69, 0x6e, 0x64, 0x12, 0x1f,
	0x0a, 0x04, 0x74, 0x61, 0x67, 0x73, 0x18, 0x05, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0b, 0x2e, 0x73,
	0x63, 0x68, 0x65, 0x6d, 0x61, 0x2e, 0x54, 0x61, 0x67, 0x52, 0x04, 0x74, 0x61, 0x67, 0x73, 0x12,
	0x18, 0x0a, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x12, 0x15, 0x0a, 0x03, 0x73, 0x69, 0x67,
	0x18, 0x07, 0x20, 0x01, 0x28, 0x0c, 0x48, 0x00, 0x52, 0x03, 0x73, 0x69, 0x67, 0x88, 0x01, 0x01,
	0x42, 0x06, 0x0a, 0x04, 0x5f, 0x73, 0x69, 0x67, 0x42, 0x09, 0x5a, 0x07, 0x73, 0x63, 0x68, 0x65,
	0x6d, 0x61, 0x2f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_np2p_event_proto_rawDescOnce sync.Once
	file_np2p_event_proto_rawDescData = file_np2p_event_proto_rawDesc
)

func file_np2p_event_proto_rawDescGZIP() []byte {
	file_np2p_event_proto_rawDescOnce.Do(func() {
		file_np2p_event_proto_rawDescData = protoimpl.X.CompressGZIP(file_np2p_event_proto_rawDescData)
	})
	return file_np2p_event_proto_rawDescData
}

var file_np2p_event_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_np2p_event_proto_goTypes = []interface{}{
	(*Tag)(nil),         // 0: schema.Tag
	(*Np2PEventPB)(nil), // 1: schema.Np2pEventPB
}
var file_np2p_event_proto_depIdxs = []int32{
	0, // 0: schema.Np2pEventPB.tags:type_name -> schema.Tag
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_np2p_event_proto_init() }
func file_np2p_event_proto_init() {
	if File_np2p_event_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_np2p_event_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Tag); i {
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
		file_np2p_event_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Np2PEventPB); i {
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
	file_np2p_event_proto_msgTypes[1].OneofWrappers = []interface{}{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_np2p_event_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_np2p_event_proto_goTypes,
		DependencyIndexes: file_np2p_event_proto_depIdxs,
		MessageInfos:      file_np2p_event_proto_msgTypes,
	}.Build()
	File_np2p_event_proto = out.File
	file_np2p_event_proto_rawDesc = nil
	file_np2p_event_proto_goTypes = nil
	file_np2p_event_proto_depIdxs = nil
}