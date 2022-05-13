package idx

import (
	"errors"
	"net"

	"github.com/sony/sonyflake"
)

var sf = sonyflake.NewSonyflake(sonyflake.Settings{MachineID: machineID})

// NextID generate id
func NextID() (uint64, error) {
	return sf.NextID()
}

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
		inet, ok := a.(*net.IPNet)
		if !ok || inet.IP.IsLoopback() {
			continue
		}

		ip := inet.IP.To4()
		// Pass ipv6 address
		if ip == nil {
			continue
		}
		return ip, nil
	}
	return nil, errors.New("no private ip address")
}
