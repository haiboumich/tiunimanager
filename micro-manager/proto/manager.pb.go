// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.17.2
// source: manager.proto

package manager

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

var File_manager_proto protoreflect.FileDescriptor

var file_manager_proto_rawDesc = []byte{
	0x0a, 0x0d, 0x6d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a,
	0x0e, 0x6d, 0x6e, 0x67, 0x5f, 0x61, 0x75, 0x74, 0x68, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a,
	0x0e, 0x6d, 0x6e, 0x67, 0x5f, 0x68, 0x6f, 0x73, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x32,
	0xbe, 0x03, 0x0a, 0x12, 0x54, 0x69, 0x43, 0x50, 0x4d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72, 0x53,
	0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x26, 0x0a, 0x05, 0x4c, 0x6f, 0x67, 0x69, 0x6e, 0x12,
	0x0d, 0x2e, 0x4c, 0x6f, 0x67, 0x69, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0e,
	0x2e, 0x4c, 0x6f, 0x67, 0x69, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x29,
	0x0a, 0x06, 0x4c, 0x6f, 0x67, 0x6f, 0x75, 0x74, 0x12, 0x0e, 0x2e, 0x4c, 0x6f, 0x67, 0x6f, 0x75,
	0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0f, 0x2e, 0x4c, 0x6f, 0x67, 0x6f, 0x75,
	0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x41, 0x0a, 0x0e, 0x56, 0x65, 0x72,
	0x69, 0x66, 0x79, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x12, 0x16, 0x2e, 0x56, 0x65,
	0x72, 0x69, 0x66, 0x79, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x17, 0x2e, 0x56, 0x65, 0x72, 0x69, 0x66, 0x79, 0x49, 0x64, 0x65, 0x6e,
	0x74, 0x69, 0x74, 0x79, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x35, 0x0a, 0x0a,
	0x49, 0x6d, 0x70, 0x6f, 0x72, 0x74, 0x48, 0x6f, 0x73, 0x74, 0x12, 0x12, 0x2e, 0x49, 0x6d, 0x70,
	0x6f, 0x72, 0x74, 0x48, 0x6f, 0x73, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x13,
	0x2e, 0x49, 0x6d, 0x70, 0x6f, 0x72, 0x74, 0x48, 0x6f, 0x73, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x12, 0x35, 0x0a, 0x0a, 0x52, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x48, 0x6f, 0x73,
	0x74, 0x12, 0x12, 0x2e, 0x52, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x48, 0x6f, 0x73, 0x74, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x13, 0x2e, 0x52, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x48, 0x6f,
	0x73, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x31, 0x0a, 0x08, 0x4c, 0x69,
	0x73, 0x74, 0x48, 0x6f, 0x73, 0x74, 0x12, 0x11, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x48, 0x6f, 0x73,
	0x74, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x12, 0x2e, 0x4c, 0x69, 0x73, 0x74,
	0x48, 0x6f, 0x73, 0x74, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x3b, 0x0a,
	0x0c, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x44, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x73, 0x12, 0x14, 0x2e,
	0x43, 0x68, 0x65, 0x63, 0x6b, 0x44, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x73, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x15, 0x2e, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x44, 0x65, 0x74, 0x61, 0x69,
	0x6c, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x34, 0x0a, 0x0a, 0x41, 0x6c,
	0x6c, 0x6f, 0x63, 0x48, 0x6f, 0x73, 0x74, 0x73, 0x12, 0x12, 0x2e, 0x41, 0x6c, 0x6c, 0x6f, 0x63,
	0x48, 0x6f, 0x73, 0x74, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x12, 0x2e, 0x41,
	0x6c, 0x6c, 0x6f, 0x63, 0x48, 0x6f, 0x73, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x42, 0x0c, 0x5a, 0x0a, 0x2e, 0x2f, 0x3b, 0x6d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72, 0x62, 0x06,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var file_manager_proto_goTypes = []interface{}{
	(*LoginRequest)(nil),           // 0: LoginRequest
	(*LogoutRequest)(nil),          // 1: LogoutRequest
	(*VerifyIdentityRequest)(nil),  // 2: VerifyIdentityRequest
	(*ImportHostRequest)(nil),      // 3: ImportHostRequest
	(*RemoveHostRequest)(nil),      // 4: RemoveHostRequest
	(*ListHostsRequest)(nil),       // 5: ListHostsRequest
	(*CheckDetailsRequest)(nil),    // 6: CheckDetailsRequest
	(*AllocHostsRequest)(nil),      // 7: AllocHostsRequest
	(*LoginResponse)(nil),          // 8: LoginResponse
	(*LogoutResponse)(nil),         // 9: LogoutResponse
	(*VerifyIdentityResponse)(nil), // 10: VerifyIdentityResponse
	(*ImportHostResponse)(nil),     // 11: ImportHostResponse
	(*RemoveHostResponse)(nil),     // 12: RemoveHostResponse
	(*ListHostsResponse)(nil),      // 13: ListHostsResponse
	(*CheckDetailsResponse)(nil),   // 14: CheckDetailsResponse
	(*AllocHostResponse)(nil),      // 15: AllocHostResponse
}
var file_manager_proto_depIdxs = []int32{
	0,  // 0: TiCPManagerService.Login:input_type -> LoginRequest
	1,  // 1: TiCPManagerService.Logout:input_type -> LogoutRequest
	2,  // 2: TiCPManagerService.VerifyIdentity:input_type -> VerifyIdentityRequest
	3,  // 3: TiCPManagerService.ImportHost:input_type -> ImportHostRequest
	4,  // 4: TiCPManagerService.RemoveHost:input_type -> RemoveHostRequest
	5,  // 5: TiCPManagerService.ListHost:input_type -> ListHostsRequest
	6,  // 6: TiCPManagerService.CheckDetails:input_type -> CheckDetailsRequest
	7,  // 7: TiCPManagerService.AllocHosts:input_type -> AllocHostsRequest
	8,  // 8: TiCPManagerService.Login:output_type -> LoginResponse
	9,  // 9: TiCPManagerService.Logout:output_type -> LogoutResponse
	10, // 10: TiCPManagerService.VerifyIdentity:output_type -> VerifyIdentityResponse
	11, // 11: TiCPManagerService.ImportHost:output_type -> ImportHostResponse
	12, // 12: TiCPManagerService.RemoveHost:output_type -> RemoveHostResponse
	13, // 13: TiCPManagerService.ListHost:output_type -> ListHostsResponse
	14, // 14: TiCPManagerService.CheckDetails:output_type -> CheckDetailsResponse
	15, // 15: TiCPManagerService.AllocHosts:output_type -> AllocHostResponse
	8,  // [8:16] is the sub-list for method output_type
	0,  // [0:8] is the sub-list for method input_type
	0,  // [0:0] is the sub-list for extension type_name
	0,  // [0:0] is the sub-list for extension extendee
	0,  // [0:0] is the sub-list for field type_name
}

func init() { file_manager_proto_init() }
func file_manager_proto_init() {
	if File_manager_proto != nil {
		return
	}
	file_mng_auth_proto_init()
	file_mng_host_proto_init()
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_manager_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_manager_proto_goTypes,
		DependencyIndexes: file_manager_proto_depIdxs,
	}.Build()
	File_manager_proto = out.File
	file_manager_proto_rawDesc = nil
	file_manager_proto_goTypes = nil
	file_manager_proto_depIdxs = nil
}
