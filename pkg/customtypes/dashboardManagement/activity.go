package dashboardManagement

import (
	"encoding/base64"
	"strings"
	"time"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
)

type ActivityItem struct {
	EventID       string    `json:"eventId"`
	EventType     string    `json:"eventType"`
	Channel       *string   `json:"channel"`
	OccurredAt    time.Time `json:"occurredAt"`
	JobID         string    `json:"jobId"`
	JobTitle      string    `json:"jobTitle"`
	CompanyName   string    `json:"companyName"`
	ContactID     *string   `json:"contactId"`
	ContactName   *string   `json:"contactName"`
	ApplicationID *string   `json:"applicationId"`
	MessageID     *string   `json:"messageId"`
}

type ListActivityParams struct {
	Limit  int    `form:"limit"  validate:"omitempty,min=1,max=100"`
	Before string `form:"before" validate:"omitempty,max=200"`
}

type ActivityPage struct {
	Items      []ActivityItem `json:"items"`
	NextCursor *string        `json:"nextCursor"`
}

// ActivityCursor marks a position in the feed, which is ordered by time and then event ID.
type ActivityCursor struct {
	OccurredAt time.Time
	EventID    string
}

func (i ActivityItem) Cursor() ActivityCursor {
	return ActivityCursor{OccurredAt: i.OccurredAt, EventID: i.EventID}
}

func (c ActivityCursor) Encode() string {
	return base64.RawURLEncoding.EncodeToString(
		[]byte(
			c.OccurredAt.UTC().
				Format(time.RFC3339Nano) +
				config.GetActivityCursorSeparator() + c.EventID,
		),
	)
}

func ParseActivityCursor(encoded string) (ActivityCursor, bool) {
	decoded, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return ActivityCursor{}, false
	}

	occurredAt, eventID, found := strings.Cut(string(decoded), config.GetActivityCursorSeparator())
	if !found || eventID == "" {
		return ActivityCursor{}, false
	}

	parsed, err := time.Parse(time.RFC3339Nano, occurredAt)
	if err != nil {
		return ActivityCursor{}, false
	}

	return ActivityCursor{OccurredAt: parsed, EventID: eventID}, true
}
