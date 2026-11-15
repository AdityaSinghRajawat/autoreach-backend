package utils

import "time"

func GetCurrTimeStamp() time.Time {
	return time.Now().UTC().Truncate(time.Microsecond)
}

func NextSendTime(
	now time.Time,
	location *time.Location,
	startHour, endHour int,
	weekends bool,
) time.Time {
	local := now.In(location)

	for day := 0; day < 8; day++ {
		date := local.AddDate(0, 0, day)
		opens := time.Date(date.Year(), date.Month(), date.Day(), startHour, 0, 0, 0, location)
		closes := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, location).
			Add(time.Duration(endHour) * time.Hour)

		if !weekends && (opens.Weekday() == time.Saturday || opens.Weekday() == time.Sunday) {
			continue
		}

		if local.Before(opens) {
			return opens.UTC()
		}

		if local.Before(closes) {
			return now
		}
	}

	return now
}
