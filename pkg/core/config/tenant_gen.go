package config

// NOTE: THIS FILE WAS PRODUCED BY THE
// MSGP CODE GENERATION TOOL (github.com/tinylib/msgp)
// DO NOT EDIT

import (
	"github.com/tinylib/msgp/msgp"
)

// DecodeMsg implements msgp.Decodable
func (z *SMTPConfiguration) DecodeMsg(dc *msgp.Reader) (err error) {
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
		case "HOST":
			z.Host, err = dc.ReadString()
			if err != nil {
				return
			}
		case "PORT":
			z.Port, err = dc.ReadInt()
			if err != nil {
				return
			}
		case "MODE":
			z.Mode, err = dc.ReadString()
			if err != nil {
				return
			}
		case "LOGIN":
			z.Login, err = dc.ReadString()
			if err != nil {
				return
			}
		case "PASSWORD":
			z.Password, err = dc.ReadString()
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
func (z *SMTPConfiguration) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 5
	// write "HOST"
	err = en.Append(0x85, 0xa4, 0x48, 0x4f, 0x53, 0x54)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Host)
	if err != nil {
		return
	}
	// write "PORT"
	err = en.Append(0xa4, 0x50, 0x4f, 0x52, 0x54)
	if err != nil {
		return err
	}
	err = en.WriteInt(z.Port)
	if err != nil {
		return
	}
	// write "MODE"
	err = en.Append(0xa4, 0x4d, 0x4f, 0x44, 0x45)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Mode)
	if err != nil {
		return
	}
	// write "LOGIN"
	err = en.Append(0xa5, 0x4c, 0x4f, 0x47, 0x49, 0x4e)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Login)
	if err != nil {
		return
	}
	// write "PASSWORD"
	err = en.Append(0xa8, 0x50, 0x41, 0x53, 0x53, 0x57, 0x4f, 0x52, 0x44)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Password)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *SMTPConfiguration) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 5
	// string "HOST"
	o = append(o, 0x85, 0xa4, 0x48, 0x4f, 0x53, 0x54)
	o = msgp.AppendString(o, z.Host)
	// string "PORT"
	o = append(o, 0xa4, 0x50, 0x4f, 0x52, 0x54)
	o = msgp.AppendInt(o, z.Port)
	// string "MODE"
	o = append(o, 0xa4, 0x4d, 0x4f, 0x44, 0x45)
	o = msgp.AppendString(o, z.Mode)
	// string "LOGIN"
	o = append(o, 0xa5, 0x4c, 0x4f, 0x47, 0x49, 0x4e)
	o = msgp.AppendString(o, z.Login)
	// string "PASSWORD"
	o = append(o, 0xa8, 0x50, 0x41, 0x53, 0x53, 0x57, 0x4f, 0x52, 0x44)
	o = msgp.AppendString(o, z.Password)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *SMTPConfiguration) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "HOST":
			z.Host, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "PORT":
			z.Port, bts, err = msgp.ReadIntBytes(bts)
			if err != nil {
				return
			}
		case "MODE":
			z.Mode, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "LOGIN":
			z.Login, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "PASSWORD":
			z.Password, bts, err = msgp.ReadStringBytes(bts)
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
func (z *SMTPConfiguration) Msgsize() (s int) {
	s = 1 + 5 + msgp.StringPrefixSize + len(z.Host) + 5 + msgp.IntSize + 5 + msgp.StringPrefixSize + len(z.Mode) + 6 + msgp.StringPrefixSize + len(z.Login) + 9 + msgp.StringPrefixSize + len(z.Password)
	return
}

// DecodeMsg implements msgp.Decodable
func (z *SSOConfiguration) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
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
		case "NAME":
			z.Name, err = dc.ReadString()
			if err != nil {
				return
			}
		case "ENABLED":
			z.Enabled, err = dc.ReadBool()
			if err != nil {
				return
			}
		case "CLIENT_ID":
			z.ClientID, err = dc.ReadString()
			if err != nil {
				return
			}
		case "CLIENT_SECRET":
			z.ClientSecret, err = dc.ReadString()
			if err != nil {
				return
			}
		case "SCOPE":
			z.Scope, err = dc.ReadString()
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
func (z *SSOConfiguration) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 5
	// write "NAME"
	err = en.Append(0x85, 0xa4, 0x4e, 0x41, 0x4d, 0x45)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Name)
	if err != nil {
		return
	}
	// write "ENABLED"
	err = en.Append(0xa7, 0x45, 0x4e, 0x41, 0x42, 0x4c, 0x45, 0x44)
	if err != nil {
		return err
	}
	err = en.WriteBool(z.Enabled)
	if err != nil {
		return
	}
	// write "CLIENT_ID"
	err = en.Append(0xa9, 0x43, 0x4c, 0x49, 0x45, 0x4e, 0x54, 0x5f, 0x49, 0x44)
	if err != nil {
		return err
	}
	err = en.WriteString(z.ClientID)
	if err != nil {
		return
	}
	// write "CLIENT_SECRET"
	err = en.Append(0xad, 0x43, 0x4c, 0x49, 0x45, 0x4e, 0x54, 0x5f, 0x53, 0x45, 0x43, 0x52, 0x45, 0x54)
	if err != nil {
		return err
	}
	err = en.WriteString(z.ClientSecret)
	if err != nil {
		return
	}
	// write "SCOPE"
	err = en.Append(0xa5, 0x53, 0x43, 0x4f, 0x50, 0x45)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Scope)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *SSOConfiguration) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 5
	// string "NAME"
	o = append(o, 0x85, 0xa4, 0x4e, 0x41, 0x4d, 0x45)
	o = msgp.AppendString(o, z.Name)
	// string "ENABLED"
	o = append(o, 0xa7, 0x45, 0x4e, 0x41, 0x42, 0x4c, 0x45, 0x44)
	o = msgp.AppendBool(o, z.Enabled)
	// string "CLIENT_ID"
	o = append(o, 0xa9, 0x43, 0x4c, 0x49, 0x45, 0x4e, 0x54, 0x5f, 0x49, 0x44)
	o = msgp.AppendString(o, z.ClientID)
	// string "CLIENT_SECRET"
	o = append(o, 0xad, 0x43, 0x4c, 0x49, 0x45, 0x4e, 0x54, 0x5f, 0x53, 0x45, 0x43, 0x52, 0x45, 0x54)
	o = msgp.AppendString(o, z.ClientSecret)
	// string "SCOPE"
	o = append(o, 0xa5, 0x53, 0x43, 0x4f, 0x50, 0x45)
	o = msgp.AppendString(o, z.Scope)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *SSOConfiguration) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zcmr uint32
	zcmr, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zcmr > 0 {
		zcmr--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "NAME":
			z.Name, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "ENABLED":
			z.Enabled, bts, err = msgp.ReadBoolBytes(bts)
			if err != nil {
				return
			}
		case "CLIENT_ID":
			z.ClientID, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "CLIENT_SECRET":
			z.ClientSecret, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "SCOPE":
			z.Scope, bts, err = msgp.ReadStringBytes(bts)
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
func (z *SSOConfiguration) Msgsize() (s int) {
	s = 1 + 5 + msgp.StringPrefixSize + len(z.Name) + 8 + msgp.BoolSize + 10 + msgp.StringPrefixSize + len(z.ClientID) + 14 + msgp.StringPrefixSize + len(z.ClientSecret) + 6 + msgp.StringPrefixSize + len(z.Scope)
	return
}

// DecodeMsg implements msgp.Decodable
func (z *SSOSetting) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zhct uint32
	zhct, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zhct > 0 {
		zhct--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "URL_PREFIX":
			z.URLPrefix, err = dc.ReadString()
			if err != nil {
				return
			}
		case "JS_SDK_CDN_URL":
			z.JSSDKCDNURL, err = dc.ReadString()
			if err != nil {
				return
			}
		case "STATE_JWT_SECRET":
			z.StateJWTSecret, err = dc.ReadString()
			if err != nil {
				return
			}
		case "AUTO_LINK_PROVIDER_KEYS":
			var zcua uint32
			zcua, err = dc.ReadArrayHeader()
			if err != nil {
				return
			}
			if cap(z.AutoLinkProviderKeys) >= int(zcua) {
				z.AutoLinkProviderKeys = (z.AutoLinkProviderKeys)[:zcua]
			} else {
				z.AutoLinkProviderKeys = make([]string, zcua)
			}
			for zajw := range z.AutoLinkProviderKeys {
				z.AutoLinkProviderKeys[zajw], err = dc.ReadString()
				if err != nil {
					return
				}
			}
		case "ALLOWED_CALLBACK_URLS":
			var zxhx uint32
			zxhx, err = dc.ReadArrayHeader()
			if err != nil {
				return
			}
			if cap(z.AllowedCallbackURLs) >= int(zxhx) {
				z.AllowedCallbackURLs = (z.AllowedCallbackURLs)[:zxhx]
			} else {
				z.AllowedCallbackURLs = make([]string, zxhx)
			}
			for zwht := range z.AllowedCallbackURLs {
				z.AllowedCallbackURLs[zwht], err = dc.ReadString()
				if err != nil {
					return
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
func (z *SSOSetting) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 5
	// write "URL_PREFIX"
	err = en.Append(0x85, 0xaa, 0x55, 0x52, 0x4c, 0x5f, 0x50, 0x52, 0x45, 0x46, 0x49, 0x58)
	if err != nil {
		return err
	}
	err = en.WriteString(z.URLPrefix)
	if err != nil {
		return
	}
	// write "JS_SDK_CDN_URL"
	err = en.Append(0xae, 0x4a, 0x53, 0x5f, 0x53, 0x44, 0x4b, 0x5f, 0x43, 0x44, 0x4e, 0x5f, 0x55, 0x52, 0x4c)
	if err != nil {
		return err
	}
	err = en.WriteString(z.JSSDKCDNURL)
	if err != nil {
		return
	}
	// write "STATE_JWT_SECRET"
	err = en.Append(0xb0, 0x53, 0x54, 0x41, 0x54, 0x45, 0x5f, 0x4a, 0x57, 0x54, 0x5f, 0x53, 0x45, 0x43, 0x52, 0x45, 0x54)
	if err != nil {
		return err
	}
	err = en.WriteString(z.StateJWTSecret)
	if err != nil {
		return
	}
	// write "AUTO_LINK_PROVIDER_KEYS"
	err = en.Append(0xb7, 0x41, 0x55, 0x54, 0x4f, 0x5f, 0x4c, 0x49, 0x4e, 0x4b, 0x5f, 0x50, 0x52, 0x4f, 0x56, 0x49, 0x44, 0x45, 0x52, 0x5f, 0x4b, 0x45, 0x59, 0x53)
	if err != nil {
		return err
	}
	err = en.WriteArrayHeader(uint32(len(z.AutoLinkProviderKeys)))
	if err != nil {
		return
	}
	for zajw := range z.AutoLinkProviderKeys {
		err = en.WriteString(z.AutoLinkProviderKeys[zajw])
		if err != nil {
			return
		}
	}
	// write "ALLOWED_CALLBACK_URLS"
	err = en.Append(0xb5, 0x41, 0x4c, 0x4c, 0x4f, 0x57, 0x45, 0x44, 0x5f, 0x43, 0x41, 0x4c, 0x4c, 0x42, 0x41, 0x43, 0x4b, 0x5f, 0x55, 0x52, 0x4c, 0x53)
	if err != nil {
		return err
	}
	err = en.WriteArrayHeader(uint32(len(z.AllowedCallbackURLs)))
	if err != nil {
		return
	}
	for zwht := range z.AllowedCallbackURLs {
		err = en.WriteString(z.AllowedCallbackURLs[zwht])
		if err != nil {
			return
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *SSOSetting) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 5
	// string "URL_PREFIX"
	o = append(o, 0x85, 0xaa, 0x55, 0x52, 0x4c, 0x5f, 0x50, 0x52, 0x45, 0x46, 0x49, 0x58)
	o = msgp.AppendString(o, z.URLPrefix)
	// string "JS_SDK_CDN_URL"
	o = append(o, 0xae, 0x4a, 0x53, 0x5f, 0x53, 0x44, 0x4b, 0x5f, 0x43, 0x44, 0x4e, 0x5f, 0x55, 0x52, 0x4c)
	o = msgp.AppendString(o, z.JSSDKCDNURL)
	// string "STATE_JWT_SECRET"
	o = append(o, 0xb0, 0x53, 0x54, 0x41, 0x54, 0x45, 0x5f, 0x4a, 0x57, 0x54, 0x5f, 0x53, 0x45, 0x43, 0x52, 0x45, 0x54)
	o = msgp.AppendString(o, z.StateJWTSecret)
	// string "AUTO_LINK_PROVIDER_KEYS"
	o = append(o, 0xb7, 0x41, 0x55, 0x54, 0x4f, 0x5f, 0x4c, 0x49, 0x4e, 0x4b, 0x5f, 0x50, 0x52, 0x4f, 0x56, 0x49, 0x44, 0x45, 0x52, 0x5f, 0x4b, 0x45, 0x59, 0x53)
	o = msgp.AppendArrayHeader(o, uint32(len(z.AutoLinkProviderKeys)))
	for zajw := range z.AutoLinkProviderKeys {
		o = msgp.AppendString(o, z.AutoLinkProviderKeys[zajw])
	}
	// string "ALLOWED_CALLBACK_URLS"
	o = append(o, 0xb5, 0x41, 0x4c, 0x4c, 0x4f, 0x57, 0x45, 0x44, 0x5f, 0x43, 0x41, 0x4c, 0x4c, 0x42, 0x41, 0x43, 0x4b, 0x5f, 0x55, 0x52, 0x4c, 0x53)
	o = msgp.AppendArrayHeader(o, uint32(len(z.AllowedCallbackURLs)))
	for zwht := range z.AllowedCallbackURLs {
		o = msgp.AppendString(o, z.AllowedCallbackURLs[zwht])
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *SSOSetting) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "URL_PREFIX":
			z.URLPrefix, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "JS_SDK_CDN_URL":
			z.JSSDKCDNURL, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "STATE_JWT_SECRET":
			z.StateJWTSecret, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "AUTO_LINK_PROVIDER_KEYS":
			var zdaf uint32
			zdaf, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.AutoLinkProviderKeys) >= int(zdaf) {
				z.AutoLinkProviderKeys = (z.AutoLinkProviderKeys)[:zdaf]
			} else {
				z.AutoLinkProviderKeys = make([]string, zdaf)
			}
			for zajw := range z.AutoLinkProviderKeys {
				z.AutoLinkProviderKeys[zajw], bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
			}
		case "ALLOWED_CALLBACK_URLS":
			var zpks uint32
			zpks, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.AllowedCallbackURLs) >= int(zpks) {
				z.AllowedCallbackURLs = (z.AllowedCallbackURLs)[:zpks]
			} else {
				z.AllowedCallbackURLs = make([]string, zpks)
			}
			for zwht := range z.AllowedCallbackURLs {
				z.AllowedCallbackURLs[zwht], bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
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
func (z *SSOSetting) Msgsize() (s int) {
	s = 1 + 11 + msgp.StringPrefixSize + len(z.URLPrefix) + 15 + msgp.StringPrefixSize + len(z.JSSDKCDNURL) + 17 + msgp.StringPrefixSize + len(z.StateJWTSecret) + 24 + msgp.ArrayHeaderSize
	for zajw := range z.AutoLinkProviderKeys {
		s += msgp.StringPrefixSize + len(z.AutoLinkProviderKeys[zajw])
	}
	s += 22 + msgp.ArrayHeaderSize
	for zwht := range z.AllowedCallbackURLs {
		s += msgp.StringPrefixSize + len(z.AllowedCallbackURLs[zwht])
	}
	return
}

// DecodeMsg implements msgp.Decodable
func (z *TenantConfiguration) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zcxo uint32
	zcxo, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zcxo > 0 {
		zcxo--
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
			var zeff uint32
			zeff, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			for zeff > 0 {
				zeff--
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
			var zrsw uint32
			zrsw, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			for zrsw > 0 {
				zrsw--
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
			var zxpk uint32
			zxpk, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			for zxpk > 0 {
				zxpk--
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
		case "SMTP":
			err = z.SMTP.DecodeMsg(dc)
			if err != nil {
				return
			}
		case "WELCOME_EMAIL":
			err = z.WelcomeEmail.DecodeMsg(dc)
			if err != nil {
				return
			}
		case "SSO_SETTING":
			err = z.SSOSetting.DecodeMsg(dc)
			if err != nil {
				return
			}
		case "SSO_CONFIGS":
			var zdnj uint32
			zdnj, err = dc.ReadArrayHeader()
			if err != nil {
				return
			}
			if cap(z.SSOConfigs) >= int(zdnj) {
				z.SSOConfigs = (z.SSOConfigs)[:zdnj]
			} else {
				z.SSOConfigs = make([]SSOConfiguration, zdnj)
			}
			for zjfb := range z.SSOConfigs {
				err = z.SSOConfigs[zjfb].DecodeMsg(dc)
				if err != nil {
					return
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
	// map header, size 12
	// write "DATABASE_URL"
	err = en.Append(0x8c, 0xac, 0x44, 0x41, 0x54, 0x41, 0x42, 0x41, 0x53, 0x45, 0x5f, 0x55, 0x52, 0x4c)
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
	// write "SMTP"
	err = en.Append(0xa4, 0x53, 0x4d, 0x54, 0x50)
	if err != nil {
		return err
	}
	err = z.SMTP.EncodeMsg(en)
	if err != nil {
		return
	}
	// write "WELCOME_EMAIL"
	err = en.Append(0xad, 0x57, 0x45, 0x4c, 0x43, 0x4f, 0x4d, 0x45, 0x5f, 0x45, 0x4d, 0x41, 0x49, 0x4c)
	if err != nil {
		return err
	}
	err = z.WelcomeEmail.EncodeMsg(en)
	if err != nil {
		return
	}
	// write "SSO_SETTING"
	err = en.Append(0xab, 0x53, 0x53, 0x4f, 0x5f, 0x53, 0x45, 0x54, 0x54, 0x49, 0x4e, 0x47)
	if err != nil {
		return err
	}
	err = z.SSOSetting.EncodeMsg(en)
	if err != nil {
		return
	}
	// write "SSO_CONFIGS"
	err = en.Append(0xab, 0x53, 0x53, 0x4f, 0x5f, 0x43, 0x4f, 0x4e, 0x46, 0x49, 0x47, 0x53)
	if err != nil {
		return err
	}
	err = en.WriteArrayHeader(uint32(len(z.SSOConfigs)))
	if err != nil {
		return
	}
	for zjfb := range z.SSOConfigs {
		err = z.SSOConfigs[zjfb].EncodeMsg(en)
		if err != nil {
			return
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *TenantConfiguration) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 12
	// string "DATABASE_URL"
	o = append(o, 0x8c, 0xac, 0x44, 0x41, 0x54, 0x41, 0x42, 0x41, 0x53, 0x45, 0x5f, 0x55, 0x52, 0x4c)
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
	// string "SMTP"
	o = append(o, 0xa4, 0x53, 0x4d, 0x54, 0x50)
	o, err = z.SMTP.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "WELCOME_EMAIL"
	o = append(o, 0xad, 0x57, 0x45, 0x4c, 0x43, 0x4f, 0x4d, 0x45, 0x5f, 0x45, 0x4d, 0x41, 0x49, 0x4c)
	o, err = z.WelcomeEmail.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "SSO_SETTING"
	o = append(o, 0xab, 0x53, 0x53, 0x4f, 0x5f, 0x53, 0x45, 0x54, 0x54, 0x49, 0x4e, 0x47)
	o, err = z.SSOSetting.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "SSO_CONFIGS"
	o = append(o, 0xab, 0x53, 0x53, 0x4f, 0x5f, 0x43, 0x4f, 0x4e, 0x46, 0x49, 0x47, 0x53)
	o = msgp.AppendArrayHeader(o, uint32(len(z.SSOConfigs)))
	for zjfb := range z.SSOConfigs {
		o, err = z.SSOConfigs[zjfb].MarshalMsg(o)
		if err != nil {
			return
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *TenantConfiguration) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zobc uint32
	zobc, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zobc > 0 {
		zobc--
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
			var zsnv uint32
			zsnv, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			for zsnv > 0 {
				zsnv--
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
			var zkgt uint32
			zkgt, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			for zkgt > 0 {
				zkgt--
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
			var zema uint32
			zema, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			for zema > 0 {
				zema--
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
		case "SMTP":
			bts, err = z.SMTP.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "WELCOME_EMAIL":
			bts, err = z.WelcomeEmail.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "SSO_SETTING":
			bts, err = z.SSOSetting.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "SSO_CONFIGS":
			var zpez uint32
			zpez, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.SSOConfigs) >= int(zpez) {
				z.SSOConfigs = (z.SSOConfigs)[:zpez]
			} else {
				z.SSOConfigs = make([]SSOConfiguration, zpez)
			}
			for zjfb := range z.SSOConfigs {
				bts, err = z.SSOConfigs[zjfb].UnmarshalMsg(bts)
				if err != nil {
					return
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
	s = 1 + 13 + msgp.StringPrefixSize + len(z.DBConnectionStr) + 8 + msgp.StringPrefixSize + len(z.APIKey) + 11 + msgp.StringPrefixSize + len(z.MasterKey) + 9 + msgp.StringPrefixSize + len(z.AppName) + 10 + msgp.StringPrefixSize + len(z.CORSHost) + 12 + 1 + 7 + msgp.StringPrefixSize + len(z.TokenStore.Secret) + 7 + msgp.Int64Size + 13 + 1 + 15 + msgp.StringPrefixSize + len(z.UserProfile.ImplName) + 15 + msgp.StringPrefixSize + len(z.UserProfile.ImplStoreURL) + 11 + 1 + 8 + msgp.BoolSize + 18 + msgp.StringPrefixSize + len(z.UserAudit.TrailHandlerURL) + 5 + z.SMTP.Msgsize() + 14 + z.WelcomeEmail.Msgsize() + 12 + z.SSOSetting.Msgsize() + 12 + msgp.ArrayHeaderSize
	for zjfb := range z.SSOConfigs {
		s += z.SSOConfigs[zjfb].Msgsize()
	}
	return
}

// DecodeMsg implements msgp.Decodable
func (z *TokenStoreConfiguration) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zqke uint32
	zqke, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zqke > 0 {
		zqke--
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
	var zqyh uint32
	zqyh, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zqyh > 0 {
		zqyh--
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
	var zyzr uint32
	zyzr, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zyzr > 0 {
		zyzr--
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
	var zywj uint32
	zywj, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zywj > 0 {
		zywj--
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
	var zjpj uint32
	zjpj, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zjpj > 0 {
		zjpj--
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
	var zzpf uint32
	zzpf, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zzpf > 0 {
		zzpf--
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

// DecodeMsg implements msgp.Decodable
func (z *WelcomeEmailConfiguration) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zrfe uint32
	zrfe, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zrfe > 0 {
		zrfe--
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
		case "SENDER_NAME":
			z.SenderName, err = dc.ReadString()
			if err != nil {
				return
			}
		case "SENDER":
			z.Sender, err = dc.ReadString()
			if err != nil {
				return
			}
		case "SUBJECT":
			z.Subject, err = dc.ReadString()
			if err != nil {
				return
			}
		case "REPLY_TO_NAME":
			z.ReplyToName, err = dc.ReadString()
			if err != nil {
				return
			}
		case "REPLY_TO":
			z.ReplyTo, err = dc.ReadString()
			if err != nil {
				return
			}
		case "TEXT_URL":
			z.TextURL, err = dc.ReadString()
			if err != nil {
				return
			}
		case "HTML_URL":
			z.HTMLURL, err = dc.ReadString()
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
func (z *WelcomeEmailConfiguration) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 8
	// write "ENABLED"
	err = en.Append(0x88, 0xa7, 0x45, 0x4e, 0x41, 0x42, 0x4c, 0x45, 0x44)
	if err != nil {
		return err
	}
	err = en.WriteBool(z.Enabled)
	if err != nil {
		return
	}
	// write "SENDER_NAME"
	err = en.Append(0xab, 0x53, 0x45, 0x4e, 0x44, 0x45, 0x52, 0x5f, 0x4e, 0x41, 0x4d, 0x45)
	if err != nil {
		return err
	}
	err = en.WriteString(z.SenderName)
	if err != nil {
		return
	}
	// write "SENDER"
	err = en.Append(0xa6, 0x53, 0x45, 0x4e, 0x44, 0x45, 0x52)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Sender)
	if err != nil {
		return
	}
	// write "SUBJECT"
	err = en.Append(0xa7, 0x53, 0x55, 0x42, 0x4a, 0x45, 0x43, 0x54)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Subject)
	if err != nil {
		return
	}
	// write "REPLY_TO_NAME"
	err = en.Append(0xad, 0x52, 0x45, 0x50, 0x4c, 0x59, 0x5f, 0x54, 0x4f, 0x5f, 0x4e, 0x41, 0x4d, 0x45)
	if err != nil {
		return err
	}
	err = en.WriteString(z.ReplyToName)
	if err != nil {
		return
	}
	// write "REPLY_TO"
	err = en.Append(0xa8, 0x52, 0x45, 0x50, 0x4c, 0x59, 0x5f, 0x54, 0x4f)
	if err != nil {
		return err
	}
	err = en.WriteString(z.ReplyTo)
	if err != nil {
		return
	}
	// write "TEXT_URL"
	err = en.Append(0xa8, 0x54, 0x45, 0x58, 0x54, 0x5f, 0x55, 0x52, 0x4c)
	if err != nil {
		return err
	}
	err = en.WriteString(z.TextURL)
	if err != nil {
		return
	}
	// write "HTML_URL"
	err = en.Append(0xa8, 0x48, 0x54, 0x4d, 0x4c, 0x5f, 0x55, 0x52, 0x4c)
	if err != nil {
		return err
	}
	err = en.WriteString(z.HTMLURL)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *WelcomeEmailConfiguration) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 8
	// string "ENABLED"
	o = append(o, 0x88, 0xa7, 0x45, 0x4e, 0x41, 0x42, 0x4c, 0x45, 0x44)
	o = msgp.AppendBool(o, z.Enabled)
	// string "SENDER_NAME"
	o = append(o, 0xab, 0x53, 0x45, 0x4e, 0x44, 0x45, 0x52, 0x5f, 0x4e, 0x41, 0x4d, 0x45)
	o = msgp.AppendString(o, z.SenderName)
	// string "SENDER"
	o = append(o, 0xa6, 0x53, 0x45, 0x4e, 0x44, 0x45, 0x52)
	o = msgp.AppendString(o, z.Sender)
	// string "SUBJECT"
	o = append(o, 0xa7, 0x53, 0x55, 0x42, 0x4a, 0x45, 0x43, 0x54)
	o = msgp.AppendString(o, z.Subject)
	// string "REPLY_TO_NAME"
	o = append(o, 0xad, 0x52, 0x45, 0x50, 0x4c, 0x59, 0x5f, 0x54, 0x4f, 0x5f, 0x4e, 0x41, 0x4d, 0x45)
	o = msgp.AppendString(o, z.ReplyToName)
	// string "REPLY_TO"
	o = append(o, 0xa8, 0x52, 0x45, 0x50, 0x4c, 0x59, 0x5f, 0x54, 0x4f)
	o = msgp.AppendString(o, z.ReplyTo)
	// string "TEXT_URL"
	o = append(o, 0xa8, 0x54, 0x45, 0x58, 0x54, 0x5f, 0x55, 0x52, 0x4c)
	o = msgp.AppendString(o, z.TextURL)
	// string "HTML_URL"
	o = append(o, 0xa8, 0x48, 0x54, 0x4d, 0x4c, 0x5f, 0x55, 0x52, 0x4c)
	o = msgp.AppendString(o, z.HTMLURL)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *WelcomeEmailConfiguration) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zgmo uint32
	zgmo, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zgmo > 0 {
		zgmo--
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
		case "SENDER_NAME":
			z.SenderName, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "SENDER":
			z.Sender, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "SUBJECT":
			z.Subject, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "REPLY_TO_NAME":
			z.ReplyToName, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "REPLY_TO":
			z.ReplyTo, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "TEXT_URL":
			z.TextURL, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "HTML_URL":
			z.HTMLURL, bts, err = msgp.ReadStringBytes(bts)
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
func (z *WelcomeEmailConfiguration) Msgsize() (s int) {
	s = 1 + 8 + msgp.BoolSize + 12 + msgp.StringPrefixSize + len(z.SenderName) + 7 + msgp.StringPrefixSize + len(z.Sender) + 8 + msgp.StringPrefixSize + len(z.Subject) + 14 + msgp.StringPrefixSize + len(z.ReplyToName) + 9 + msgp.StringPrefixSize + len(z.ReplyTo) + 9 + msgp.StringPrefixSize + len(z.TextURL) + 9 + msgp.StringPrefixSize + len(z.HTMLURL)
	return
}
