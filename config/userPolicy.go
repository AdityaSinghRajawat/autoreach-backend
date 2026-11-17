package config

const (
	emailChangeKeyLabel       = "email-change"
	emailChangeEmailParam     = "email"
	emailChangeBoundSeparator = "\x00"
)

type userConfig struct {
	emailChangeURL string
}

var userConfigI userConfig

func loadUserConfig() {
	userConfigI = userConfig{
		emailChangeURL: getEnvWithDefault("AUTH_EMAIL_CHANGE_URL", ""),
	}
}

func GetEmailChangeURL() string            { return userConfigI.emailChangeURL }
func GetEmailChangeKeyLabel() string       { return emailChangeKeyLabel }
func GetEmailChangeEmailParam() string     { return emailChangeEmailParam }
func GetEmailChangeBoundSeparator() string { return emailChangeBoundSeparator }
