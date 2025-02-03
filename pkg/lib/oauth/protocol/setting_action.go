package protocol

type SettingAction string

const (
	SettingActionChangePassword SettingAction = "change_password"
	SettingActionDeleteAccount  SettingAction = "delete_account"
	SettingActionAddEmail       SettingAction = "add_email"
	SettingActionAddPhone       SettingAction = "add_phone"
	SettingActionAddUsername    SettingAction = "add_username"
	SettingActionChangeEmail    SettingAction = "change_email"
	SettingActionChangePhone    SettingAction = "change_phone"
	SettingActionChangeUsername SettingAction = "change_username"
)
