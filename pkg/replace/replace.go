package replace

import (
	"fmt"
	"regexp"
	"strings"

	"template/pkg/json"
)

var pwdReplacer = NewManagerReplacer("password", "vncPwd", "vnc_pwd", "adminPass", "admin_pass", "virtualMachinePwd", "tenantPassword", "user_data", "userData")

func PwdReplacerReplaceInterface(in interface{}) string {
	jsonStr, err := json.Marshal(in)
	if err != nil {
		return ""
	}
	return pwdReplacer.Replace(string(jsonStr))
}

func PwdReplacerReplaceStr(in string) string {
	return pwdReplacer.Replace(in)
}

type Replacer interface {
	Replace(input string) string
}

type NopReplacer struct {
}

func (NopReplacer) Replace(input string) string {
	return input
}

type PasswordReplacer struct {
}

func (p PasswordReplacer) Replace(input string) string {
	return strings.Repeat("*", 6)
}

func NewManagerReplacer(keys ...string) Replacer {
	if len(keys) == 0 {
		return NopReplacer{}
	}
	return &messageReplacer{
		match: regexp.MustCompile(
			fmt.Sprintf(`(?U)"(%s)":\s*"(.*)",?`, strings.Join(keys, "|"))),
	}
}

type messageReplacer struct {
	match *regexp.Regexp
}

func (m *messageReplacer) Replace(input string) string {
	p := m.match.FindAllStringSubmatchIndex(input, -1)
	var (
		output   string
		oldIndex int
	)
	for _, fIndex := range p {
		if length := len(fIndex); length != 6 {
			continue
		}
		temp := PasswordReplacer{}.Replace(input[fIndex[4]:fIndex[5]])
		output += input[oldIndex:fIndex[4]] + temp
		oldIndex = fIndex[5]
	}
	output += input[oldIndex:]
	return output
}
