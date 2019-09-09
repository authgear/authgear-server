package config

// NOTE: THIS FILE WAS PRODUCED BY THE
// MSGP CODE GENERATION TOOL (github.com/tinylib/msgp)
// DO NOT EDIT

import (
	"github.com/tinylib/msgp/msgp"
)

// DecodeMsg implements msgp.Decodable
func (z *APIClientConfiguration) DecodeMsg(dc *msgp.Reader) (err error) {
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
		case "name":
			z.Name, err = dc.ReadString()
			if err != nil {
				return
			}
		case "disabled":
			z.Disabled, err = dc.ReadBool()
			if err != nil {
				return
			}
		case "api_key":
			z.APIKey, err = dc.ReadString()
			if err != nil {
				return
			}
		case "session_transport":
			{
				var zbzg string
				zbzg, err = dc.ReadString()
				z.SessionTransport = SessionTransportType(zbzg)
			}
			if err != nil {
				return
			}
		case "access_token_lifetime":
			z.AccessTokenLifetime, err = dc.ReadInt()
			if err != nil {
				return
			}
		case "session_idle_timeout_enabled":
			z.SessionIdleTimeoutEnabled, err = dc.ReadBool()
			if err != nil {
				return
			}
		case "session_idle_timeout":
			z.SessionIdleTimeout, err = dc.ReadInt()
			if err != nil {
				return
			}
		case "refresh_token_disabled":
			z.RefreshTokenDisabled, err = dc.ReadBool()
			if err != nil {
				return
			}
		case "refresh_token_lifetime":
			z.RefreshTokenLifetime, err = dc.ReadInt()
			if err != nil {
				return
			}
		case "same_site":
			{
				var zbai string
				zbai, err = dc.ReadString()
				z.SameSite = SessionCookieSameSite(zbai)
			}
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
func (z *APIClientConfiguration) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 10
	// write "name"
	err = en.Append(0x8a, 0xa4, 0x6e, 0x61, 0x6d, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Name)
	if err != nil {
		return
	}
	// write "disabled"
	err = en.Append(0xa8, 0x64, 0x69, 0x73, 0x61, 0x62, 0x6c, 0x65, 0x64)
	if err != nil {
		return err
	}
	err = en.WriteBool(z.Disabled)
	if err != nil {
		return
	}
	// write "api_key"
	err = en.Append(0xa7, 0x61, 0x70, 0x69, 0x5f, 0x6b, 0x65, 0x79)
	if err != nil {
		return err
	}
	err = en.WriteString(z.APIKey)
	if err != nil {
		return
	}
	// write "session_transport"
	err = en.Append(0xb1, 0x73, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x5f, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74)
	if err != nil {
		return err
	}
	err = en.WriteString(string(z.SessionTransport))
	if err != nil {
		return
	}
	// write "access_token_lifetime"
	err = en.Append(0xb5, 0x61, 0x63, 0x63, 0x65, 0x73, 0x73, 0x5f, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x5f, 0x6c, 0x69, 0x66, 0x65, 0x74, 0x69, 0x6d, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteInt(z.AccessTokenLifetime)
	if err != nil {
		return
	}
	// write "session_idle_timeout_enabled"
	err = en.Append(0xbc, 0x73, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x5f, 0x69, 0x64, 0x6c, 0x65, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x5f, 0x65, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64)
	if err != nil {
		return err
	}
	err = en.WriteBool(z.SessionIdleTimeoutEnabled)
	if err != nil {
		return
	}
	// write "session_idle_timeout"
	err = en.Append(0xb4, 0x73, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x5f, 0x69, 0x64, 0x6c, 0x65, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74)
	if err != nil {
		return err
	}
	err = en.WriteInt(z.SessionIdleTimeout)
	if err != nil {
		return
	}
	// write "refresh_token_disabled"
	err = en.Append(0xb6, 0x72, 0x65, 0x66, 0x72, 0x65, 0x73, 0x68, 0x5f, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x5f, 0x64, 0x69, 0x73, 0x61, 0x62, 0x6c, 0x65, 0x64)
	if err != nil {
		return err
	}
	err = en.WriteBool(z.RefreshTokenDisabled)
	if err != nil {
		return
	}
	// write "refresh_token_lifetime"
	err = en.Append(0xb6, 0x72, 0x65, 0x66, 0x72, 0x65, 0x73, 0x68, 0x5f, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x5f, 0x6c, 0x69, 0x66, 0x65, 0x74, 0x69, 0x6d, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteInt(z.RefreshTokenLifetime)
	if err != nil {
		return
	}
	// write "same_site"
	err = en.Append(0xa9, 0x73, 0x61, 0x6d, 0x65, 0x5f, 0x73, 0x69, 0x74, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteString(string(z.SameSite))
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *APIClientConfiguration) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 10
	// string "name"
	o = append(o, 0x8a, 0xa4, 0x6e, 0x61, 0x6d, 0x65)
	o = msgp.AppendString(o, z.Name)
	// string "disabled"
	o = append(o, 0xa8, 0x64, 0x69, 0x73, 0x61, 0x62, 0x6c, 0x65, 0x64)
	o = msgp.AppendBool(o, z.Disabled)
	// string "api_key"
	o = append(o, 0xa7, 0x61, 0x70, 0x69, 0x5f, 0x6b, 0x65, 0x79)
	o = msgp.AppendString(o, z.APIKey)
	// string "session_transport"
	o = append(o, 0xb1, 0x73, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x5f, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74)
	o = msgp.AppendString(o, string(z.SessionTransport))
	// string "access_token_lifetime"
	o = append(o, 0xb5, 0x61, 0x63, 0x63, 0x65, 0x73, 0x73, 0x5f, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x5f, 0x6c, 0x69, 0x66, 0x65, 0x74, 0x69, 0x6d, 0x65)
	o = msgp.AppendInt(o, z.AccessTokenLifetime)
	// string "session_idle_timeout_enabled"
	o = append(o, 0xbc, 0x73, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x5f, 0x69, 0x64, 0x6c, 0x65, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x5f, 0x65, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64)
	o = msgp.AppendBool(o, z.SessionIdleTimeoutEnabled)
	// string "session_idle_timeout"
	o = append(o, 0xb4, 0x73, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x5f, 0x69, 0x64, 0x6c, 0x65, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74)
	o = msgp.AppendInt(o, z.SessionIdleTimeout)
	// string "refresh_token_disabled"
	o = append(o, 0xb6, 0x72, 0x65, 0x66, 0x72, 0x65, 0x73, 0x68, 0x5f, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x5f, 0x64, 0x69, 0x73, 0x61, 0x62, 0x6c, 0x65, 0x64)
	o = msgp.AppendBool(o, z.RefreshTokenDisabled)
	// string "refresh_token_lifetime"
	o = append(o, 0xb6, 0x72, 0x65, 0x66, 0x72, 0x65, 0x73, 0x68, 0x5f, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x5f, 0x6c, 0x69, 0x66, 0x65, 0x74, 0x69, 0x6d, 0x65)
	o = msgp.AppendInt(o, z.RefreshTokenLifetime)
	// string "same_site"
	o = append(o, 0xa9, 0x73, 0x61, 0x6d, 0x65, 0x5f, 0x73, 0x69, 0x74, 0x65)
	o = msgp.AppendString(o, string(z.SameSite))
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *APIClientConfiguration) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "name":
			z.Name, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "disabled":
			z.Disabled, bts, err = msgp.ReadBoolBytes(bts)
			if err != nil {
				return
			}
		case "api_key":
			z.APIKey, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "session_transport":
			{
				var zajw string
				zajw, bts, err = msgp.ReadStringBytes(bts)
				z.SessionTransport = SessionTransportType(zajw)
			}
			if err != nil {
				return
			}
		case "access_token_lifetime":
			z.AccessTokenLifetime, bts, err = msgp.ReadIntBytes(bts)
			if err != nil {
				return
			}
		case "session_idle_timeout_enabled":
			z.SessionIdleTimeoutEnabled, bts, err = msgp.ReadBoolBytes(bts)
			if err != nil {
				return
			}
		case "session_idle_timeout":
			z.SessionIdleTimeout, bts, err = msgp.ReadIntBytes(bts)
			if err != nil {
				return
			}
		case "refresh_token_disabled":
			z.RefreshTokenDisabled, bts, err = msgp.ReadBoolBytes(bts)
			if err != nil {
				return
			}
		case "refresh_token_lifetime":
			z.RefreshTokenLifetime, bts, err = msgp.ReadIntBytes(bts)
			if err != nil {
				return
			}
		case "same_site":
			{
				var zwht string
				zwht, bts, err = msgp.ReadStringBytes(bts)
				z.SameSite = SessionCookieSameSite(zwht)
			}
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
func (z *APIClientConfiguration) Msgsize() (s int) {
	s = 1 + 5 + msgp.StringPrefixSize + len(z.Name) + 9 + msgp.BoolSize + 8 + msgp.StringPrefixSize + len(z.APIKey) + 18 + msgp.StringPrefixSize + len(string(z.SessionTransport)) + 22 + msgp.IntSize + 29 + msgp.BoolSize + 21 + msgp.IntSize + 23 + msgp.BoolSize + 23 + msgp.IntSize + 10 + msgp.StringPrefixSize + len(string(z.SameSite))
	return
}

// DecodeMsg implements msgp.Decodable
func (z *AppConfiguration) DecodeMsg(dc *msgp.Reader) (err error) {
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
		case "database_url":
			z.DatabaseURL, err = dc.ReadString()
			if err != nil {
				return
			}
		case "database_schema":
			z.DatabaseSchema, err = dc.ReadString()
			if err != nil {
				return
			}
		case "smtp":
			err = z.SMTP.DecodeMsg(dc)
			if err != nil {
				return
			}
		case "twilio":
			var zcua uint32
			zcua, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			for zcua > 0 {
				zcua--
				field, err = dc.ReadMapKeyPtr()
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "account_sid":
					z.Twilio.AccountSID, err = dc.ReadString()
					if err != nil {
						return
					}
				case "auth_token":
					z.Twilio.AuthToken, err = dc.ReadString()
					if err != nil {
						return
					}
				case "from":
					z.Twilio.From, err = dc.ReadString()
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
		case "nexmo":
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
				case "api_key":
					z.Nexmo.APIKey, err = dc.ReadString()
					if err != nil {
						return
					}
				case "secret":
					z.Nexmo.APISecret, err = dc.ReadString()
					if err != nil {
						return
					}
				case "from":
					z.Nexmo.From, err = dc.ReadString()
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
		case "hook":
			var zlqf uint32
			zlqf, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			for zlqf > 0 {
				zlqf--
				field, err = dc.ReadMapKeyPtr()
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "sync_hook_timeout_second":
					z.Hook.SyncHookTimeout, err = dc.ReadInt()
					if err != nil {
						return
					}
				case "sync_hook_total_timeout_second":
					z.Hook.SyncHookTotalTimeout, err = dc.ReadInt()
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
func (z *AppConfiguration) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 6
	// write "database_url"
	err = en.Append(0x86, 0xac, 0x64, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x5f, 0x75, 0x72, 0x6c)
	if err != nil {
		return err
	}
	err = en.WriteString(z.DatabaseURL)
	if err != nil {
		return
	}
	// write "database_schema"
	err = en.Append(0xaf, 0x64, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x5f, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61)
	if err != nil {
		return err
	}
	err = en.WriteString(z.DatabaseSchema)
	if err != nil {
		return
	}
	// write "smtp"
	err = en.Append(0xa4, 0x73, 0x6d, 0x74, 0x70)
	if err != nil {
		return err
	}
	err = z.SMTP.EncodeMsg(en)
	if err != nil {
		return
	}
	// write "twilio"
	// map header, size 3
	// write "account_sid"
	err = en.Append(0xa6, 0x74, 0x77, 0x69, 0x6c, 0x69, 0x6f, 0x83, 0xab, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x5f, 0x73, 0x69, 0x64)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Twilio.AccountSID)
	if err != nil {
		return
	}
	// write "auth_token"
	err = en.Append(0xaa, 0x61, 0x75, 0x74, 0x68, 0x5f, 0x74, 0x6f, 0x6b, 0x65, 0x6e)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Twilio.AuthToken)
	if err != nil {
		return
	}
	// write "from"
	err = en.Append(0xa4, 0x66, 0x72, 0x6f, 0x6d)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Twilio.From)
	if err != nil {
		return
	}
	// write "nexmo"
	// map header, size 3
	// write "api_key"
	err = en.Append(0xa5, 0x6e, 0x65, 0x78, 0x6d, 0x6f, 0x83, 0xa7, 0x61, 0x70, 0x69, 0x5f, 0x6b, 0x65, 0x79)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Nexmo.APIKey)
	if err != nil {
		return
	}
	// write "secret"
	err = en.Append(0xa6, 0x73, 0x65, 0x63, 0x72, 0x65, 0x74)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Nexmo.APISecret)
	if err != nil {
		return
	}
	// write "from"
	err = en.Append(0xa4, 0x66, 0x72, 0x6f, 0x6d)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Nexmo.From)
	if err != nil {
		return
	}
	// write "hook"
	// map header, size 2
	// write "sync_hook_timeout_second"
	err = en.Append(0xa4, 0x68, 0x6f, 0x6f, 0x6b, 0x82, 0xb8, 0x73, 0x79, 0x6e, 0x63, 0x5f, 0x68, 0x6f, 0x6f, 0x6b, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x5f, 0x73, 0x65, 0x63, 0x6f, 0x6e, 0x64)
	if err != nil {
		return err
	}
	err = en.WriteInt(z.Hook.SyncHookTimeout)
	if err != nil {
		return
	}
	// write "sync_hook_total_timeout_second"
	err = en.Append(0xbe, 0x73, 0x79, 0x6e, 0x63, 0x5f, 0x68, 0x6f, 0x6f, 0x6b, 0x5f, 0x74, 0x6f, 0x74, 0x61, 0x6c, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x5f, 0x73, 0x65, 0x63, 0x6f, 0x6e, 0x64)
	if err != nil {
		return err
	}
	err = en.WriteInt(z.Hook.SyncHookTotalTimeout)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *AppConfiguration) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 6
	// string "database_url"
	o = append(o, 0x86, 0xac, 0x64, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x5f, 0x75, 0x72, 0x6c)
	o = msgp.AppendString(o, z.DatabaseURL)
	// string "database_schema"
	o = append(o, 0xaf, 0x64, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x5f, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61)
	o = msgp.AppendString(o, z.DatabaseSchema)
	// string "smtp"
	o = append(o, 0xa4, 0x73, 0x6d, 0x74, 0x70)
	o, err = z.SMTP.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "twilio"
	// map header, size 3
	// string "account_sid"
	o = append(o, 0xa6, 0x74, 0x77, 0x69, 0x6c, 0x69, 0x6f, 0x83, 0xab, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x5f, 0x73, 0x69, 0x64)
	o = msgp.AppendString(o, z.Twilio.AccountSID)
	// string "auth_token"
	o = append(o, 0xaa, 0x61, 0x75, 0x74, 0x68, 0x5f, 0x74, 0x6f, 0x6b, 0x65, 0x6e)
	o = msgp.AppendString(o, z.Twilio.AuthToken)
	// string "from"
	o = append(o, 0xa4, 0x66, 0x72, 0x6f, 0x6d)
	o = msgp.AppendString(o, z.Twilio.From)
	// string "nexmo"
	// map header, size 3
	// string "api_key"
	o = append(o, 0xa5, 0x6e, 0x65, 0x78, 0x6d, 0x6f, 0x83, 0xa7, 0x61, 0x70, 0x69, 0x5f, 0x6b, 0x65, 0x79)
	o = msgp.AppendString(o, z.Nexmo.APIKey)
	// string "secret"
	o = append(o, 0xa6, 0x73, 0x65, 0x63, 0x72, 0x65, 0x74)
	o = msgp.AppendString(o, z.Nexmo.APISecret)
	// string "from"
	o = append(o, 0xa4, 0x66, 0x72, 0x6f, 0x6d)
	o = msgp.AppendString(o, z.Nexmo.From)
	// string "hook"
	// map header, size 2
	// string "sync_hook_timeout_second"
	o = append(o, 0xa4, 0x68, 0x6f, 0x6f, 0x6b, 0x82, 0xb8, 0x73, 0x79, 0x6e, 0x63, 0x5f, 0x68, 0x6f, 0x6f, 0x6b, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x5f, 0x73, 0x65, 0x63, 0x6f, 0x6e, 0x64)
	o = msgp.AppendInt(o, z.Hook.SyncHookTimeout)
	// string "sync_hook_total_timeout_second"
	o = append(o, 0xbe, 0x73, 0x79, 0x6e, 0x63, 0x5f, 0x68, 0x6f, 0x6f, 0x6b, 0x5f, 0x74, 0x6f, 0x74, 0x61, 0x6c, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x5f, 0x73, 0x65, 0x63, 0x6f, 0x6e, 0x64)
	o = msgp.AppendInt(o, z.Hook.SyncHookTotalTimeout)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *AppConfiguration) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zdaf uint32
	zdaf, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zdaf > 0 {
		zdaf--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "database_url":
			z.DatabaseURL, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "database_schema":
			z.DatabaseSchema, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "smtp":
			bts, err = z.SMTP.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "twilio":
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
				case "account_sid":
					z.Twilio.AccountSID, bts, err = msgp.ReadStringBytes(bts)
					if err != nil {
						return
					}
				case "auth_token":
					z.Twilio.AuthToken, bts, err = msgp.ReadStringBytes(bts)
					if err != nil {
						return
					}
				case "from":
					z.Twilio.From, bts, err = msgp.ReadStringBytes(bts)
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
		case "nexmo":
			var zjfb uint32
			zjfb, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			for zjfb > 0 {
				zjfb--
				field, bts, err = msgp.ReadMapKeyZC(bts)
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "api_key":
					z.Nexmo.APIKey, bts, err = msgp.ReadStringBytes(bts)
					if err != nil {
						return
					}
				case "secret":
					z.Nexmo.APISecret, bts, err = msgp.ReadStringBytes(bts)
					if err != nil {
						return
					}
				case "from":
					z.Nexmo.From, bts, err = msgp.ReadStringBytes(bts)
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
		case "hook":
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
				case "sync_hook_timeout_second":
					z.Hook.SyncHookTimeout, bts, err = msgp.ReadIntBytes(bts)
					if err != nil {
						return
					}
				case "sync_hook_total_timeout_second":
					z.Hook.SyncHookTotalTimeout, bts, err = msgp.ReadIntBytes(bts)
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
func (z *AppConfiguration) Msgsize() (s int) {
	s = 1 + 13 + msgp.StringPrefixSize + len(z.DatabaseURL) + 16 + msgp.StringPrefixSize + len(z.DatabaseSchema) + 5 + z.SMTP.Msgsize() + 7 + 1 + 12 + msgp.StringPrefixSize + len(z.Twilio.AccountSID) + 11 + msgp.StringPrefixSize + len(z.Twilio.AuthToken) + 5 + msgp.StringPrefixSize + len(z.Twilio.From) + 6 + 1 + 8 + msgp.StringPrefixSize + len(z.Nexmo.APIKey) + 7 + msgp.StringPrefixSize + len(z.Nexmo.APISecret) + 5 + msgp.StringPrefixSize + len(z.Nexmo.From) + 5 + 1 + 25 + msgp.IntSize + 31 + msgp.IntSize
	return
}

// DecodeMsg implements msgp.Decodable
func (z *AuthConfiguration) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zdnj uint32
	zdnj, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zdnj > 0 {
		zdnj--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "authentication_session":
			var zobc uint32
			zobc, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			for zobc > 0 {
				zobc--
				field, err = dc.ReadMapKeyPtr()
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "secret":
					z.AuthenticationSession.Secret, err = dc.ReadString()
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
		case "login_id_keys":
			var zsnv uint32
			zsnv, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			if z.LoginIDKeys == nil && zsnv > 0 {
				z.LoginIDKeys = make(map[string]LoginIDKeyConfiguration, zsnv)
			} else if len(z.LoginIDKeys) > 0 {
				for key, _ := range z.LoginIDKeys {
					delete(z.LoginIDKeys, key)
				}
			}
			for zsnv > 0 {
				zsnv--
				var zeff string
				var zrsw LoginIDKeyConfiguration
				zeff, err = dc.ReadString()
				if err != nil {
					return
				}
				err = zrsw.DecodeMsg(dc)
				if err != nil {
					return
				}
				z.LoginIDKeys[zeff] = zrsw
			}
		case "allowed_realms":
			var zkgt uint32
			zkgt, err = dc.ReadArrayHeader()
			if err != nil {
				return
			}
			if cap(z.AllowedRealms) >= int(zkgt) {
				z.AllowedRealms = (z.AllowedRealms)[:zkgt]
			} else {
				z.AllowedRealms = make([]string, zkgt)
			}
			for zxpk := range z.AllowedRealms {
				z.AllowedRealms[zxpk], err = dc.ReadString()
				if err != nil {
					return
				}
			}
		case "on_user_duplicate_allow_create":
			z.OnUserDuplicateAllowCreate, err = dc.ReadBool()
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
func (z *AuthConfiguration) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 4
	// write "authentication_session"
	// map header, size 1
	// write "secret"
	err = en.Append(0x84, 0xb6, 0x61, 0x75, 0x74, 0x68, 0x65, 0x6e, 0x74, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x73, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x81, 0xa6, 0x73, 0x65, 0x63, 0x72, 0x65, 0x74)
	if err != nil {
		return err
	}
	err = en.WriteString(z.AuthenticationSession.Secret)
	if err != nil {
		return
	}
	// write "login_id_keys"
	err = en.Append(0xad, 0x6c, 0x6f, 0x67, 0x69, 0x6e, 0x5f, 0x69, 0x64, 0x5f, 0x6b, 0x65, 0x79, 0x73)
	if err != nil {
		return err
	}
	err = en.WriteMapHeader(uint32(len(z.LoginIDKeys)))
	if err != nil {
		return
	}
	for zeff, zrsw := range z.LoginIDKeys {
		err = en.WriteString(zeff)
		if err != nil {
			return
		}
		err = zrsw.EncodeMsg(en)
		if err != nil {
			return
		}
	}
	// write "allowed_realms"
	err = en.Append(0xae, 0x61, 0x6c, 0x6c, 0x6f, 0x77, 0x65, 0x64, 0x5f, 0x72, 0x65, 0x61, 0x6c, 0x6d, 0x73)
	if err != nil {
		return err
	}
	err = en.WriteArrayHeader(uint32(len(z.AllowedRealms)))
	if err != nil {
		return
	}
	for zxpk := range z.AllowedRealms {
		err = en.WriteString(z.AllowedRealms[zxpk])
		if err != nil {
			return
		}
	}
	// write "on_user_duplicate_allow_create"
	err = en.Append(0xbe, 0x6f, 0x6e, 0x5f, 0x75, 0x73, 0x65, 0x72, 0x5f, 0x64, 0x75, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x65, 0x5f, 0x61, 0x6c, 0x6c, 0x6f, 0x77, 0x5f, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteBool(z.OnUserDuplicateAllowCreate)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *AuthConfiguration) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 4
	// string "authentication_session"
	// map header, size 1
	// string "secret"
	o = append(o, 0x84, 0xb6, 0x61, 0x75, 0x74, 0x68, 0x65, 0x6e, 0x74, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x73, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x81, 0xa6, 0x73, 0x65, 0x63, 0x72, 0x65, 0x74)
	o = msgp.AppendString(o, z.AuthenticationSession.Secret)
	// string "login_id_keys"
	o = append(o, 0xad, 0x6c, 0x6f, 0x67, 0x69, 0x6e, 0x5f, 0x69, 0x64, 0x5f, 0x6b, 0x65, 0x79, 0x73)
	o = msgp.AppendMapHeader(o, uint32(len(z.LoginIDKeys)))
	for zeff, zrsw := range z.LoginIDKeys {
		o = msgp.AppendString(o, zeff)
		o, err = zrsw.MarshalMsg(o)
		if err != nil {
			return
		}
	}
	// string "allowed_realms"
	o = append(o, 0xae, 0x61, 0x6c, 0x6c, 0x6f, 0x77, 0x65, 0x64, 0x5f, 0x72, 0x65, 0x61, 0x6c, 0x6d, 0x73)
	o = msgp.AppendArrayHeader(o, uint32(len(z.AllowedRealms)))
	for zxpk := range z.AllowedRealms {
		o = msgp.AppendString(o, z.AllowedRealms[zxpk])
	}
	// string "on_user_duplicate_allow_create"
	o = append(o, 0xbe, 0x6f, 0x6e, 0x5f, 0x75, 0x73, 0x65, 0x72, 0x5f, 0x64, 0x75, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x65, 0x5f, 0x61, 0x6c, 0x6c, 0x6f, 0x77, 0x5f, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65)
	o = msgp.AppendBool(o, z.OnUserDuplicateAllowCreate)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *AuthConfiguration) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
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
		case "authentication_session":
			var zpez uint32
			zpez, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			for zpez > 0 {
				zpez--
				field, bts, err = msgp.ReadMapKeyZC(bts)
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "secret":
					z.AuthenticationSession.Secret, bts, err = msgp.ReadStringBytes(bts)
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
		case "login_id_keys":
			var zqke uint32
			zqke, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			if z.LoginIDKeys == nil && zqke > 0 {
				z.LoginIDKeys = make(map[string]LoginIDKeyConfiguration, zqke)
			} else if len(z.LoginIDKeys) > 0 {
				for key, _ := range z.LoginIDKeys {
					delete(z.LoginIDKeys, key)
				}
			}
			for zqke > 0 {
				var zeff string
				var zrsw LoginIDKeyConfiguration
				zqke--
				zeff, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
				bts, err = zrsw.UnmarshalMsg(bts)
				if err != nil {
					return
				}
				z.LoginIDKeys[zeff] = zrsw
			}
		case "allowed_realms":
			var zqyh uint32
			zqyh, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.AllowedRealms) >= int(zqyh) {
				z.AllowedRealms = (z.AllowedRealms)[:zqyh]
			} else {
				z.AllowedRealms = make([]string, zqyh)
			}
			for zxpk := range z.AllowedRealms {
				z.AllowedRealms[zxpk], bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
			}
		case "on_user_duplicate_allow_create":
			z.OnUserDuplicateAllowCreate, bts, err = msgp.ReadBoolBytes(bts)
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
func (z *AuthConfiguration) Msgsize() (s int) {
	s = 1 + 23 + 1 + 7 + msgp.StringPrefixSize + len(z.AuthenticationSession.Secret) + 14 + msgp.MapHeaderSize
	if z.LoginIDKeys != nil {
		for zeff, zrsw := range z.LoginIDKeys {
			_ = zrsw
			s += msgp.StringPrefixSize + len(zeff) + zrsw.Msgsize()
		}
	}
	s += 15 + msgp.ArrayHeaderSize
	for zxpk := range z.AllowedRealms {
		s += msgp.StringPrefixSize + len(z.AllowedRealms[zxpk])
	}
	s += 31 + msgp.BoolSize
	return
}

// DecodeMsg implements msgp.Decodable
func (z *AuthenticationSessionConfiguration) DecodeMsg(dc *msgp.Reader) (err error) {
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
		case "secret":
			z.Secret, err = dc.ReadString()
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
func (z AuthenticationSessionConfiguration) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 1
	// write "secret"
	err = en.Append(0x81, 0xa6, 0x73, 0x65, 0x63, 0x72, 0x65, 0x74)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Secret)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z AuthenticationSessionConfiguration) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 1
	// string "secret"
	o = append(o, 0x81, 0xa6, 0x73, 0x65, 0x63, 0x72, 0x65, 0x74)
	o = msgp.AppendString(o, z.Secret)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *AuthenticationSessionConfiguration) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "secret":
			z.Secret, bts, err = msgp.ReadStringBytes(bts)
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
func (z AuthenticationSessionConfiguration) Msgsize() (s int) {
	s = 1 + 7 + msgp.StringPrefixSize + len(z.Secret)
	return
}

// DecodeMsg implements msgp.Decodable
func (z *CORSConfiguration) DecodeMsg(dc *msgp.Reader) (err error) {
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
		case "origin":
			z.Origin, err = dc.ReadString()
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
func (z CORSConfiguration) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 1
	// write "origin"
	err = en.Append(0x81, 0xa6, 0x6f, 0x72, 0x69, 0x67, 0x69, 0x6e)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Origin)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z CORSConfiguration) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 1
	// string "origin"
	o = append(o, 0x81, 0xa6, 0x6f, 0x72, 0x69, 0x67, 0x69, 0x6e)
	o = msgp.AppendString(o, z.Origin)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *CORSConfiguration) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "origin":
			z.Origin, bts, err = msgp.ReadStringBytes(bts)
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
func (z CORSConfiguration) Msgsize() (s int) {
	s = 1 + 7 + msgp.StringPrefixSize + len(z.Origin)
	return
}

// DecodeMsg implements msgp.Decodable
func (z *CustomTokenConfiguration) DecodeMsg(dc *msgp.Reader) (err error) {
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
		case "enabled":
			z.Enabled, err = dc.ReadBool()
			if err != nil {
				return
			}
		case "issuer":
			z.Issuer, err = dc.ReadString()
			if err != nil {
				return
			}
		case "audience":
			z.Audience, err = dc.ReadString()
			if err != nil {
				return
			}
		case "secret":
			z.Secret, err = dc.ReadString()
			if err != nil {
				return
			}
		case "on_user_duplicate_allow_merge":
			z.OnUserDuplicateAllowMerge, err = dc.ReadBool()
			if err != nil {
				return
			}
		case "on_user_duplicate_allow_create":
			z.OnUserDuplicateAllowCreate, err = dc.ReadBool()
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
func (z *CustomTokenConfiguration) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 6
	// write "enabled"
	err = en.Append(0x86, 0xa7, 0x65, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64)
	if err != nil {
		return err
	}
	err = en.WriteBool(z.Enabled)
	if err != nil {
		return
	}
	// write "issuer"
	err = en.Append(0xa6, 0x69, 0x73, 0x73, 0x75, 0x65, 0x72)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Issuer)
	if err != nil {
		return
	}
	// write "audience"
	err = en.Append(0xa8, 0x61, 0x75, 0x64, 0x69, 0x65, 0x6e, 0x63, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Audience)
	if err != nil {
		return
	}
	// write "secret"
	err = en.Append(0xa6, 0x73, 0x65, 0x63, 0x72, 0x65, 0x74)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Secret)
	if err != nil {
		return
	}
	// write "on_user_duplicate_allow_merge"
	err = en.Append(0xbd, 0x6f, 0x6e, 0x5f, 0x75, 0x73, 0x65, 0x72, 0x5f, 0x64, 0x75, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x65, 0x5f, 0x61, 0x6c, 0x6c, 0x6f, 0x77, 0x5f, 0x6d, 0x65, 0x72, 0x67, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteBool(z.OnUserDuplicateAllowMerge)
	if err != nil {
		return
	}
	// write "on_user_duplicate_allow_create"
	err = en.Append(0xbe, 0x6f, 0x6e, 0x5f, 0x75, 0x73, 0x65, 0x72, 0x5f, 0x64, 0x75, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x65, 0x5f, 0x61, 0x6c, 0x6c, 0x6f, 0x77, 0x5f, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteBool(z.OnUserDuplicateAllowCreate)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *CustomTokenConfiguration) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 6
	// string "enabled"
	o = append(o, 0x86, 0xa7, 0x65, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64)
	o = msgp.AppendBool(o, z.Enabled)
	// string "issuer"
	o = append(o, 0xa6, 0x69, 0x73, 0x73, 0x75, 0x65, 0x72)
	o = msgp.AppendString(o, z.Issuer)
	// string "audience"
	o = append(o, 0xa8, 0x61, 0x75, 0x64, 0x69, 0x65, 0x6e, 0x63, 0x65)
	o = msgp.AppendString(o, z.Audience)
	// string "secret"
	o = append(o, 0xa6, 0x73, 0x65, 0x63, 0x72, 0x65, 0x74)
	o = msgp.AppendString(o, z.Secret)
	// string "on_user_duplicate_allow_merge"
	o = append(o, 0xbd, 0x6f, 0x6e, 0x5f, 0x75, 0x73, 0x65, 0x72, 0x5f, 0x64, 0x75, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x65, 0x5f, 0x61, 0x6c, 0x6c, 0x6f, 0x77, 0x5f, 0x6d, 0x65, 0x72, 0x67, 0x65)
	o = msgp.AppendBool(o, z.OnUserDuplicateAllowMerge)
	// string "on_user_duplicate_allow_create"
	o = append(o, 0xbe, 0x6f, 0x6e, 0x5f, 0x75, 0x73, 0x65, 0x72, 0x5f, 0x64, 0x75, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x65, 0x5f, 0x61, 0x6c, 0x6c, 0x6f, 0x77, 0x5f, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65)
	o = msgp.AppendBool(o, z.OnUserDuplicateAllowCreate)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *CustomTokenConfiguration) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "enabled":
			z.Enabled, bts, err = msgp.ReadBoolBytes(bts)
			if err != nil {
				return
			}
		case "issuer":
			z.Issuer, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "audience":
			z.Audience, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "secret":
			z.Secret, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "on_user_duplicate_allow_merge":
			z.OnUserDuplicateAllowMerge, bts, err = msgp.ReadBoolBytes(bts)
			if err != nil {
				return
			}
		case "on_user_duplicate_allow_create":
			z.OnUserDuplicateAllowCreate, bts, err = msgp.ReadBoolBytes(bts)
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
func (z *CustomTokenConfiguration) Msgsize() (s int) {
	s = 1 + 8 + msgp.BoolSize + 7 + msgp.StringPrefixSize + len(z.Issuer) + 9 + msgp.StringPrefixSize + len(z.Audience) + 7 + msgp.StringPrefixSize + len(z.Secret) + 30 + msgp.BoolSize + 31 + msgp.BoolSize
	return
}

// DecodeMsg implements msgp.Decodable
func (z *DeploymentRoute) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zsbz uint32
	zsbz, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zsbz > 0 {
		zsbz--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "version":
			z.Version, err = dc.ReadString()
			if err != nil {
				return
			}
		case "path":
			z.Path, err = dc.ReadString()
			if err != nil {
				return
			}
		case "type":
			z.Type, err = dc.ReadString()
			if err != nil {
				return
			}
		case "type_config":
			var zrjx uint32
			zrjx, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			if z.TypeConfig == nil && zrjx > 0 {
				z.TypeConfig = make(map[string]interface{}, zrjx)
			} else if len(z.TypeConfig) > 0 {
				for key, _ := range z.TypeConfig {
					delete(z.TypeConfig, key)
				}
			}
			for zrjx > 0 {
				zrjx--
				var ztaf string
				var zeth interface{}
				ztaf, err = dc.ReadString()
				if err != nil {
					return
				}
				zeth, err = dc.ReadIntf()
				if err != nil {
					return
				}
				z.TypeConfig[ztaf] = zeth
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
func (z *DeploymentRoute) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 4
	// write "version"
	err = en.Append(0x84, 0xa7, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Version)
	if err != nil {
		return
	}
	// write "path"
	err = en.Append(0xa4, 0x70, 0x61, 0x74, 0x68)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Path)
	if err != nil {
		return
	}
	// write "type"
	err = en.Append(0xa4, 0x74, 0x79, 0x70, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Type)
	if err != nil {
		return
	}
	// write "type_config"
	err = en.Append(0xab, 0x74, 0x79, 0x70, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67)
	if err != nil {
		return err
	}
	err = en.WriteMapHeader(uint32(len(z.TypeConfig)))
	if err != nil {
		return
	}
	for ztaf, zeth := range z.TypeConfig {
		err = en.WriteString(ztaf)
		if err != nil {
			return
		}
		err = en.WriteIntf(zeth)
		if err != nil {
			return
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *DeploymentRoute) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 4
	// string "version"
	o = append(o, 0x84, 0xa7, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e)
	o = msgp.AppendString(o, z.Version)
	// string "path"
	o = append(o, 0xa4, 0x70, 0x61, 0x74, 0x68)
	o = msgp.AppendString(o, z.Path)
	// string "type"
	o = append(o, 0xa4, 0x74, 0x79, 0x70, 0x65)
	o = msgp.AppendString(o, z.Type)
	// string "type_config"
	o = append(o, 0xab, 0x74, 0x79, 0x70, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67)
	o = msgp.AppendMapHeader(o, uint32(len(z.TypeConfig)))
	for ztaf, zeth := range z.TypeConfig {
		o = msgp.AppendString(o, ztaf)
		o, err = msgp.AppendIntf(o, zeth)
		if err != nil {
			return
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *DeploymentRoute) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zawn uint32
	zawn, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zawn > 0 {
		zawn--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "version":
			z.Version, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "path":
			z.Path, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "type":
			z.Type, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "type_config":
			var zwel uint32
			zwel, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			if z.TypeConfig == nil && zwel > 0 {
				z.TypeConfig = make(map[string]interface{}, zwel)
			} else if len(z.TypeConfig) > 0 {
				for key, _ := range z.TypeConfig {
					delete(z.TypeConfig, key)
				}
			}
			for zwel > 0 {
				var ztaf string
				var zeth interface{}
				zwel--
				ztaf, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
				zeth, bts, err = msgp.ReadIntfBytes(bts)
				if err != nil {
					return
				}
				z.TypeConfig[ztaf] = zeth
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
func (z *DeploymentRoute) Msgsize() (s int) {
	s = 1 + 8 + msgp.StringPrefixSize + len(z.Version) + 5 + msgp.StringPrefixSize + len(z.Path) + 5 + msgp.StringPrefixSize + len(z.Type) + 12 + msgp.MapHeaderSize
	if z.TypeConfig != nil {
		for ztaf, zeth := range z.TypeConfig {
			_ = zeth
			s += msgp.StringPrefixSize + len(ztaf) + msgp.GuessSize(zeth)
		}
	}
	return
}

// DecodeMsg implements msgp.Decodable
func (z *ForgotPasswordConfiguration) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zrbe uint32
	zrbe, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zrbe > 0 {
		zrbe--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "app_name":
			z.AppName, err = dc.ReadString()
			if err != nil {
				return
			}
		case "url_prefix":
			z.URLPrefix, err = dc.ReadString()
			if err != nil {
				return
			}
		case "secure_match":
			z.SecureMatch, err = dc.ReadBool()
			if err != nil {
				return
			}
		case "sender":
			z.Sender, err = dc.ReadString()
			if err != nil {
				return
			}
		case "subject":
			z.Subject, err = dc.ReadString()
			if err != nil {
				return
			}
		case "reply_to":
			z.ReplyTo, err = dc.ReadString()
			if err != nil {
				return
			}
		case "reset_url_lifetime":
			z.ResetURLLifetime, err = dc.ReadInt()
			if err != nil {
				return
			}
		case "success_redirect":
			z.SuccessRedirect, err = dc.ReadString()
			if err != nil {
				return
			}
		case "error_redirect":
			z.ErrorRedirect, err = dc.ReadString()
			if err != nil {
				return
			}
		case "email_text_url":
			z.EmailTextURL, err = dc.ReadString()
			if err != nil {
				return
			}
		case "email_html_url":
			z.EmailHTMLURL, err = dc.ReadString()
			if err != nil {
				return
			}
		case "reset_html_url":
			z.ResetHTMLURL, err = dc.ReadString()
			if err != nil {
				return
			}
		case "reset_success_html_url":
			z.ResetSuccessHTMLURL, err = dc.ReadString()
			if err != nil {
				return
			}
		case "reset_error_html_url":
			z.ResetErrorHTMLURL, err = dc.ReadString()
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
func (z *ForgotPasswordConfiguration) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 14
	// write "app_name"
	err = en.Append(0x8e, 0xa8, 0x61, 0x70, 0x70, 0x5f, 0x6e, 0x61, 0x6d, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteString(z.AppName)
	if err != nil {
		return
	}
	// write "url_prefix"
	err = en.Append(0xaa, 0x75, 0x72, 0x6c, 0x5f, 0x70, 0x72, 0x65, 0x66, 0x69, 0x78)
	if err != nil {
		return err
	}
	err = en.WriteString(z.URLPrefix)
	if err != nil {
		return
	}
	// write "secure_match"
	err = en.Append(0xac, 0x73, 0x65, 0x63, 0x75, 0x72, 0x65, 0x5f, 0x6d, 0x61, 0x74, 0x63, 0x68)
	if err != nil {
		return err
	}
	err = en.WriteBool(z.SecureMatch)
	if err != nil {
		return
	}
	// write "sender"
	err = en.Append(0xa6, 0x73, 0x65, 0x6e, 0x64, 0x65, 0x72)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Sender)
	if err != nil {
		return
	}
	// write "subject"
	err = en.Append(0xa7, 0x73, 0x75, 0x62, 0x6a, 0x65, 0x63, 0x74)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Subject)
	if err != nil {
		return
	}
	// write "reply_to"
	err = en.Append(0xa8, 0x72, 0x65, 0x70, 0x6c, 0x79, 0x5f, 0x74, 0x6f)
	if err != nil {
		return err
	}
	err = en.WriteString(z.ReplyTo)
	if err != nil {
		return
	}
	// write "reset_url_lifetime"
	err = en.Append(0xb2, 0x72, 0x65, 0x73, 0x65, 0x74, 0x5f, 0x75, 0x72, 0x6c, 0x5f, 0x6c, 0x69, 0x66, 0x65, 0x74, 0x69, 0x6d, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteInt(z.ResetURLLifetime)
	if err != nil {
		return
	}
	// write "success_redirect"
	err = en.Append(0xb0, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x5f, 0x72, 0x65, 0x64, 0x69, 0x72, 0x65, 0x63, 0x74)
	if err != nil {
		return err
	}
	err = en.WriteString(z.SuccessRedirect)
	if err != nil {
		return
	}
	// write "error_redirect"
	err = en.Append(0xae, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x5f, 0x72, 0x65, 0x64, 0x69, 0x72, 0x65, 0x63, 0x74)
	if err != nil {
		return err
	}
	err = en.WriteString(z.ErrorRedirect)
	if err != nil {
		return
	}
	// write "email_text_url"
	err = en.Append(0xae, 0x65, 0x6d, 0x61, 0x69, 0x6c, 0x5f, 0x74, 0x65, 0x78, 0x74, 0x5f, 0x75, 0x72, 0x6c)
	if err != nil {
		return err
	}
	err = en.WriteString(z.EmailTextURL)
	if err != nil {
		return
	}
	// write "email_html_url"
	err = en.Append(0xae, 0x65, 0x6d, 0x61, 0x69, 0x6c, 0x5f, 0x68, 0x74, 0x6d, 0x6c, 0x5f, 0x75, 0x72, 0x6c)
	if err != nil {
		return err
	}
	err = en.WriteString(z.EmailHTMLURL)
	if err != nil {
		return
	}
	// write "reset_html_url"
	err = en.Append(0xae, 0x72, 0x65, 0x73, 0x65, 0x74, 0x5f, 0x68, 0x74, 0x6d, 0x6c, 0x5f, 0x75, 0x72, 0x6c)
	if err != nil {
		return err
	}
	err = en.WriteString(z.ResetHTMLURL)
	if err != nil {
		return
	}
	// write "reset_success_html_url"
	err = en.Append(0xb6, 0x72, 0x65, 0x73, 0x65, 0x74, 0x5f, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x5f, 0x68, 0x74, 0x6d, 0x6c, 0x5f, 0x75, 0x72, 0x6c)
	if err != nil {
		return err
	}
	err = en.WriteString(z.ResetSuccessHTMLURL)
	if err != nil {
		return
	}
	// write "reset_error_html_url"
	err = en.Append(0xb4, 0x72, 0x65, 0x73, 0x65, 0x74, 0x5f, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x5f, 0x68, 0x74, 0x6d, 0x6c, 0x5f, 0x75, 0x72, 0x6c)
	if err != nil {
		return err
	}
	err = en.WriteString(z.ResetErrorHTMLURL)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *ForgotPasswordConfiguration) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 14
	// string "app_name"
	o = append(o, 0x8e, 0xa8, 0x61, 0x70, 0x70, 0x5f, 0x6e, 0x61, 0x6d, 0x65)
	o = msgp.AppendString(o, z.AppName)
	// string "url_prefix"
	o = append(o, 0xaa, 0x75, 0x72, 0x6c, 0x5f, 0x70, 0x72, 0x65, 0x66, 0x69, 0x78)
	o = msgp.AppendString(o, z.URLPrefix)
	// string "secure_match"
	o = append(o, 0xac, 0x73, 0x65, 0x63, 0x75, 0x72, 0x65, 0x5f, 0x6d, 0x61, 0x74, 0x63, 0x68)
	o = msgp.AppendBool(o, z.SecureMatch)
	// string "sender"
	o = append(o, 0xa6, 0x73, 0x65, 0x6e, 0x64, 0x65, 0x72)
	o = msgp.AppendString(o, z.Sender)
	// string "subject"
	o = append(o, 0xa7, 0x73, 0x75, 0x62, 0x6a, 0x65, 0x63, 0x74)
	o = msgp.AppendString(o, z.Subject)
	// string "reply_to"
	o = append(o, 0xa8, 0x72, 0x65, 0x70, 0x6c, 0x79, 0x5f, 0x74, 0x6f)
	o = msgp.AppendString(o, z.ReplyTo)
	// string "reset_url_lifetime"
	o = append(o, 0xb2, 0x72, 0x65, 0x73, 0x65, 0x74, 0x5f, 0x75, 0x72, 0x6c, 0x5f, 0x6c, 0x69, 0x66, 0x65, 0x74, 0x69, 0x6d, 0x65)
	o = msgp.AppendInt(o, z.ResetURLLifetime)
	// string "success_redirect"
	o = append(o, 0xb0, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x5f, 0x72, 0x65, 0x64, 0x69, 0x72, 0x65, 0x63, 0x74)
	o = msgp.AppendString(o, z.SuccessRedirect)
	// string "error_redirect"
	o = append(o, 0xae, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x5f, 0x72, 0x65, 0x64, 0x69, 0x72, 0x65, 0x63, 0x74)
	o = msgp.AppendString(o, z.ErrorRedirect)
	// string "email_text_url"
	o = append(o, 0xae, 0x65, 0x6d, 0x61, 0x69, 0x6c, 0x5f, 0x74, 0x65, 0x78, 0x74, 0x5f, 0x75, 0x72, 0x6c)
	o = msgp.AppendString(o, z.EmailTextURL)
	// string "email_html_url"
	o = append(o, 0xae, 0x65, 0x6d, 0x61, 0x69, 0x6c, 0x5f, 0x68, 0x74, 0x6d, 0x6c, 0x5f, 0x75, 0x72, 0x6c)
	o = msgp.AppendString(o, z.EmailHTMLURL)
	// string "reset_html_url"
	o = append(o, 0xae, 0x72, 0x65, 0x73, 0x65, 0x74, 0x5f, 0x68, 0x74, 0x6d, 0x6c, 0x5f, 0x75, 0x72, 0x6c)
	o = msgp.AppendString(o, z.ResetHTMLURL)
	// string "reset_success_html_url"
	o = append(o, 0xb6, 0x72, 0x65, 0x73, 0x65, 0x74, 0x5f, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x5f, 0x68, 0x74, 0x6d, 0x6c, 0x5f, 0x75, 0x72, 0x6c)
	o = msgp.AppendString(o, z.ResetSuccessHTMLURL)
	// string "reset_error_html_url"
	o = append(o, 0xb4, 0x72, 0x65, 0x73, 0x65, 0x74, 0x5f, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x5f, 0x68, 0x74, 0x6d, 0x6c, 0x5f, 0x75, 0x72, 0x6c)
	o = msgp.AppendString(o, z.ResetErrorHTMLURL)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *ForgotPasswordConfiguration) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zmfd uint32
	zmfd, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zmfd > 0 {
		zmfd--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "app_name":
			z.AppName, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "url_prefix":
			z.URLPrefix, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "secure_match":
			z.SecureMatch, bts, err = msgp.ReadBoolBytes(bts)
			if err != nil {
				return
			}
		case "sender":
			z.Sender, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "subject":
			z.Subject, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "reply_to":
			z.ReplyTo, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "reset_url_lifetime":
			z.ResetURLLifetime, bts, err = msgp.ReadIntBytes(bts)
			if err != nil {
				return
			}
		case "success_redirect":
			z.SuccessRedirect, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "error_redirect":
			z.ErrorRedirect, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "email_text_url":
			z.EmailTextURL, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "email_html_url":
			z.EmailHTMLURL, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "reset_html_url":
			z.ResetHTMLURL, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "reset_success_html_url":
			z.ResetSuccessHTMLURL, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "reset_error_html_url":
			z.ResetErrorHTMLURL, bts, err = msgp.ReadStringBytes(bts)
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
func (z *ForgotPasswordConfiguration) Msgsize() (s int) {
	s = 1 + 9 + msgp.StringPrefixSize + len(z.AppName) + 11 + msgp.StringPrefixSize + len(z.URLPrefix) + 13 + msgp.BoolSize + 7 + msgp.StringPrefixSize + len(z.Sender) + 8 + msgp.StringPrefixSize + len(z.Subject) + 9 + msgp.StringPrefixSize + len(z.ReplyTo) + 19 + msgp.IntSize + 17 + msgp.StringPrefixSize + len(z.SuccessRedirect) + 15 + msgp.StringPrefixSize + len(z.ErrorRedirect) + 15 + msgp.StringPrefixSize + len(z.EmailTextURL) + 15 + msgp.StringPrefixSize + len(z.EmailHTMLURL) + 15 + msgp.StringPrefixSize + len(z.ResetHTMLURL) + 23 + msgp.StringPrefixSize + len(z.ResetSuccessHTMLURL) + 21 + msgp.StringPrefixSize + len(z.ResetErrorHTMLURL)
	return
}

// DecodeMsg implements msgp.Decodable
func (z *Hook) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zzdc uint32
	zzdc, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zzdc > 0 {
		zzdc--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "event":
			z.Event, err = dc.ReadString()
			if err != nil {
				return
			}
		case "url":
			z.URL, err = dc.ReadString()
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
func (z Hook) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 2
	// write "event"
	err = en.Append(0x82, 0xa5, 0x65, 0x76, 0x65, 0x6e, 0x74)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Event)
	if err != nil {
		return
	}
	// write "url"
	err = en.Append(0xa3, 0x75, 0x72, 0x6c)
	if err != nil {
		return err
	}
	err = en.WriteString(z.URL)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z Hook) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "event"
	o = append(o, 0x82, 0xa5, 0x65, 0x76, 0x65, 0x6e, 0x74)
	o = msgp.AppendString(o, z.Event)
	// string "url"
	o = append(o, 0xa3, 0x75, 0x72, 0x6c)
	o = msgp.AppendString(o, z.URL)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Hook) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zelx uint32
	zelx, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zelx > 0 {
		zelx--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "event":
			z.Event, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "url":
			z.URL, bts, err = msgp.ReadStringBytes(bts)
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
func (z Hook) Msgsize() (s int) {
	s = 1 + 6 + msgp.StringPrefixSize + len(z.Event) + 4 + msgp.StringPrefixSize + len(z.URL)
	return
}

// DecodeMsg implements msgp.Decodable
func (z *HookAppConfiguration) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zbal uint32
	zbal, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zbal > 0 {
		zbal--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "sync_hook_timeout_second":
			z.SyncHookTimeout, err = dc.ReadInt()
			if err != nil {
				return
			}
		case "sync_hook_total_timeout_second":
			z.SyncHookTotalTimeout, err = dc.ReadInt()
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
func (z HookAppConfiguration) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 2
	// write "sync_hook_timeout_second"
	err = en.Append(0x82, 0xb8, 0x73, 0x79, 0x6e, 0x63, 0x5f, 0x68, 0x6f, 0x6f, 0x6b, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x5f, 0x73, 0x65, 0x63, 0x6f, 0x6e, 0x64)
	if err != nil {
		return err
	}
	err = en.WriteInt(z.SyncHookTimeout)
	if err != nil {
		return
	}
	// write "sync_hook_total_timeout_second"
	err = en.Append(0xbe, 0x73, 0x79, 0x6e, 0x63, 0x5f, 0x68, 0x6f, 0x6f, 0x6b, 0x5f, 0x74, 0x6f, 0x74, 0x61, 0x6c, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x5f, 0x73, 0x65, 0x63, 0x6f, 0x6e, 0x64)
	if err != nil {
		return err
	}
	err = en.WriteInt(z.SyncHookTotalTimeout)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z HookAppConfiguration) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "sync_hook_timeout_second"
	o = append(o, 0x82, 0xb8, 0x73, 0x79, 0x6e, 0x63, 0x5f, 0x68, 0x6f, 0x6f, 0x6b, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x5f, 0x73, 0x65, 0x63, 0x6f, 0x6e, 0x64)
	o = msgp.AppendInt(o, z.SyncHookTimeout)
	// string "sync_hook_total_timeout_second"
	o = append(o, 0xbe, 0x73, 0x79, 0x6e, 0x63, 0x5f, 0x68, 0x6f, 0x6f, 0x6b, 0x5f, 0x74, 0x6f, 0x74, 0x61, 0x6c, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x5f, 0x73, 0x65, 0x63, 0x6f, 0x6e, 0x64)
	o = msgp.AppendInt(o, z.SyncHookTotalTimeout)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *HookAppConfiguration) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zjqz uint32
	zjqz, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zjqz > 0 {
		zjqz--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "sync_hook_timeout_second":
			z.SyncHookTimeout, bts, err = msgp.ReadIntBytes(bts)
			if err != nil {
				return
			}
		case "sync_hook_total_timeout_second":
			z.SyncHookTotalTimeout, bts, err = msgp.ReadIntBytes(bts)
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
func (z HookAppConfiguration) Msgsize() (s int) {
	s = 1 + 25 + msgp.IntSize + 31 + msgp.IntSize
	return
}

// DecodeMsg implements msgp.Decodable
func (z *HookUserConfiguration) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zkct uint32
	zkct, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zkct > 0 {
		zkct--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "secret":
			z.Secret, err = dc.ReadString()
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
func (z HookUserConfiguration) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 1
	// write "secret"
	err = en.Append(0x81, 0xa6, 0x73, 0x65, 0x63, 0x72, 0x65, 0x74)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Secret)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z HookUserConfiguration) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 1
	// string "secret"
	o = append(o, 0x81, 0xa6, 0x73, 0x65, 0x63, 0x72, 0x65, 0x74)
	o = msgp.AppendString(o, z.Secret)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *HookUserConfiguration) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var ztmt uint32
	ztmt, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for ztmt > 0 {
		ztmt--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "secret":
			z.Secret, bts, err = msgp.ReadStringBytes(bts)
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
func (z HookUserConfiguration) Msgsize() (s int) {
	s = 1 + 7 + msgp.StringPrefixSize + len(z.Secret)
	return
}

// DecodeMsg implements msgp.Decodable
func (z *LoginIDKeyConfiguration) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var ztco uint32
	ztco, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for ztco > 0 {
		ztco--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "type":
			{
				var zana string
				zana, err = dc.ReadString()
				z.Type = LoginIDKeyType(zana)
			}
			if err != nil {
				return
			}
		case "minimum":
			if dc.IsNil() {
				err = dc.ReadNil()
				if err != nil {
					return
				}
				z.Minimum = nil
			} else {
				if z.Minimum == nil {
					z.Minimum = new(int)
				}
				*z.Minimum, err = dc.ReadInt()
				if err != nil {
					return
				}
			}
		case "maximum":
			if dc.IsNil() {
				err = dc.ReadNil()
				if err != nil {
					return
				}
				z.Maximum = nil
			} else {
				if z.Maximum == nil {
					z.Maximum = new(int)
				}
				*z.Maximum, err = dc.ReadInt()
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
func (z *LoginIDKeyConfiguration) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 3
	// write "type"
	err = en.Append(0x83, 0xa4, 0x74, 0x79, 0x70, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteString(string(z.Type))
	if err != nil {
		return
	}
	// write "minimum"
	err = en.Append(0xa7, 0x6d, 0x69, 0x6e, 0x69, 0x6d, 0x75, 0x6d)
	if err != nil {
		return err
	}
	if z.Minimum == nil {
		err = en.WriteNil()
		if err != nil {
			return
		}
	} else {
		err = en.WriteInt(*z.Minimum)
		if err != nil {
			return
		}
	}
	// write "maximum"
	err = en.Append(0xa7, 0x6d, 0x61, 0x78, 0x69, 0x6d, 0x75, 0x6d)
	if err != nil {
		return err
	}
	if z.Maximum == nil {
		err = en.WriteNil()
		if err != nil {
			return
		}
	} else {
		err = en.WriteInt(*z.Maximum)
		if err != nil {
			return
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *LoginIDKeyConfiguration) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 3
	// string "type"
	o = append(o, 0x83, 0xa4, 0x74, 0x79, 0x70, 0x65)
	o = msgp.AppendString(o, string(z.Type))
	// string "minimum"
	o = append(o, 0xa7, 0x6d, 0x69, 0x6e, 0x69, 0x6d, 0x75, 0x6d)
	if z.Minimum == nil {
		o = msgp.AppendNil(o)
	} else {
		o = msgp.AppendInt(o, *z.Minimum)
	}
	// string "maximum"
	o = append(o, 0xa7, 0x6d, 0x61, 0x78, 0x69, 0x6d, 0x75, 0x6d)
	if z.Maximum == nil {
		o = msgp.AppendNil(o)
	} else {
		o = msgp.AppendInt(o, *z.Maximum)
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *LoginIDKeyConfiguration) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var ztyy uint32
	ztyy, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for ztyy > 0 {
		ztyy--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "type":
			{
				var zinl string
				zinl, bts, err = msgp.ReadStringBytes(bts)
				z.Type = LoginIDKeyType(zinl)
			}
			if err != nil {
				return
			}
		case "minimum":
			if msgp.IsNil(bts) {
				bts, err = msgp.ReadNilBytes(bts)
				if err != nil {
					return
				}
				z.Minimum = nil
			} else {
				if z.Minimum == nil {
					z.Minimum = new(int)
				}
				*z.Minimum, bts, err = msgp.ReadIntBytes(bts)
				if err != nil {
					return
				}
			}
		case "maximum":
			if msgp.IsNil(bts) {
				bts, err = msgp.ReadNilBytes(bts)
				if err != nil {
					return
				}
				z.Maximum = nil
			} else {
				if z.Maximum == nil {
					z.Maximum = new(int)
				}
				*z.Maximum, bts, err = msgp.ReadIntBytes(bts)
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
func (z *LoginIDKeyConfiguration) Msgsize() (s int) {
	s = 1 + 5 + msgp.StringPrefixSize + len(string(z.Type)) + 8
	if z.Minimum == nil {
		s += msgp.NilSize
	} else {
		s += msgp.IntSize
	}
	s += 8
	if z.Maximum == nil {
		s += msgp.NilSize
	} else {
		s += msgp.IntSize
	}
	return
}

// DecodeMsg implements msgp.Decodable
func (z *LoginIDKeyType) DecodeMsg(dc *msgp.Reader) (err error) {
	{
		var zare string
		zare, err = dc.ReadString()
		(*z) = LoginIDKeyType(zare)
	}
	if err != nil {
		return
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z LoginIDKeyType) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteString(string(z))
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z LoginIDKeyType) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	o = msgp.AppendString(o, string(z))
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *LoginIDKeyType) UnmarshalMsg(bts []byte) (o []byte, err error) {
	{
		var zljy string
		zljy, bts, err = msgp.ReadStringBytes(bts)
		(*z) = LoginIDKeyType(zljy)
	}
	if err != nil {
		return
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z LoginIDKeyType) Msgsize() (s int) {
	s = msgp.StringPrefixSize + len(string(z))
	return
}

// DecodeMsg implements msgp.Decodable
func (z *MFABearerTokenConfiguration) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zixj uint32
	zixj, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zixj > 0 {
		zixj--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "expire_in_days":
			z.ExpireInDays, err = dc.ReadInt()
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
func (z MFABearerTokenConfiguration) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 1
	// write "expire_in_days"
	err = en.Append(0x81, 0xae, 0x65, 0x78, 0x70, 0x69, 0x72, 0x65, 0x5f, 0x69, 0x6e, 0x5f, 0x64, 0x61, 0x79, 0x73)
	if err != nil {
		return err
	}
	err = en.WriteInt(z.ExpireInDays)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z MFABearerTokenConfiguration) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 1
	// string "expire_in_days"
	o = append(o, 0x81, 0xae, 0x65, 0x78, 0x70, 0x69, 0x72, 0x65, 0x5f, 0x69, 0x6e, 0x5f, 0x64, 0x61, 0x79, 0x73)
	o = msgp.AppendInt(o, z.ExpireInDays)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *MFABearerTokenConfiguration) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zrsc uint32
	zrsc, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zrsc > 0 {
		zrsc--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "expire_in_days":
			z.ExpireInDays, bts, err = msgp.ReadIntBytes(bts)
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
func (z MFABearerTokenConfiguration) Msgsize() (s int) {
	s = 1 + 15 + msgp.IntSize
	return
}

// DecodeMsg implements msgp.Decodable
func (z *MFAConfiguration) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zctn uint32
	zctn, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zctn > 0 {
		zctn--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "enforcement":
			{
				var zswy string
				zswy, err = dc.ReadString()
				z.Enforcement = MFAEnforcement(zswy)
			}
			if err != nil {
				return
			}
		case "maximum":
			if dc.IsNil() {
				err = dc.ReadNil()
				if err != nil {
					return
				}
				z.Maximum = nil
			} else {
				if z.Maximum == nil {
					z.Maximum = new(int)
				}
				*z.Maximum, err = dc.ReadInt()
				if err != nil {
					return
				}
			}
		case "totp":
			var znsg uint32
			znsg, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			for znsg > 0 {
				znsg--
				field, err = dc.ReadMapKeyPtr()
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "maximum":
					z.TOTP.Maximum, err = dc.ReadInt()
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
		case "oob":
			var zrus uint32
			zrus, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			for zrus > 0 {
				zrus--
				field, err = dc.ReadMapKeyPtr()
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "sms":
					var zsvm uint32
					zsvm, err = dc.ReadMapHeader()
					if err != nil {
						return
					}
					for zsvm > 0 {
						zsvm--
						field, err = dc.ReadMapKeyPtr()
						if err != nil {
							return
						}
						switch msgp.UnsafeString(field) {
						case "maximum":
							z.OOB.SMS.Maximum, err = dc.ReadInt()
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
				case "email":
					var zaoz uint32
					zaoz, err = dc.ReadMapHeader()
					if err != nil {
						return
					}
					for zaoz > 0 {
						zaoz--
						field, err = dc.ReadMapKeyPtr()
						if err != nil {
							return
						}
						switch msgp.UnsafeString(field) {
						case "maximum":
							z.OOB.Email.Maximum, err = dc.ReadInt()
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
		case "bearer_token":
			var zfzb uint32
			zfzb, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			for zfzb > 0 {
				zfzb--
				field, err = dc.ReadMapKeyPtr()
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "expire_in_days":
					z.BearerToken.ExpireInDays, err = dc.ReadInt()
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
		case "recovery_code":
			var zsbo uint32
			zsbo, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			for zsbo > 0 {
				zsbo--
				field, err = dc.ReadMapKeyPtr()
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "count":
					z.RecoveryCode.Count, err = dc.ReadInt()
					if err != nil {
						return
					}
				case "list_enabled":
					z.RecoveryCode.ListEnabled, err = dc.ReadBool()
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
func (z *MFAConfiguration) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 6
	// write "enforcement"
	err = en.Append(0x86, 0xab, 0x65, 0x6e, 0x66, 0x6f, 0x72, 0x63, 0x65, 0x6d, 0x65, 0x6e, 0x74)
	if err != nil {
		return err
	}
	err = en.WriteString(string(z.Enforcement))
	if err != nil {
		return
	}
	// write "maximum"
	err = en.Append(0xa7, 0x6d, 0x61, 0x78, 0x69, 0x6d, 0x75, 0x6d)
	if err != nil {
		return err
	}
	if z.Maximum == nil {
		err = en.WriteNil()
		if err != nil {
			return
		}
	} else {
		err = en.WriteInt(*z.Maximum)
		if err != nil {
			return
		}
	}
	// write "totp"
	// map header, size 1
	// write "maximum"
	err = en.Append(0xa4, 0x74, 0x6f, 0x74, 0x70, 0x81, 0xa7, 0x6d, 0x61, 0x78, 0x69, 0x6d, 0x75, 0x6d)
	if err != nil {
		return err
	}
	err = en.WriteInt(z.TOTP.Maximum)
	if err != nil {
		return
	}
	// write "oob"
	// map header, size 2
	// write "sms"
	// map header, size 1
	// write "maximum"
	err = en.Append(0xa3, 0x6f, 0x6f, 0x62, 0x82, 0xa3, 0x73, 0x6d, 0x73, 0x81, 0xa7, 0x6d, 0x61, 0x78, 0x69, 0x6d, 0x75, 0x6d)
	if err != nil {
		return err
	}
	err = en.WriteInt(z.OOB.SMS.Maximum)
	if err != nil {
		return
	}
	// write "email"
	// map header, size 1
	// write "maximum"
	err = en.Append(0xa5, 0x65, 0x6d, 0x61, 0x69, 0x6c, 0x81, 0xa7, 0x6d, 0x61, 0x78, 0x69, 0x6d, 0x75, 0x6d)
	if err != nil {
		return err
	}
	err = en.WriteInt(z.OOB.Email.Maximum)
	if err != nil {
		return
	}
	// write "bearer_token"
	// map header, size 1
	// write "expire_in_days"
	err = en.Append(0xac, 0x62, 0x65, 0x61, 0x72, 0x65, 0x72, 0x5f, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x81, 0xae, 0x65, 0x78, 0x70, 0x69, 0x72, 0x65, 0x5f, 0x69, 0x6e, 0x5f, 0x64, 0x61, 0x79, 0x73)
	if err != nil {
		return err
	}
	err = en.WriteInt(z.BearerToken.ExpireInDays)
	if err != nil {
		return
	}
	// write "recovery_code"
	// map header, size 2
	// write "count"
	err = en.Append(0xad, 0x72, 0x65, 0x63, 0x6f, 0x76, 0x65, 0x72, 0x79, 0x5f, 0x63, 0x6f, 0x64, 0x65, 0x82, 0xa5, 0x63, 0x6f, 0x75, 0x6e, 0x74)
	if err != nil {
		return err
	}
	err = en.WriteInt(z.RecoveryCode.Count)
	if err != nil {
		return
	}
	// write "list_enabled"
	err = en.Append(0xac, 0x6c, 0x69, 0x73, 0x74, 0x5f, 0x65, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64)
	if err != nil {
		return err
	}
	err = en.WriteBool(z.RecoveryCode.ListEnabled)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *MFAConfiguration) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 6
	// string "enforcement"
	o = append(o, 0x86, 0xab, 0x65, 0x6e, 0x66, 0x6f, 0x72, 0x63, 0x65, 0x6d, 0x65, 0x6e, 0x74)
	o = msgp.AppendString(o, string(z.Enforcement))
	// string "maximum"
	o = append(o, 0xa7, 0x6d, 0x61, 0x78, 0x69, 0x6d, 0x75, 0x6d)
	if z.Maximum == nil {
		o = msgp.AppendNil(o)
	} else {
		o = msgp.AppendInt(o, *z.Maximum)
	}
	// string "totp"
	// map header, size 1
	// string "maximum"
	o = append(o, 0xa4, 0x74, 0x6f, 0x74, 0x70, 0x81, 0xa7, 0x6d, 0x61, 0x78, 0x69, 0x6d, 0x75, 0x6d)
	o = msgp.AppendInt(o, z.TOTP.Maximum)
	// string "oob"
	// map header, size 2
	// string "sms"
	// map header, size 1
	// string "maximum"
	o = append(o, 0xa3, 0x6f, 0x6f, 0x62, 0x82, 0xa3, 0x73, 0x6d, 0x73, 0x81, 0xa7, 0x6d, 0x61, 0x78, 0x69, 0x6d, 0x75, 0x6d)
	o = msgp.AppendInt(o, z.OOB.SMS.Maximum)
	// string "email"
	// map header, size 1
	// string "maximum"
	o = append(o, 0xa5, 0x65, 0x6d, 0x61, 0x69, 0x6c, 0x81, 0xa7, 0x6d, 0x61, 0x78, 0x69, 0x6d, 0x75, 0x6d)
	o = msgp.AppendInt(o, z.OOB.Email.Maximum)
	// string "bearer_token"
	// map header, size 1
	// string "expire_in_days"
	o = append(o, 0xac, 0x62, 0x65, 0x61, 0x72, 0x65, 0x72, 0x5f, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x81, 0xae, 0x65, 0x78, 0x70, 0x69, 0x72, 0x65, 0x5f, 0x69, 0x6e, 0x5f, 0x64, 0x61, 0x79, 0x73)
	o = msgp.AppendInt(o, z.BearerToken.ExpireInDays)
	// string "recovery_code"
	// map header, size 2
	// string "count"
	o = append(o, 0xad, 0x72, 0x65, 0x63, 0x6f, 0x76, 0x65, 0x72, 0x79, 0x5f, 0x63, 0x6f, 0x64, 0x65, 0x82, 0xa5, 0x63, 0x6f, 0x75, 0x6e, 0x74)
	o = msgp.AppendInt(o, z.RecoveryCode.Count)
	// string "list_enabled"
	o = append(o, 0xac, 0x6c, 0x69, 0x73, 0x74, 0x5f, 0x65, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64)
	o = msgp.AppendBool(o, z.RecoveryCode.ListEnabled)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *MFAConfiguration) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zjif uint32
	zjif, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zjif > 0 {
		zjif--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "enforcement":
			{
				var zqgz string
				zqgz, bts, err = msgp.ReadStringBytes(bts)
				z.Enforcement = MFAEnforcement(zqgz)
			}
			if err != nil {
				return
			}
		case "maximum":
			if msgp.IsNil(bts) {
				bts, err = msgp.ReadNilBytes(bts)
				if err != nil {
					return
				}
				z.Maximum = nil
			} else {
				if z.Maximum == nil {
					z.Maximum = new(int)
				}
				*z.Maximum, bts, err = msgp.ReadIntBytes(bts)
				if err != nil {
					return
				}
			}
		case "totp":
			var zsnw uint32
			zsnw, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			for zsnw > 0 {
				zsnw--
				field, bts, err = msgp.ReadMapKeyZC(bts)
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "maximum":
					z.TOTP.Maximum, bts, err = msgp.ReadIntBytes(bts)
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
		case "oob":
			var ztls uint32
			ztls, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			for ztls > 0 {
				ztls--
				field, bts, err = msgp.ReadMapKeyZC(bts)
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "sms":
					var zmvo uint32
					zmvo, bts, err = msgp.ReadMapHeaderBytes(bts)
					if err != nil {
						return
					}
					for zmvo > 0 {
						zmvo--
						field, bts, err = msgp.ReadMapKeyZC(bts)
						if err != nil {
							return
						}
						switch msgp.UnsafeString(field) {
						case "maximum":
							z.OOB.SMS.Maximum, bts, err = msgp.ReadIntBytes(bts)
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
				case "email":
					var zigk uint32
					zigk, bts, err = msgp.ReadMapHeaderBytes(bts)
					if err != nil {
						return
					}
					for zigk > 0 {
						zigk--
						field, bts, err = msgp.ReadMapKeyZC(bts)
						if err != nil {
							return
						}
						switch msgp.UnsafeString(field) {
						case "maximum":
							z.OOB.Email.Maximum, bts, err = msgp.ReadIntBytes(bts)
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
		case "bearer_token":
			var zopb uint32
			zopb, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			for zopb > 0 {
				zopb--
				field, bts, err = msgp.ReadMapKeyZC(bts)
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "expire_in_days":
					z.BearerToken.ExpireInDays, bts, err = msgp.ReadIntBytes(bts)
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
		case "recovery_code":
			var zuop uint32
			zuop, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			for zuop > 0 {
				zuop--
				field, bts, err = msgp.ReadMapKeyZC(bts)
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "count":
					z.RecoveryCode.Count, bts, err = msgp.ReadIntBytes(bts)
					if err != nil {
						return
					}
				case "list_enabled":
					z.RecoveryCode.ListEnabled, bts, err = msgp.ReadBoolBytes(bts)
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
func (z *MFAConfiguration) Msgsize() (s int) {
	s = 1 + 12 + msgp.StringPrefixSize + len(string(z.Enforcement)) + 8
	if z.Maximum == nil {
		s += msgp.NilSize
	} else {
		s += msgp.IntSize
	}
	s += 5 + 1 + 8 + msgp.IntSize + 4 + 1 + 4 + 1 + 8 + msgp.IntSize + 6 + 1 + 8 + msgp.IntSize + 13 + 1 + 15 + msgp.IntSize + 14 + 1 + 6 + msgp.IntSize + 13 + msgp.BoolSize
	return
}

// DecodeMsg implements msgp.Decodable
func (z *MFAEnforcement) DecodeMsg(dc *msgp.Reader) (err error) {
	{
		var zedl string
		zedl, err = dc.ReadString()
		(*z) = MFAEnforcement(zedl)
	}
	if err != nil {
		return
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z MFAEnforcement) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteString(string(z))
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z MFAEnforcement) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	o = msgp.AppendString(o, string(z))
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *MFAEnforcement) UnmarshalMsg(bts []byte) (o []byte, err error) {
	{
		var zupd string
		zupd, bts, err = msgp.ReadStringBytes(bts)
		(*z) = MFAEnforcement(zupd)
	}
	if err != nil {
		return
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z MFAEnforcement) Msgsize() (s int) {
	s = msgp.StringPrefixSize + len(string(z))
	return
}

// DecodeMsg implements msgp.Decodable
func (z *MFAOOBConfiguration) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zome uint32
	zome, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zome > 0 {
		zome--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "sms":
			var zrvj uint32
			zrvj, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			for zrvj > 0 {
				zrvj--
				field, err = dc.ReadMapKeyPtr()
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "maximum":
					z.SMS.Maximum, err = dc.ReadInt()
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
		case "email":
			var zarz uint32
			zarz, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			for zarz > 0 {
				zarz--
				field, err = dc.ReadMapKeyPtr()
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "maximum":
					z.Email.Maximum, err = dc.ReadInt()
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
func (z *MFAOOBConfiguration) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 2
	// write "sms"
	// map header, size 1
	// write "maximum"
	err = en.Append(0x82, 0xa3, 0x73, 0x6d, 0x73, 0x81, 0xa7, 0x6d, 0x61, 0x78, 0x69, 0x6d, 0x75, 0x6d)
	if err != nil {
		return err
	}
	err = en.WriteInt(z.SMS.Maximum)
	if err != nil {
		return
	}
	// write "email"
	// map header, size 1
	// write "maximum"
	err = en.Append(0xa5, 0x65, 0x6d, 0x61, 0x69, 0x6c, 0x81, 0xa7, 0x6d, 0x61, 0x78, 0x69, 0x6d, 0x75, 0x6d)
	if err != nil {
		return err
	}
	err = en.WriteInt(z.Email.Maximum)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *MFAOOBConfiguration) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "sms"
	// map header, size 1
	// string "maximum"
	o = append(o, 0x82, 0xa3, 0x73, 0x6d, 0x73, 0x81, 0xa7, 0x6d, 0x61, 0x78, 0x69, 0x6d, 0x75, 0x6d)
	o = msgp.AppendInt(o, z.SMS.Maximum)
	// string "email"
	// map header, size 1
	// string "maximum"
	o = append(o, 0xa5, 0x65, 0x6d, 0x61, 0x69, 0x6c, 0x81, 0xa7, 0x6d, 0x61, 0x78, 0x69, 0x6d, 0x75, 0x6d)
	o = msgp.AppendInt(o, z.Email.Maximum)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *MFAOOBConfiguration) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zknt uint32
	zknt, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zknt > 0 {
		zknt--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "sms":
			var zxye uint32
			zxye, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			for zxye > 0 {
				zxye--
				field, bts, err = msgp.ReadMapKeyZC(bts)
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "maximum":
					z.SMS.Maximum, bts, err = msgp.ReadIntBytes(bts)
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
		case "email":
			var zucw uint32
			zucw, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			for zucw > 0 {
				zucw--
				field, bts, err = msgp.ReadMapKeyZC(bts)
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "maximum":
					z.Email.Maximum, bts, err = msgp.ReadIntBytes(bts)
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
func (z *MFAOOBConfiguration) Msgsize() (s int) {
	s = 1 + 4 + 1 + 8 + msgp.IntSize + 6 + 1 + 8 + msgp.IntSize
	return
}

// DecodeMsg implements msgp.Decodable
func (z *MFAOOBEmailConfiguration) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zlsx uint32
	zlsx, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zlsx > 0 {
		zlsx--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "maximum":
			z.Maximum, err = dc.ReadInt()
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
func (z MFAOOBEmailConfiguration) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 1
	// write "maximum"
	err = en.Append(0x81, 0xa7, 0x6d, 0x61, 0x78, 0x69, 0x6d, 0x75, 0x6d)
	if err != nil {
		return err
	}
	err = en.WriteInt(z.Maximum)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z MFAOOBEmailConfiguration) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 1
	// string "maximum"
	o = append(o, 0x81, 0xa7, 0x6d, 0x61, 0x78, 0x69, 0x6d, 0x75, 0x6d)
	o = msgp.AppendInt(o, z.Maximum)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *MFAOOBEmailConfiguration) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zbgy uint32
	zbgy, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zbgy > 0 {
		zbgy--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "maximum":
			z.Maximum, bts, err = msgp.ReadIntBytes(bts)
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
func (z MFAOOBEmailConfiguration) Msgsize() (s int) {
	s = 1 + 8 + msgp.IntSize
	return
}

// DecodeMsg implements msgp.Decodable
func (z *MFAOOBSMSConfiguration) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zrao uint32
	zrao, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zrao > 0 {
		zrao--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "maximum":
			z.Maximum, err = dc.ReadInt()
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
func (z MFAOOBSMSConfiguration) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 1
	// write "maximum"
	err = en.Append(0x81, 0xa7, 0x6d, 0x61, 0x78, 0x69, 0x6d, 0x75, 0x6d)
	if err != nil {
		return err
	}
	err = en.WriteInt(z.Maximum)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z MFAOOBSMSConfiguration) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 1
	// string "maximum"
	o = append(o, 0x81, 0xa7, 0x6d, 0x61, 0x78, 0x69, 0x6d, 0x75, 0x6d)
	o = msgp.AppendInt(o, z.Maximum)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *MFAOOBSMSConfiguration) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zmbt uint32
	zmbt, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zmbt > 0 {
		zmbt--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "maximum":
			z.Maximum, bts, err = msgp.ReadIntBytes(bts)
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
func (z MFAOOBSMSConfiguration) Msgsize() (s int) {
	s = 1 + 8 + msgp.IntSize
	return
}

// DecodeMsg implements msgp.Decodable
func (z *MFARecoveryCodeConfiguration) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zvls uint32
	zvls, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zvls > 0 {
		zvls--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "count":
			z.Count, err = dc.ReadInt()
			if err != nil {
				return
			}
		case "list_enabled":
			z.ListEnabled, err = dc.ReadBool()
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
func (z MFARecoveryCodeConfiguration) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 2
	// write "count"
	err = en.Append(0x82, 0xa5, 0x63, 0x6f, 0x75, 0x6e, 0x74)
	if err != nil {
		return err
	}
	err = en.WriteInt(z.Count)
	if err != nil {
		return
	}
	// write "list_enabled"
	err = en.Append(0xac, 0x6c, 0x69, 0x73, 0x74, 0x5f, 0x65, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64)
	if err != nil {
		return err
	}
	err = en.WriteBool(z.ListEnabled)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z MFARecoveryCodeConfiguration) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "count"
	o = append(o, 0x82, 0xa5, 0x63, 0x6f, 0x75, 0x6e, 0x74)
	o = msgp.AppendInt(o, z.Count)
	// string "list_enabled"
	o = append(o, 0xac, 0x6c, 0x69, 0x73, 0x74, 0x5f, 0x65, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64)
	o = msgp.AppendBool(o, z.ListEnabled)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *MFARecoveryCodeConfiguration) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zjfj uint32
	zjfj, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zjfj > 0 {
		zjfj--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "count":
			z.Count, bts, err = msgp.ReadIntBytes(bts)
			if err != nil {
				return
			}
		case "list_enabled":
			z.ListEnabled, bts, err = msgp.ReadBoolBytes(bts)
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
func (z MFARecoveryCodeConfiguration) Msgsize() (s int) {
	s = 1 + 6 + msgp.IntSize + 13 + msgp.BoolSize
	return
}

// DecodeMsg implements msgp.Decodable
func (z *MFATOTPConfiguration) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zzak uint32
	zzak, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zzak > 0 {
		zzak--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "maximum":
			z.Maximum, err = dc.ReadInt()
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
func (z MFATOTPConfiguration) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 1
	// write "maximum"
	err = en.Append(0x81, 0xa7, 0x6d, 0x61, 0x78, 0x69, 0x6d, 0x75, 0x6d)
	if err != nil {
		return err
	}
	err = en.WriteInt(z.Maximum)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z MFATOTPConfiguration) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 1
	// string "maximum"
	o = append(o, 0x81, 0xa7, 0x6d, 0x61, 0x78, 0x69, 0x6d, 0x75, 0x6d)
	o = msgp.AppendInt(o, z.Maximum)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *MFATOTPConfiguration) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zbtz uint32
	zbtz, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zbtz > 0 {
		zbtz--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "maximum":
			z.Maximum, bts, err = msgp.ReadIntBytes(bts)
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
func (z MFATOTPConfiguration) Msgsize() (s int) {
	s = 1 + 8 + msgp.IntSize
	return
}

// DecodeMsg implements msgp.Decodable
func (z *NexmoConfiguration) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zsym uint32
	zsym, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zsym > 0 {
		zsym--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "api_key":
			z.APIKey, err = dc.ReadString()
			if err != nil {
				return
			}
		case "secret":
			z.APISecret, err = dc.ReadString()
			if err != nil {
				return
			}
		case "from":
			z.From, err = dc.ReadString()
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
func (z NexmoConfiguration) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 3
	// write "api_key"
	err = en.Append(0x83, 0xa7, 0x61, 0x70, 0x69, 0x5f, 0x6b, 0x65, 0x79)
	if err != nil {
		return err
	}
	err = en.WriteString(z.APIKey)
	if err != nil {
		return
	}
	// write "secret"
	err = en.Append(0xa6, 0x73, 0x65, 0x63, 0x72, 0x65, 0x74)
	if err != nil {
		return err
	}
	err = en.WriteString(z.APISecret)
	if err != nil {
		return
	}
	// write "from"
	err = en.Append(0xa4, 0x66, 0x72, 0x6f, 0x6d)
	if err != nil {
		return err
	}
	err = en.WriteString(z.From)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z NexmoConfiguration) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 3
	// string "api_key"
	o = append(o, 0x83, 0xa7, 0x61, 0x70, 0x69, 0x5f, 0x6b, 0x65, 0x79)
	o = msgp.AppendString(o, z.APIKey)
	// string "secret"
	o = append(o, 0xa6, 0x73, 0x65, 0x63, 0x72, 0x65, 0x74)
	o = msgp.AppendString(o, z.APISecret)
	// string "from"
	o = append(o, 0xa4, 0x66, 0x72, 0x6f, 0x6d)
	o = msgp.AppendString(o, z.From)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *NexmoConfiguration) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zgeu uint32
	zgeu, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zgeu > 0 {
		zgeu--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "api_key":
			z.APIKey, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "secret":
			z.APISecret, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "from":
			z.From, bts, err = msgp.ReadStringBytes(bts)
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
func (z NexmoConfiguration) Msgsize() (s int) {
	s = 1 + 8 + msgp.StringPrefixSize + len(z.APIKey) + 7 + msgp.StringPrefixSize + len(z.APISecret) + 5 + msgp.StringPrefixSize + len(z.From)
	return
}

// DecodeMsg implements msgp.Decodable
func (z *OAuthConfiguration) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zdqi uint32
	zdqi, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zdqi > 0 {
		zdqi--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "url_prefix":
			z.URLPrefix, err = dc.ReadString()
			if err != nil {
				return
			}
		case "state_jwt_secret":
			z.StateJWTSecret, err = dc.ReadString()
			if err != nil {
				return
			}
		case "allowed_callback_urls":
			var zyco uint32
			zyco, err = dc.ReadArrayHeader()
			if err != nil {
				return
			}
			if cap(z.AllowedCallbackURLs) >= int(zyco) {
				z.AllowedCallbackURLs = (z.AllowedCallbackURLs)[:zyco]
			} else {
				z.AllowedCallbackURLs = make([]string, zyco)
			}
			for zdtr := range z.AllowedCallbackURLs {
				z.AllowedCallbackURLs[zdtr], err = dc.ReadString()
				if err != nil {
					return
				}
			}
		case "external_access_token_flow_enabled":
			z.ExternalAccessTokenFlowEnabled, err = dc.ReadBool()
			if err != nil {
				return
			}
		case "on_user_duplicate_allow_merge":
			z.OnUserDuplicateAllowMerge, err = dc.ReadBool()
			if err != nil {
				return
			}
		case "on_user_duplicate_allow_create":
			z.OnUserDuplicateAllowCreate, err = dc.ReadBool()
			if err != nil {
				return
			}
		case "providers":
			var zhgh uint32
			zhgh, err = dc.ReadArrayHeader()
			if err != nil {
				return
			}
			if cap(z.Providers) >= int(zhgh) {
				z.Providers = (z.Providers)[:zhgh]
			} else {
				z.Providers = make([]OAuthProviderConfiguration, zhgh)
			}
			for zzqm := range z.Providers {
				err = z.Providers[zzqm].DecodeMsg(dc)
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
func (z *OAuthConfiguration) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 7
	// write "url_prefix"
	err = en.Append(0x87, 0xaa, 0x75, 0x72, 0x6c, 0x5f, 0x70, 0x72, 0x65, 0x66, 0x69, 0x78)
	if err != nil {
		return err
	}
	err = en.WriteString(z.URLPrefix)
	if err != nil {
		return
	}
	// write "state_jwt_secret"
	err = en.Append(0xb0, 0x73, 0x74, 0x61, 0x74, 0x65, 0x5f, 0x6a, 0x77, 0x74, 0x5f, 0x73, 0x65, 0x63, 0x72, 0x65, 0x74)
	if err != nil {
		return err
	}
	err = en.WriteString(z.StateJWTSecret)
	if err != nil {
		return
	}
	// write "allowed_callback_urls"
	err = en.Append(0xb5, 0x61, 0x6c, 0x6c, 0x6f, 0x77, 0x65, 0x64, 0x5f, 0x63, 0x61, 0x6c, 0x6c, 0x62, 0x61, 0x63, 0x6b, 0x5f, 0x75, 0x72, 0x6c, 0x73)
	if err != nil {
		return err
	}
	err = en.WriteArrayHeader(uint32(len(z.AllowedCallbackURLs)))
	if err != nil {
		return
	}
	for zdtr := range z.AllowedCallbackURLs {
		err = en.WriteString(z.AllowedCallbackURLs[zdtr])
		if err != nil {
			return
		}
	}
	// write "external_access_token_flow_enabled"
	err = en.Append(0xd9, 0x22, 0x65, 0x78, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x5f, 0x61, 0x63, 0x63, 0x65, 0x73, 0x73, 0x5f, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x5f, 0x66, 0x6c, 0x6f, 0x77, 0x5f, 0x65, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64)
	if err != nil {
		return err
	}
	err = en.WriteBool(z.ExternalAccessTokenFlowEnabled)
	if err != nil {
		return
	}
	// write "on_user_duplicate_allow_merge"
	err = en.Append(0xbd, 0x6f, 0x6e, 0x5f, 0x75, 0x73, 0x65, 0x72, 0x5f, 0x64, 0x75, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x65, 0x5f, 0x61, 0x6c, 0x6c, 0x6f, 0x77, 0x5f, 0x6d, 0x65, 0x72, 0x67, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteBool(z.OnUserDuplicateAllowMerge)
	if err != nil {
		return
	}
	// write "on_user_duplicate_allow_create"
	err = en.Append(0xbe, 0x6f, 0x6e, 0x5f, 0x75, 0x73, 0x65, 0x72, 0x5f, 0x64, 0x75, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x65, 0x5f, 0x61, 0x6c, 0x6c, 0x6f, 0x77, 0x5f, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteBool(z.OnUserDuplicateAllowCreate)
	if err != nil {
		return
	}
	// write "providers"
	err = en.Append(0xa9, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x64, 0x65, 0x72, 0x73)
	if err != nil {
		return err
	}
	err = en.WriteArrayHeader(uint32(len(z.Providers)))
	if err != nil {
		return
	}
	for zzqm := range z.Providers {
		err = z.Providers[zzqm].EncodeMsg(en)
		if err != nil {
			return
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *OAuthConfiguration) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 7
	// string "url_prefix"
	o = append(o, 0x87, 0xaa, 0x75, 0x72, 0x6c, 0x5f, 0x70, 0x72, 0x65, 0x66, 0x69, 0x78)
	o = msgp.AppendString(o, z.URLPrefix)
	// string "state_jwt_secret"
	o = append(o, 0xb0, 0x73, 0x74, 0x61, 0x74, 0x65, 0x5f, 0x6a, 0x77, 0x74, 0x5f, 0x73, 0x65, 0x63, 0x72, 0x65, 0x74)
	o = msgp.AppendString(o, z.StateJWTSecret)
	// string "allowed_callback_urls"
	o = append(o, 0xb5, 0x61, 0x6c, 0x6c, 0x6f, 0x77, 0x65, 0x64, 0x5f, 0x63, 0x61, 0x6c, 0x6c, 0x62, 0x61, 0x63, 0x6b, 0x5f, 0x75, 0x72, 0x6c, 0x73)
	o = msgp.AppendArrayHeader(o, uint32(len(z.AllowedCallbackURLs)))
	for zdtr := range z.AllowedCallbackURLs {
		o = msgp.AppendString(o, z.AllowedCallbackURLs[zdtr])
	}
	// string "external_access_token_flow_enabled"
	o = append(o, 0xd9, 0x22, 0x65, 0x78, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x5f, 0x61, 0x63, 0x63, 0x65, 0x73, 0x73, 0x5f, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x5f, 0x66, 0x6c, 0x6f, 0x77, 0x5f, 0x65, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64)
	o = msgp.AppendBool(o, z.ExternalAccessTokenFlowEnabled)
	// string "on_user_duplicate_allow_merge"
	o = append(o, 0xbd, 0x6f, 0x6e, 0x5f, 0x75, 0x73, 0x65, 0x72, 0x5f, 0x64, 0x75, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x65, 0x5f, 0x61, 0x6c, 0x6c, 0x6f, 0x77, 0x5f, 0x6d, 0x65, 0x72, 0x67, 0x65)
	o = msgp.AppendBool(o, z.OnUserDuplicateAllowMerge)
	// string "on_user_duplicate_allow_create"
	o = append(o, 0xbe, 0x6f, 0x6e, 0x5f, 0x75, 0x73, 0x65, 0x72, 0x5f, 0x64, 0x75, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x65, 0x5f, 0x61, 0x6c, 0x6c, 0x6f, 0x77, 0x5f, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65)
	o = msgp.AppendBool(o, z.OnUserDuplicateAllowCreate)
	// string "providers"
	o = append(o, 0xa9, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x64, 0x65, 0x72, 0x73)
	o = msgp.AppendArrayHeader(o, uint32(len(z.Providers)))
	for zzqm := range z.Providers {
		o, err = z.Providers[zzqm].MarshalMsg(o)
		if err != nil {
			return
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *OAuthConfiguration) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zovg uint32
	zovg, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zovg > 0 {
		zovg--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "url_prefix":
			z.URLPrefix, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "state_jwt_secret":
			z.StateJWTSecret, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "allowed_callback_urls":
			var zsey uint32
			zsey, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.AllowedCallbackURLs) >= int(zsey) {
				z.AllowedCallbackURLs = (z.AllowedCallbackURLs)[:zsey]
			} else {
				z.AllowedCallbackURLs = make([]string, zsey)
			}
			for zdtr := range z.AllowedCallbackURLs {
				z.AllowedCallbackURLs[zdtr], bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
			}
		case "external_access_token_flow_enabled":
			z.ExternalAccessTokenFlowEnabled, bts, err = msgp.ReadBoolBytes(bts)
			if err != nil {
				return
			}
		case "on_user_duplicate_allow_merge":
			z.OnUserDuplicateAllowMerge, bts, err = msgp.ReadBoolBytes(bts)
			if err != nil {
				return
			}
		case "on_user_duplicate_allow_create":
			z.OnUserDuplicateAllowCreate, bts, err = msgp.ReadBoolBytes(bts)
			if err != nil {
				return
			}
		case "providers":
			var zcjp uint32
			zcjp, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.Providers) >= int(zcjp) {
				z.Providers = (z.Providers)[:zcjp]
			} else {
				z.Providers = make([]OAuthProviderConfiguration, zcjp)
			}
			for zzqm := range z.Providers {
				bts, err = z.Providers[zzqm].UnmarshalMsg(bts)
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
func (z *OAuthConfiguration) Msgsize() (s int) {
	s = 1 + 11 + msgp.StringPrefixSize + len(z.URLPrefix) + 17 + msgp.StringPrefixSize + len(z.StateJWTSecret) + 22 + msgp.ArrayHeaderSize
	for zdtr := range z.AllowedCallbackURLs {
		s += msgp.StringPrefixSize + len(z.AllowedCallbackURLs[zdtr])
	}
	s += 36 + msgp.BoolSize + 30 + msgp.BoolSize + 31 + msgp.BoolSize + 10 + msgp.ArrayHeaderSize
	for zzqm := range z.Providers {
		s += z.Providers[zzqm].Msgsize()
	}
	return
}

// DecodeMsg implements msgp.Decodable
func (z *OAuthProviderConfiguration) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zjhy uint32
	zjhy, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zjhy > 0 {
		zjhy--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "id":
			z.ID, err = dc.ReadString()
			if err != nil {
				return
			}
		case "type":
			{
				var znuf string
				znuf, err = dc.ReadString()
				z.Type = OAuthProviderType(znuf)
			}
			if err != nil {
				return
			}
		case "client_id":
			z.ClientID, err = dc.ReadString()
			if err != nil {
				return
			}
		case "client_secret":
			z.ClientSecret, err = dc.ReadString()
			if err != nil {
				return
			}
		case "scope":
			z.Scope, err = dc.ReadString()
			if err != nil {
				return
			}
		case "tenant":
			z.Tenant, err = dc.ReadString()
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
func (z *OAuthProviderConfiguration) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 6
	// write "id"
	err = en.Append(0x86, 0xa2, 0x69, 0x64)
	if err != nil {
		return err
	}
	err = en.WriteString(z.ID)
	if err != nil {
		return
	}
	// write "type"
	err = en.Append(0xa4, 0x74, 0x79, 0x70, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteString(string(z.Type))
	if err != nil {
		return
	}
	// write "client_id"
	err = en.Append(0xa9, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x5f, 0x69, 0x64)
	if err != nil {
		return err
	}
	err = en.WriteString(z.ClientID)
	if err != nil {
		return
	}
	// write "client_secret"
	err = en.Append(0xad, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x5f, 0x73, 0x65, 0x63, 0x72, 0x65, 0x74)
	if err != nil {
		return err
	}
	err = en.WriteString(z.ClientSecret)
	if err != nil {
		return
	}
	// write "scope"
	err = en.Append(0xa5, 0x73, 0x63, 0x6f, 0x70, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Scope)
	if err != nil {
		return
	}
	// write "tenant"
	err = en.Append(0xa6, 0x74, 0x65, 0x6e, 0x61, 0x6e, 0x74)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Tenant)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *OAuthProviderConfiguration) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 6
	// string "id"
	o = append(o, 0x86, 0xa2, 0x69, 0x64)
	o = msgp.AppendString(o, z.ID)
	// string "type"
	o = append(o, 0xa4, 0x74, 0x79, 0x70, 0x65)
	o = msgp.AppendString(o, string(z.Type))
	// string "client_id"
	o = append(o, 0xa9, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x5f, 0x69, 0x64)
	o = msgp.AppendString(o, z.ClientID)
	// string "client_secret"
	o = append(o, 0xad, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x5f, 0x73, 0x65, 0x63, 0x72, 0x65, 0x74)
	o = msgp.AppendString(o, z.ClientSecret)
	// string "scope"
	o = append(o, 0xa5, 0x73, 0x63, 0x6f, 0x70, 0x65)
	o = msgp.AppendString(o, z.Scope)
	// string "tenant"
	o = append(o, 0xa6, 0x74, 0x65, 0x6e, 0x61, 0x6e, 0x74)
	o = msgp.AppendString(o, z.Tenant)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *OAuthProviderConfiguration) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var znjj uint32
	znjj, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for znjj > 0 {
		znjj--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "id":
			z.ID, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "type":
			{
				var zhhj string
				zhhj, bts, err = msgp.ReadStringBytes(bts)
				z.Type = OAuthProviderType(zhhj)
			}
			if err != nil {
				return
			}
		case "client_id":
			z.ClientID, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "client_secret":
			z.ClientSecret, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "scope":
			z.Scope, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "tenant":
			z.Tenant, bts, err = msgp.ReadStringBytes(bts)
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
func (z *OAuthProviderConfiguration) Msgsize() (s int) {
	s = 1 + 3 + msgp.StringPrefixSize + len(z.ID) + 5 + msgp.StringPrefixSize + len(string(z.Type)) + 10 + msgp.StringPrefixSize + len(z.ClientID) + 14 + msgp.StringPrefixSize + len(z.ClientSecret) + 6 + msgp.StringPrefixSize + len(z.Scope) + 7 + msgp.StringPrefixSize + len(z.Tenant)
	return
}

// DecodeMsg implements msgp.Decodable
func (z *OAuthProviderType) DecodeMsg(dc *msgp.Reader) (err error) {
	{
		var zuvr string
		zuvr, err = dc.ReadString()
		(*z) = OAuthProviderType(zuvr)
	}
	if err != nil {
		return
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z OAuthProviderType) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteString(string(z))
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z OAuthProviderType) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	o = msgp.AppendString(o, string(z))
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *OAuthProviderType) UnmarshalMsg(bts []byte) (o []byte, err error) {
	{
		var zusq string
		zusq, bts, err = msgp.ReadStringBytes(bts)
		(*z) = OAuthProviderType(zusq)
	}
	if err != nil {
		return
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z OAuthProviderType) Msgsize() (s int) {
	s = msgp.StringPrefixSize + len(string(z))
	return
}

// DecodeMsg implements msgp.Decodable
func (z *PasswordConfiguration) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zvml uint32
	zvml, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zvml > 0 {
		zvml--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "min_length":
			z.MinLength, err = dc.ReadInt()
			if err != nil {
				return
			}
		case "uppercase_required":
			z.UppercaseRequired, err = dc.ReadBool()
			if err != nil {
				return
			}
		case "lowercase_required":
			z.LowercaseRequired, err = dc.ReadBool()
			if err != nil {
				return
			}
		case "digit_required":
			z.DigitRequired, err = dc.ReadBool()
			if err != nil {
				return
			}
		case "symbol_required":
			z.SymbolRequired, err = dc.ReadBool()
			if err != nil {
				return
			}
		case "minimum_guessable_level":
			z.MinimumGuessableLevel, err = dc.ReadInt()
			if err != nil {
				return
			}
		case "excluded_keywords":
			var zpyv uint32
			zpyv, err = dc.ReadArrayHeader()
			if err != nil {
				return
			}
			if cap(z.ExcludedKeywords) >= int(zpyv) {
				z.ExcludedKeywords = (z.ExcludedKeywords)[:zpyv]
			} else {
				z.ExcludedKeywords = make([]string, zpyv)
			}
			for zfgq := range z.ExcludedKeywords {
				z.ExcludedKeywords[zfgq], err = dc.ReadString()
				if err != nil {
					return
				}
			}
		case "history_size":
			z.HistorySize, err = dc.ReadInt()
			if err != nil {
				return
			}
		case "history_days":
			z.HistoryDays, err = dc.ReadInt()
			if err != nil {
				return
			}
		case "expiry_days":
			z.ExpiryDays, err = dc.ReadInt()
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
func (z *PasswordConfiguration) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 10
	// write "min_length"
	err = en.Append(0x8a, 0xaa, 0x6d, 0x69, 0x6e, 0x5f, 0x6c, 0x65, 0x6e, 0x67, 0x74, 0x68)
	if err != nil {
		return err
	}
	err = en.WriteInt(z.MinLength)
	if err != nil {
		return
	}
	// write "uppercase_required"
	err = en.Append(0xb2, 0x75, 0x70, 0x70, 0x65, 0x72, 0x63, 0x61, 0x73, 0x65, 0x5f, 0x72, 0x65, 0x71, 0x75, 0x69, 0x72, 0x65, 0x64)
	if err != nil {
		return err
	}
	err = en.WriteBool(z.UppercaseRequired)
	if err != nil {
		return
	}
	// write "lowercase_required"
	err = en.Append(0xb2, 0x6c, 0x6f, 0x77, 0x65, 0x72, 0x63, 0x61, 0x73, 0x65, 0x5f, 0x72, 0x65, 0x71, 0x75, 0x69, 0x72, 0x65, 0x64)
	if err != nil {
		return err
	}
	err = en.WriteBool(z.LowercaseRequired)
	if err != nil {
		return
	}
	// write "digit_required"
	err = en.Append(0xae, 0x64, 0x69, 0x67, 0x69, 0x74, 0x5f, 0x72, 0x65, 0x71, 0x75, 0x69, 0x72, 0x65, 0x64)
	if err != nil {
		return err
	}
	err = en.WriteBool(z.DigitRequired)
	if err != nil {
		return
	}
	// write "symbol_required"
	err = en.Append(0xaf, 0x73, 0x79, 0x6d, 0x62, 0x6f, 0x6c, 0x5f, 0x72, 0x65, 0x71, 0x75, 0x69, 0x72, 0x65, 0x64)
	if err != nil {
		return err
	}
	err = en.WriteBool(z.SymbolRequired)
	if err != nil {
		return
	}
	// write "minimum_guessable_level"
	err = en.Append(0xb7, 0x6d, 0x69, 0x6e, 0x69, 0x6d, 0x75, 0x6d, 0x5f, 0x67, 0x75, 0x65, 0x73, 0x73, 0x61, 0x62, 0x6c, 0x65, 0x5f, 0x6c, 0x65, 0x76, 0x65, 0x6c)
	if err != nil {
		return err
	}
	err = en.WriteInt(z.MinimumGuessableLevel)
	if err != nil {
		return
	}
	// write "excluded_keywords"
	err = en.Append(0xb1, 0x65, 0x78, 0x63, 0x6c, 0x75, 0x64, 0x65, 0x64, 0x5f, 0x6b, 0x65, 0x79, 0x77, 0x6f, 0x72, 0x64, 0x73)
	if err != nil {
		return err
	}
	err = en.WriteArrayHeader(uint32(len(z.ExcludedKeywords)))
	if err != nil {
		return
	}
	for zfgq := range z.ExcludedKeywords {
		err = en.WriteString(z.ExcludedKeywords[zfgq])
		if err != nil {
			return
		}
	}
	// write "history_size"
	err = en.Append(0xac, 0x68, 0x69, 0x73, 0x74, 0x6f, 0x72, 0x79, 0x5f, 0x73, 0x69, 0x7a, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteInt(z.HistorySize)
	if err != nil {
		return
	}
	// write "history_days"
	err = en.Append(0xac, 0x68, 0x69, 0x73, 0x74, 0x6f, 0x72, 0x79, 0x5f, 0x64, 0x61, 0x79, 0x73)
	if err != nil {
		return err
	}
	err = en.WriteInt(z.HistoryDays)
	if err != nil {
		return
	}
	// write "expiry_days"
	err = en.Append(0xab, 0x65, 0x78, 0x70, 0x69, 0x72, 0x79, 0x5f, 0x64, 0x61, 0x79, 0x73)
	if err != nil {
		return err
	}
	err = en.WriteInt(z.ExpiryDays)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *PasswordConfiguration) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 10
	// string "min_length"
	o = append(o, 0x8a, 0xaa, 0x6d, 0x69, 0x6e, 0x5f, 0x6c, 0x65, 0x6e, 0x67, 0x74, 0x68)
	o = msgp.AppendInt(o, z.MinLength)
	// string "uppercase_required"
	o = append(o, 0xb2, 0x75, 0x70, 0x70, 0x65, 0x72, 0x63, 0x61, 0x73, 0x65, 0x5f, 0x72, 0x65, 0x71, 0x75, 0x69, 0x72, 0x65, 0x64)
	o = msgp.AppendBool(o, z.UppercaseRequired)
	// string "lowercase_required"
	o = append(o, 0xb2, 0x6c, 0x6f, 0x77, 0x65, 0x72, 0x63, 0x61, 0x73, 0x65, 0x5f, 0x72, 0x65, 0x71, 0x75, 0x69, 0x72, 0x65, 0x64)
	o = msgp.AppendBool(o, z.LowercaseRequired)
	// string "digit_required"
	o = append(o, 0xae, 0x64, 0x69, 0x67, 0x69, 0x74, 0x5f, 0x72, 0x65, 0x71, 0x75, 0x69, 0x72, 0x65, 0x64)
	o = msgp.AppendBool(o, z.DigitRequired)
	// string "symbol_required"
	o = append(o, 0xaf, 0x73, 0x79, 0x6d, 0x62, 0x6f, 0x6c, 0x5f, 0x72, 0x65, 0x71, 0x75, 0x69, 0x72, 0x65, 0x64)
	o = msgp.AppendBool(o, z.SymbolRequired)
	// string "minimum_guessable_level"
	o = append(o, 0xb7, 0x6d, 0x69, 0x6e, 0x69, 0x6d, 0x75, 0x6d, 0x5f, 0x67, 0x75, 0x65, 0x73, 0x73, 0x61, 0x62, 0x6c, 0x65, 0x5f, 0x6c, 0x65, 0x76, 0x65, 0x6c)
	o = msgp.AppendInt(o, z.MinimumGuessableLevel)
	// string "excluded_keywords"
	o = append(o, 0xb1, 0x65, 0x78, 0x63, 0x6c, 0x75, 0x64, 0x65, 0x64, 0x5f, 0x6b, 0x65, 0x79, 0x77, 0x6f, 0x72, 0x64, 0x73)
	o = msgp.AppendArrayHeader(o, uint32(len(z.ExcludedKeywords)))
	for zfgq := range z.ExcludedKeywords {
		o = msgp.AppendString(o, z.ExcludedKeywords[zfgq])
	}
	// string "history_size"
	o = append(o, 0xac, 0x68, 0x69, 0x73, 0x74, 0x6f, 0x72, 0x79, 0x5f, 0x73, 0x69, 0x7a, 0x65)
	o = msgp.AppendInt(o, z.HistorySize)
	// string "history_days"
	o = append(o, 0xac, 0x68, 0x69, 0x73, 0x74, 0x6f, 0x72, 0x79, 0x5f, 0x64, 0x61, 0x79, 0x73)
	o = msgp.AppendInt(o, z.HistoryDays)
	// string "expiry_days"
	o = append(o, 0xab, 0x65, 0x78, 0x70, 0x69, 0x72, 0x79, 0x5f, 0x64, 0x61, 0x79, 0x73)
	o = msgp.AppendInt(o, z.ExpiryDays)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *PasswordConfiguration) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zlur uint32
	zlur, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zlur > 0 {
		zlur--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "min_length":
			z.MinLength, bts, err = msgp.ReadIntBytes(bts)
			if err != nil {
				return
			}
		case "uppercase_required":
			z.UppercaseRequired, bts, err = msgp.ReadBoolBytes(bts)
			if err != nil {
				return
			}
		case "lowercase_required":
			z.LowercaseRequired, bts, err = msgp.ReadBoolBytes(bts)
			if err != nil {
				return
			}
		case "digit_required":
			z.DigitRequired, bts, err = msgp.ReadBoolBytes(bts)
			if err != nil {
				return
			}
		case "symbol_required":
			z.SymbolRequired, bts, err = msgp.ReadBoolBytes(bts)
			if err != nil {
				return
			}
		case "minimum_guessable_level":
			z.MinimumGuessableLevel, bts, err = msgp.ReadIntBytes(bts)
			if err != nil {
				return
			}
		case "excluded_keywords":
			var zupi uint32
			zupi, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.ExcludedKeywords) >= int(zupi) {
				z.ExcludedKeywords = (z.ExcludedKeywords)[:zupi]
			} else {
				z.ExcludedKeywords = make([]string, zupi)
			}
			for zfgq := range z.ExcludedKeywords {
				z.ExcludedKeywords[zfgq], bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
			}
		case "history_size":
			z.HistorySize, bts, err = msgp.ReadIntBytes(bts)
			if err != nil {
				return
			}
		case "history_days":
			z.HistoryDays, bts, err = msgp.ReadIntBytes(bts)
			if err != nil {
				return
			}
		case "expiry_days":
			z.ExpiryDays, bts, err = msgp.ReadIntBytes(bts)
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
func (z *PasswordConfiguration) Msgsize() (s int) {
	s = 1 + 11 + msgp.IntSize + 19 + msgp.BoolSize + 19 + msgp.BoolSize + 15 + msgp.BoolSize + 16 + msgp.BoolSize + 24 + msgp.IntSize + 18 + msgp.ArrayHeaderSize
	for zfgq := range z.ExcludedKeywords {
		s += msgp.StringPrefixSize + len(z.ExcludedKeywords[zfgq])
	}
	s += 13 + msgp.IntSize + 13 + msgp.IntSize + 12 + msgp.IntSize
	return
}

// DecodeMsg implements msgp.Decodable
func (z *SMTPConfiguration) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zfvi uint32
	zfvi, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zfvi > 0 {
		zfvi--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "host":
			z.Host, err = dc.ReadString()
			if err != nil {
				return
			}
		case "port":
			z.Port, err = dc.ReadInt()
			if err != nil {
				return
			}
		case "mode":
			{
				var zzrg string
				zzrg, err = dc.ReadString()
				z.Mode = SMTPMode(zzrg)
			}
			if err != nil {
				return
			}
		case "login":
			z.Login, err = dc.ReadString()
			if err != nil {
				return
			}
		case "password":
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
	// write "host"
	err = en.Append(0x85, 0xa4, 0x68, 0x6f, 0x73, 0x74)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Host)
	if err != nil {
		return
	}
	// write "port"
	err = en.Append(0xa4, 0x70, 0x6f, 0x72, 0x74)
	if err != nil {
		return err
	}
	err = en.WriteInt(z.Port)
	if err != nil {
		return
	}
	// write "mode"
	err = en.Append(0xa4, 0x6d, 0x6f, 0x64, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteString(string(z.Mode))
	if err != nil {
		return
	}
	// write "login"
	err = en.Append(0xa5, 0x6c, 0x6f, 0x67, 0x69, 0x6e)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Login)
	if err != nil {
		return
	}
	// write "password"
	err = en.Append(0xa8, 0x70, 0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64)
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
	// string "host"
	o = append(o, 0x85, 0xa4, 0x68, 0x6f, 0x73, 0x74)
	o = msgp.AppendString(o, z.Host)
	// string "port"
	o = append(o, 0xa4, 0x70, 0x6f, 0x72, 0x74)
	o = msgp.AppendInt(o, z.Port)
	// string "mode"
	o = append(o, 0xa4, 0x6d, 0x6f, 0x64, 0x65)
	o = msgp.AppendString(o, string(z.Mode))
	// string "login"
	o = append(o, 0xa5, 0x6c, 0x6f, 0x67, 0x69, 0x6e)
	o = msgp.AppendString(o, z.Login)
	// string "password"
	o = append(o, 0xa8, 0x70, 0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64)
	o = msgp.AppendString(o, z.Password)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *SMTPConfiguration) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zbmy uint32
	zbmy, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zbmy > 0 {
		zbmy--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "host":
			z.Host, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "port":
			z.Port, bts, err = msgp.ReadIntBytes(bts)
			if err != nil {
				return
			}
		case "mode":
			{
				var zarl string
				zarl, bts, err = msgp.ReadStringBytes(bts)
				z.Mode = SMTPMode(zarl)
			}
			if err != nil {
				return
			}
		case "login":
			z.Login, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "password":
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
	s = 1 + 5 + msgp.StringPrefixSize + len(z.Host) + 5 + msgp.IntSize + 5 + msgp.StringPrefixSize + len(string(z.Mode)) + 6 + msgp.StringPrefixSize + len(z.Login) + 9 + msgp.StringPrefixSize + len(z.Password)
	return
}

// DecodeMsg implements msgp.Decodable
func (z *SMTPMode) DecodeMsg(dc *msgp.Reader) (err error) {
	{
		var zctz string
		zctz, err = dc.ReadString()
		(*z) = SMTPMode(zctz)
	}
	if err != nil {
		return
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z SMTPMode) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteString(string(z))
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z SMTPMode) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	o = msgp.AppendString(o, string(z))
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *SMTPMode) UnmarshalMsg(bts []byte) (o []byte, err error) {
	{
		var zljl string
		zljl, bts, err = msgp.ReadStringBytes(bts)
		(*z) = SMTPMode(zljl)
	}
	if err != nil {
		return
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z SMTPMode) Msgsize() (s int) {
	s = msgp.StringPrefixSize + len(string(z))
	return
}

// DecodeMsg implements msgp.Decodable
func (z *SSOConfiguration) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zziv uint32
	zziv, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zziv > 0 {
		zziv--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "custom_token":
			err = z.CustomToken.DecodeMsg(dc)
			if err != nil {
				return
			}
		case "oauth":
			err = z.OAuth.DecodeMsg(dc)
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
	// map header, size 2
	// write "custom_token"
	err = en.Append(0x82, 0xac, 0x63, 0x75, 0x73, 0x74, 0x6f, 0x6d, 0x5f, 0x74, 0x6f, 0x6b, 0x65, 0x6e)
	if err != nil {
		return err
	}
	err = z.CustomToken.EncodeMsg(en)
	if err != nil {
		return
	}
	// write "oauth"
	err = en.Append(0xa5, 0x6f, 0x61, 0x75, 0x74, 0x68)
	if err != nil {
		return err
	}
	err = z.OAuth.EncodeMsg(en)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *SSOConfiguration) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "custom_token"
	o = append(o, 0x82, 0xac, 0x63, 0x75, 0x73, 0x74, 0x6f, 0x6d, 0x5f, 0x74, 0x6f, 0x6b, 0x65, 0x6e)
	o, err = z.CustomToken.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "oauth"
	o = append(o, 0xa5, 0x6f, 0x61, 0x75, 0x74, 0x68)
	o, err = z.OAuth.MarshalMsg(o)
	if err != nil {
		return
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *SSOConfiguration) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zabj uint32
	zabj, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zabj > 0 {
		zabj--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "custom_token":
			bts, err = z.CustomToken.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "oauth":
			bts, err = z.OAuth.UnmarshalMsg(bts)
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
	s = 1 + 13 + z.CustomToken.Msgsize() + 6 + z.OAuth.Msgsize()
	return
}

// DecodeMsg implements msgp.Decodable
func (z *SessionCookieSameSite) DecodeMsg(dc *msgp.Reader) (err error) {
	{
		var zmlx string
		zmlx, err = dc.ReadString()
		(*z) = SessionCookieSameSite(zmlx)
	}
	if err != nil {
		return
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z SessionCookieSameSite) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteString(string(z))
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z SessionCookieSameSite) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	o = msgp.AppendString(o, string(z))
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *SessionCookieSameSite) UnmarshalMsg(bts []byte) (o []byte, err error) {
	{
		var zvbw string
		zvbw, bts, err = msgp.ReadStringBytes(bts)
		(*z) = SessionCookieSameSite(zvbw)
	}
	if err != nil {
		return
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z SessionCookieSameSite) Msgsize() (s int) {
	s = msgp.StringPrefixSize + len(string(z))
	return
}

// DecodeMsg implements msgp.Decodable
func (z *SessionTransportType) DecodeMsg(dc *msgp.Reader) (err error) {
	{
		var zgvb string
		zgvb, err = dc.ReadString()
		(*z) = SessionTransportType(zgvb)
	}
	if err != nil {
		return
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z SessionTransportType) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteString(string(z))
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z SessionTransportType) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	o = msgp.AppendString(o, string(z))
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *SessionTransportType) UnmarshalMsg(bts []byte) (o []byte, err error) {
	{
		var zqzg string
		zqzg, bts, err = msgp.ReadStringBytes(bts)
		(*z) = SessionTransportType(zqzg)
	}
	if err != nil {
		return
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z SessionTransportType) Msgsize() (s int) {
	s = msgp.StringPrefixSize + len(string(z))
	return
}

// DecodeMsg implements msgp.Decodable
func (z *TenantConfiguration) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zsdj uint32
	zsdj, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zsdj > 0 {
		zsdj--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "version":
			z.Version, err = dc.ReadString()
			if err != nil {
				return
			}
		case "app_id":
			z.AppID, err = dc.ReadString()
			if err != nil {
				return
			}
		case "app_name":
			z.AppName, err = dc.ReadString()
			if err != nil {
				return
			}
		case "app_config":
			err = z.AppConfig.DecodeMsg(dc)
			if err != nil {
				return
			}
		case "user_config":
			err = z.UserConfig.DecodeMsg(dc)
			if err != nil {
				return
			}
		case "hooks":
			var zsgp uint32
			zsgp, err = dc.ReadArrayHeader()
			if err != nil {
				return
			}
			if cap(z.Hooks) >= int(zsgp) {
				z.Hooks = (z.Hooks)[:zsgp]
			} else {
				z.Hooks = make([]Hook, zsgp)
			}
			for zexy := range z.Hooks {
				var zngc uint32
				zngc, err = dc.ReadMapHeader()
				if err != nil {
					return
				}
				for zngc > 0 {
					zngc--
					field, err = dc.ReadMapKeyPtr()
					if err != nil {
						return
					}
					switch msgp.UnsafeString(field) {
					case "event":
						z.Hooks[zexy].Event, err = dc.ReadString()
						if err != nil {
							return
						}
					case "url":
						z.Hooks[zexy].URL, err = dc.ReadString()
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
			}
		case "deployment_routes":
			var zwfl uint32
			zwfl, err = dc.ReadArrayHeader()
			if err != nil {
				return
			}
			if cap(z.DeploymentRoutes) >= int(zwfl) {
				z.DeploymentRoutes = (z.DeploymentRoutes)[:zwfl]
			} else {
				z.DeploymentRoutes = make([]DeploymentRoute, zwfl)
			}
			for zakb := range z.DeploymentRoutes {
				err = z.DeploymentRoutes[zakb].DecodeMsg(dc)
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
	// map header, size 7
	// write "version"
	err = en.Append(0x87, 0xa7, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Version)
	if err != nil {
		return
	}
	// write "app_id"
	err = en.Append(0xa6, 0x61, 0x70, 0x70, 0x5f, 0x69, 0x64)
	if err != nil {
		return err
	}
	err = en.WriteString(z.AppID)
	if err != nil {
		return
	}
	// write "app_name"
	err = en.Append(0xa8, 0x61, 0x70, 0x70, 0x5f, 0x6e, 0x61, 0x6d, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteString(z.AppName)
	if err != nil {
		return
	}
	// write "app_config"
	err = en.Append(0xaa, 0x61, 0x70, 0x70, 0x5f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67)
	if err != nil {
		return err
	}
	err = z.AppConfig.EncodeMsg(en)
	if err != nil {
		return
	}
	// write "user_config"
	err = en.Append(0xab, 0x75, 0x73, 0x65, 0x72, 0x5f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67)
	if err != nil {
		return err
	}
	err = z.UserConfig.EncodeMsg(en)
	if err != nil {
		return
	}
	// write "hooks"
	err = en.Append(0xa5, 0x68, 0x6f, 0x6f, 0x6b, 0x73)
	if err != nil {
		return err
	}
	err = en.WriteArrayHeader(uint32(len(z.Hooks)))
	if err != nil {
		return
	}
	for zexy := range z.Hooks {
		// map header, size 2
		// write "event"
		err = en.Append(0x82, 0xa5, 0x65, 0x76, 0x65, 0x6e, 0x74)
		if err != nil {
			return err
		}
		err = en.WriteString(z.Hooks[zexy].Event)
		if err != nil {
			return
		}
		// write "url"
		err = en.Append(0xa3, 0x75, 0x72, 0x6c)
		if err != nil {
			return err
		}
		err = en.WriteString(z.Hooks[zexy].URL)
		if err != nil {
			return
		}
	}
	// write "deployment_routes"
	err = en.Append(0xb1, 0x64, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x5f, 0x72, 0x6f, 0x75, 0x74, 0x65, 0x73)
	if err != nil {
		return err
	}
	err = en.WriteArrayHeader(uint32(len(z.DeploymentRoutes)))
	if err != nil {
		return
	}
	for zakb := range z.DeploymentRoutes {
		err = z.DeploymentRoutes[zakb].EncodeMsg(en)
		if err != nil {
			return
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *TenantConfiguration) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 7
	// string "version"
	o = append(o, 0x87, 0xa7, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e)
	o = msgp.AppendString(o, z.Version)
	// string "app_id"
	o = append(o, 0xa6, 0x61, 0x70, 0x70, 0x5f, 0x69, 0x64)
	o = msgp.AppendString(o, z.AppID)
	// string "app_name"
	o = append(o, 0xa8, 0x61, 0x70, 0x70, 0x5f, 0x6e, 0x61, 0x6d, 0x65)
	o = msgp.AppendString(o, z.AppName)
	// string "app_config"
	o = append(o, 0xaa, 0x61, 0x70, 0x70, 0x5f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67)
	o, err = z.AppConfig.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "user_config"
	o = append(o, 0xab, 0x75, 0x73, 0x65, 0x72, 0x5f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67)
	o, err = z.UserConfig.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "hooks"
	o = append(o, 0xa5, 0x68, 0x6f, 0x6f, 0x6b, 0x73)
	o = msgp.AppendArrayHeader(o, uint32(len(z.Hooks)))
	for zexy := range z.Hooks {
		// map header, size 2
		// string "event"
		o = append(o, 0x82, 0xa5, 0x65, 0x76, 0x65, 0x6e, 0x74)
		o = msgp.AppendString(o, z.Hooks[zexy].Event)
		// string "url"
		o = append(o, 0xa3, 0x75, 0x72, 0x6c)
		o = msgp.AppendString(o, z.Hooks[zexy].URL)
	}
	// string "deployment_routes"
	o = append(o, 0xb1, 0x64, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x5f, 0x72, 0x6f, 0x75, 0x74, 0x65, 0x73)
	o = msgp.AppendArrayHeader(o, uint32(len(z.DeploymentRoutes)))
	for zakb := range z.DeploymentRoutes {
		o, err = z.DeploymentRoutes[zakb].MarshalMsg(o)
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
	var zdif uint32
	zdif, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zdif > 0 {
		zdif--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "version":
			z.Version, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "app_id":
			z.AppID, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "app_name":
			z.AppName, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "app_config":
			bts, err = z.AppConfig.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "user_config":
			bts, err = z.UserConfig.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "hooks":
			var zibu uint32
			zibu, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.Hooks) >= int(zibu) {
				z.Hooks = (z.Hooks)[:zibu]
			} else {
				z.Hooks = make([]Hook, zibu)
			}
			for zexy := range z.Hooks {
				var zuff uint32
				zuff, bts, err = msgp.ReadMapHeaderBytes(bts)
				if err != nil {
					return
				}
				for zuff > 0 {
					zuff--
					field, bts, err = msgp.ReadMapKeyZC(bts)
					if err != nil {
						return
					}
					switch msgp.UnsafeString(field) {
					case "event":
						z.Hooks[zexy].Event, bts, err = msgp.ReadStringBytes(bts)
						if err != nil {
							return
						}
					case "url":
						z.Hooks[zexy].URL, bts, err = msgp.ReadStringBytes(bts)
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
			}
		case "deployment_routes":
			var zmow uint32
			zmow, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.DeploymentRoutes) >= int(zmow) {
				z.DeploymentRoutes = (z.DeploymentRoutes)[:zmow]
			} else {
				z.DeploymentRoutes = make([]DeploymentRoute, zmow)
			}
			for zakb := range z.DeploymentRoutes {
				bts, err = z.DeploymentRoutes[zakb].UnmarshalMsg(bts)
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
	s = 1 + 8 + msgp.StringPrefixSize + len(z.Version) + 7 + msgp.StringPrefixSize + len(z.AppID) + 9 + msgp.StringPrefixSize + len(z.AppName) + 11 + z.AppConfig.Msgsize() + 12 + z.UserConfig.Msgsize() + 6 + msgp.ArrayHeaderSize
	for zexy := range z.Hooks {
		s += 1 + 6 + msgp.StringPrefixSize + len(z.Hooks[zexy].Event) + 4 + msgp.StringPrefixSize + len(z.Hooks[zexy].URL)
	}
	s += 18 + msgp.ArrayHeaderSize
	for zakb := range z.DeploymentRoutes {
		s += z.DeploymentRoutes[zakb].Msgsize()
	}
	return
}

// DecodeMsg implements msgp.Decodable
func (z *TwilioConfiguration) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zdit uint32
	zdit, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zdit > 0 {
		zdit--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "account_sid":
			z.AccountSID, err = dc.ReadString()
			if err != nil {
				return
			}
		case "auth_token":
			z.AuthToken, err = dc.ReadString()
			if err != nil {
				return
			}
		case "from":
			z.From, err = dc.ReadString()
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
func (z TwilioConfiguration) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 3
	// write "account_sid"
	err = en.Append(0x83, 0xab, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x5f, 0x73, 0x69, 0x64)
	if err != nil {
		return err
	}
	err = en.WriteString(z.AccountSID)
	if err != nil {
		return
	}
	// write "auth_token"
	err = en.Append(0xaa, 0x61, 0x75, 0x74, 0x68, 0x5f, 0x74, 0x6f, 0x6b, 0x65, 0x6e)
	if err != nil {
		return err
	}
	err = en.WriteString(z.AuthToken)
	if err != nil {
		return
	}
	// write "from"
	err = en.Append(0xa4, 0x66, 0x72, 0x6f, 0x6d)
	if err != nil {
		return err
	}
	err = en.WriteString(z.From)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z TwilioConfiguration) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 3
	// string "account_sid"
	o = append(o, 0x83, 0xab, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x5f, 0x73, 0x69, 0x64)
	o = msgp.AppendString(o, z.AccountSID)
	// string "auth_token"
	o = append(o, 0xaa, 0x61, 0x75, 0x74, 0x68, 0x5f, 0x74, 0x6f, 0x6b, 0x65, 0x6e)
	o = msgp.AppendString(o, z.AuthToken)
	// string "from"
	o = append(o, 0xa4, 0x66, 0x72, 0x6f, 0x6d)
	o = msgp.AppendString(o, z.From)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *TwilioConfiguration) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zslz uint32
	zslz, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zslz > 0 {
		zslz--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "account_sid":
			z.AccountSID, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "auth_token":
			z.AuthToken, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "from":
			z.From, bts, err = msgp.ReadStringBytes(bts)
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
func (z TwilioConfiguration) Msgsize() (s int) {
	s = 1 + 12 + msgp.StringPrefixSize + len(z.AccountSID) + 11 + msgp.StringPrefixSize + len(z.AuthToken) + 5 + msgp.StringPrefixSize + len(z.From)
	return
}

// DecodeMsg implements msgp.Decodable
func (z *UserAuditConfiguration) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zoqj uint32
	zoqj, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zoqj > 0 {
		zoqj--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "enabled":
			z.Enabled, err = dc.ReadBool()
			if err != nil {
				return
			}
		case "trail_handler_url":
			z.TrailHandlerURL, err = dc.ReadString()
			if err != nil {
				return
			}
		case "password":
			err = z.Password.DecodeMsg(dc)
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
func (z *UserAuditConfiguration) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 3
	// write "enabled"
	err = en.Append(0x83, 0xa7, 0x65, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64)
	if err != nil {
		return err
	}
	err = en.WriteBool(z.Enabled)
	if err != nil {
		return
	}
	// write "trail_handler_url"
	err = en.Append(0xb1, 0x74, 0x72, 0x61, 0x69, 0x6c, 0x5f, 0x68, 0x61, 0x6e, 0x64, 0x6c, 0x65, 0x72, 0x5f, 0x75, 0x72, 0x6c)
	if err != nil {
		return err
	}
	err = en.WriteString(z.TrailHandlerURL)
	if err != nil {
		return
	}
	// write "password"
	err = en.Append(0xa8, 0x70, 0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64)
	if err != nil {
		return err
	}
	err = z.Password.EncodeMsg(en)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *UserAuditConfiguration) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 3
	// string "enabled"
	o = append(o, 0x83, 0xa7, 0x65, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64)
	o = msgp.AppendBool(o, z.Enabled)
	// string "trail_handler_url"
	o = append(o, 0xb1, 0x74, 0x72, 0x61, 0x69, 0x6c, 0x5f, 0x68, 0x61, 0x6e, 0x64, 0x6c, 0x65, 0x72, 0x5f, 0x75, 0x72, 0x6c)
	o = msgp.AppendString(o, z.TrailHandlerURL)
	// string "password"
	o = append(o, 0xa8, 0x70, 0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64)
	o, err = z.Password.MarshalMsg(o)
	if err != nil {
		return
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *UserAuditConfiguration) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zmqr uint32
	zmqr, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zmqr > 0 {
		zmqr--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "enabled":
			z.Enabled, bts, err = msgp.ReadBoolBytes(bts)
			if err != nil {
				return
			}
		case "trail_handler_url":
			z.TrailHandlerURL, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "password":
			bts, err = z.Password.UnmarshalMsg(bts)
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
func (z *UserAuditConfiguration) Msgsize() (s int) {
	s = 1 + 8 + msgp.BoolSize + 18 + msgp.StringPrefixSize + len(z.TrailHandlerURL) + 9 + z.Password.Msgsize()
	return
}

// DecodeMsg implements msgp.Decodable
func (z *UserConfiguration) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var ziyx uint32
	ziyx, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for ziyx > 0 {
		ziyx--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "clients":
			var zyes uint32
			zyes, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			if z.Clients == nil && zyes > 0 {
				z.Clients = make(map[string]APIClientConfiguration, zyes)
			} else if len(z.Clients) > 0 {
				for key, _ := range z.Clients {
					delete(z.Clients, key)
				}
			}
			for zyes > 0 {
				zyes--
				var ztic string
				var ztoj APIClientConfiguration
				ztic, err = dc.ReadString()
				if err != nil {
					return
				}
				err = ztoj.DecodeMsg(dc)
				if err != nil {
					return
				}
				z.Clients[ztic] = ztoj
			}
		case "master_key":
			z.MasterKey, err = dc.ReadString()
			if err != nil {
				return
			}
		case "url_prefix":
			z.URLPrefix, err = dc.ReadString()
			if err != nil {
				return
			}
		case "cors":
			var zxzy uint32
			zxzy, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			for zxzy > 0 {
				zxzy--
				field, err = dc.ReadMapKeyPtr()
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "origin":
					z.CORS.Origin, err = dc.ReadString()
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
		case "auth":
			err = z.Auth.DecodeMsg(dc)
			if err != nil {
				return
			}
		case "mfa":
			err = z.MFA.DecodeMsg(dc)
			if err != nil {
				return
			}
		case "user_audit":
			var zfro uint32
			zfro, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			for zfro > 0 {
				zfro--
				field, err = dc.ReadMapKeyPtr()
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "enabled":
					z.UserAudit.Enabled, err = dc.ReadBool()
					if err != nil {
						return
					}
				case "trail_handler_url":
					z.UserAudit.TrailHandlerURL, err = dc.ReadString()
					if err != nil {
						return
					}
				case "password":
					err = z.UserAudit.Password.DecodeMsg(dc)
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
		case "forgot_password":
			err = z.ForgotPassword.DecodeMsg(dc)
			if err != nil {
				return
			}
		case "welcome_email":
			err = z.WelcomeEmail.DecodeMsg(dc)
			if err != nil {
				return
			}
		case "sso":
			var zrod uint32
			zrod, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			for zrod > 0 {
				zrod--
				field, err = dc.ReadMapKeyPtr()
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "custom_token":
					err = z.SSO.CustomToken.DecodeMsg(dc)
					if err != nil {
						return
					}
				case "oauth":
					err = z.SSO.OAuth.DecodeMsg(dc)
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
		case "user_verification":
			err = z.UserVerification.DecodeMsg(dc)
			if err != nil {
				return
			}
		case "hook":
			var zmbn uint32
			zmbn, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			for zmbn > 0 {
				zmbn--
				field, err = dc.ReadMapKeyPtr()
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "secret":
					z.Hook.Secret, err = dc.ReadString()
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
func (z *UserConfiguration) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 12
	// write "clients"
	err = en.Append(0x8c, 0xa7, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x73)
	if err != nil {
		return err
	}
	err = en.WriteMapHeader(uint32(len(z.Clients)))
	if err != nil {
		return
	}
	for ztic, ztoj := range z.Clients {
		err = en.WriteString(ztic)
		if err != nil {
			return
		}
		err = ztoj.EncodeMsg(en)
		if err != nil {
			return
		}
	}
	// write "master_key"
	err = en.Append(0xaa, 0x6d, 0x61, 0x73, 0x74, 0x65, 0x72, 0x5f, 0x6b, 0x65, 0x79)
	if err != nil {
		return err
	}
	err = en.WriteString(z.MasterKey)
	if err != nil {
		return
	}
	// write "url_prefix"
	err = en.Append(0xaa, 0x75, 0x72, 0x6c, 0x5f, 0x70, 0x72, 0x65, 0x66, 0x69, 0x78)
	if err != nil {
		return err
	}
	err = en.WriteString(z.URLPrefix)
	if err != nil {
		return
	}
	// write "cors"
	// map header, size 1
	// write "origin"
	err = en.Append(0xa4, 0x63, 0x6f, 0x72, 0x73, 0x81, 0xa6, 0x6f, 0x72, 0x69, 0x67, 0x69, 0x6e)
	if err != nil {
		return err
	}
	err = en.WriteString(z.CORS.Origin)
	if err != nil {
		return
	}
	// write "auth"
	err = en.Append(0xa4, 0x61, 0x75, 0x74, 0x68)
	if err != nil {
		return err
	}
	err = z.Auth.EncodeMsg(en)
	if err != nil {
		return
	}
	// write "mfa"
	err = en.Append(0xa3, 0x6d, 0x66, 0x61)
	if err != nil {
		return err
	}
	err = z.MFA.EncodeMsg(en)
	if err != nil {
		return
	}
	// write "user_audit"
	// map header, size 3
	// write "enabled"
	err = en.Append(0xaa, 0x75, 0x73, 0x65, 0x72, 0x5f, 0x61, 0x75, 0x64, 0x69, 0x74, 0x83, 0xa7, 0x65, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64)
	if err != nil {
		return err
	}
	err = en.WriteBool(z.UserAudit.Enabled)
	if err != nil {
		return
	}
	// write "trail_handler_url"
	err = en.Append(0xb1, 0x74, 0x72, 0x61, 0x69, 0x6c, 0x5f, 0x68, 0x61, 0x6e, 0x64, 0x6c, 0x65, 0x72, 0x5f, 0x75, 0x72, 0x6c)
	if err != nil {
		return err
	}
	err = en.WriteString(z.UserAudit.TrailHandlerURL)
	if err != nil {
		return
	}
	// write "password"
	err = en.Append(0xa8, 0x70, 0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64)
	if err != nil {
		return err
	}
	err = z.UserAudit.Password.EncodeMsg(en)
	if err != nil {
		return
	}
	// write "forgot_password"
	err = en.Append(0xaf, 0x66, 0x6f, 0x72, 0x67, 0x6f, 0x74, 0x5f, 0x70, 0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64)
	if err != nil {
		return err
	}
	err = z.ForgotPassword.EncodeMsg(en)
	if err != nil {
		return
	}
	// write "welcome_email"
	err = en.Append(0xad, 0x77, 0x65, 0x6c, 0x63, 0x6f, 0x6d, 0x65, 0x5f, 0x65, 0x6d, 0x61, 0x69, 0x6c)
	if err != nil {
		return err
	}
	err = z.WelcomeEmail.EncodeMsg(en)
	if err != nil {
		return
	}
	// write "sso"
	// map header, size 2
	// write "custom_token"
	err = en.Append(0xa3, 0x73, 0x73, 0x6f, 0x82, 0xac, 0x63, 0x75, 0x73, 0x74, 0x6f, 0x6d, 0x5f, 0x74, 0x6f, 0x6b, 0x65, 0x6e)
	if err != nil {
		return err
	}
	err = z.SSO.CustomToken.EncodeMsg(en)
	if err != nil {
		return
	}
	// write "oauth"
	err = en.Append(0xa5, 0x6f, 0x61, 0x75, 0x74, 0x68)
	if err != nil {
		return err
	}
	err = z.SSO.OAuth.EncodeMsg(en)
	if err != nil {
		return
	}
	// write "user_verification"
	err = en.Append(0xb1, 0x75, 0x73, 0x65, 0x72, 0x5f, 0x76, 0x65, 0x72, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e)
	if err != nil {
		return err
	}
	err = z.UserVerification.EncodeMsg(en)
	if err != nil {
		return
	}
	// write "hook"
	// map header, size 1
	// write "secret"
	err = en.Append(0xa4, 0x68, 0x6f, 0x6f, 0x6b, 0x81, 0xa6, 0x73, 0x65, 0x63, 0x72, 0x65, 0x74)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Hook.Secret)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *UserConfiguration) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 12
	// string "clients"
	o = append(o, 0x8c, 0xa7, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x73)
	o = msgp.AppendMapHeader(o, uint32(len(z.Clients)))
	for ztic, ztoj := range z.Clients {
		o = msgp.AppendString(o, ztic)
		o, err = ztoj.MarshalMsg(o)
		if err != nil {
			return
		}
	}
	// string "master_key"
	o = append(o, 0xaa, 0x6d, 0x61, 0x73, 0x74, 0x65, 0x72, 0x5f, 0x6b, 0x65, 0x79)
	o = msgp.AppendString(o, z.MasterKey)
	// string "url_prefix"
	o = append(o, 0xaa, 0x75, 0x72, 0x6c, 0x5f, 0x70, 0x72, 0x65, 0x66, 0x69, 0x78)
	o = msgp.AppendString(o, z.URLPrefix)
	// string "cors"
	// map header, size 1
	// string "origin"
	o = append(o, 0xa4, 0x63, 0x6f, 0x72, 0x73, 0x81, 0xa6, 0x6f, 0x72, 0x69, 0x67, 0x69, 0x6e)
	o = msgp.AppendString(o, z.CORS.Origin)
	// string "auth"
	o = append(o, 0xa4, 0x61, 0x75, 0x74, 0x68)
	o, err = z.Auth.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "mfa"
	o = append(o, 0xa3, 0x6d, 0x66, 0x61)
	o, err = z.MFA.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "user_audit"
	// map header, size 3
	// string "enabled"
	o = append(o, 0xaa, 0x75, 0x73, 0x65, 0x72, 0x5f, 0x61, 0x75, 0x64, 0x69, 0x74, 0x83, 0xa7, 0x65, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64)
	o = msgp.AppendBool(o, z.UserAudit.Enabled)
	// string "trail_handler_url"
	o = append(o, 0xb1, 0x74, 0x72, 0x61, 0x69, 0x6c, 0x5f, 0x68, 0x61, 0x6e, 0x64, 0x6c, 0x65, 0x72, 0x5f, 0x75, 0x72, 0x6c)
	o = msgp.AppendString(o, z.UserAudit.TrailHandlerURL)
	// string "password"
	o = append(o, 0xa8, 0x70, 0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64)
	o, err = z.UserAudit.Password.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "forgot_password"
	o = append(o, 0xaf, 0x66, 0x6f, 0x72, 0x67, 0x6f, 0x74, 0x5f, 0x70, 0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64)
	o, err = z.ForgotPassword.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "welcome_email"
	o = append(o, 0xad, 0x77, 0x65, 0x6c, 0x63, 0x6f, 0x6d, 0x65, 0x5f, 0x65, 0x6d, 0x61, 0x69, 0x6c)
	o, err = z.WelcomeEmail.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "sso"
	// map header, size 2
	// string "custom_token"
	o = append(o, 0xa3, 0x73, 0x73, 0x6f, 0x82, 0xac, 0x63, 0x75, 0x73, 0x74, 0x6f, 0x6d, 0x5f, 0x74, 0x6f, 0x6b, 0x65, 0x6e)
	o, err = z.SSO.CustomToken.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "oauth"
	o = append(o, 0xa5, 0x6f, 0x61, 0x75, 0x74, 0x68)
	o, err = z.SSO.OAuth.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "user_verification"
	o = append(o, 0xb1, 0x75, 0x73, 0x65, 0x72, 0x5f, 0x76, 0x65, 0x72, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e)
	o, err = z.UserVerification.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "hook"
	// map header, size 1
	// string "secret"
	o = append(o, 0xa4, 0x68, 0x6f, 0x6f, 0x6b, 0x81, 0xa6, 0x73, 0x65, 0x63, 0x72, 0x65, 0x74)
	o = msgp.AppendString(o, z.Hook.Secret)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *UserConfiguration) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zdrz uint32
	zdrz, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zdrz > 0 {
		zdrz--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "clients":
			var znpn uint32
			znpn, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			if z.Clients == nil && znpn > 0 {
				z.Clients = make(map[string]APIClientConfiguration, znpn)
			} else if len(z.Clients) > 0 {
				for key, _ := range z.Clients {
					delete(z.Clients, key)
				}
			}
			for znpn > 0 {
				var ztic string
				var ztoj APIClientConfiguration
				znpn--
				ztic, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
				bts, err = ztoj.UnmarshalMsg(bts)
				if err != nil {
					return
				}
				z.Clients[ztic] = ztoj
			}
		case "master_key":
			z.MasterKey, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "url_prefix":
			z.URLPrefix, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "cors":
			var zrwc uint32
			zrwc, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			for zrwc > 0 {
				zrwc--
				field, bts, err = msgp.ReadMapKeyZC(bts)
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "origin":
					z.CORS.Origin, bts, err = msgp.ReadStringBytes(bts)
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
		case "auth":
			bts, err = z.Auth.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "mfa":
			bts, err = z.MFA.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "user_audit":
			var zjpm uint32
			zjpm, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			for zjpm > 0 {
				zjpm--
				field, bts, err = msgp.ReadMapKeyZC(bts)
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "enabled":
					z.UserAudit.Enabled, bts, err = msgp.ReadBoolBytes(bts)
					if err != nil {
						return
					}
				case "trail_handler_url":
					z.UserAudit.TrailHandlerURL, bts, err = msgp.ReadStringBytes(bts)
					if err != nil {
						return
					}
				case "password":
					bts, err = z.UserAudit.Password.UnmarshalMsg(bts)
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
		case "forgot_password":
			bts, err = z.ForgotPassword.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "welcome_email":
			bts, err = z.WelcomeEmail.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "sso":
			var zhdt uint32
			zhdt, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			for zhdt > 0 {
				zhdt--
				field, bts, err = msgp.ReadMapKeyZC(bts)
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "custom_token":
					bts, err = z.SSO.CustomToken.UnmarshalMsg(bts)
					if err != nil {
						return
					}
				case "oauth":
					bts, err = z.SSO.OAuth.UnmarshalMsg(bts)
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
		case "user_verification":
			bts, err = z.UserVerification.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "hook":
			var zjmh uint32
			zjmh, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			for zjmh > 0 {
				zjmh--
				field, bts, err = msgp.ReadMapKeyZC(bts)
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "secret":
					z.Hook.Secret, bts, err = msgp.ReadStringBytes(bts)
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
func (z *UserConfiguration) Msgsize() (s int) {
	s = 1 + 8 + msgp.MapHeaderSize
	if z.Clients != nil {
		for ztic, ztoj := range z.Clients {
			_ = ztoj
			s += msgp.StringPrefixSize + len(ztic) + ztoj.Msgsize()
		}
	}
	s += 11 + msgp.StringPrefixSize + len(z.MasterKey) + 11 + msgp.StringPrefixSize + len(z.URLPrefix) + 5 + 1 + 7 + msgp.StringPrefixSize + len(z.CORS.Origin) + 5 + z.Auth.Msgsize() + 4 + z.MFA.Msgsize() + 11 + 1 + 8 + msgp.BoolSize + 18 + msgp.StringPrefixSize + len(z.UserAudit.TrailHandlerURL) + 9 + z.UserAudit.Password.Msgsize() + 16 + z.ForgotPassword.Msgsize() + 14 + z.WelcomeEmail.Msgsize() + 4 + 1 + 13 + z.SSO.CustomToken.Msgsize() + 6 + z.SSO.OAuth.Msgsize() + 18 + z.UserVerification.Msgsize() + 5 + 1 + 7 + msgp.StringPrefixSize + len(z.Hook.Secret)
	return
}

// DecodeMsg implements msgp.Decodable
func (z *UserVerificationCodeFormat) DecodeMsg(dc *msgp.Reader) (err error) {
	{
		var zayo string
		zayo, err = dc.ReadString()
		(*z) = UserVerificationCodeFormat(zayo)
	}
	if err != nil {
		return
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z UserVerificationCodeFormat) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteString(string(z))
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z UserVerificationCodeFormat) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	o = msgp.AppendString(o, string(z))
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *UserVerificationCodeFormat) UnmarshalMsg(bts []byte) (o []byte, err error) {
	{
		var zrsu string
		zrsu, bts, err = msgp.ReadStringBytes(bts)
		(*z) = UserVerificationCodeFormat(zrsu)
	}
	if err != nil {
		return
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z UserVerificationCodeFormat) Msgsize() (s int) {
	s = msgp.StringPrefixSize + len(string(z))
	return
}

// DecodeMsg implements msgp.Decodable
func (z *UserVerificationConfiguration) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zvgz uint32
	zvgz, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zvgz > 0 {
		zvgz--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "url_prefix":
			z.URLPrefix, err = dc.ReadString()
			if err != nil {
				return
			}
		case "auto_send_on_signup":
			z.AutoSendOnSignup, err = dc.ReadBool()
			if err != nil {
				return
			}
		case "criteria":
			{
				var zhbk string
				zhbk, err = dc.ReadString()
				z.Criteria = UserVerificationCriteria(zhbk)
			}
			if err != nil {
				return
			}
		case "error_redirect":
			z.ErrorRedirect, err = dc.ReadString()
			if err != nil {
				return
			}
		case "error_html_url":
			z.ErrorHTMLURL, err = dc.ReadString()
			if err != nil {
				return
			}
		case "login_id_keys":
			var zmyy uint32
			zmyy, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			if z.LoginIDKeys == nil && zmyy > 0 {
				z.LoginIDKeys = make(map[string]UserVerificationKeyConfiguration, zmyy)
			} else if len(z.LoginIDKeys) > 0 {
				for key, _ := range z.LoginIDKeys {
					delete(z.LoginIDKeys, key)
				}
			}
			for zmyy > 0 {
				zmyy--
				var zfum string
				var zaps UserVerificationKeyConfiguration
				zfum, err = dc.ReadString()
				if err != nil {
					return
				}
				err = zaps.DecodeMsg(dc)
				if err != nil {
					return
				}
				z.LoginIDKeys[zfum] = zaps
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
func (z *UserVerificationConfiguration) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 6
	// write "url_prefix"
	err = en.Append(0x86, 0xaa, 0x75, 0x72, 0x6c, 0x5f, 0x70, 0x72, 0x65, 0x66, 0x69, 0x78)
	if err != nil {
		return err
	}
	err = en.WriteString(z.URLPrefix)
	if err != nil {
		return
	}
	// write "auto_send_on_signup"
	err = en.Append(0xb3, 0x61, 0x75, 0x74, 0x6f, 0x5f, 0x73, 0x65, 0x6e, 0x64, 0x5f, 0x6f, 0x6e, 0x5f, 0x73, 0x69, 0x67, 0x6e, 0x75, 0x70)
	if err != nil {
		return err
	}
	err = en.WriteBool(z.AutoSendOnSignup)
	if err != nil {
		return
	}
	// write "criteria"
	err = en.Append(0xa8, 0x63, 0x72, 0x69, 0x74, 0x65, 0x72, 0x69, 0x61)
	if err != nil {
		return err
	}
	err = en.WriteString(string(z.Criteria))
	if err != nil {
		return
	}
	// write "error_redirect"
	err = en.Append(0xae, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x5f, 0x72, 0x65, 0x64, 0x69, 0x72, 0x65, 0x63, 0x74)
	if err != nil {
		return err
	}
	err = en.WriteString(z.ErrorRedirect)
	if err != nil {
		return
	}
	// write "error_html_url"
	err = en.Append(0xae, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x5f, 0x68, 0x74, 0x6d, 0x6c, 0x5f, 0x75, 0x72, 0x6c)
	if err != nil {
		return err
	}
	err = en.WriteString(z.ErrorHTMLURL)
	if err != nil {
		return
	}
	// write "login_id_keys"
	err = en.Append(0xad, 0x6c, 0x6f, 0x67, 0x69, 0x6e, 0x5f, 0x69, 0x64, 0x5f, 0x6b, 0x65, 0x79, 0x73)
	if err != nil {
		return err
	}
	err = en.WriteMapHeader(uint32(len(z.LoginIDKeys)))
	if err != nil {
		return
	}
	for zfum, zaps := range z.LoginIDKeys {
		err = en.WriteString(zfum)
		if err != nil {
			return
		}
		err = zaps.EncodeMsg(en)
		if err != nil {
			return
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *UserVerificationConfiguration) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 6
	// string "url_prefix"
	o = append(o, 0x86, 0xaa, 0x75, 0x72, 0x6c, 0x5f, 0x70, 0x72, 0x65, 0x66, 0x69, 0x78)
	o = msgp.AppendString(o, z.URLPrefix)
	// string "auto_send_on_signup"
	o = append(o, 0xb3, 0x61, 0x75, 0x74, 0x6f, 0x5f, 0x73, 0x65, 0x6e, 0x64, 0x5f, 0x6f, 0x6e, 0x5f, 0x73, 0x69, 0x67, 0x6e, 0x75, 0x70)
	o = msgp.AppendBool(o, z.AutoSendOnSignup)
	// string "criteria"
	o = append(o, 0xa8, 0x63, 0x72, 0x69, 0x74, 0x65, 0x72, 0x69, 0x61)
	o = msgp.AppendString(o, string(z.Criteria))
	// string "error_redirect"
	o = append(o, 0xae, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x5f, 0x72, 0x65, 0x64, 0x69, 0x72, 0x65, 0x63, 0x74)
	o = msgp.AppendString(o, z.ErrorRedirect)
	// string "error_html_url"
	o = append(o, 0xae, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x5f, 0x68, 0x74, 0x6d, 0x6c, 0x5f, 0x75, 0x72, 0x6c)
	o = msgp.AppendString(o, z.ErrorHTMLURL)
	// string "login_id_keys"
	o = append(o, 0xad, 0x6c, 0x6f, 0x67, 0x69, 0x6e, 0x5f, 0x69, 0x64, 0x5f, 0x6b, 0x65, 0x79, 0x73)
	o = msgp.AppendMapHeader(o, uint32(len(z.LoginIDKeys)))
	for zfum, zaps := range z.LoginIDKeys {
		o = msgp.AppendString(o, zfum)
		o, err = zaps.MarshalMsg(o)
		if err != nil {
			return
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *UserVerificationConfiguration) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var ztej uint32
	ztej, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for ztej > 0 {
		ztej--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "url_prefix":
			z.URLPrefix, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "auto_send_on_signup":
			z.AutoSendOnSignup, bts, err = msgp.ReadBoolBytes(bts)
			if err != nil {
				return
			}
		case "criteria":
			{
				var zvgw string
				zvgw, bts, err = msgp.ReadStringBytes(bts)
				z.Criteria = UserVerificationCriteria(zvgw)
			}
			if err != nil {
				return
			}
		case "error_redirect":
			z.ErrorRedirect, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "error_html_url":
			z.ErrorHTMLURL, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "login_id_keys":
			var zffb uint32
			zffb, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			if z.LoginIDKeys == nil && zffb > 0 {
				z.LoginIDKeys = make(map[string]UserVerificationKeyConfiguration, zffb)
			} else if len(z.LoginIDKeys) > 0 {
				for key, _ := range z.LoginIDKeys {
					delete(z.LoginIDKeys, key)
				}
			}
			for zffb > 0 {
				var zfum string
				var zaps UserVerificationKeyConfiguration
				zffb--
				zfum, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
				bts, err = zaps.UnmarshalMsg(bts)
				if err != nil {
					return
				}
				z.LoginIDKeys[zfum] = zaps
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
func (z *UserVerificationConfiguration) Msgsize() (s int) {
	s = 1 + 11 + msgp.StringPrefixSize + len(z.URLPrefix) + 20 + msgp.BoolSize + 9 + msgp.StringPrefixSize + len(string(z.Criteria)) + 15 + msgp.StringPrefixSize + len(z.ErrorRedirect) + 15 + msgp.StringPrefixSize + len(z.ErrorHTMLURL) + 14 + msgp.MapHeaderSize
	if z.LoginIDKeys != nil {
		for zfum, zaps := range z.LoginIDKeys {
			_ = zaps
			s += msgp.StringPrefixSize + len(zfum) + zaps.Msgsize()
		}
	}
	return
}

// DecodeMsg implements msgp.Decodable
func (z *UserVerificationCriteria) DecodeMsg(dc *msgp.Reader) (err error) {
	{
		var zbgg string
		zbgg, err = dc.ReadString()
		(*z) = UserVerificationCriteria(zbgg)
	}
	if err != nil {
		return
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z UserVerificationCriteria) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteString(string(z))
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z UserVerificationCriteria) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	o = msgp.AppendString(o, string(z))
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *UserVerificationCriteria) UnmarshalMsg(bts []byte) (o []byte, err error) {
	{
		var zcnq string
		zcnq, bts, err = msgp.ReadStringBytes(bts)
		(*z) = UserVerificationCriteria(zcnq)
	}
	if err != nil {
		return
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z UserVerificationCriteria) Msgsize() (s int) {
	s = msgp.StringPrefixSize + len(string(z))
	return
}

// DecodeMsg implements msgp.Decodable
func (z *UserVerificationKeyConfiguration) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zbae uint32
	zbae, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zbae > 0 {
		zbae--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "code_format":
			{
				var zreu string
				zreu, err = dc.ReadString()
				z.CodeFormat = UserVerificationCodeFormat(zreu)
			}
			if err != nil {
				return
			}
		case "expiry":
			z.Expiry, err = dc.ReadInt64()
			if err != nil {
				return
			}
		case "success_redirect":
			z.SuccessRedirect, err = dc.ReadString()
			if err != nil {
				return
			}
		case "success_html_url":
			z.SuccessHTMLURL, err = dc.ReadString()
			if err != nil {
				return
			}
		case "error_redirect":
			z.ErrorRedirect, err = dc.ReadString()
			if err != nil {
				return
			}
		case "error_html_url":
			z.ErrorHTMLURL, err = dc.ReadString()
			if err != nil {
				return
			}
		case "provider":
			{
				var znuz string
				znuz, err = dc.ReadString()
				z.Provider = UserVerificationProvider(znuz)
			}
			if err != nil {
				return
			}
		case "provider_config":
			err = z.ProviderConfig.DecodeMsg(dc)
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
func (z *UserVerificationKeyConfiguration) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 8
	// write "code_format"
	err = en.Append(0x88, 0xab, 0x63, 0x6f, 0x64, 0x65, 0x5f, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x74)
	if err != nil {
		return err
	}
	err = en.WriteString(string(z.CodeFormat))
	if err != nil {
		return
	}
	// write "expiry"
	err = en.Append(0xa6, 0x65, 0x78, 0x70, 0x69, 0x72, 0x79)
	if err != nil {
		return err
	}
	err = en.WriteInt64(z.Expiry)
	if err != nil {
		return
	}
	// write "success_redirect"
	err = en.Append(0xb0, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x5f, 0x72, 0x65, 0x64, 0x69, 0x72, 0x65, 0x63, 0x74)
	if err != nil {
		return err
	}
	err = en.WriteString(z.SuccessRedirect)
	if err != nil {
		return
	}
	// write "success_html_url"
	err = en.Append(0xb0, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x5f, 0x68, 0x74, 0x6d, 0x6c, 0x5f, 0x75, 0x72, 0x6c)
	if err != nil {
		return err
	}
	err = en.WriteString(z.SuccessHTMLURL)
	if err != nil {
		return
	}
	// write "error_redirect"
	err = en.Append(0xae, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x5f, 0x72, 0x65, 0x64, 0x69, 0x72, 0x65, 0x63, 0x74)
	if err != nil {
		return err
	}
	err = en.WriteString(z.ErrorRedirect)
	if err != nil {
		return
	}
	// write "error_html_url"
	err = en.Append(0xae, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x5f, 0x68, 0x74, 0x6d, 0x6c, 0x5f, 0x75, 0x72, 0x6c)
	if err != nil {
		return err
	}
	err = en.WriteString(z.ErrorHTMLURL)
	if err != nil {
		return
	}
	// write "provider"
	err = en.Append(0xa8, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x64, 0x65, 0x72)
	if err != nil {
		return err
	}
	err = en.WriteString(string(z.Provider))
	if err != nil {
		return
	}
	// write "provider_config"
	err = en.Append(0xaf, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x64, 0x65, 0x72, 0x5f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67)
	if err != nil {
		return err
	}
	err = z.ProviderConfig.EncodeMsg(en)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *UserVerificationKeyConfiguration) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 8
	// string "code_format"
	o = append(o, 0x88, 0xab, 0x63, 0x6f, 0x64, 0x65, 0x5f, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x74)
	o = msgp.AppendString(o, string(z.CodeFormat))
	// string "expiry"
	o = append(o, 0xa6, 0x65, 0x78, 0x70, 0x69, 0x72, 0x79)
	o = msgp.AppendInt64(o, z.Expiry)
	// string "success_redirect"
	o = append(o, 0xb0, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x5f, 0x72, 0x65, 0x64, 0x69, 0x72, 0x65, 0x63, 0x74)
	o = msgp.AppendString(o, z.SuccessRedirect)
	// string "success_html_url"
	o = append(o, 0xb0, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x5f, 0x68, 0x74, 0x6d, 0x6c, 0x5f, 0x75, 0x72, 0x6c)
	o = msgp.AppendString(o, z.SuccessHTMLURL)
	// string "error_redirect"
	o = append(o, 0xae, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x5f, 0x72, 0x65, 0x64, 0x69, 0x72, 0x65, 0x63, 0x74)
	o = msgp.AppendString(o, z.ErrorRedirect)
	// string "error_html_url"
	o = append(o, 0xae, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x5f, 0x68, 0x74, 0x6d, 0x6c, 0x5f, 0x75, 0x72, 0x6c)
	o = msgp.AppendString(o, z.ErrorHTMLURL)
	// string "provider"
	o = append(o, 0xa8, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x64, 0x65, 0x72)
	o = msgp.AppendString(o, string(z.Provider))
	// string "provider_config"
	o = append(o, 0xaf, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x64, 0x65, 0x72, 0x5f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67)
	o, err = z.ProviderConfig.MarshalMsg(o)
	if err != nil {
		return
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *UserVerificationKeyConfiguration) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zjqx uint32
	zjqx, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zjqx > 0 {
		zjqx--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "code_format":
			{
				var zmzo string
				zmzo, bts, err = msgp.ReadStringBytes(bts)
				z.CodeFormat = UserVerificationCodeFormat(zmzo)
			}
			if err != nil {
				return
			}
		case "expiry":
			z.Expiry, bts, err = msgp.ReadInt64Bytes(bts)
			if err != nil {
				return
			}
		case "success_redirect":
			z.SuccessRedirect, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "success_html_url":
			z.SuccessHTMLURL, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "error_redirect":
			z.ErrorRedirect, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "error_html_url":
			z.ErrorHTMLURL, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "provider":
			{
				var ztar string
				ztar, bts, err = msgp.ReadStringBytes(bts)
				z.Provider = UserVerificationProvider(ztar)
			}
			if err != nil {
				return
			}
		case "provider_config":
			bts, err = z.ProviderConfig.UnmarshalMsg(bts)
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
func (z *UserVerificationKeyConfiguration) Msgsize() (s int) {
	s = 1 + 12 + msgp.StringPrefixSize + len(string(z.CodeFormat)) + 7 + msgp.Int64Size + 17 + msgp.StringPrefixSize + len(z.SuccessRedirect) + 17 + msgp.StringPrefixSize + len(z.SuccessHTMLURL) + 15 + msgp.StringPrefixSize + len(z.ErrorRedirect) + 15 + msgp.StringPrefixSize + len(z.ErrorHTMLURL) + 9 + msgp.StringPrefixSize + len(string(z.Provider)) + 16 + z.ProviderConfig.Msgsize()
	return
}

// DecodeMsg implements msgp.Decodable
func (z *UserVerificationProvider) DecodeMsg(dc *msgp.Reader) (err error) {
	{
		var zkut string
		zkut, err = dc.ReadString()
		(*z) = UserVerificationProvider(zkut)
	}
	if err != nil {
		return
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z UserVerificationProvider) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteString(string(z))
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z UserVerificationProvider) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	o = msgp.AppendString(o, string(z))
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *UserVerificationProvider) UnmarshalMsg(bts []byte) (o []byte, err error) {
	{
		var zmyg string
		zmyg, bts, err = msgp.ReadStringBytes(bts)
		(*z) = UserVerificationProvider(zmyg)
	}
	if err != nil {
		return
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z UserVerificationProvider) Msgsize() (s int) {
	s = msgp.StringPrefixSize + len(string(z))
	return
}

// DecodeMsg implements msgp.Decodable
func (z *UserVerificationProviderConfiguration) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zmsv uint32
	zmsv, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zmsv > 0 {
		zmsv--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "subject":
			z.Subject, err = dc.ReadString()
			if err != nil {
				return
			}
		case "sender":
			z.Sender, err = dc.ReadString()
			if err != nil {
				return
			}
		case "reply_to":
			z.ReplyTo, err = dc.ReadString()
			if err != nil {
				return
			}
		case "text_url":
			z.TextURL, err = dc.ReadString()
			if err != nil {
				return
			}
		case "html_url":
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
func (z *UserVerificationProviderConfiguration) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 5
	// write "subject"
	err = en.Append(0x85, 0xa7, 0x73, 0x75, 0x62, 0x6a, 0x65, 0x63, 0x74)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Subject)
	if err != nil {
		return
	}
	// write "sender"
	err = en.Append(0xa6, 0x73, 0x65, 0x6e, 0x64, 0x65, 0x72)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Sender)
	if err != nil {
		return
	}
	// write "reply_to"
	err = en.Append(0xa8, 0x72, 0x65, 0x70, 0x6c, 0x79, 0x5f, 0x74, 0x6f)
	if err != nil {
		return err
	}
	err = en.WriteString(z.ReplyTo)
	if err != nil {
		return
	}
	// write "text_url"
	err = en.Append(0xa8, 0x74, 0x65, 0x78, 0x74, 0x5f, 0x75, 0x72, 0x6c)
	if err != nil {
		return err
	}
	err = en.WriteString(z.TextURL)
	if err != nil {
		return
	}
	// write "html_url"
	err = en.Append(0xa8, 0x68, 0x74, 0x6d, 0x6c, 0x5f, 0x75, 0x72, 0x6c)
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
func (z *UserVerificationProviderConfiguration) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 5
	// string "subject"
	o = append(o, 0x85, 0xa7, 0x73, 0x75, 0x62, 0x6a, 0x65, 0x63, 0x74)
	o = msgp.AppendString(o, z.Subject)
	// string "sender"
	o = append(o, 0xa6, 0x73, 0x65, 0x6e, 0x64, 0x65, 0x72)
	o = msgp.AppendString(o, z.Sender)
	// string "reply_to"
	o = append(o, 0xa8, 0x72, 0x65, 0x70, 0x6c, 0x79, 0x5f, 0x74, 0x6f)
	o = msgp.AppendString(o, z.ReplyTo)
	// string "text_url"
	o = append(o, 0xa8, 0x74, 0x65, 0x78, 0x74, 0x5f, 0x75, 0x72, 0x6c)
	o = msgp.AppendString(o, z.TextURL)
	// string "html_url"
	o = append(o, 0xa8, 0x68, 0x74, 0x6d, 0x6c, 0x5f, 0x75, 0x72, 0x6c)
	o = msgp.AppendString(o, z.HTMLURL)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *UserVerificationProviderConfiguration) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zyba uint32
	zyba, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zyba > 0 {
		zyba--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "subject":
			z.Subject, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "sender":
			z.Sender, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "reply_to":
			z.ReplyTo, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "text_url":
			z.TextURL, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "html_url":
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
func (z *UserVerificationProviderConfiguration) Msgsize() (s int) {
	s = 1 + 8 + msgp.StringPrefixSize + len(z.Subject) + 7 + msgp.StringPrefixSize + len(z.Sender) + 9 + msgp.StringPrefixSize + len(z.ReplyTo) + 9 + msgp.StringPrefixSize + len(z.TextURL) + 9 + msgp.StringPrefixSize + len(z.HTMLURL)
	return
}

// DecodeMsg implements msgp.Decodable
func (z *WelcomeEmailConfiguration) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zddv uint32
	zddv, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zddv > 0 {
		zddv--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "enabled":
			z.Enabled, err = dc.ReadBool()
			if err != nil {
				return
			}
		case "url_prefix":
			z.URLPrefix, err = dc.ReadString()
			if err != nil {
				return
			}
		case "sender":
			z.Sender, err = dc.ReadString()
			if err != nil {
				return
			}
		case "subject":
			z.Subject, err = dc.ReadString()
			if err != nil {
				return
			}
		case "reply_to":
			z.ReplyTo, err = dc.ReadString()
			if err != nil {
				return
			}
		case "text_url":
			z.TextURL, err = dc.ReadString()
			if err != nil {
				return
			}
		case "html_url":
			z.HTMLURL, err = dc.ReadString()
			if err != nil {
				return
			}
		case "destination":
			{
				var zoxi string
				zoxi, err = dc.ReadString()
				z.Destination = WelcomeEmailDestination(zoxi)
			}
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
	// write "enabled"
	err = en.Append(0x88, 0xa7, 0x65, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64)
	if err != nil {
		return err
	}
	err = en.WriteBool(z.Enabled)
	if err != nil {
		return
	}
	// write "url_prefix"
	err = en.Append(0xaa, 0x75, 0x72, 0x6c, 0x5f, 0x70, 0x72, 0x65, 0x66, 0x69, 0x78)
	if err != nil {
		return err
	}
	err = en.WriteString(z.URLPrefix)
	if err != nil {
		return
	}
	// write "sender"
	err = en.Append(0xa6, 0x73, 0x65, 0x6e, 0x64, 0x65, 0x72)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Sender)
	if err != nil {
		return
	}
	// write "subject"
	err = en.Append(0xa7, 0x73, 0x75, 0x62, 0x6a, 0x65, 0x63, 0x74)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Subject)
	if err != nil {
		return
	}
	// write "reply_to"
	err = en.Append(0xa8, 0x72, 0x65, 0x70, 0x6c, 0x79, 0x5f, 0x74, 0x6f)
	if err != nil {
		return err
	}
	err = en.WriteString(z.ReplyTo)
	if err != nil {
		return
	}
	// write "text_url"
	err = en.Append(0xa8, 0x74, 0x65, 0x78, 0x74, 0x5f, 0x75, 0x72, 0x6c)
	if err != nil {
		return err
	}
	err = en.WriteString(z.TextURL)
	if err != nil {
		return
	}
	// write "html_url"
	err = en.Append(0xa8, 0x68, 0x74, 0x6d, 0x6c, 0x5f, 0x75, 0x72, 0x6c)
	if err != nil {
		return err
	}
	err = en.WriteString(z.HTMLURL)
	if err != nil {
		return
	}
	// write "destination"
	err = en.Append(0xab, 0x64, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e)
	if err != nil {
		return err
	}
	err = en.WriteString(string(z.Destination))
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *WelcomeEmailConfiguration) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 8
	// string "enabled"
	o = append(o, 0x88, 0xa7, 0x65, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64)
	o = msgp.AppendBool(o, z.Enabled)
	// string "url_prefix"
	o = append(o, 0xaa, 0x75, 0x72, 0x6c, 0x5f, 0x70, 0x72, 0x65, 0x66, 0x69, 0x78)
	o = msgp.AppendString(o, z.URLPrefix)
	// string "sender"
	o = append(o, 0xa6, 0x73, 0x65, 0x6e, 0x64, 0x65, 0x72)
	o = msgp.AppendString(o, z.Sender)
	// string "subject"
	o = append(o, 0xa7, 0x73, 0x75, 0x62, 0x6a, 0x65, 0x63, 0x74)
	o = msgp.AppendString(o, z.Subject)
	// string "reply_to"
	o = append(o, 0xa8, 0x72, 0x65, 0x70, 0x6c, 0x79, 0x5f, 0x74, 0x6f)
	o = msgp.AppendString(o, z.ReplyTo)
	// string "text_url"
	o = append(o, 0xa8, 0x74, 0x65, 0x78, 0x74, 0x5f, 0x75, 0x72, 0x6c)
	o = msgp.AppendString(o, z.TextURL)
	// string "html_url"
	o = append(o, 0xa8, 0x68, 0x74, 0x6d, 0x6c, 0x5f, 0x75, 0x72, 0x6c)
	o = msgp.AppendString(o, z.HTMLURL)
	// string "destination"
	o = append(o, 0xab, 0x64, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e)
	o = msgp.AppendString(o, string(z.Destination))
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *WelcomeEmailConfiguration) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zfsf uint32
	zfsf, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zfsf > 0 {
		zfsf--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "enabled":
			z.Enabled, bts, err = msgp.ReadBoolBytes(bts)
			if err != nil {
				return
			}
		case "url_prefix":
			z.URLPrefix, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "sender":
			z.Sender, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "subject":
			z.Subject, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "reply_to":
			z.ReplyTo, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "text_url":
			z.TextURL, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "html_url":
			z.HTMLURL, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "destination":
			{
				var zgpy string
				zgpy, bts, err = msgp.ReadStringBytes(bts)
				z.Destination = WelcomeEmailDestination(zgpy)
			}
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
	s = 1 + 8 + msgp.BoolSize + 11 + msgp.StringPrefixSize + len(z.URLPrefix) + 7 + msgp.StringPrefixSize + len(z.Sender) + 8 + msgp.StringPrefixSize + len(z.Subject) + 9 + msgp.StringPrefixSize + len(z.ReplyTo) + 9 + msgp.StringPrefixSize + len(z.TextURL) + 9 + msgp.StringPrefixSize + len(z.HTMLURL) + 12 + msgp.StringPrefixSize + len(string(z.Destination))
	return
}

// DecodeMsg implements msgp.Decodable
func (z *WelcomeEmailDestination) DecodeMsg(dc *msgp.Reader) (err error) {
	{
		var zclm string
		zclm, err = dc.ReadString()
		(*z) = WelcomeEmailDestination(zclm)
	}
	if err != nil {
		return
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z WelcomeEmailDestination) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteString(string(z))
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z WelcomeEmailDestination) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	o = msgp.AppendString(o, string(z))
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *WelcomeEmailDestination) UnmarshalMsg(bts []byte) (o []byte, err error) {
	{
		var zxiu string
		zxiu, bts, err = msgp.ReadStringBytes(bts)
		(*z) = WelcomeEmailDestination(zxiu)
	}
	if err != nil {
		return
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z WelcomeEmailDestination) Msgsize() (s int) {
	s = msgp.StringPrefixSize + len(string(z))
	return
}
