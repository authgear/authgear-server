package model

type UsageName string

const (
	UsageNameUserExport UsageName = "user_export"
	UsageNameUserImport UsageName = "user_import"
	UsageNameEmail      UsageName = "email"
	UsageNameWhatsapp   UsageName = "whatsapp"
	UsageNameSMS        UsageName = "sms"
)

type UsageLimitPeriod string

const (
	UsageLimitPeriodDay   UsageLimitPeriod = "day"
	UsageLimitPeriodMonth UsageLimitPeriod = "month"
)

type UsageLimitAction string

const (
	UsageLimitActionAlert UsageLimitAction = "alert"
	UsageLimitActionBlock UsageLimitAction = "block"
)
