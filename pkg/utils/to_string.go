package utils

import "strconv"

func ToString(num int) string {
	if num == 0 {
		return ""
	}
	return strconv.Itoa(num)
}
