package utils

import (
	"errors"
	"net"

	"github.com/sony/sonyflake"
)

var sf *sonyflake.Sonyflake

func machineID() (uint16, error) {
	ip, err := lower16BitIPV4()
	if err != nil {
		return 0, err
	}

	return uint16(ip[2])<<8 + uint16(ip[3]), nil
}

func lower16BitIPV4() (net.IP, error) {
	as, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, a := range as {
		ipnet, ok := a.(*net.IPNet)
		if !ok || ipnet.IP.IsLoopback() {
			continue
		}

		ip := ipnet.IP.To4()
		// Pass ipv6 address
		if ip == nil {
			continue
		}
		return ip, nil
	}
	return nil, errors.New("no private ip address")
}

func FlakeInit() {
	st := sonyflake.Settings{MachineID: machineID}
	sf = sonyflake.NewSonyflake(st)
	if sf == nil {
		panic("sonyflake not created")
	}
}

func GetSnowID() (uint64, error) {
	return sf.NextID()
}
