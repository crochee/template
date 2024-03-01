package net

import (
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"net"
)

/*
	https://en.wikipedia.org/wiki/Classless_Inter-Domain_Routing
	CIDR表示法:
	IPv4   	网络号/前缀长度		192.168.1.0/24
	IPv6	接口号/前缀长度		2001:db8::/64
*/
type CIDR struct {
	ip    net.IP
	ipNet *net.IPNet
}

// 解析CIDR网段
func ParseCIDR(s string) (*CIDR, error) {
	i, n, err := net.ParseCIDR(s)
	if err != nil {
		return nil, err
	}
	return &CIDR{ip: i, ipNet: n}, nil
}

// 判断网段是否相等
func (c CIDR) Equal(ns string) bool {
	c2, err := ParseCIDR(ns)
	if err != nil {
		return false
	}
	return c.ipNet.IP.Equal(c2.ipNet.IP) /* && c.ipNet.IP.Equal(c2.ip) */
}

// 判断是否IPv4
func (c CIDR) IsIPv4() bool {
	_, bits := c.ipNet.Mask.Size()
	return bits/8 == net.IPv4len
}

// 判断是否IPv6
func (c CIDR) IsIPv6() bool {
	_, bits := c.ipNet.Mask.Size()
	return bits/8 == net.IPv6len
}

// 判断IP是否包含在网段中
func (c CIDR) Contains(ip string) bool {
	return c.ipNet.Contains(net.ParseIP(ip))
}

func (c CIDR) ContainsCIDR(cidr CIDR) bool {
	startIp, endIp := cidr.IPRange()
	return c.Contains(startIp) && c.Contains(endIp)
}

// 根据子网掩码长度校准后的CIDR
func (c CIDR) CIDR() string {
	return c.ipNet.String()
}

// CIDR字符串中的IP部分
func (c CIDR) IP() string {
	return c.ip.String()
}

func (c CIDR) RawIP() net.IP {
	return c.ip
}

// 网络号
func (c CIDR) Network() string {
	return c.ipNet.IP.String()
}

// 子网掩码位数
func (c CIDR) MaskSize() (ones, bits int) {
	ones, bits = c.ipNet.Mask.Size()
	return
}

// 子网掩码
func (c CIDR) Mask() string {
	mask, _ := hex.DecodeString(c.ipNet.Mask.String())
	return net.IP(mask).String()
}

// 网关(网段第二个IP)
func (c CIDR) Gateway() (gateway string) {
	gw := make(net.IP, len(c.ipNet.IP))
	copy(gw, c.ipNet.IP)
	for step := 0; step < 2 && c.ipNet.Contains(gw); step++ {
		gateway = gw.String()
		IncrIP(gw)
	}
	return
}

// 广播地址(网段最后一个IP)
func (c CIDR) Broadcast() string {
	mask := c.ipNet.Mask
	bcst := make(net.IP, len(c.ipNet.IP))
	copy(bcst, c.ipNet.IP)
	for i := 0; i < len(mask); i++ {
		ipIdx := len(bcst) - i - 1
		bcst[ipIdx] = c.ipNet.IP[ipIdx] | ^mask[len(mask)-i-1]
	}
	return bcst.String()
}

// 起始IP、结束IP
func (c CIDR) IPRange() (start, end string) {
	return c.Network(), c.Broadcast()
}

// IP数量
func (c CIDR) IPCount() *big.Int {
	ones, bits := c.ipNet.Mask.Size()
	return big.NewInt(0).Lsh(big.NewInt(1), uint(bits-ones))
}

// 遍历网段下所有IP
func (c CIDR) ForEachIP(iterator func(ip string) error) error {
	next := make(net.IP, len(c.ipNet.IP))
	copy(next, c.ipNet.IP)
	for c.ipNet.Contains(next) {
		if err := iterator(next.String()); err != nil {
			return err
		}
		IncrIP(next)
	}
	return nil
}

// 从指定IP开始遍历网段下后续的IP
func (c CIDR) ForEachIPBeginWith(beginIP string, iterator func(ip string) error) error {
	next := net.ParseIP(beginIP)
	for c.ipNet.Contains(next) {
		if err := iterator(next.String()); err != nil {
			return err
		}
		IncrIP(next)
	}
	return nil
}

// 裂解子网的方式
const (
	SubnetMethodSubnetNum = 0 // 基于子网数量
	SubnetMethodHostNum   = 1 // 基于主机数量
)

// 裂解网段
func (c CIDR) SubNetting(method, num int) ([]*CIDR, error) {
	if num < 1 || (num&(num-1)) != 0 {
		return nil, fmt.Errorf("裂解数量必须是2的次方")
	}

	newOnes := int(math.Log2(float64(num)))
	ones, bits := c.MaskSize()
	switch method {
	default:
		return nil, fmt.Errorf("不支持的裂解方式")
	case SubnetMethodSubnetNum:
		newOnes = ones + newOnes
		// 如果子网的掩码长度大于父网段的长度，则无法裂解
		if newOnes > bits {
			return nil, nil
		}
	case SubnetMethodHostNum:
		newOnes = bits - newOnes
		// 如果子网的掩码长度小于等于父网段的掩码长度，则无法裂解
		if newOnes <= ones {
			return nil, nil
		}
		// 主机数量转换为子网数量
		num = int(math.Pow(float64(2), float64(newOnes-ones)))
	}

	cidrs := []*CIDR{}
	network := make(net.IP, len(c.ipNet.IP))
	copy(network, c.ipNet.IP)
	for i := 0; i < num; i++ {
		cidr, _ := ParseCIDR(fmt.Sprintf("%v/%v", network.String(), newOnes))
		cidrs = append(cidrs, cidr)

		// 广播地址的下一个IP即为下一段的网络号
		network = net.ParseIP(cidr.Broadcast())
		IncrIP(network)
	}

	return cidrs, nil
}

// 合并网段
func SuperNetting(ns []string) (*CIDR, error) {
	num := len(ns)
	if num < 1 || (num&(num-1)) != 0 {
		return nil, fmt.Errorf("子网数量必须是2的次方")
	}

	mask := ""
	cidrs := []*CIDR{}
	for _, n := range ns {
		// 检查子网CIDR有效性
		c, err := ParseCIDR(n)
		if err != nil {
			return nil, fmt.Errorf("网段%v格式错误", n)
		}
		cidrs = append(cidrs, c)

		// TODO 暂只考虑相同子网掩码的网段合并
		if mask == "" {
			mask = c.Mask()
		} else if c.Mask() != mask {
			return nil, fmt.Errorf("子网掩码不一致")
		}
	}
	AscSortCIDRs(cidrs)

	// 检查网段是否连续
	var network net.IP
	for _, c := range cidrs {
		if len(network) > 0 {
			if !network.Equal(c.ipNet.IP) {
				return nil, fmt.Errorf("必须是连续的网段")
			}
		}
		network = net.ParseIP(c.Broadcast())
		IncrIP(network)
	}

	// 子网掩码左移，得到共同的父网段
	c := cidrs[0]
	ones, bits := c.MaskSize()
	ones -= int(math.Log2(float64(num)))
	c.ipNet.Mask = net.CIDRMask(ones, bits)
	c.ipNet.IP.Mask(c.ipNet.Mask)

	return c, nil
}
