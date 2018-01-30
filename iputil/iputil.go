package iputil

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"strings"
)

// AddrToList parse a IP address to a list of IPs.
func AddrToList(addr string) (ips []string) {
	// trim the last '/' in format such as '1.1.1.1/'
	addr = strings.TrimRight(addr, "/")

	// addr format as '1.1.1.1-1.1.1.255'
	if strings.Contains(addr, "-") {
		ipr := strings.SplitN(addr, "-", 2)
		return range2list(ipr[0], ipr[1])
	}

	// addr format as '1.1.1.1/xx'
	if strings.Contains(addr, "/") {
		if strings.Contains(addr, "/32") {
			ips = append(ips, strings.Replace(addr, "/32", "", -1))
			return
		}
		return addr2listx(addr)
	}

	ips = append(ips, addr)
	return
}

// addr2listx converts IP address to a list of IPs.
func addr2listx(addr string) (ips []string) {
	cidr := AddrToCidr(strings.TrimSpace(addr))

	_, ipnet, _ := net.ParseCIDR(cidr)

	ip, _ := ipnet2iprange(ipnet)
	ipNum := ip2int(ip)
	max := getIPMaskSize(ipnet.Mask) - 2 // to remove gateway and broadcast

	var (
		pos     int32 = 1
		attempt int32 = 1
		iterNum int32
	)

	for attempt < max {
		attempt++
		iterNum = ipNum + pos
		pos = pos%max + 1

		ips = append(ips, int2ip(iterNum).String())
	}

	return
}

// range2list converts ip range from start to end into a list of IPs.
func range2list(start, end string) (ips []string) {
	sip := net.ParseIP(start)
	eip := net.ParseIP(end)

	// return nil if either of the pass in ip string is invalid
	if sip.To4() == nil || eip.To4() == nil {
		return
	}

	sipNum := ip2int(sip.To4())
	eipNum := ip2int(eip.To4())

	for num := sipNum; num <= eipNum; num++ {
		ips = append(ips, int2ip(num).String())
	}

	return
}

// AddrToCidr convert IP address to IP address in CIDR format.
func AddrToCidr(addr string) string {
	if !strings.Contains(addr, "/") {
		return addr
	}

	ip := strings.Split(addr, "/")[0]
	mask := strings.Split(addr, "/")[1]
	if strings.Contains(mask, ".") {
		return fmt.Sprintf("%s/%s", ip, mask2cidr(mask))
	}

	return addr

}

// mask2cidr converts IP Mask format to CIDR format as string.
func mask2cidr(mask string) string {
	maskList := strings.Split(mask, ".")

	var m []int
	for _, v := range maskList {
		i, _ := strconv.Atoi(v)
		m = append(m, i)
	}

	ipMask := net.IPv4Mask(byte(m[0]), byte(m[1]), byte(m[2]), byte(m[3]))
	s, _ := ipMask.Size()

	return strconv.Itoa(s)
}

// cidr2mask converts IP CIDR format to IP Mask format as string.
func cidr2mask(cidr string) string {
	i, _ := strconv.Atoi(cidr)
	cidrMask := net.CIDRMask(i, 32)

	var s []string
	for _, v := range cidrMask {
		s = append(s, strconv.Itoa(int(v)))
	}

	return strings.Join(s, ".")
}

// ipnetToiprange returns the first and last ip of the IPNet's ip range.
func ipnet2iprange(ipnet *net.IPNet) (net.IP, net.IP) {
	netIP := ipnet.IP.To4()
	fip := netIP.Mask(ipnet.Mask)
	lip := net.IPv4(0, 0, 0, 0).To4()
	for i := 0; i < len(lip); i++ {
		lip[i] = netIP[i] | ^ipnet.Mask[i]
	}
	return fip, lip
}

// getIPMaskSize returns the size of IPMask.
func getIPMaskSize(mask net.IPMask) int32 {
	m := net.IPv4Mask(0, 0, 0, 0)
	for i := 0; i < net.IPv4len; i++ {
		m[i] = ^mask[i]
	}
	return int32(binary.BigEndian.Uint32(m)) + 1
}

// ip2int converts a 4 bytes IP into 32 bit integer
func ip2int(ip net.IP) int32 {
	return int32(binary.BigEndian.Uint32(ip.To4()))
}

// int2ip converts 32 bit integer into 4 bytes IP address
func int2ip(n int32) net.IP {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(n))
	return net.IP(b)
}
