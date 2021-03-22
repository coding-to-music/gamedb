// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.15.6
// source: github.proto

package generated

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type CommitsRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Pagination *PaginationRequest2 `protobuf:"bytes,1,opt,name=pagination,proto3" json:"pagination,omitempty"`
}

func (x *CommitsRequest) Reset() {
	*x = CommitsRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CommitsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CommitsRequest) ProtoMessage() {}

func (x *CommitsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_github_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CommitsRequest.ProtoReflect.Descriptor instead.
func (*CommitsRequest) Descriptor() ([]byte, []int) {
	return file_github_proto_rawDescGZIP(), []int{0}
}

func (x *CommitsRequest) GetPagination() *PaginationRequest2 {
	if x != nil {
		return x.Pagination
	}
	return nil
}

type CommitsResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Commits []*CommitResponse `protobuf:"bytes,1,rep,name=commits,proto3" json:"commits,omitempty"`
}

func (x *CommitsResponse) Reset() {
	*x = CommitsResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CommitsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CommitsResponse) ProtoMessage() {}

func (x *CommitsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_github_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CommitsResponse.ProtoReflect.Descriptor instead.
func (*CommitsResponse) Descriptor() ([]byte, []int) {
	return file_github_proto_rawDescGZIP(), []int{1}
}

func (x *CommitsResponse) GetCommits() []*CommitResponse {
	if x != nil {
		return x.Commits
	}
	return nil
}

type CommitResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Message string                 `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
	Time    *timestamppb.Timestamp `protobuf:"bytes,2,opt,name=time,proto3" json:"time,omitempty"`
	Link    string                 `protobuf:"bytes,3,opt,name=link,proto3" json:"link,omitempty"`
	Hash    string                 `protobuf:"bytes,4,opt,name=hash,proto3" json:"hash,omitempty"`
}

func (x *CommitResponse) Reset() {
	*x = CommitResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CommitResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CommitResponse) ProtoMessage() {}

func (x *CommitResponse) ProtoReflect() protoreflect.Message {
	mi := &file_github_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CommitResponse.ProtoReflect.Descriptor instead.
func (*CommitResponse) Descriptor() ([]byte, []int) {
	return file_github_proto_rawDescGZIP(), []int{2}
}

func (x *CommitResponse) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

func (x *CommitResponse) GetTime() *timestamppb.Timestamp {
	if x != nil {
		return x.Time
	}
	return nil
}

func (x *CommitResponse) GetLink() string {
	if x != nil {
		return x.Link
	}
	return ""
}

func (x *CommitResponse) GetHash() string {
	if x != nil {
		return x.Hash
	}
	return ""
}

var File_github_proto protoreflect.FileDescriptor

var file_github_proto_rawDesc = []byte{
	0x0a, 0x0c, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x09,
	0x67, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x65, 0x64, 0x1a, 0x0c, 0x73, 0x68, 0x61, 0x72, 0x65,
	0x64, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61,
	0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x4f, 0x0a, 0x0e, 0x43, 0x6f, 0x6d, 0x6d,
	0x69, 0x74, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x3d, 0x0a, 0x0a, 0x70, 0x61,
	0x67, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1d,
	0x2e, 0x67, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x65, 0x64, 0x2e, 0x50, 0x61, 0x67, 0x69, 0x6e,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x32, 0x52, 0x0a, 0x70,
	0x61, 0x67, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x22, 0x46, 0x0a, 0x0f, 0x43, 0x6f, 0x6d,
	0x6d, 0x69, 0x74, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x33, 0x0a, 0x07,
	0x63, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x19, 0x2e,
	0x67, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x65, 0x64, 0x2e, 0x43, 0x6f, 0x6d, 0x6d, 0x69, 0x74,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x52, 0x07, 0x63, 0x6f, 0x6d, 0x6d, 0x69, 0x74,
	0x73, 0x22, 0x82, 0x01, 0x0a, 0x0e, 0x43, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x2e,
	0x0a, 0x04, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54,
	0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x04, 0x74, 0x69, 0x6d, 0x65, 0x12, 0x12,
	0x0a, 0x04, 0x6c, 0x69, 0x6e, 0x6b, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6c, 0x69,
	0x6e, 0x6b, 0x12, 0x12, 0x0a, 0x04, 0x68, 0x61, 0x73, 0x68, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x04, 0x68, 0x61, 0x73, 0x68, 0x32, 0x53, 0x0a, 0x0d, 0x47, 0x69, 0x74, 0x48, 0x75, 0x62,
	0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x42, 0x0a, 0x07, 0x43, 0x6f, 0x6d, 0x6d, 0x69,
	0x74, 0x73, 0x12, 0x19, 0x2e, 0x67, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x65, 0x64, 0x2e, 0x43,
	0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1a, 0x2e,
	0x67, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x65, 0x64, 0x2e, 0x43, 0x6f, 0x6d, 0x6d, 0x69, 0x74,
	0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0x30, 0x5a, 0x2e, 0x67,
	0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x67, 0x61, 0x6d, 0x65, 0x64, 0x62,
	0x2f, 0x67, 0x61, 0x6d, 0x65, 0x64, 0x62, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x62, 0x61, 0x63, 0x6b,
	0x65, 0x6e, 0x64, 0x2f, 0x67, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x65, 0x64, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_github_proto_rawDescOnce sync.Once
	file_github_proto_rawDescData = file_github_proto_rawDesc
)

func file_github_proto_rawDescGZIP() []byte {
	file_github_proto_rawDescOnce.Do(func() {
		file_github_proto_rawDescData = protoimpl.X.CompressGZIP(file_github_proto_rawDescData)
	})
	return file_github_proto_rawDescData
}

var file_github_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_github_proto_goTypes = []interface{}{
	(*CommitsRequest)(nil),        // 0: generated.CommitsRequest
	(*CommitsResponse)(nil),       // 1: generated.CommitsResponse
	(*CommitResponse)(nil),        // 2: generated.CommitResponse
	(*PaginationRequest2)(nil),    // 3: generated.PaginationRequest2
	(*timestamppb.Timestamp)(nil), // 4: google.protobuf.Timestamp
}
var file_github_proto_depIdxs = []int32{
	3, // 0: generated.CommitsRequest.pagination:type_name -> generated.PaginationRequest2
	2, // 1: generated.CommitsResponse.commits:type_name -> generated.CommitResponse
	4, // 2: generated.CommitResponse.time:type_name -> google.protobuf.Timestamp
	0, // 3: generated.GitHubService.Commits:input_type -> generated.CommitsRequest
	1, // 4: generated.GitHubService.Commits:output_type -> generated.CommitsResponse
	4, // [4:5] is the sub-list for method output_type
	3, // [3:4] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_github_proto_init() }
func file_github_proto_init() {
	if File_github_proto != nil {
		return
	}
	file_shared_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_github_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CommitsRequest); i {
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
		file_github_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CommitsResponse); i {
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
		file_github_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CommitResponse); i {
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
			RawDescriptor: file_github_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_github_proto_goTypes,
		DependencyIndexes: file_github_proto_depIdxs,
		MessageInfos:      file_github_proto_msgTypes,
	}.Build()
	File_github_proto = out.File
	file_github_proto_rawDesc = nil
	file_github_proto_goTypes = nil
	file_github_proto_depIdxs = nil
}
