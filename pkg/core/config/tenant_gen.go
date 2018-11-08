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
		case "MASTER_KEY":
			z.MasterKey, err = dc.ReadString()
			if err != nil {
				return
			}
		case "APP_NAME":
			z.AppName, err = dc.ReadString()
			if err != nil {
				return
			}
		case "CORS_HOST":
			z.CORSHost, err = dc.ReadString()
			if err != nil {
				return
			}
		case "TOKEN_STORE":
			var zbzg uint32
			zbzg, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			for zbzg > 0 {
				zbzg--
				field, err = dc.ReadMapKeyPtr()
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "SECRET":
					z.TokenStore.Secret, err = dc.ReadString()
					if err != nil {
						return
					}
				case "EXPIRY":
					z.TokenStore.Expiry, err = dc.ReadInt64()
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
		case "USER_PROFILE":
			var zbai uint32
			zbai, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			for zbai > 0 {
				zbai--
				field, err = dc.ReadMapKeyPtr()
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "IMPLEMENTATION":
					z.UserProfile.ImplName, err = dc.ReadString()
					if err != nil {
						return
					}
				case "IMPL_STORE_URL":
					z.UserProfile.ImplStoreURL, err = dc.ReadString()
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
		case "USER_AUDIT":
			var zcmr uint32
			zcmr, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			for zcmr > 0 {
				zcmr--
				field, err = dc.ReadMapKeyPtr()
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "ENABLED":
					z.UserAudit.Enabled, err = dc.ReadBool()
					if err != nil {
						return
					}
				case "TRAIL_HANDLER_URL":
					z.UserAudit.TrailHandlerURL, err = dc.ReadString()
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
func (z *TenantConfiguration) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 8
	// write "DATABASE_URL"
	err = en.Append(0x88, 0xac, 0x44, 0x41, 0x54, 0x41, 0x42, 0x41, 0x53, 0x45, 0x5f, 0x55, 0x52, 0x4c)
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
	// write "MASTER_KEY"
	err = en.Append(0xaa, 0x4d, 0x41, 0x53, 0x54, 0x45, 0x52, 0x5f, 0x4b, 0x45, 0x59)
	if err != nil {
		return err
	}
	err = en.WriteString(z.MasterKey)
	if err != nil {
		return
	}
	// write "APP_NAME"
	err = en.Append(0xa8, 0x41, 0x50, 0x50, 0x5f, 0x4e, 0x41, 0x4d, 0x45)
	if err != nil {
		return err
	}
	err = en.WriteString(z.AppName)
	if err != nil {
		return
	}
	// write "CORS_HOST"
	err = en.Append(0xa9, 0x43, 0x4f, 0x52, 0x53, 0x5f, 0x48, 0x4f, 0x53, 0x54)
	if err != nil {
		return err
	}
	err = en.WriteString(z.CORSHost)
	if err != nil {
		return
	}
	// write "TOKEN_STORE"
	// map header, size 2
	// write "SECRET"
	err = en.Append(0xab, 0x54, 0x4f, 0x4b, 0x45, 0x4e, 0x5f, 0x53, 0x54, 0x4f, 0x52, 0x45, 0x82, 0xa6, 0x53, 0x45, 0x43, 0x52, 0x45, 0x54)
	if err != nil {
		return err
	}
	err = en.WriteString(z.TokenStore.Secret)
	if err != nil {
		return
	}
	// write "EXPIRY"
	err = en.Append(0xa6, 0x45, 0x58, 0x50, 0x49, 0x52, 0x59)
	if err != nil {
		return err
	}
	err = en.WriteInt64(z.TokenStore.Expiry)
	if err != nil {
		return
	}
	// write "USER_PROFILE"
	// map header, size 2
	// write "IMPLEMENTATION"
	err = en.Append(0xac, 0x55, 0x53, 0x45, 0x52, 0x5f, 0x50, 0x52, 0x4f, 0x46, 0x49, 0x4c, 0x45, 0x82, 0xae, 0x49, 0x4d, 0x50, 0x4c, 0x45, 0x4d, 0x45, 0x4e, 0x54, 0x41, 0x54, 0x49, 0x4f, 0x4e)
	if err != nil {
		return err
	}
	err = en.WriteString(z.UserProfile.ImplName)
	if err != nil {
		return
	}
	// write "IMPL_STORE_URL"
	err = en.Append(0xae, 0x49, 0x4d, 0x50, 0x4c, 0x5f, 0x53, 0x54, 0x4f, 0x52, 0x45, 0x5f, 0x55, 0x52, 0x4c)
	if err != nil {
		return err
	}
	err = en.WriteString(z.UserProfile.ImplStoreURL)
	if err != nil {
		return
	}
	// write "USER_AUDIT"
	// map header, size 2
	// write "ENABLED"
	err = en.Append(0xaa, 0x55, 0x53, 0x45, 0x52, 0x5f, 0x41, 0x55, 0x44, 0x49, 0x54, 0x82, 0xa7, 0x45, 0x4e, 0x41, 0x42, 0x4c, 0x45, 0x44)
	if err != nil {
		return err
	}
	err = en.WriteBool(z.UserAudit.Enabled)
	if err != nil {
		return
	}
	// write "TRAIL_HANDLER_URL"
	err = en.Append(0xb1, 0x54, 0x52, 0x41, 0x49, 0x4c, 0x5f, 0x48, 0x41, 0x4e, 0x44, 0x4c, 0x45, 0x52, 0x5f, 0x55, 0x52, 0x4c)
	if err != nil {
		return err
	}
	err = en.WriteString(z.UserAudit.TrailHandlerURL)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *TenantConfiguration) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 8
	// string "DATABASE_URL"
	o = append(o, 0x88, 0xac, 0x44, 0x41, 0x54, 0x41, 0x42, 0x41, 0x53, 0x45, 0x5f, 0x55, 0x52, 0x4c)
	o = msgp.AppendString(o, z.DBConnectionStr)
	// string "API_KEY"
	o = append(o, 0xa7, 0x41, 0x50, 0x49, 0x5f, 0x4b, 0x45, 0x59)
	o = msgp.AppendString(o, z.APIKey)
	// string "MASTER_KEY"
	o = append(o, 0xaa, 0x4d, 0x41, 0x53, 0x54, 0x45, 0x52, 0x5f, 0x4b, 0x45, 0x59)
	o = msgp.AppendString(o, z.MasterKey)
	// string "APP_NAME"
	o = append(o, 0xa8, 0x41, 0x50, 0x50, 0x5f, 0x4e, 0x41, 0x4d, 0x45)
	o = msgp.AppendString(o, z.AppName)
	// string "CORS_HOST"
	o = append(o, 0xa9, 0x43, 0x4f, 0x52, 0x53, 0x5f, 0x48, 0x4f, 0x53, 0x54)
	o = msgp.AppendString(o, z.CORSHost)
	// string "TOKEN_STORE"
	// map header, size 2
	// string "SECRET"
	o = append(o, 0xab, 0x54, 0x4f, 0x4b, 0x45, 0x4e, 0x5f, 0x53, 0x54, 0x4f, 0x52, 0x45, 0x82, 0xa6, 0x53, 0x45, 0x43, 0x52, 0x45, 0x54)
	o = msgp.AppendString(o, z.TokenStore.Secret)
	// string "EXPIRY"
	o = append(o, 0xa6, 0x45, 0x58, 0x50, 0x49, 0x52, 0x59)
	o = msgp.AppendInt64(o, z.TokenStore.Expiry)
	// string "USER_PROFILE"
	// map header, size 2
	// string "IMPLEMENTATION"
	o = append(o, 0xac, 0x55, 0x53, 0x45, 0x52, 0x5f, 0x50, 0x52, 0x4f, 0x46, 0x49, 0x4c, 0x45, 0x82, 0xae, 0x49, 0x4d, 0x50, 0x4c, 0x45, 0x4d, 0x45, 0x4e, 0x54, 0x41, 0x54, 0x49, 0x4f, 0x4e)
	o = msgp.AppendString(o, z.UserProfile.ImplName)
	// string "IMPL_STORE_URL"
	o = append(o, 0xae, 0x49, 0x4d, 0x50, 0x4c, 0x5f, 0x53, 0x54, 0x4f, 0x52, 0x45, 0x5f, 0x55, 0x52, 0x4c)
	o = msgp.AppendString(o, z.UserProfile.ImplStoreURL)
	// string "USER_AUDIT"
	// map header, size 2
	// string "ENABLED"
	o = append(o, 0xaa, 0x55, 0x53, 0x45, 0x52, 0x5f, 0x41, 0x55, 0x44, 0x49, 0x54, 0x82, 0xa7, 0x45, 0x4e, 0x41, 0x42, 0x4c, 0x45, 0x44)
	o = msgp.AppendBool(o, z.UserAudit.Enabled)
	// string "TRAIL_HANDLER_URL"
	o = append(o, 0xb1, 0x54, 0x52, 0x41, 0x49, 0x4c, 0x5f, 0x48, 0x41, 0x4e, 0x44, 0x4c, 0x45, 0x52, 0x5f, 0x55, 0x52, 0x4c)
	o = msgp.AppendString(o, z.UserAudit.TrailHandlerURL)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *TenantConfiguration) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zajw uint32
	zajw, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zajw > 0 {
		zajw--
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
		case "MASTER_KEY":
			z.MasterKey, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "APP_NAME":
			z.AppName, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "CORS_HOST":
			z.CORSHost, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "TOKEN_STORE":
			var zwht uint32
			zwht, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			for zwht > 0 {
				zwht--
				field, bts, err = msgp.ReadMapKeyZC(bts)
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "SECRET":
					z.TokenStore.Secret, bts, err = msgp.ReadStringBytes(bts)
					if err != nil {
						return
					}
				case "EXPIRY":
					z.TokenStore.Expiry, bts, err = msgp.ReadInt64Bytes(bts)
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
		case "USER_PROFILE":
			var zhct uint32
			zhct, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			for zhct > 0 {
				zhct--
				field, bts, err = msgp.ReadMapKeyZC(bts)
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "IMPLEMENTATION":
					z.UserProfile.ImplName, bts, err = msgp.ReadStringBytes(bts)
					if err != nil {
						return
					}
				case "IMPL_STORE_URL":
					z.UserProfile.ImplStoreURL, bts, err = msgp.ReadStringBytes(bts)
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
		case "USER_AUDIT":
			var zcua uint32
			zcua, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			for zcua > 0 {
				zcua--
				field, bts, err = msgp.ReadMapKeyZC(bts)
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "ENABLED":
					z.UserAudit.Enabled, bts, err = msgp.ReadBoolBytes(bts)
					if err != nil {
						return
					}
				case "TRAIL_HANDLER_URL":
					z.UserAudit.TrailHandlerURL, bts, err = msgp.ReadStringBytes(bts)
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
func (z *TenantConfiguration) Msgsize() (s int) {
	s = 1 + 13 + msgp.StringPrefixSize + len(z.DBConnectionStr) + 8 + msgp.StringPrefixSize + len(z.APIKey) + 11 + msgp.StringPrefixSize + len(z.MasterKey) + 9 + msgp.StringPrefixSize + len(z.AppName) + 10 + msgp.StringPrefixSize + len(z.CORSHost) + 12 + 1 + 7 + msgp.StringPrefixSize + len(z.TokenStore.Secret) + 7 + msgp.Int64Size + 13 + 1 + 15 + msgp.StringPrefixSize + len(z.UserProfile.ImplName) + 15 + msgp.StringPrefixSize + len(z.UserProfile.ImplStoreURL) + 11 + 1 + 8 + msgp.BoolSize + 18 + msgp.StringPrefixSize + len(z.UserAudit.TrailHandlerURL)
	return
}

// DecodeMsg implements msgp.Decodable
func (z *TokenStoreConfiguration) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zxhx uint32
	zxhx, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zxhx > 0 {
		zxhx--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "SECRET":
			z.Secret, err = dc.ReadString()
			if err != nil {
				return
			}
		case "EXPIRY":
			z.Expiry, err = dc.ReadInt64()
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
func (z TokenStoreConfiguration) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 2
	// write "SECRET"
	err = en.Append(0x82, 0xa6, 0x53, 0x45, 0x43, 0x52, 0x45, 0x54)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Secret)
	if err != nil {
		return
	}
	// write "EXPIRY"
	err = en.Append(0xa6, 0x45, 0x58, 0x50, 0x49, 0x52, 0x59)
	if err != nil {
		return err
	}
	err = en.WriteInt64(z.Expiry)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z TokenStoreConfiguration) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "SECRET"
	o = append(o, 0x82, 0xa6, 0x53, 0x45, 0x43, 0x52, 0x45, 0x54)
	o = msgp.AppendString(o, z.Secret)
	// string "EXPIRY"
	o = append(o, 0xa6, 0x45, 0x58, 0x50, 0x49, 0x52, 0x59)
	o = msgp.AppendInt64(o, z.Expiry)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *TokenStoreConfiguration) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zlqf uint32
	zlqf, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zlqf > 0 {
		zlqf--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "SECRET":
			z.Secret, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "EXPIRY":
			z.Expiry, bts, err = msgp.ReadInt64Bytes(bts)
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
func (z TokenStoreConfiguration) Msgsize() (s int) {
	s = 1 + 7 + msgp.StringPrefixSize + len(z.Secret) + 7 + msgp.Int64Size
	return
}

// DecodeMsg implements msgp.Decodable
func (z *UserAuditConfiguration) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zdaf uint32
	zdaf, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zdaf > 0 {
		zdaf--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "ENABLED":
			z.Enabled, err = dc.ReadBool()
			if err != nil {
				return
			}
		case "TRAIL_HANDLER_URL":
			z.TrailHandlerURL, err = dc.ReadString()
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
func (z UserAuditConfiguration) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 2
	// write "ENABLED"
	err = en.Append(0x82, 0xa7, 0x45, 0x4e, 0x41, 0x42, 0x4c, 0x45, 0x44)
	if err != nil {
		return err
	}
	err = en.WriteBool(z.Enabled)
	if err != nil {
		return
	}
	// write "TRAIL_HANDLER_URL"
	err = en.Append(0xb1, 0x54, 0x52, 0x41, 0x49, 0x4c, 0x5f, 0x48, 0x41, 0x4e, 0x44, 0x4c, 0x45, 0x52, 0x5f, 0x55, 0x52, 0x4c)
	if err != nil {
		return err
	}
	err = en.WriteString(z.TrailHandlerURL)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z UserAuditConfiguration) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "ENABLED"
	o = append(o, 0x82, 0xa7, 0x45, 0x4e, 0x41, 0x42, 0x4c, 0x45, 0x44)
	o = msgp.AppendBool(o, z.Enabled)
	// string "TRAIL_HANDLER_URL"
	o = append(o, 0xb1, 0x54, 0x52, 0x41, 0x49, 0x4c, 0x5f, 0x48, 0x41, 0x4e, 0x44, 0x4c, 0x45, 0x52, 0x5f, 0x55, 0x52, 0x4c)
	o = msgp.AppendString(o, z.TrailHandlerURL)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *UserAuditConfiguration) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zpks uint32
	zpks, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zpks > 0 {
		zpks--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "ENABLED":
			z.Enabled, bts, err = msgp.ReadBoolBytes(bts)
			if err != nil {
				return
			}
		case "TRAIL_HANDLER_URL":
			z.TrailHandlerURL, bts, err = msgp.ReadStringBytes(bts)
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
func (z UserAuditConfiguration) Msgsize() (s int) {
	s = 1 + 8 + msgp.BoolSize + 18 + msgp.StringPrefixSize + len(z.TrailHandlerURL)
	return
}

// DecodeMsg implements msgp.Decodable
func (z *UserProfileConfiguration) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zjfb uint32
	zjfb, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zjfb > 0 {
		zjfb--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "IMPLEMENTATION":
			z.ImplName, err = dc.ReadString()
			if err != nil {
				return
			}
		case "IMPL_STORE_URL":
			z.ImplStoreURL, err = dc.ReadString()
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
func (z UserProfileConfiguration) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 2
	// write "IMPLEMENTATION"
	err = en.Append(0x82, 0xae, 0x49, 0x4d, 0x50, 0x4c, 0x45, 0x4d, 0x45, 0x4e, 0x54, 0x41, 0x54, 0x49, 0x4f, 0x4e)
	if err != nil {
		return err
	}
	err = en.WriteString(z.ImplName)
	if err != nil {
		return
	}
	// write "IMPL_STORE_URL"
	err = en.Append(0xae, 0x49, 0x4d, 0x50, 0x4c, 0x5f, 0x53, 0x54, 0x4f, 0x52, 0x45, 0x5f, 0x55, 0x52, 0x4c)
	if err != nil {
		return err
	}
	err = en.WriteString(z.ImplStoreURL)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z UserProfileConfiguration) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "IMPLEMENTATION"
	o = append(o, 0x82, 0xae, 0x49, 0x4d, 0x50, 0x4c, 0x45, 0x4d, 0x45, 0x4e, 0x54, 0x41, 0x54, 0x49, 0x4f, 0x4e)
	o = msgp.AppendString(o, z.ImplName)
	// string "IMPL_STORE_URL"
	o = append(o, 0xae, 0x49, 0x4d, 0x50, 0x4c, 0x5f, 0x53, 0x54, 0x4f, 0x52, 0x45, 0x5f, 0x55, 0x52, 0x4c)
	o = msgp.AppendString(o, z.ImplStoreURL)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *UserProfileConfiguration) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zcxo uint32
	zcxo, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zcxo > 0 {
		zcxo--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "IMPLEMENTATION":
			z.ImplName, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "IMPL_STORE_URL":
			z.ImplStoreURL, bts, err = msgp.ReadStringBytes(bts)
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
func (z UserProfileConfiguration) Msgsize() (s int) {
	s = 1 + 15 + msgp.StringPrefixSize + len(z.ImplName) + 15 + msgp.StringPrefixSize + len(z.ImplStoreURL)
	return
}
