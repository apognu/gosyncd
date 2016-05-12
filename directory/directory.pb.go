// Code generated by protoc-gen-go.
// source: directory.proto
// DO NOT EDIT!

/*
Package directory is a generated protocol buffer package.

It is generated from these files:
	directory.proto

It has these top-level messages:
	Directory
*/
package directory

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
const _ = proto.ProtoPackageIsVersion1

type Directory struct {
	Files            []*Directory_File `protobuf:"bytes,1,rep,name=files" json:"files,omitempty"`
	XXX_unrecognized []byte            `json:"-"`
}

func (m *Directory) Reset()                    { *m = Directory{} }
func (m *Directory) String() string            { return proto.CompactTextString(m) }
func (*Directory) ProtoMessage()               {}
func (*Directory) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *Directory) GetFiles() []*Directory_File {
	if m != nil {
		return m.Files
	}
	return nil
}

type Directory_File struct {
	Path             *string `protobuf:"bytes,1,req,name=path" json:"path,omitempty"`
	Mode             *uint32 `protobuf:"varint,2,req,name=mode" json:"mode,omitempty"`
	Checksum         *string `protobuf:"bytes,3,req,name=checksum" json:"checksum,omitempty"`
	LastUpdate       *int64  `protobuf:"varint,4,req,name=lastUpdate" json:"lastUpdate,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *Directory_File) Reset()                    { *m = Directory_File{} }
func (m *Directory_File) String() string            { return proto.CompactTextString(m) }
func (*Directory_File) ProtoMessage()               {}
func (*Directory_File) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0, 0} }

func (m *Directory_File) GetPath() string {
	if m != nil && m.Path != nil {
		return *m.Path
	}
	return ""
}

func (m *Directory_File) GetMode() uint32 {
	if m != nil && m.Mode != nil {
		return *m.Mode
	}
	return 0
}

func (m *Directory_File) GetChecksum() string {
	if m != nil && m.Checksum != nil {
		return *m.Checksum
	}
	return ""
}

func (m *Directory_File) GetLastUpdate() int64 {
	if m != nil && m.LastUpdate != nil {
		return *m.LastUpdate
	}
	return 0
}

func init() {
	proto.RegisterType((*Directory)(nil), "Directory")
	proto.RegisterType((*Directory_File)(nil), "Directory.File")
}

var fileDescriptor0 = []byte{
	// 135 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0xe2, 0xe2, 0x4f, 0xc9, 0x2c, 0x4a,
	0x4d, 0x2e, 0xc9, 0x2f, 0xaa, 0xd4, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x57, 0x2a, 0xe5, 0xe2, 0x74,
	0x81, 0x09, 0x09, 0xc9, 0x71, 0xb1, 0xa6, 0x65, 0xe6, 0xa4, 0x16, 0x4b, 0x30, 0x2a, 0x30, 0x6b,
	0x70, 0x1b, 0xf1, 0xeb, 0xc1, 0xa5, 0xf4, 0xdc, 0x80, 0xe2, 0x52, 0x1e, 0x5c, 0x2c, 0x20, 0x5a,
	0x88, 0x87, 0x8b, 0xa5, 0x20, 0xb1, 0x24, 0x03, 0xa8, 0x8c, 0x49, 0x83, 0x13, 0xc4, 0xcb, 0xcd,
	0x4f, 0x49, 0x95, 0x60, 0x02, 0xf2, 0x78, 0x85, 0x04, 0xb8, 0x38, 0x92, 0x33, 0x52, 0x93, 0xb3,
	0x8b, 0x4b, 0x73, 0x25, 0x98, 0xc1, 0xf2, 0x42, 0x5c, 0x5c, 0x39, 0x89, 0xc5, 0x25, 0xa1, 0x05,
	0x29, 0x89, 0x25, 0xa9, 0x12, 0x2c, 0x40, 0x31, 0x66, 0x40, 0x00, 0x00, 0x00, 0xff, 0xff, 0x3d,
	0x59, 0xd4, 0x60, 0x88, 0x00, 0x00, 0x00,
}