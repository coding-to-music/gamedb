// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.14.0
// source: articles.proto

package generated

import (
	proto "github.com/golang/protobuf/proto"
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

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

type ListArticlesRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Pagination *PaginationRequest `protobuf:"bytes,1,opt,name=pagination,proto3" json:"pagination,omitempty"`
	Ids        []int32            `protobuf:"varint,2,rep,packed,name=ids,proto3" json:"ids,omitempty"`
	FeedLabel  []int32            `protobuf:"varint,3,rep,packed,name=feedLabel,proto3" json:"feedLabel,omitempty"`
}

func (x *ListArticlesRequest) Reset() {
	*x = ListArticlesRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_articles_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListArticlesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListArticlesRequest) ProtoMessage() {}

func (x *ListArticlesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_articles_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListArticlesRequest.ProtoReflect.Descriptor instead.
func (*ListArticlesRequest) Descriptor() ([]byte, []int) {
	return file_articles_proto_rawDescGZIP(), []int{0}
}

func (x *ListArticlesRequest) GetPagination() *PaginationRequest {
	if x != nil {
		return x.Pagination
	}
	return nil
}

func (x *ListArticlesRequest) GetIds() []int32 {
	if x != nil {
		return x.Ids
	}
	return nil
}

func (x *ListArticlesRequest) GetFeedLabel() []int32 {
	if x != nil {
		return x.FeedLabel
	}
	return nil
}

type ArticlesResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Pagination *PaginationResponse `protobuf:"bytes,1,opt,name=pagination,proto3" json:"pagination,omitempty"`
	Articles   []*ArticleResponse  `protobuf:"bytes,2,rep,name=articles,proto3" json:"articles,omitempty"`
}

func (x *ArticlesResponse) Reset() {
	*x = ArticlesResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_articles_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ArticlesResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ArticlesResponse) ProtoMessage() {}

func (x *ArticlesResponse) ProtoReflect() protoreflect.Message {
	mi := &file_articles_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ArticlesResponse.ProtoReflect.Descriptor instead.
func (*ArticlesResponse) Descriptor() ([]byte, []int) {
	return file_articles_proto_rawDescGZIP(), []int{1}
}

func (x *ArticlesResponse) GetPagination() *PaginationResponse {
	if x != nil {
		return x.Pagination
	}
	return nil
}

func (x *ArticlesResponse) GetArticles() []*ArticleResponse {
	if x != nil {
		return x.Articles
	}
	return nil
}

type ArticleResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id          int32                  `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Title       string                 `protobuf:"bytes,2,opt,name=title,proto3" json:"title,omitempty"`
	Url         string                 `protobuf:"bytes,3,opt,name=url,proto3" json:"url,omitempty"`
	IsExternal  string                 `protobuf:"bytes,4,opt,name=isExternal,proto3" json:"isExternal,omitempty"`
	Author      string                 `protobuf:"bytes,5,opt,name=author,proto3" json:"author,omitempty"`
	Contents    string                 `protobuf:"bytes,6,opt,name=contents,proto3" json:"contents,omitempty"`
	Date        *timestamppb.Timestamp `protobuf:"bytes,7,opt,name=date,proto3" json:"date,omitempty"`
	FeedLabel   string                 `protobuf:"bytes,8,opt,name=feedLabel,proto3" json:"feedLabel,omitempty"`
	FeedName    string                 `protobuf:"bytes,9,opt,name=feedName,proto3" json:"feedName,omitempty"`
	FeedType    string                 `protobuf:"bytes,10,opt,name=feedType,proto3" json:"feedType,omitempty"`
	AppID       string                 `protobuf:"bytes,11,opt,name=appID,proto3" json:"appID,omitempty"`
	AppName     string                 `protobuf:"bytes,12,opt,name=appName,proto3" json:"appName,omitempty"`
	AppIcon     string                 `protobuf:"bytes,13,opt,name=appIcon,proto3" json:"appIcon,omitempty"`
	ArticleIcon string                 `protobuf:"bytes,14,opt,name=articleIcon,proto3" json:"articleIcon,omitempty"`
}

func (x *ArticleResponse) Reset() {
	*x = ArticleResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_articles_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ArticleResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ArticleResponse) ProtoMessage() {}

func (x *ArticleResponse) ProtoReflect() protoreflect.Message {
	mi := &file_articles_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ArticleResponse.ProtoReflect.Descriptor instead.
func (*ArticleResponse) Descriptor() ([]byte, []int) {
	return file_articles_proto_rawDescGZIP(), []int{2}
}

func (x *ArticleResponse) GetId() int32 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *ArticleResponse) GetTitle() string {
	if x != nil {
		return x.Title
	}
	return ""
}

func (x *ArticleResponse) GetUrl() string {
	if x != nil {
		return x.Url
	}
	return ""
}

func (x *ArticleResponse) GetIsExternal() string {
	if x != nil {
		return x.IsExternal
	}
	return ""
}

func (x *ArticleResponse) GetAuthor() string {
	if x != nil {
		return x.Author
	}
	return ""
}

func (x *ArticleResponse) GetContents() string {
	if x != nil {
		return x.Contents
	}
	return ""
}

func (x *ArticleResponse) GetDate() *timestamppb.Timestamp {
	if x != nil {
		return x.Date
	}
	return nil
}

func (x *ArticleResponse) GetFeedLabel() string {
	if x != nil {
		return x.FeedLabel
	}
	return ""
}

func (x *ArticleResponse) GetFeedName() string {
	if x != nil {
		return x.FeedName
	}
	return ""
}

func (x *ArticleResponse) GetFeedType() string {
	if x != nil {
		return x.FeedType
	}
	return ""
}

func (x *ArticleResponse) GetAppID() string {
	if x != nil {
		return x.AppID
	}
	return ""
}

func (x *ArticleResponse) GetAppName() string {
	if x != nil {
		return x.AppName
	}
	return ""
}

func (x *ArticleResponse) GetAppIcon() string {
	if x != nil {
		return x.AppIcon
	}
	return ""
}

func (x *ArticleResponse) GetArticleIcon() string {
	if x != nil {
		return x.ArticleIcon
	}
	return ""
}

var File_articles_proto protoreflect.FileDescriptor

var file_articles_proto_rawDesc = []byte{
	0x0a, 0x0e, 0x61, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x09, 0x67, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x65, 0x64, 0x1a, 0x0c, 0x73, 0x68, 0x61,
	0x72, 0x65, 0x64, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73,
	0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x83, 0x01, 0x0a, 0x13, 0x4c,
	0x69, 0x73, 0x74, 0x41, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x12, 0x3c, 0x0a, 0x0a, 0x70, 0x61, 0x67, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x67, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74,
	0x65, 0x64, 0x2e, 0x50, 0x61, 0x67, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x52, 0x0a, 0x70, 0x61, 0x67, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x12, 0x10, 0x0a, 0x03, 0x69, 0x64, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x05, 0x52, 0x03, 0x69,
	0x64, 0x73, 0x12, 0x1c, 0x0a, 0x09, 0x66, 0x65, 0x65, 0x64, 0x4c, 0x61, 0x62, 0x65, 0x6c, 0x18,
	0x03, 0x20, 0x03, 0x28, 0x05, 0x52, 0x09, 0x66, 0x65, 0x65, 0x64, 0x4c, 0x61, 0x62, 0x65, 0x6c,
	0x22, 0x89, 0x01, 0x0a, 0x10, 0x41, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x73, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x3d, 0x0a, 0x0a, 0x70, 0x61, 0x67, 0x69, 0x6e, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x67, 0x65, 0x6e, 0x65,
	0x72, 0x61, 0x74, 0x65, 0x64, 0x2e, 0x50, 0x61, 0x67, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x52, 0x0a, 0x70, 0x61, 0x67, 0x69, 0x6e, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x12, 0x36, 0x0a, 0x08, 0x61, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x73,
	0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74,
	0x65, 0x64, 0x2e, 0x41, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x52, 0x08, 0x61, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x73, 0x22, 0x8f, 0x03, 0x0a,
	0x0f, 0x41, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x02, 0x69, 0x64,
	0x12, 0x14, 0x0a, 0x05, 0x74, 0x69, 0x74, 0x6c, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x05, 0x74, 0x69, 0x74, 0x6c, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x75, 0x72, 0x6c, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x03, 0x75, 0x72, 0x6c, 0x12, 0x1e, 0x0a, 0x0a, 0x69, 0x73, 0x45, 0x78,
	0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x69, 0x73,
	0x45, 0x78, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x12, 0x16, 0x0a, 0x06, 0x61, 0x75, 0x74, 0x68,
	0x6f, 0x72, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x61, 0x75, 0x74, 0x68, 0x6f, 0x72,
	0x12, 0x1a, 0x0a, 0x08, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x18, 0x06, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x08, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x12, 0x2e, 0x0a, 0x04,
	0x64, 0x61, 0x74, 0x65, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d,
	0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x04, 0x64, 0x61, 0x74, 0x65, 0x12, 0x1c, 0x0a, 0x09,
	0x66, 0x65, 0x65, 0x64, 0x4c, 0x61, 0x62, 0x65, 0x6c, 0x18, 0x08, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x09, 0x66, 0x65, 0x65, 0x64, 0x4c, 0x61, 0x62, 0x65, 0x6c, 0x12, 0x1a, 0x0a, 0x08, 0x66, 0x65,
	0x65, 0x64, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x09, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x66, 0x65,
	0x65, 0x64, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x66, 0x65, 0x65, 0x64, 0x54, 0x79,
	0x70, 0x65, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x66, 0x65, 0x65, 0x64, 0x54, 0x79,
	0x70, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x61, 0x70, 0x70, 0x49, 0x44, 0x18, 0x0b, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x05, 0x61, 0x70, 0x70, 0x49, 0x44, 0x12, 0x18, 0x0a, 0x07, 0x61, 0x70, 0x70, 0x4e,
	0x61, 0x6d, 0x65, 0x18, 0x0c, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x61, 0x70, 0x70, 0x4e, 0x61,
	0x6d, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x61, 0x70, 0x70, 0x49, 0x63, 0x6f, 0x6e, 0x18, 0x0d, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x07, 0x61, 0x70, 0x70, 0x49, 0x63, 0x6f, 0x6e, 0x12, 0x20, 0x0a, 0x0b,
	0x61, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x49, 0x63, 0x6f, 0x6e, 0x18, 0x0e, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x0b, 0x61, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x49, 0x63, 0x6f, 0x6e, 0x32, 0x58,
	0x0a, 0x0f, 0x41, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x73, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63,
	0x65, 0x12, 0x45, 0x0a, 0x04, 0x4c, 0x69, 0x73, 0x74, 0x12, 0x1e, 0x2e, 0x67, 0x65, 0x6e, 0x65,
	0x72, 0x61, 0x74, 0x65, 0x64, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x41, 0x72, 0x74, 0x69, 0x63, 0x6c,
	0x65, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1b, 0x2e, 0x67, 0x65, 0x6e, 0x65,
	0x72, 0x61, 0x74, 0x65, 0x64, 0x2e, 0x41, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x73, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0x30, 0x5a, 0x2e, 0x67, 0x69, 0x74, 0x68,
	0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x67, 0x61, 0x6d, 0x65, 0x64, 0x62, 0x2f, 0x67, 0x61,
	0x6d, 0x65, 0x64, 0x62, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x62, 0x61, 0x63, 0x6b, 0x65, 0x6e, 0x64,
	0x2f, 0x67, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x65, 0x64, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_articles_proto_rawDescOnce sync.Once
	file_articles_proto_rawDescData = file_articles_proto_rawDesc
)

func file_articles_proto_rawDescGZIP() []byte {
	file_articles_proto_rawDescOnce.Do(func() {
		file_articles_proto_rawDescData = protoimpl.X.CompressGZIP(file_articles_proto_rawDescData)
	})
	return file_articles_proto_rawDescData
}

var file_articles_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_articles_proto_goTypes = []interface{}{
	(*ListArticlesRequest)(nil),   // 0: generated.ListArticlesRequest
	(*ArticlesResponse)(nil),      // 1: generated.ArticlesResponse
	(*ArticleResponse)(nil),       // 2: generated.ArticleResponse
	(*PaginationRequest)(nil),     // 3: generated.PaginationRequest
	(*PaginationResponse)(nil),    // 4: generated.PaginationResponse
	(*timestamppb.Timestamp)(nil), // 5: google.protobuf.Timestamp
}
var file_articles_proto_depIdxs = []int32{
	3, // 0: generated.ListArticlesRequest.pagination:type_name -> generated.PaginationRequest
	4, // 1: generated.ArticlesResponse.pagination:type_name -> generated.PaginationResponse
	2, // 2: generated.ArticlesResponse.articles:type_name -> generated.ArticleResponse
	5, // 3: generated.ArticleResponse.date:type_name -> google.protobuf.Timestamp
	0, // 4: generated.ArticlesService.List:input_type -> generated.ListArticlesRequest
	1, // 5: generated.ArticlesService.List:output_type -> generated.ArticlesResponse
	5, // [5:6] is the sub-list for method output_type
	4, // [4:5] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_articles_proto_init() }
func file_articles_proto_init() {
	if File_articles_proto != nil {
		return
	}
	file_shared_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_articles_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListArticlesRequest); i {
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
		file_articles_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ArticlesResponse); i {
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
		file_articles_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ArticleResponse); i {
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
			RawDescriptor: file_articles_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_articles_proto_goTypes,
		DependencyIndexes: file_articles_proto_depIdxs,
		MessageInfos:      file_articles_proto_msgTypes,
	}.Build()
	File_articles_proto = out.File
	file_articles_proto_rawDesc = nil
	file_articles_proto_goTypes = nil
	file_articles_proto_depIdxs = nil
}