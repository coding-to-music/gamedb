// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.15.6
// source: shared.proto

package generated

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

type Empty struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *Empty) Reset() {
	*x = Empty{}
	if protoimpl.UnsafeEnabled {
		mi := &file_shared_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Empty) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Empty) ProtoMessage() {}

func (x *Empty) ProtoReflect() protoreflect.Message {
	mi := &file_shared_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Empty.ProtoReflect.Descriptor instead.
func (*Empty) Descriptor() ([]byte, []int) {
	return file_shared_proto_rawDescGZIP(), []int{0}
}

type Price struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Exists          bool   `protobuf:"varint,1,opt,name=exists,proto3" json:"exists,omitempty"`
	Currency        string `protobuf:"bytes,2,opt,name=currency,proto3" json:"currency,omitempty"`
	Initial         int32  `protobuf:"varint,3,opt,name=initial,proto3" json:"initial,omitempty"`
	Final           int32  `protobuf:"varint,4,opt,name=final,proto3" json:"final,omitempty"`
	DiscountPercent int32  `protobuf:"varint,5,opt,name=discount_percent,json=discountPercent,proto3" json:"discount_percent,omitempty"`
	Individual      int32  `protobuf:"varint,6,opt,name=individual,proto3" json:"individual,omitempty"`
	Free            bool   `protobuf:"varint,7,opt,name=free,proto3" json:"free,omitempty"`
}

func (x *Price) Reset() {
	*x = Price{}
	if protoimpl.UnsafeEnabled {
		mi := &file_shared_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Price) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Price) ProtoMessage() {}

func (x *Price) ProtoReflect() protoreflect.Message {
	mi := &file_shared_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Price.ProtoReflect.Descriptor instead.
func (*Price) Descriptor() ([]byte, []int) {
	return file_shared_proto_rawDescGZIP(), []int{1}
}

func (x *Price) GetExists() bool {
	if x != nil {
		return x.Exists
	}
	return false
}

func (x *Price) GetCurrency() string {
	if x != nil {
		return x.Currency
	}
	return ""
}

func (x *Price) GetInitial() int32 {
	if x != nil {
		return x.Initial
	}
	return 0
}

func (x *Price) GetFinal() int32 {
	if x != nil {
		return x.Final
	}
	return 0
}

func (x *Price) GetDiscountPercent() int32 {
	if x != nil {
		return x.DiscountPercent
	}
	return 0
}

func (x *Price) GetIndividual() int32 {
	if x != nil {
		return x.Individual
	}
	return 0
}

func (x *Price) GetFree() bool {
	if x != nil {
		return x.Free
	}
	return false
}

type PaginationRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Offset    int64  `protobuf:"varint,1,opt,name=offset,proto3" json:"offset,omitempty"`
	Limit     int64  `protobuf:"varint,2,opt,name=limit,proto3" json:"limit,omitempty"`
	SortField string `protobuf:"bytes,3,opt,name=sortField,proto3" json:"sortField,omitempty"`
	SortOrder string `protobuf:"bytes,4,opt,name=sortOrder,proto3" json:"sortOrder,omitempty"`
}

func (x *PaginationRequest) Reset() {
	*x = PaginationRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_shared_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PaginationRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PaginationRequest) ProtoMessage() {}

func (x *PaginationRequest) ProtoReflect() protoreflect.Message {
	mi := &file_shared_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PaginationRequest.ProtoReflect.Descriptor instead.
func (*PaginationRequest) Descriptor() ([]byte, []int) {
	return file_shared_proto_rawDescGZIP(), []int{2}
}

func (x *PaginationRequest) GetOffset() int64 {
	if x != nil {
		return x.Offset
	}
	return 0
}

func (x *PaginationRequest) GetLimit() int64 {
	if x != nil {
		return x.Limit
	}
	return 0
}

func (x *PaginationRequest) GetSortField() string {
	if x != nil {
		return x.SortField
	}
	return ""
}

func (x *PaginationRequest) GetSortOrder() string {
	if x != nil {
		return x.SortOrder
	}
	return ""
}

type PaginationResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Offset        int64 `protobuf:"varint,1,opt,name=offset,proto3" json:"offset,omitempty"`
	Limit         int64 `protobuf:"varint,2,opt,name=limit,proto3" json:"limit,omitempty"`
	Total         int64 `protobuf:"varint,3,opt,name=total,proto3" json:"total,omitempty"`
	TotalFiltered int64 `protobuf:"varint,4,opt,name=totalFiltered,proto3" json:"totalFiltered,omitempty"`
	PagesTotal    int64 `protobuf:"varint,5,opt,name=pagesTotal,proto3" json:"pagesTotal,omitempty"`
	PagesCurrent  int64 `protobuf:"varint,6,opt,name=pagesCurrent,proto3" json:"pagesCurrent,omitempty"`
}

func (x *PaginationResponse) Reset() {
	*x = PaginationResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_shared_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PaginationResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PaginationResponse) ProtoMessage() {}

func (x *PaginationResponse) ProtoReflect() protoreflect.Message {
	mi := &file_shared_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PaginationResponse.ProtoReflect.Descriptor instead.
func (*PaginationResponse) Descriptor() ([]byte, []int) {
	return file_shared_proto_rawDescGZIP(), []int{3}
}

func (x *PaginationResponse) GetOffset() int64 {
	if x != nil {
		return x.Offset
	}
	return 0
}

func (x *PaginationResponse) GetLimit() int64 {
	if x != nil {
		return x.Limit
	}
	return 0
}

func (x *PaginationResponse) GetTotal() int64 {
	if x != nil {
		return x.Total
	}
	return 0
}

func (x *PaginationResponse) GetTotalFiltered() int64 {
	if x != nil {
		return x.TotalFiltered
	}
	return 0
}

func (x *PaginationResponse) GetPagesTotal() int64 {
	if x != nil {
		return x.PagesTotal
	}
	return 0
}

func (x *PaginationResponse) GetPagesCurrent() int64 {
	if x != nil {
		return x.PagesCurrent
	}
	return 0
}

type PaginationRequest2 struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Page  int64 `protobuf:"varint,1,opt,name=page,proto3" json:"page,omitempty"`
	Limit int64 `protobuf:"varint,2,opt,name=limit,proto3" json:"limit,omitempty"`
}

func (x *PaginationRequest2) Reset() {
	*x = PaginationRequest2{}
	if protoimpl.UnsafeEnabled {
		mi := &file_shared_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PaginationRequest2) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PaginationRequest2) ProtoMessage() {}

func (x *PaginationRequest2) ProtoReflect() protoreflect.Message {
	mi := &file_shared_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PaginationRequest2.ProtoReflect.Descriptor instead.
func (*PaginationRequest2) Descriptor() ([]byte, []int) {
	return file_shared_proto_rawDescGZIP(), []int{4}
}

func (x *PaginationRequest2) GetPage() int64 {
	if x != nil {
		return x.Page
	}
	return 0
}

func (x *PaginationRequest2) GetLimit() int64 {
	if x != nil {
		return x.Limit
	}
	return 0
}

var File_shared_proto protoreflect.FileDescriptor

var file_shared_proto_rawDesc = []byte{
	0x0a, 0x0c, 0x73, 0x68, 0x61, 0x72, 0x65, 0x64, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x09,
	0x67, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x65, 0x64, 0x22, 0x07, 0x0a, 0x05, 0x45, 0x6d, 0x70,
	0x74, 0x79, 0x22, 0xca, 0x01, 0x0a, 0x05, 0x50, 0x72, 0x69, 0x63, 0x65, 0x12, 0x16, 0x0a, 0x06,
	0x65, 0x78, 0x69, 0x73, 0x74, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x06, 0x65, 0x78,
	0x69, 0x73, 0x74, 0x73, 0x12, 0x1a, 0x0a, 0x08, 0x63, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x63, 0x79,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x63, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x63, 0x79,
	0x12, 0x18, 0x0a, 0x07, 0x69, 0x6e, 0x69, 0x74, 0x69, 0x61, 0x6c, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x05, 0x52, 0x07, 0x69, 0x6e, 0x69, 0x74, 0x69, 0x61, 0x6c, 0x12, 0x14, 0x0a, 0x05, 0x66, 0x69,
	0x6e, 0x61, 0x6c, 0x18, 0x04, 0x20, 0x01, 0x28, 0x05, 0x52, 0x05, 0x66, 0x69, 0x6e, 0x61, 0x6c,
	0x12, 0x29, 0x0a, 0x10, 0x64, 0x69, 0x73, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x5f, 0x70, 0x65, 0x72,
	0x63, 0x65, 0x6e, 0x74, 0x18, 0x05, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0f, 0x64, 0x69, 0x73, 0x63,
	0x6f, 0x75, 0x6e, 0x74, 0x50, 0x65, 0x72, 0x63, 0x65, 0x6e, 0x74, 0x12, 0x1e, 0x0a, 0x0a, 0x69,
	0x6e, 0x64, 0x69, 0x76, 0x69, 0x64, 0x75, 0x61, 0x6c, 0x18, 0x06, 0x20, 0x01, 0x28, 0x05, 0x52,
	0x0a, 0x69, 0x6e, 0x64, 0x69, 0x76, 0x69, 0x64, 0x75, 0x61, 0x6c, 0x12, 0x12, 0x0a, 0x04, 0x66,
	0x72, 0x65, 0x65, 0x18, 0x07, 0x20, 0x01, 0x28, 0x08, 0x52, 0x04, 0x66, 0x72, 0x65, 0x65, 0x22,
	0x7d, 0x0a, 0x11, 0x50, 0x61, 0x67, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x6f, 0x66, 0x66, 0x73, 0x65, 0x74, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x03, 0x52, 0x06, 0x6f, 0x66, 0x66, 0x73, 0x65, 0x74, 0x12, 0x14, 0x0a, 0x05,
	0x6c, 0x69, 0x6d, 0x69, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x05, 0x6c, 0x69, 0x6d,
	0x69, 0x74, 0x12, 0x1c, 0x0a, 0x09, 0x73, 0x6f, 0x72, 0x74, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x73, 0x6f, 0x72, 0x74, 0x46, 0x69, 0x65, 0x6c, 0x64,
	0x12, 0x1c, 0x0a, 0x09, 0x73, 0x6f, 0x72, 0x74, 0x4f, 0x72, 0x64, 0x65, 0x72, 0x18, 0x04, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x09, 0x73, 0x6f, 0x72, 0x74, 0x4f, 0x72, 0x64, 0x65, 0x72, 0x22, 0xc2,
	0x01, 0x0a, 0x12, 0x50, 0x61, 0x67, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x6f, 0x66, 0x66, 0x73, 0x65, 0x74, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x06, 0x6f, 0x66, 0x66, 0x73, 0x65, 0x74, 0x12, 0x14, 0x0a,
	0x05, 0x6c, 0x69, 0x6d, 0x69, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x05, 0x6c, 0x69,
	0x6d, 0x69, 0x74, 0x12, 0x14, 0x0a, 0x05, 0x74, 0x6f, 0x74, 0x61, 0x6c, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x03, 0x52, 0x05, 0x74, 0x6f, 0x74, 0x61, 0x6c, 0x12, 0x24, 0x0a, 0x0d, 0x74, 0x6f, 0x74,
	0x61, 0x6c, 0x46, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x65, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x03,
	0x52, 0x0d, 0x74, 0x6f, 0x74, 0x61, 0x6c, 0x46, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x65, 0x64, 0x12,
	0x1e, 0x0a, 0x0a, 0x70, 0x61, 0x67, 0x65, 0x73, 0x54, 0x6f, 0x74, 0x61, 0x6c, 0x18, 0x05, 0x20,
	0x01, 0x28, 0x03, 0x52, 0x0a, 0x70, 0x61, 0x67, 0x65, 0x73, 0x54, 0x6f, 0x74, 0x61, 0x6c, 0x12,
	0x22, 0x0a, 0x0c, 0x70, 0x61, 0x67, 0x65, 0x73, 0x43, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x74, 0x18,
	0x06, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0c, 0x70, 0x61, 0x67, 0x65, 0x73, 0x43, 0x75, 0x72, 0x72,
	0x65, 0x6e, 0x74, 0x22, 0x3e, 0x0a, 0x12, 0x50, 0x61, 0x67, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x32, 0x12, 0x12, 0x0a, 0x04, 0x70, 0x61, 0x67,
	0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x04, 0x70, 0x61, 0x67, 0x65, 0x12, 0x14, 0x0a,
	0x05, 0x6c, 0x69, 0x6d, 0x69, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x05, 0x6c, 0x69,
	0x6d, 0x69, 0x74, 0x42, 0x30, 0x5a, 0x2e, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f,
	0x6d, 0x2f, 0x67, 0x61, 0x6d, 0x65, 0x64, 0x62, 0x2f, 0x67, 0x61, 0x6d, 0x65, 0x64, 0x62, 0x2f,
	0x70, 0x6b, 0x67, 0x2f, 0x62, 0x61, 0x63, 0x6b, 0x65, 0x6e, 0x64, 0x2f, 0x67, 0x65, 0x6e, 0x65,
	0x72, 0x61, 0x74, 0x65, 0x64, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_shared_proto_rawDescOnce sync.Once
	file_shared_proto_rawDescData = file_shared_proto_rawDesc
)

func file_shared_proto_rawDescGZIP() []byte {
	file_shared_proto_rawDescOnce.Do(func() {
		file_shared_proto_rawDescData = protoimpl.X.CompressGZIP(file_shared_proto_rawDescData)
	})
	return file_shared_proto_rawDescData
}

var file_shared_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_shared_proto_goTypes = []interface{}{
	(*Empty)(nil),              // 0: generated.Empty
	(*Price)(nil),              // 1: generated.Price
	(*PaginationRequest)(nil),  // 2: generated.PaginationRequest
	(*PaginationResponse)(nil), // 3: generated.PaginationResponse
	(*PaginationRequest2)(nil), // 4: generated.PaginationRequest2
}
var file_shared_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_shared_proto_init() }
func file_shared_proto_init() {
	if File_shared_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_shared_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Empty); i {
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
		file_shared_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Price); i {
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
		file_shared_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PaginationRequest); i {
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
		file_shared_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PaginationResponse); i {
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
		file_shared_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PaginationRequest2); i {
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
			RawDescriptor: file_shared_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_shared_proto_goTypes,
		DependencyIndexes: file_shared_proto_depIdxs,
		MessageInfos:      file_shared_proto_msgTypes,
	}.Build()
	File_shared_proto = out.File
	file_shared_proto_rawDesc = nil
	file_shared_proto_goTypes = nil
	file_shared_proto_depIdxs = nil
}
