package config

const (
	trackingTokenParam     = "token"
	trackingURLParam       = "u"
	trackingSignatureParam = "s"
	trackingOpenPath       = "/t/o/"
	trackingClickPath      = "/t/c/"
	trackingSettingKey     = "emailTracking"
	trackingKeyLabel       = "outreach-tracking"
	trackingTokenBytes     = 16
	noStoreValue           = "no-store"
	gifContentType         = "image/gif"
)

// A 1x1 transparent GIF served by the open-tracking endpoint.
var trackingPixelI = []byte{
	0x47, 0x49, 0x46, 0x38, 0x39, 0x61, 0x01, 0x00, 0x01, 0x00, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00,
	0xff, 0xff, 0xff, 0x21, 0xf9, 0x04, 0x01, 0x00, 0x00, 0x00, 0x00, 0x2c, 0x00, 0x00, 0x00, 0x00,
	0x01, 0x00, 0x01, 0x00, 0x00, 0x02, 0x02, 0x44, 0x01, 0x00, 0x3b,
}

func GetTrackingTokenParam() string     { return trackingTokenParam }
func GetTrackingURLParam() string       { return trackingURLParam }
func GetTrackingSignatureParam() string { return trackingSignatureParam }
func GetTrackingOpenPath() string       { return trackingOpenPath }
func GetTrackingClickPath() string      { return trackingClickPath }
func GetTrackingSettingKey() string     { return trackingSettingKey }
func GetTrackingKeyLabel() string       { return trackingKeyLabel }
func GetTrackingTokenBytes() int        { return trackingTokenBytes }
func GetTrackingPixel() []byte          { return append([]byte(nil), trackingPixelI...) }
func GetNoStoreValue() string           { return noStoreValue }
func GetGIFContentType() string         { return gifContentType }
