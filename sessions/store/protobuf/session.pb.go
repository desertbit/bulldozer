// Code generated by protoc-gen-gogo.
// source: session.proto
// DO NOT EDIT!

/*
Package protobuf is a generated protocol buffer package.

It is generated from these files:
	session.proto

It has these top-level messages:
	Session
*/
package protobuf

import proto "code.google.com/p/gogoprotobuf/proto"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = math.Inf

type Session struct {
	Values           []byte `protobuf:"bytes,1,opt" json:"Values,omitempty"`
	ExpiresAt        *int64 `protobuf:"varint,2,opt" json:"ExpiresAt,omitempty"`
	XXX_unrecognized []byte `json:"-"`
}

func (m *Session) Reset()         { *m = Session{} }
func (m *Session) String() string { return proto.CompactTextString(m) }
func (*Session) ProtoMessage()    {}

func (m *Session) GetValues() []byte {
	if m != nil {
		return m.Values
	}
	return nil
}

func (m *Session) GetExpiresAt() int64 {
	if m != nil && m.ExpiresAt != nil {
		return *m.ExpiresAt
	}
	return 0
}

func init() {
}
