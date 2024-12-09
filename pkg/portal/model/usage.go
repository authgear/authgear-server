package model

type Usage struct {
	Items []UsageItem `json:"items"`
}

type UsageItem struct {
	UsageType      UsageType      `json:"usageType"`
	SMSRegion      SMSRegion      `json:"smsRegion"`
	WhatsappRegion WhatsappRegion `json:"whatsappRegion"`
	Quantity       int            `json:"quantity"`
}
