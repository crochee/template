package timex

import (
	"time"
)

const (
	DefaultLocation  = "Asia/Shanghai"
	TimeFormatLayout = "2006-01-02 15:04:05"
)

var CST = func() *time.Location {
	loc, err := time.LoadLocation(DefaultLocation)
	if err != nil {
		panic(err)
	}
	return loc
}()

func TimeFormat(t time.Time) string {
	return t.In(CST).Format(TimeFormatLayout)
}

func GetUTC(timeInfo string) string {
	local, _ := time.ParseInLocation(TimeFormatLayout, timeInfo, time.Local)
	return local.UTC().Format(TimeFormatLayout)
}

func GetAddTime(timeInfo string) string {
	if timeInfo == "" {
		return ""
	}
	local, _ := time.ParseInLocation(TimeFormatLayout, timeInfo, time.Local)
	add, _ := time.ParseDuration("1s")
	return local.Add(add).Format(TimeFormatLayout)
}
