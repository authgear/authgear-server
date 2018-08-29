package config

// NOTE: THIS FILE WAS PRODUCED BY THE
// MSGP CODE GENERATION TOOL (github.com/tinylib/msgp)
// DO NOT EDIT

import (
	"github.com/tinylib/msgp/msgp"
)

// DecodeMsg implements msgp.Decodable
func (z *TenantConfiguration) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zxvk uint32
	zxvk, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zxvk > 0 {
		zxvk--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "DATABASE_URL":
			z.DBConnectionStr, err = dc.ReadString()
			if err != nil {
				return
			}
		case "API_KEY":
			z.APIKey, err = dc.ReadString()
			if err != nil {
				return
			}
		case "MASRER_KEY":
			z.MasterKey, err = dc.ReadString()
			if err != nil {
				return
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z TenantConfiguration) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 3
	// write "DATABASE_URL"
	err = en.Append(0x83, 0xac, 0x44, 0x41, 0x54, 0x41, 0x42, 0x41, 0x53, 0x45, 0x5f, 0x55, 0x52, 0x4c)
	if err != nil {
		return err
	}
	err = en.WriteString(z.DBConnectionStr)
	if err != nil {
		return
	}
	// write "API_KEY"
	err = en.Append(0xa7, 0x41, 0x50, 0x49, 0x5f, 0x4b, 0x45, 0x59)
	if err != nil {
		return err
	}
	err = en.WriteString(z.APIKey)
	if err != nil {
		return
	}
	// write "MASRER_KEY"
	err = en.Append(0xaa, 0x4d, 0x41, 0x53, 0x52, 0x45, 0x52, 0x5f, 0x4b, 0x45, 0x59)
	if err != nil {
		return err
	}
	err = en.WriteString(z.MasterKey)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z TenantConfiguration) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 3
	// string "DATABASE_URL"
	o = append(o, 0x83, 0xac, 0x44, 0x41, 0x54, 0x41, 0x42, 0x41, 0x53, 0x45, 0x5f, 0x55, 0x52, 0x4c)
	o = msgp.AppendString(o, z.DBConnectionStr)
	// string "API_KEY"
	o = append(o, 0xa7, 0x41, 0x50, 0x49, 0x5f, 0x4b, 0x45, 0x59)
	o = msgp.AppendString(o, z.APIKey)
	// string "MASRER_KEY"
	o = append(o, 0xaa, 0x4d, 0x41, 0x53, 0x52, 0x45, 0x52, 0x5f, 0x4b, 0x45, 0x59)
	o = msgp.AppendString(o, z.MasterKey)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *TenantConfiguration) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zbzg uint32
	zbzg, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zbzg > 0 {
		zbzg--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "DATABASE_URL":
			z.DBConnectionStr, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "API_KEY":
			z.APIKey, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "MASRER_KEY":
			z.MasterKey, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z TenantConfiguration) Msgsize() (s int) {
	s = 1 + 13 + msgp.StringPrefixSize + len(z.DBConnectionStr) + 8 + msgp.StringPrefixSize + len(z.APIKey) + 11 + msgp.StringPrefixSize + len(z.MasterKey)
	return
}
