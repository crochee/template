package net

import (
	"bytes"
	"sort"
)

// 升序
func AscSortCIDRs(cs []*CIDR) {
	sort.Slice(cs, func(i, j int) bool {
		if n := bytes.Compare(cs[i].ipNet.IP, cs[j].ipNet.IP); n != 0 {
			return n < 0
		}

		if n := bytes.Compare(cs[i].ipNet.Mask, cs[j].ipNet.Mask); n != 0 {
			return n < 0
		}

		return false
	})
}

// 降序
func DescSortCIDRs(cs []*CIDR) {
	sort.Slice(cs, func(i, j int) bool {
		if n := bytes.Compare(cs[i].ipNet.IP, cs[j].ipNet.IP); n != 0 {
			return n >= 0
		}

		if n := bytes.Compare(cs[i].ipNet.Mask, cs[j].ipNet.Mask); n != 0 {
			return n >= 0
		}

		return false
	})
}
