package config

var atsHostSuffixesI = map[string][]string{
	"GREENHOUSE":      {"greenhouse.io"},
	"LEVER":           {"lever.co"},
	"WORKDAY":         {"myworkdayjobs.com", "myworkdaysite.com"},
	"ASHBY":           {"ashbyhq.com"},
	"SMARTRECRUITERS": {"smartrecruiters.com"},
}

var retryableErrCodesI = map[CustomErrorCode]struct{}{
	errCodesI.generic.internalServerError: {},
	errCodesI.generic.serviceUnavailable:  {},
	errCodesI.generic.rateLimited:         {},
}

func GetATSHostSuffixes(atsType string) ([]string, bool) {
	suffixes, restricted := atsHostSuffixesI[atsType]

	return append([]string(nil), suffixes...), restricted
}

func IsRetryableErrCode(errCode CustomErrorCode) bool {
	_, retryable := retryableErrCodesI[errCode]

	return retryable
}

const defaultMinMatchScore = 0.80

func GetDefaultMinMatchScore() float64 {
	return defaultMinMatchScore
}
