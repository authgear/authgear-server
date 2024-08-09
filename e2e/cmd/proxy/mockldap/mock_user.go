package mockldap

type User struct {
	DN         string `json:"dn"`
	UID        string `json:"uid"`
	Credential struct {
		Password string `json:"password"`
	} `json:"credential"`
	Attributes map[string]string `json:"attributes"`
}
