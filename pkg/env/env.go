package env

import (
	"github.com/spf13/viper"
)

const Environment = "environment"

func IsPrivate() bool {
	env := viper.GetString(Environment)
	return env == "private"
}
