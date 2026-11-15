package config

import "strings"

const (
	unsubscribePath           = "/u/"
	unsubscribeKeyLabel       = "outreach-unsubscribe"
	unsubscribeKeyword        = "unsubscribe"
	replySubjectPrefixes      = "re:,fwd:,fw:"
	listUnsubscribeHeader     = "List-Unsubscribe"
	listUnsubscribePostHeader = "List-Unsubscribe-Post"
	listUnsubscribeOneClick   = "List-Unsubscribe=One-Click"
	htmlContentType           = "text/html; charset=utf-8"
)

func GetUnsubscribePath() string              { return unsubscribePath }
func GetUnsubscribeKeyLabel() string          { return unsubscribeKeyLabel }
func GetUnsubscribeKeyword() string           { return unsubscribeKeyword }
func GetReplySubjectPrefixes() []string       { return strings.Split(replySubjectPrefixes, ",") }
func GetListUnsubscribeHeader() string        { return listUnsubscribeHeader }
func GetListUnsubscribePostHeader() string    { return listUnsubscribePostHeader }
func GetListUnsubscribeOneClickValue() string { return listUnsubscribeOneClick }
func GetHTMLContentType() string              { return htmlContentType }
