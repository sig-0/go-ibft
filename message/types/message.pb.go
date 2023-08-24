// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.21.12
// source: message/types/proto/message.proto

package types

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

// View represents a time interval in which validators attempt to reach consensus for a particular sequence.
// Initially for a given sequence, round starts at 0. Whenever consensus is not reached in a round, validators move on to the next.
type View struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Sequence uint64 `protobuf:"varint,1,opt,name=sequence,proto3" json:"sequence,omitempty"`
	Round    uint64 `protobuf:"varint,2,opt,name=round,proto3" json:"round,omitempty"`
}

func (x *View) Reset() {
	*x = View{}
	if protoimpl.UnsafeEnabled {
		mi := &file_message_types_proto_message_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *View) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*View) ProtoMessage() {}

func (x *View) ProtoReflect() protoreflect.Message {
	mi := &file_message_types_proto_message_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use View.ProtoReflect.Descriptor instead.
func (*View) Descriptor() ([]byte, []int) {
	return file_message_types_proto_message_proto_rawDescGZIP(), []int{0}
}

func (x *View) GetSequence() uint64 {
	if x != nil {
		return x.Sequence
	}
	return 0
}

func (x *View) GetRound() uint64 {
	if x != nil {
		return x.Round
	}
	return 0
}

// ProposedBlock is a block constructed by the proposer in a specific round
type ProposedBlock struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Data  []byte `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
	Round uint64 `protobuf:"varint,2,opt,name=round,proto3" json:"round,omitempty"`
}

func (x *ProposedBlock) Reset() {
	*x = ProposedBlock{}
	if protoimpl.UnsafeEnabled {
		mi := &file_message_types_proto_message_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ProposedBlock) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProposedBlock) ProtoMessage() {}

func (x *ProposedBlock) ProtoReflect() protoreflect.Message {
	mi := &file_message_types_proto_message_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProposedBlock.ProtoReflect.Descriptor instead.
func (*ProposedBlock) Descriptor() ([]byte, []int) {
	return file_message_types_proto_message_proto_rawDescGZIP(), []int{1}
}

func (x *ProposedBlock) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

func (x *ProposedBlock) GetRound() uint64 {
	if x != nil {
		return x.Round
	}
	return 0
}

// PreparedCertificate is the last PROPOSAL message to reach Quorum-1 PREPARE messages
type PreparedCertificate struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ProposalMessage *MsgProposal  `protobuf:"bytes,1,opt,name=proposal_message,json=proposalMessage,proto3" json:"proposal_message,omitempty"`
	PrepareMessages []*MsgPrepare `protobuf:"bytes,2,rep,name=prepare_messages,json=prepareMessages,proto3" json:"prepare_messages,omitempty"`
}

func (x *PreparedCertificate) Reset() {
	*x = PreparedCertificate{}
	if protoimpl.UnsafeEnabled {
		mi := &file_message_types_proto_message_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PreparedCertificate) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PreparedCertificate) ProtoMessage() {}

func (x *PreparedCertificate) ProtoReflect() protoreflect.Message {
	mi := &file_message_types_proto_message_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PreparedCertificate.ProtoReflect.Descriptor instead.
func (*PreparedCertificate) Descriptor() ([]byte, []int) {
	return file_message_types_proto_message_proto_rawDescGZIP(), []int{2}
}

func (x *PreparedCertificate) GetProposalMessage() *MsgProposal {
	if x != nil {
		return x.ProposalMessage
	}
	return nil
}

func (x *PreparedCertificate) GetPrepareMessages() []*MsgPrepare {
	if x != nil {
		return x.PrepareMessages
	}
	return nil
}

// RoundChangeCertificate is a collection of ROUND-CHANGE messages multicasted by validators in a specific view
type RoundChangeCertificate struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Messages []*MsgRoundChange `protobuf:"bytes,1,rep,name=messages,proto3" json:"messages,omitempty"`
}

func (x *RoundChangeCertificate) Reset() {
	*x = RoundChangeCertificate{}
	if protoimpl.UnsafeEnabled {
		mi := &file_message_types_proto_message_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RoundChangeCertificate) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RoundChangeCertificate) ProtoMessage() {}

func (x *RoundChangeCertificate) ProtoReflect() protoreflect.Message {
	mi := &file_message_types_proto_message_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RoundChangeCertificate.ProtoReflect.Descriptor instead.
func (*RoundChangeCertificate) Descriptor() ([]byte, []int) {
	return file_message_types_proto_message_proto_rawDescGZIP(), []int{3}
}

func (x *RoundChangeCertificate) GetMessages() []*MsgRoundChange {
	if x != nil {
		return x.Messages
	}
	return nil
}

// MsgProposal is a proposed block multicasted by the proposer in a specific view.
// If round_change_certificate is not empty, the block was extracted from previous rounds
type MsgProposal struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	View                   *View                   `protobuf:"bytes,1,opt,name=view,proto3" json:"view,omitempty"`
	From                   []byte                  `protobuf:"bytes,2,opt,name=from,proto3" json:"from,omitempty"`
	Signature              []byte                  `protobuf:"bytes,3,opt,name=signature,proto3" json:"signature,omitempty"`
	ProposedBlock          *ProposedBlock          `protobuf:"bytes,4,opt,name=proposed_block,json=proposedBlock,proto3" json:"proposed_block,omitempty"`
	ProposalHash           []byte                  `protobuf:"bytes,5,opt,name=proposal_hash,json=proposalHash,proto3" json:"proposal_hash,omitempty"`
	RoundChangeCertificate *RoundChangeCertificate `protobuf:"bytes,6,opt,name=round_change_certificate,json=roundChangeCertificate,proto3" json:"round_change_certificate,omitempty"`
}

func (x *MsgProposal) Reset() {
	*x = MsgProposal{}
	if protoimpl.UnsafeEnabled {
		mi := &file_message_types_proto_message_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MsgProposal) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MsgProposal) ProtoMessage() {}

func (x *MsgProposal) ProtoReflect() protoreflect.Message {
	mi := &file_message_types_proto_message_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MsgProposal.ProtoReflect.Descriptor instead.
func (*MsgProposal) Descriptor() ([]byte, []int) {
	return file_message_types_proto_message_proto_rawDescGZIP(), []int{4}
}

func (x *MsgProposal) GetView() *View {
	if x != nil {
		return x.View
	}
	return nil
}

func (x *MsgProposal) GetFrom() []byte {
	if x != nil {
		return x.From
	}
	return nil
}

func (x *MsgProposal) GetSignature() []byte {
	if x != nil {
		return x.Signature
	}
	return nil
}

func (x *MsgProposal) GetProposedBlock() *ProposedBlock {
	if x != nil {
		return x.ProposedBlock
	}
	return nil
}

func (x *MsgProposal) GetProposalHash() []byte {
	if x != nil {
		return x.ProposalHash
	}
	return nil
}

func (x *MsgProposal) GetRoundChangeCertificate() *RoundChangeCertificate {
	if x != nil {
		return x.RoundChangeCertificate
	}
	return nil
}

// MsgPrepare is the keccak hash corresponding to a proposed block in a specific view
type MsgPrepare struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	View         *View  `protobuf:"bytes,1,opt,name=view,proto3" json:"view,omitempty"`
	From         []byte `protobuf:"bytes,2,opt,name=from,proto3" json:"from,omitempty"`
	Signature    []byte `protobuf:"bytes,3,opt,name=signature,proto3" json:"signature,omitempty"`
	ProposalHash []byte `protobuf:"bytes,4,opt,name=proposal_hash,json=proposalHash,proto3" json:"proposal_hash,omitempty"`
}

func (x *MsgPrepare) Reset() {
	*x = MsgPrepare{}
	if protoimpl.UnsafeEnabled {
		mi := &file_message_types_proto_message_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MsgPrepare) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MsgPrepare) ProtoMessage() {}

func (x *MsgPrepare) ProtoReflect() protoreflect.Message {
	mi := &file_message_types_proto_message_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MsgPrepare.ProtoReflect.Descriptor instead.
func (*MsgPrepare) Descriptor() ([]byte, []int) {
	return file_message_types_proto_message_proto_rawDescGZIP(), []int{5}
}

func (x *MsgPrepare) GetView() *View {
	if x != nil {
		return x.View
	}
	return nil
}

func (x *MsgPrepare) GetFrom() []byte {
	if x != nil {
		return x.From
	}
	return nil
}

func (x *MsgPrepare) GetSignature() []byte {
	if x != nil {
		return x.Signature
	}
	return nil
}

func (x *MsgPrepare) GetProposalHash() []byte {
	if x != nil {
		return x.ProposalHash
	}
	return nil
}

// MsgCommit is a validator's seal(signature) over a proposed block in a specific view
type MsgCommit struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	View         *View  `protobuf:"bytes,1,opt,name=view,proto3" json:"view,omitempty"`
	From         []byte `protobuf:"bytes,2,opt,name=from,proto3" json:"from,omitempty"`
	Signature    []byte `protobuf:"bytes,3,opt,name=signature,proto3" json:"signature,omitempty"`
	ProposalHash []byte `protobuf:"bytes,4,opt,name=proposal_hash,json=proposalHash,proto3" json:"proposal_hash,omitempty"`
	CommitSeal   []byte `protobuf:"bytes,5,opt,name=commit_seal,json=commitSeal,proto3" json:"commit_seal,omitempty"`
}

func (x *MsgCommit) Reset() {
	*x = MsgCommit{}
	if protoimpl.UnsafeEnabled {
		mi := &file_message_types_proto_message_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MsgCommit) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MsgCommit) ProtoMessage() {}

func (x *MsgCommit) ProtoReflect() protoreflect.Message {
	mi := &file_message_types_proto_message_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MsgCommit.ProtoReflect.Descriptor instead.
func (*MsgCommit) Descriptor() ([]byte, []int) {
	return file_message_types_proto_message_proto_rawDescGZIP(), []int{6}
}

func (x *MsgCommit) GetView() *View {
	if x != nil {
		return x.View
	}
	return nil
}

func (x *MsgCommit) GetFrom() []byte {
	if x != nil {
		return x.From
	}
	return nil
}

func (x *MsgCommit) GetSignature() []byte {
	if x != nil {
		return x.Signature
	}
	return nil
}

func (x *MsgCommit) GetProposalHash() []byte {
	if x != nil {
		return x.ProposalHash
	}
	return nil
}

func (x *MsgCommit) GetCommitSeal() []byte {
	if x != nil {
		return x.CommitSeal
	}
	return nil
}

// MsgRoundChange is a message multicasted when the round timer expire.
type MsgRoundChange struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	View                        *View                `protobuf:"bytes,1,opt,name=view,proto3" json:"view,omitempty"`
	From                        []byte               `protobuf:"bytes,2,opt,name=from,proto3" json:"from,omitempty"`
	Signature                   []byte               `protobuf:"bytes,3,opt,name=signature,proto3" json:"signature,omitempty"`
	LatestPreparedProposedBlock *ProposedBlock       `protobuf:"bytes,4,opt,name=latest_prepared_proposed_block,json=latestPreparedProposedBlock,proto3" json:"latest_prepared_proposed_block,omitempty"`
	LatestPreparedCertificate   *PreparedCertificate `protobuf:"bytes,5,opt,name=latest_prepared_certificate,json=latestPreparedCertificate,proto3" json:"latest_prepared_certificate,omitempty"`
}

func (x *MsgRoundChange) Reset() {
	*x = MsgRoundChange{}
	if protoimpl.UnsafeEnabled {
		mi := &file_message_types_proto_message_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MsgRoundChange) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MsgRoundChange) ProtoMessage() {}

func (x *MsgRoundChange) ProtoReflect() protoreflect.Message {
	mi := &file_message_types_proto_message_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MsgRoundChange.ProtoReflect.Descriptor instead.
func (*MsgRoundChange) Descriptor() ([]byte, []int) {
	return file_message_types_proto_message_proto_rawDescGZIP(), []int{7}
}

func (x *MsgRoundChange) GetView() *View {
	if x != nil {
		return x.View
	}
	return nil
}

func (x *MsgRoundChange) GetFrom() []byte {
	if x != nil {
		return x.From
	}
	return nil
}

func (x *MsgRoundChange) GetSignature() []byte {
	if x != nil {
		return x.Signature
	}
	return nil
}

func (x *MsgRoundChange) GetLatestPreparedProposedBlock() *ProposedBlock {
	if x != nil {
		return x.LatestPreparedProposedBlock
	}
	return nil
}

func (x *MsgRoundChange) GetLatestPreparedCertificate() *PreparedCertificate {
	if x != nil {
		return x.LatestPreparedCertificate
	}
	return nil
}

var File_message_types_proto_message_proto protoreflect.FileDescriptor

var file_message_types_proto_message_proto_rawDesc = []byte{
	0x0a, 0x21, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x2f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2f,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x22, 0x38, 0x0a, 0x04, 0x56, 0x69, 0x65, 0x77, 0x12, 0x1a, 0x0a, 0x08, 0x73,
	0x65, 0x71, 0x75, 0x65, 0x6e, 0x63, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x08, 0x73,
	0x65, 0x71, 0x75, 0x65, 0x6e, 0x63, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x72, 0x6f, 0x75, 0x6e, 0x64,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x52, 0x05, 0x72, 0x6f, 0x75, 0x6e, 0x64, 0x22, 0x39, 0x0a,
	0x0d, 0x50, 0x72, 0x6f, 0x70, 0x6f, 0x73, 0x65, 0x64, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x12, 0x12,
	0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x64, 0x61,
	0x74, 0x61, 0x12, 0x14, 0x0a, 0x05, 0x72, 0x6f, 0x75, 0x6e, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x04, 0x52, 0x05, 0x72, 0x6f, 0x75, 0x6e, 0x64, 0x22, 0x86, 0x01, 0x0a, 0x13, 0x50, 0x72, 0x65,
	0x70, 0x61, 0x72, 0x65, 0x64, 0x43, 0x65, 0x72, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x65,
	0x12, 0x37, 0x0a, 0x10, 0x70, 0x72, 0x6f, 0x70, 0x6f, 0x73, 0x61, 0x6c, 0x5f, 0x6d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0c, 0x2e, 0x4d, 0x73, 0x67,
	0x50, 0x72, 0x6f, 0x70, 0x6f, 0x73, 0x61, 0x6c, 0x52, 0x0f, 0x70, 0x72, 0x6f, 0x70, 0x6f, 0x73,
	0x61, 0x6c, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x36, 0x0a, 0x10, 0x70, 0x72, 0x65,
	0x70, 0x61, 0x72, 0x65, 0x5f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73, 0x18, 0x02, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x0b, 0x2e, 0x4d, 0x73, 0x67, 0x50, 0x72, 0x65, 0x70, 0x61, 0x72, 0x65,
	0x52, 0x0f, 0x70, 0x72, 0x65, 0x70, 0x61, 0x72, 0x65, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x73, 0x22, 0x45, 0x0a, 0x16, 0x52, 0x6f, 0x75, 0x6e, 0x64, 0x43, 0x68, 0x61, 0x6e, 0x67, 0x65,
	0x43, 0x65, 0x72, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x65, 0x12, 0x2b, 0x0a, 0x08, 0x6d,
	0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0f, 0x2e,
	0x4d, 0x73, 0x67, 0x52, 0x6f, 0x75, 0x6e, 0x64, 0x43, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x52, 0x08,
	0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73, 0x22, 0x89, 0x02, 0x0a, 0x0b, 0x4d, 0x73, 0x67,
	0x50, 0x72, 0x6f, 0x70, 0x6f, 0x73, 0x61, 0x6c, 0x12, 0x19, 0x0a, 0x04, 0x76, 0x69, 0x65, 0x77,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x05, 0x2e, 0x56, 0x69, 0x65, 0x77, 0x52, 0x04, 0x76,
	0x69, 0x65, 0x77, 0x12, 0x12, 0x0a, 0x04, 0x66, 0x72, 0x6f, 0x6d, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0c, 0x52, 0x04, 0x66, 0x72, 0x6f, 0x6d, 0x12, 0x1c, 0x0a, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61,
	0x74, 0x75, 0x72, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x09, 0x73, 0x69, 0x67, 0x6e,
	0x61, 0x74, 0x75, 0x72, 0x65, 0x12, 0x35, 0x0a, 0x0e, 0x70, 0x72, 0x6f, 0x70, 0x6f, 0x73, 0x65,
	0x64, 0x5f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0e, 0x2e,
	0x50, 0x72, 0x6f, 0x70, 0x6f, 0x73, 0x65, 0x64, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x52, 0x0d, 0x70,
	0x72, 0x6f, 0x70, 0x6f, 0x73, 0x65, 0x64, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x12, 0x23, 0x0a, 0x0d,
	0x70, 0x72, 0x6f, 0x70, 0x6f, 0x73, 0x61, 0x6c, 0x5f, 0x68, 0x61, 0x73, 0x68, 0x18, 0x05, 0x20,
	0x01, 0x28, 0x0c, 0x52, 0x0c, 0x70, 0x72, 0x6f, 0x70, 0x6f, 0x73, 0x61, 0x6c, 0x48, 0x61, 0x73,
	0x68, 0x12, 0x51, 0x0a, 0x18, 0x72, 0x6f, 0x75, 0x6e, 0x64, 0x5f, 0x63, 0x68, 0x61, 0x6e, 0x67,
	0x65, 0x5f, 0x63, 0x65, 0x72, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x65, 0x18, 0x06, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x52, 0x6f, 0x75, 0x6e, 0x64, 0x43, 0x68, 0x61, 0x6e, 0x67,
	0x65, 0x43, 0x65, 0x72, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x65, 0x52, 0x16, 0x72, 0x6f,
	0x75, 0x6e, 0x64, 0x43, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x43, 0x65, 0x72, 0x74, 0x69, 0x66, 0x69,
	0x63, 0x61, 0x74, 0x65, 0x22, 0x7e, 0x0a, 0x0a, 0x4d, 0x73, 0x67, 0x50, 0x72, 0x65, 0x70, 0x61,
	0x72, 0x65, 0x12, 0x19, 0x0a, 0x04, 0x76, 0x69, 0x65, 0x77, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x05, 0x2e, 0x56, 0x69, 0x65, 0x77, 0x52, 0x04, 0x76, 0x69, 0x65, 0x77, 0x12, 0x12, 0x0a,
	0x04, 0x66, 0x72, 0x6f, 0x6d, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x66, 0x72, 0x6f,
	0x6d, 0x12, 0x1c, 0x0a, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x12,
	0x23, 0x0a, 0x0d, 0x70, 0x72, 0x6f, 0x70, 0x6f, 0x73, 0x61, 0x6c, 0x5f, 0x68, 0x61, 0x73, 0x68,
	0x18, 0x04, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0c, 0x70, 0x72, 0x6f, 0x70, 0x6f, 0x73, 0x61, 0x6c,
	0x48, 0x61, 0x73, 0x68, 0x22, 0x9e, 0x01, 0x0a, 0x09, 0x4d, 0x73, 0x67, 0x43, 0x6f, 0x6d, 0x6d,
	0x69, 0x74, 0x12, 0x19, 0x0a, 0x04, 0x76, 0x69, 0x65, 0x77, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x05, 0x2e, 0x56, 0x69, 0x65, 0x77, 0x52, 0x04, 0x76, 0x69, 0x65, 0x77, 0x12, 0x12, 0x0a,
	0x04, 0x66, 0x72, 0x6f, 0x6d, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x66, 0x72, 0x6f,
	0x6d, 0x12, 0x1c, 0x0a, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x12,
	0x23, 0x0a, 0x0d, 0x70, 0x72, 0x6f, 0x70, 0x6f, 0x73, 0x61, 0x6c, 0x5f, 0x68, 0x61, 0x73, 0x68,
	0x18, 0x04, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0c, 0x70, 0x72, 0x6f, 0x70, 0x6f, 0x73, 0x61, 0x6c,
	0x48, 0x61, 0x73, 0x68, 0x12, 0x1f, 0x0a, 0x0b, 0x63, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x5f, 0x73,
	0x65, 0x61, 0x6c, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0a, 0x63, 0x6f, 0x6d, 0x6d, 0x69,
	0x74, 0x53, 0x65, 0x61, 0x6c, 0x22, 0x88, 0x02, 0x0a, 0x0e, 0x4d, 0x73, 0x67, 0x52, 0x6f, 0x75,
	0x6e, 0x64, 0x43, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x12, 0x19, 0x0a, 0x04, 0x76, 0x69, 0x65, 0x77,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x05, 0x2e, 0x56, 0x69, 0x65, 0x77, 0x52, 0x04, 0x76,
	0x69, 0x65, 0x77, 0x12, 0x12, 0x0a, 0x04, 0x66, 0x72, 0x6f, 0x6d, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0c, 0x52, 0x04, 0x66, 0x72, 0x6f, 0x6d, 0x12, 0x1c, 0x0a, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61,
	0x74, 0x75, 0x72, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x09, 0x73, 0x69, 0x67, 0x6e,
	0x61, 0x74, 0x75, 0x72, 0x65, 0x12, 0x53, 0x0a, 0x1e, 0x6c, 0x61, 0x74, 0x65, 0x73, 0x74, 0x5f,
	0x70, 0x72, 0x65, 0x70, 0x61, 0x72, 0x65, 0x64, 0x5f, 0x70, 0x72, 0x6f, 0x70, 0x6f, 0x73, 0x65,
	0x64, 0x5f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0e, 0x2e,
	0x50, 0x72, 0x6f, 0x70, 0x6f, 0x73, 0x65, 0x64, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x52, 0x1b, 0x6c,
	0x61, 0x74, 0x65, 0x73, 0x74, 0x50, 0x72, 0x65, 0x70, 0x61, 0x72, 0x65, 0x64, 0x50, 0x72, 0x6f,
	0x70, 0x6f, 0x73, 0x65, 0x64, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x12, 0x54, 0x0a, 0x1b, 0x6c, 0x61,
	0x74, 0x65, 0x73, 0x74, 0x5f, 0x70, 0x72, 0x65, 0x70, 0x61, 0x72, 0x65, 0x64, 0x5f, 0x63, 0x65,
	0x72, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x14, 0x2e, 0x50, 0x72, 0x65, 0x70, 0x61, 0x72, 0x65, 0x64, 0x43, 0x65, 0x72, 0x74, 0x69, 0x66,
	0x69, 0x63, 0x61, 0x74, 0x65, 0x52, 0x19, 0x6c, 0x61, 0x74, 0x65, 0x73, 0x74, 0x50, 0x72, 0x65,
	0x70, 0x61, 0x72, 0x65, 0x64, 0x43, 0x65, 0x72, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x65,
	0x42, 0x0f, 0x5a, 0x0d, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x2f, 0x74, 0x79, 0x70, 0x65,
	0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_message_types_proto_message_proto_rawDescOnce sync.Once
	file_message_types_proto_message_proto_rawDescData = file_message_types_proto_message_proto_rawDesc
)

func file_message_types_proto_message_proto_rawDescGZIP() []byte {
	file_message_types_proto_message_proto_rawDescOnce.Do(func() {
		file_message_types_proto_message_proto_rawDescData = protoimpl.X.CompressGZIP(file_message_types_proto_message_proto_rawDescData)
	})
	return file_message_types_proto_message_proto_rawDescData
}

var file_message_types_proto_message_proto_msgTypes = make([]protoimpl.MessageInfo, 8)
var file_message_types_proto_message_proto_goTypes = []interface{}{
	(*View)(nil),                   // 0: View
	(*ProposedBlock)(nil),          // 1: ProposedBlock
	(*PreparedCertificate)(nil),    // 2: PreparedCertificate
	(*RoundChangeCertificate)(nil), // 3: RoundChangeCertificate
	(*MsgProposal)(nil),            // 4: MsgProposal
	(*MsgPrepare)(nil),             // 5: MsgPrepare
	(*MsgCommit)(nil),              // 6: MsgCommit
	(*MsgRoundChange)(nil),         // 7: MsgRoundChange
}
var file_message_types_proto_message_proto_depIdxs = []int32{
	4,  // 0: PreparedCertificate.proposal_message:type_name -> MsgProposal
	5,  // 1: PreparedCertificate.prepare_messages:type_name -> MsgPrepare
	7,  // 2: RoundChangeCertificate.messages:type_name -> MsgRoundChange
	0,  // 3: MsgProposal.view:type_name -> View
	1,  // 4: MsgProposal.proposed_block:type_name -> ProposedBlock
	3,  // 5: MsgProposal.round_change_certificate:type_name -> RoundChangeCertificate
	0,  // 6: MsgPrepare.view:type_name -> View
	0,  // 7: MsgCommit.view:type_name -> View
	0,  // 8: MsgRoundChange.view:type_name -> View
	1,  // 9: MsgRoundChange.latest_prepared_proposed_block:type_name -> ProposedBlock
	2,  // 10: MsgRoundChange.latest_prepared_certificate:type_name -> PreparedCertificate
	11, // [11:11] is the sub-list for method output_type
	11, // [11:11] is the sub-list for method input_type
	11, // [11:11] is the sub-list for extension type_name
	11, // [11:11] is the sub-list for extension extendee
	0,  // [0:11] is the sub-list for field type_name
}

func init() { file_message_types_proto_message_proto_init() }
func file_message_types_proto_message_proto_init() {
	if File_message_types_proto_message_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_message_types_proto_message_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*View); i {
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
		file_message_types_proto_message_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ProposedBlock); i {
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
		file_message_types_proto_message_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PreparedCertificate); i {
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
		file_message_types_proto_message_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RoundChangeCertificate); i {
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
		file_message_types_proto_message_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MsgProposal); i {
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
		file_message_types_proto_message_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MsgPrepare); i {
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
		file_message_types_proto_message_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MsgCommit); i {
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
		file_message_types_proto_message_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MsgRoundChange); i {
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
			RawDescriptor: file_message_types_proto_message_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   8,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_message_types_proto_message_proto_goTypes,
		DependencyIndexes: file_message_types_proto_message_proto_depIdxs,
		MessageInfos:      file_message_types_proto_message_proto_msgTypes,
	}.Build()
	File_message_types_proto_message_proto = out.File
	file_message_types_proto_message_proto_rawDesc = nil
	file_message_types_proto_message_proto_goTypes = nil
	file_message_types_proto_message_proto_depIdxs = nil
}
