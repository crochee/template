package client

import (
	"strings"

	"template/pkg/json"
	"template/pkg/replace"
)

type TransportContent struct {
	Request  string `json:"request"`
	Response string `json:"response"`
	Status   string `json:"status"`
}

func FormatContent(content *TransportContent) string {
	content.Request = replace.PwdReplacerReplaceStr(content.Request)
	content.Request = func(req string) string {
		if strings.Contains(req, "baseInfo") &&
			strings.Contains(req, "properties") &&
			strings.Contains(req, "Source: DCS") {
			return "调用云畅接口, " + req
		}
		if strings.Contains(req, "X-Openstack-Request-Id") {
			return "调用OpenStack接口, " + req
		}
		return req
	}(content.Request)

	content.Response = replace.PwdReplacerReplaceStr(content.Response)
	result, _ := json.Marshal(content)
	return string(result)
}
