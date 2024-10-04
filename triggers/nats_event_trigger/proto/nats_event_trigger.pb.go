// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v3.21.12
// source: triggers/nats_event_trigger/proto/nats_event_trigger.proto

package nats_event_trigger

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type RunCompletionFeed struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	SpecVersion     string              `protobuf:"bytes,1,opt,name=spec_version,json=specVersion,proto3" json:"spec_version,omitempty"`
	Id              string              `protobuf:"bytes,2,opt,name=id,proto3" json:"id,omitempty"`
	Source          string              `protobuf:"bytes,3,opt,name=source,proto3" json:"source,omitempty"`
	Type            string              `protobuf:"bytes,4,opt,name=type,proto3" json:"type,omitempty"`
	DataContentType string              `protobuf:"bytes,5,opt,name=data_content_type,json=dataContentType,proto3" json:"data_content_type,omitempty"`
	Data            *RunCompletionEvent `protobuf:"bytes,6,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *RunCompletionFeed) Reset() {
	*x = RunCompletionFeed{}
	if protoimpl.UnsafeEnabled {
		mi := &file_triggers_nats_event_trigger_proto_nats_event_trigger_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RunCompletionFeed) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RunCompletionFeed) ProtoMessage() {}

func (x *RunCompletionFeed) ProtoReflect() protoreflect.Message {
	mi := &file_triggers_nats_event_trigger_proto_nats_event_trigger_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RunCompletionFeed.ProtoReflect.Descriptor instead.
func (*RunCompletionFeed) Descriptor() ([]byte, []int) {
	return file_triggers_nats_event_trigger_proto_nats_event_trigger_proto_rawDescGZIP(), []int{0}
}

func (x *RunCompletionFeed) GetSpecVersion() string {
	if x != nil {
		return x.SpecVersion
	}
	return ""
}

func (x *RunCompletionFeed) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *RunCompletionFeed) GetSource() string {
	if x != nil {
		return x.Source
	}
	return ""
}

func (x *RunCompletionFeed) GetType() string {
	if x != nil {
		return x.Type
	}
	return ""
}

func (x *RunCompletionFeed) GetDataContentType() string {
	if x != nil {
		return x.DataContentType
	}
	return ""
}

func (x *RunCompletionFeed) GetData() *RunCompletionEvent {
	if x != nil {
		return x.Data
	}
	return nil
}

type RunCompletionEvent struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	PipelineName          string                  `protobuf:"bytes,1,opt,name=pipeline_name,json=pipelineName,proto3" json:"pipeline_name,omitempty"`
	Provider              string                  `protobuf:"bytes,2,opt,name=provider,proto3" json:"provider,omitempty"`
	RunConfigurationName  string                  `protobuf:"bytes,3,opt,name=run_configuration_name,json=runConfigurationName,proto3" json:"run_configuration_name,omitempty"`
	RunId                 string                  `protobuf:"bytes,4,opt,name=run_id,json=runId,proto3" json:"run_id,omitempty"`
	RunName               string                  `protobuf:"bytes,5,opt,name=run_name,json=runName,proto3" json:"run_name,omitempty"`
	ServingModelArtifacts []*ServingModelArtifact `protobuf:"bytes,6,rep,name=serving_model_artifacts,json=servingModelArtifacts,proto3" json:"serving_model_artifacts,omitempty"`
	Status                string                  `protobuf:"bytes,7,opt,name=status,proto3" json:"status,omitempty"`
}

func (x *RunCompletionEvent) Reset() {
	*x = RunCompletionEvent{}
	if protoimpl.UnsafeEnabled {
		mi := &file_triggers_nats_event_trigger_proto_nats_event_trigger_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RunCompletionEvent) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RunCompletionEvent) ProtoMessage() {}

func (x *RunCompletionEvent) ProtoReflect() protoreflect.Message {
	mi := &file_triggers_nats_event_trigger_proto_nats_event_trigger_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RunCompletionEvent.ProtoReflect.Descriptor instead.
func (*RunCompletionEvent) Descriptor() ([]byte, []int) {
	return file_triggers_nats_event_trigger_proto_nats_event_trigger_proto_rawDescGZIP(), []int{1}
}

func (x *RunCompletionEvent) GetPipelineName() string {
	if x != nil {
		return x.PipelineName
	}
	return ""
}

func (x *RunCompletionEvent) GetProvider() string {
	if x != nil {
		return x.Provider
	}
	return ""
}

func (x *RunCompletionEvent) GetRunConfigurationName() string {
	if x != nil {
		return x.RunConfigurationName
	}
	return ""
}

func (x *RunCompletionEvent) GetRunId() string {
	if x != nil {
		return x.RunId
	}
	return ""
}

func (x *RunCompletionEvent) GetRunName() string {
	if x != nil {
		return x.RunName
	}
	return ""
}

func (x *RunCompletionEvent) GetServingModelArtifacts() []*ServingModelArtifact {
	if x != nil {
		return x.ServingModelArtifacts
	}
	return nil
}

func (x *RunCompletionEvent) GetStatus() string {
	if x != nil {
		return x.Status
	}
	return ""
}

type ServingModelArtifact struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Location string `protobuf:"bytes,1,opt,name=location,proto3" json:"location,omitempty"`
	Name     string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
}

func (x *ServingModelArtifact) Reset() {
	*x = ServingModelArtifact{}
	if protoimpl.UnsafeEnabled {
		mi := &file_triggers_nats_event_trigger_proto_nats_event_trigger_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ServingModelArtifact) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ServingModelArtifact) ProtoMessage() {}

func (x *ServingModelArtifact) ProtoReflect() protoreflect.Message {
	mi := &file_triggers_nats_event_trigger_proto_nats_event_trigger_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ServingModelArtifact.ProtoReflect.Descriptor instead.
func (*ServingModelArtifact) Descriptor() ([]byte, []int) {
	return file_triggers_nats_event_trigger_proto_nats_event_trigger_proto_rawDescGZIP(), []int{2}
}

func (x *ServingModelArtifact) GetLocation() string {
	if x != nil {
		return x.Location
	}
	return ""
}

func (x *ServingModelArtifact) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

var File_triggers_nats_event_trigger_proto_nats_event_trigger_proto protoreflect.FileDescriptor

var file_triggers_nats_event_trigger_proto_nats_event_trigger_proto_rawDesc = []byte{
	0x0a, 0x3a, 0x74, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x73, 0x2f, 0x6e, 0x61, 0x74, 0x73, 0x5f,
	0x65, 0x76, 0x65, 0x6e, 0x74, 0x5f, 0x74, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x2f, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x2f, 0x6e, 0x61, 0x74, 0x73, 0x5f, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x5f, 0x74,
	0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x12, 0x6e, 0x61,
	0x74, 0x73, 0x5f, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x5f, 0x74, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72,
	0x1a, 0x1b, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2f, 0x65, 0x6d, 0x70, 0x74, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xda, 0x01,
	0x0a, 0x11, 0x52, 0x75, 0x6e, 0x43, 0x6f, 0x6d, 0x70, 0x6c, 0x65, 0x74, 0x69, 0x6f, 0x6e, 0x46,
	0x65, 0x65, 0x64, 0x12, 0x21, 0x0a, 0x0c, 0x73, 0x70, 0x65, 0x63, 0x5f, 0x76, 0x65, 0x72, 0x73,
	0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x73, 0x70, 0x65, 0x63, 0x56,
	0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x12, 0x12,
	0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x74, 0x79,
	0x70, 0x65, 0x12, 0x2a, 0x0a, 0x11, 0x64, 0x61, 0x74, 0x61, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65,
	0x6e, 0x74, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0f, 0x64,
	0x61, 0x74, 0x61, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65, 0x12, 0x3a,
	0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x26, 0x2e, 0x6e,
	0x61, 0x74, 0x73, 0x5f, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x5f, 0x74, 0x72, 0x69, 0x67, 0x67, 0x65,
	0x72, 0x2e, 0x52, 0x75, 0x6e, 0x43, 0x6f, 0x6d, 0x70, 0x6c, 0x65, 0x74, 0x69, 0x6f, 0x6e, 0x45,
	0x76, 0x65, 0x6e, 0x74, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x22, 0xb7, 0x02, 0x0a, 0x12, 0x52,
	0x75, 0x6e, 0x43, 0x6f, 0x6d, 0x70, 0x6c, 0x65, 0x74, 0x69, 0x6f, 0x6e, 0x45, 0x76, 0x65, 0x6e,
	0x74, 0x12, 0x23, 0x0a, 0x0d, 0x70, 0x69, 0x70, 0x65, 0x6c, 0x69, 0x6e, 0x65, 0x5f, 0x6e, 0x61,
	0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x70, 0x69, 0x70, 0x65, 0x6c, 0x69,
	0x6e, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x64,
	0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x64,
	0x65, 0x72, 0x12, 0x34, 0x0a, 0x16, 0x72, 0x75, 0x6e, 0x5f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67,
	0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x14, 0x72, 0x75, 0x6e, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x75, 0x72, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x15, 0x0a, 0x06, 0x72, 0x75, 0x6e, 0x5f,
	0x69, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x72, 0x75, 0x6e, 0x49, 0x64, 0x12,
	0x19, 0x0a, 0x08, 0x72, 0x75, 0x6e, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x07, 0x72, 0x75, 0x6e, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x60, 0x0a, 0x17, 0x73, 0x65,
	0x72, 0x76, 0x69, 0x6e, 0x67, 0x5f, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x5f, 0x61, 0x72, 0x74, 0x69,
	0x66, 0x61, 0x63, 0x74, 0x73, 0x18, 0x06, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x28, 0x2e, 0x6e, 0x61,
	0x74, 0x73, 0x5f, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x5f, 0x74, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72,
	0x2e, 0x53, 0x65, 0x72, 0x76, 0x69, 0x6e, 0x67, 0x4d, 0x6f, 0x64, 0x65, 0x6c, 0x41, 0x72, 0x74,
	0x69, 0x66, 0x61, 0x63, 0x74, 0x52, 0x15, 0x73, 0x65, 0x72, 0x76, 0x69, 0x6e, 0x67, 0x4d, 0x6f,
	0x64, 0x65, 0x6c, 0x41, 0x72, 0x74, 0x69, 0x66, 0x61, 0x63, 0x74, 0x73, 0x12, 0x16, 0x0a, 0x06,
	0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x22, 0x46, 0x0a, 0x14, 0x53, 0x65, 0x72, 0x76, 0x69, 0x6e, 0x67, 0x4d,
	0x6f, 0x64, 0x65, 0x6c, 0x41, 0x72, 0x74, 0x69, 0x66, 0x61, 0x63, 0x74, 0x12, 0x1a, 0x0a, 0x08,
	0x6c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08,
	0x6c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x32, 0x67, 0x0a, 0x10,
	0x4e, 0x41, 0x54, 0x53, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x54, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72,
	0x12, 0x53, 0x0a, 0x10, 0x50, 0x72, 0x6f, 0x63, 0x65, 0x73, 0x73, 0x45, 0x76, 0x65, 0x6e, 0x74,
	0x46, 0x65, 0x65, 0x64, 0x12, 0x25, 0x2e, 0x6e, 0x61, 0x74, 0x73, 0x5f, 0x65, 0x76, 0x65, 0x6e,
	0x74, 0x5f, 0x74, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x2e, 0x52, 0x75, 0x6e, 0x43, 0x6f, 0x6d,
	0x70, 0x6c, 0x65, 0x74, 0x69, 0x6f, 0x6e, 0x46, 0x65, 0x65, 0x64, 0x1a, 0x16, 0x2e, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d,
	0x70, 0x74, 0x79, 0x22, 0x00, 0x42, 0x55, 0x5a, 0x53, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e,
	0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x6b, 0x79, 0x2d, 0x75, 0x6b, 0x2f, 0x6b, 0x66, 0x70, 0x2d, 0x6f,
	0x70, 0x65, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x2f, 0x74, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x73,
	0x2f, 0x6e, 0x61, 0x74, 0x73, 0x5f, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x5f, 0x74, 0x72, 0x69, 0x67,
	0x67, 0x65, 0x72, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6e, 0x61, 0x74, 0x73, 0x5f, 0x65,
	0x76, 0x65, 0x6e, 0x74, 0x5f, 0x74, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_triggers_nats_event_trigger_proto_nats_event_trigger_proto_rawDescOnce sync.Once
	file_triggers_nats_event_trigger_proto_nats_event_trigger_proto_rawDescData = file_triggers_nats_event_trigger_proto_nats_event_trigger_proto_rawDesc
)

func file_triggers_nats_event_trigger_proto_nats_event_trigger_proto_rawDescGZIP() []byte {
	file_triggers_nats_event_trigger_proto_nats_event_trigger_proto_rawDescOnce.Do(func() {
		file_triggers_nats_event_trigger_proto_nats_event_trigger_proto_rawDescData = protoimpl.X.CompressGZIP(file_triggers_nats_event_trigger_proto_nats_event_trigger_proto_rawDescData)
	})
	return file_triggers_nats_event_trigger_proto_nats_event_trigger_proto_rawDescData
}

var file_triggers_nats_event_trigger_proto_nats_event_trigger_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_triggers_nats_event_trigger_proto_nats_event_trigger_proto_goTypes = []any{
	(*RunCompletionFeed)(nil),    // 0: nats_event_trigger.RunCompletionFeed
	(*RunCompletionEvent)(nil),   // 1: nats_event_trigger.RunCompletionEvent
	(*ServingModelArtifact)(nil), // 2: nats_event_trigger.ServingModelArtifact
	(*emptypb.Empty)(nil),        // 3: google.protobuf.Empty
}
var file_triggers_nats_event_trigger_proto_nats_event_trigger_proto_depIdxs = []int32{
	1, // 0: nats_event_trigger.RunCompletionFeed.data:type_name -> nats_event_trigger.RunCompletionEvent
	2, // 1: nats_event_trigger.RunCompletionEvent.serving_model_artifacts:type_name -> nats_event_trigger.ServingModelArtifact
	0, // 2: nats_event_trigger.NATSEventTrigger.ProcessEventFeed:input_type -> nats_event_trigger.RunCompletionFeed
	3, // 3: nats_event_trigger.NATSEventTrigger.ProcessEventFeed:output_type -> google.protobuf.Empty
	3, // [3:4] is the sub-list for method output_type
	2, // [2:3] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_triggers_nats_event_trigger_proto_nats_event_trigger_proto_init() }
func file_triggers_nats_event_trigger_proto_nats_event_trigger_proto_init() {
	if File_triggers_nats_event_trigger_proto_nats_event_trigger_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_triggers_nats_event_trigger_proto_nats_event_trigger_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*RunCompletionFeed); i {
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
		file_triggers_nats_event_trigger_proto_nats_event_trigger_proto_msgTypes[1].Exporter = func(v any, i int) any {
			switch v := v.(*RunCompletionEvent); i {
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
		file_triggers_nats_event_trigger_proto_nats_event_trigger_proto_msgTypes[2].Exporter = func(v any, i int) any {
			switch v := v.(*ServingModelArtifact); i {
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
			RawDescriptor: file_triggers_nats_event_trigger_proto_nats_event_trigger_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_triggers_nats_event_trigger_proto_nats_event_trigger_proto_goTypes,
		DependencyIndexes: file_triggers_nats_event_trigger_proto_nats_event_trigger_proto_depIdxs,
		MessageInfos:      file_triggers_nats_event_trigger_proto_nats_event_trigger_proto_msgTypes,
	}.Build()
	File_triggers_nats_event_trigger_proto_nats_event_trigger_proto = out.File
	file_triggers_nats_event_trigger_proto_nats_event_trigger_proto_rawDesc = nil
	file_triggers_nats_event_trigger_proto_nats_event_trigger_proto_goTypes = nil
	file_triggers_nats_event_trigger_proto_nats_event_trigger_proto_depIdxs = nil
}
