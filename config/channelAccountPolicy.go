package config

import (
	"encoding/base64"
	"log"
	"strconv"
	"strings"
)

const (
	accountIDParam         = "accountId"
	credentialKeySize      = 32
	credentialKeyPairSep   = ":"
	maxChannelAccountError = 500
)

type channelAccountConfig struct {
	credentialKeys          map[int][]byte
	activeCredentialKey     int
	maxDailyEmailsPerSender int
}

var channelAccountConfigI channelAccountConfig

func loadChannelAccountConfig() {
	keys := parseCredentialKeys(getEnvListWithDefault("CHANNEL_CREDENTIAL_KEYS", nil))
	active := getEnvIntWithDefault("CHANNEL_CREDENTIAL_KEY_VERSION", 0)

	if _, found := keys[active]; len(keys) > 0 && !found {
		log.Fatalf("CHANNEL_CREDENTIAL_KEY_VERSION %d is not in CHANNEL_CREDENTIAL_KEYS", active)
	}

	if len(keys) == 0 && IsProduction() {
		log.Fatalf("CHANNEL_CREDENTIAL_KEYS is required in production")
	}

	channelAccountConfigI = channelAccountConfig{
		credentialKeys:          keys,
		activeCredentialKey:     active,
		maxDailyEmailsPerSender: max(getEnvIntWithDefault("OUTREACH_MAX_DAILY_EMAILS", 30), 1),
	}
}

func parseCredentialKeys(entries []string) map[int][]byte {
	keys := make(map[int][]byte, len(entries))

	for _, entry := range entries {
		rawVersion, rawKey, found := strings.Cut(entry, credentialKeyPairSep)

		version, err := strconv.Atoi(rawVersion)
		if !found || err != nil || version <= 0 {
			log.Fatalf("CHANNEL_CREDENTIAL_KEYS entries must look like <version>:<base64 key>")
		}

		key, err := base64.StdEncoding.DecodeString(rawKey)
		if err != nil || len(key) != credentialKeySize {
			log.Fatalf(
				"CHANNEL_CREDENTIAL_KEYS version %d must be %d bytes of base64",
				version,
				credentialKeySize,
			)
		}

		keys[version] = key
	}

	return keys
}

func GetCredentialKeys() map[int][]byte    { return channelAccountConfigI.credentialKeys }
func GetActiveCredentialKeyVersion() int   { return channelAccountConfigI.activeCredentialKey }
func GetMaxDailyEmailsPerSender() int      { return channelAccountConfigI.maxDailyEmailsPerSender }
func GetAccountIDParam() string            { return accountIDParam }
func GetMaxChannelAccountErrorLength() int { return maxChannelAccountError }
