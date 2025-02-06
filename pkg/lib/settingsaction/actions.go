package settingsaction

type SettingsAction string

const (
	SettingsActionChangePassword SettingsAction = "change_password"
	SettingsActionDeleteAccount  SettingsAction = "delete_account"
	SettingsActionAddEmail       SettingsAction = "add_email"
	SettingsActionAddPhone       SettingsAction = "add_phone"
	SettingsActionAddUsername    SettingsAction = "add_username"
	SettingsActionChangeEmail    SettingsAction = "change_email"
	SettingsActionChangePhone    SettingsAction = "change_phone"
	SettingsActionChangeUsername SettingsAction = "change_username"
)
