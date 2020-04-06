package config

//go:generate msgp -tests=false
type AuthAPIConfiguration struct {
	Enabled            bool                                  `json:"enabled,omitempty" yaml:"enabled" msg:"enabled"`
	OnIdentityConflict *AuthAPIIdentityConflictConfiguration `json:"on_identity_conflict,omitempty" yaml:"on_identity_conflict" msg:"on_identity_conflict" default_zero_value:"true"`
}

type AuthAPIIdentityConflictConfiguration struct {
	LoginID *AuthAPILoginIDConflictConfiguration `json:"login_id,omitempty" yaml:"login_id" msg:"login_id" default_zero_value:"true"`
	OAuth   *AuthAPIOAuthConflictConfiguration   `json:"oauth,omitempty" yaml:"oauth" msg:"oauth" default_zero_value:"true"`
}

type AuthAPILoginIDConflictConfiguration struct {
	AllowCreateNewUser bool `json:"allow_create_new_user,omitempty" yaml:"allow_create_new_user" msg:"allow_create_new_user"`
}

type AuthAPIOAuthConflictConfiguration struct {
	AllowCreateNewUser bool `json:"allow_create_new_user,omitempty" yaml:"allow_create_new_user" msg:"allow_create_new_user"`
	AllowAutoMergeUser bool `json:"allow_auto_merge_user,omitempty" yaml:"allow_auto_merge_user" msg:"allow_auto_merge_user"`
}
