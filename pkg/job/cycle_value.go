package job

import (
	"fmt"
	"strconv"
	"strings"
)

type CycleValue struct {
	Day   int   `json:"day,omitempty" binding:"excluded_with=Weeks,omitempty,min=1,max=30"`
	Weeks []int `json:"weeks" binding:"excluded_with=Day,dive,min=0,max=6"`
	Hours []int `json:"hours" binding:"required,dive,min=0,max=23"`
}

func (c CycleValue) ToCronJobExpr() string {
	fields := make([]string, 6)
	fields[0] = "0"
	fields[1] = "0"
	hours := make([]string, 0, len(c.Hours))
	for _, hour := range c.Hours {
		hours = append(hours, strconv.Itoa(hour))
	}
	fields[2] = strings.Join(hours, ",")
	if c.Day != 0 {
		fields[3] = "*/" + strconv.Itoa(c.Day)
		fields[4] = "*"
		fields[5] = "?"
	} else {
		fields[3] = "?"
		fields[4] = "*"
		weeks := make([]string, 0, len(c.Weeks))
		for _, week := range c.Weeks {
			weeks = append(weeks, strconv.Itoa(week))
		}
		fields[5] = strings.Join(weeks, ",")
	}
	return strings.Join(fields, " ")
}

func (c *CycleValue) Parse(v string) error {
	fields := strings.Fields(v)
	count := len(fields)
	switch count {
	case 7:
		fields = fields[:6]
	case 6:
	default:
		return fmt.Errorf("%s is not cron", v)
	}
	if fields[0] != "0" {
		return fmt.Errorf("%s is not 0", fields[0])
	}
	if fields[1] != "0" {
		return fmt.Errorf("%s is not 0", fields[1])
	}
	if fields[4] != "*" {
		return fmt.Errorf("%s is not *", fields[4])
	}
	hours := strings.Split(fields[2], ",")
	for _, hour := range hours {
		num, err := strconv.Atoi(hour)
		if err != nil {
			return fmt.Errorf("parse hour failed,%w", err)
		}
		c.Hours = append(c.Hours, num)
	}

	if fields[3] == "?" {
		weeks := strings.Split(fields[5], ",")
		for _, week := range weeks {
			num, err := strconv.Atoi(week)
			if err != nil {
				return fmt.Errorf("parse week failed,%w", err)
			}
			c.Weeks = append(c.Weeks, num)
		}
	} else {
		if !strings.Contains(fields[3], "*/") {
			return fmt.Errorf("%s not found */", fields[3])
		}
		day := strings.TrimLeft(fields[3], "*/")
		var err error
		if c.Day, err = strconv.Atoi(day); err != nil {
			return fmt.Errorf("parse day failed,%w", err)
		}
		if fields[5] != "?" {
			return fmt.Errorf("%s is not *", fields[5])
		}
	}
	return nil
}
