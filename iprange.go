package ppsh

import (
	"strings"

	"github.com/haotrr/ppsh/iputil"
)

// ParseIPRange parses addr(s) into a list of IPs.
func ParseIPRange(addrs string) (ips []string) {
	var addrSlice []string

	if strings.Contains(addrs, ",") {
		addrSlice = strings.Split(addrs, ",")
	} else if strings.Contains(addrs, ";") {
		addrSlice = strings.Split(addrs, ";")
	} else {
		addrSlice = append(addrSlice, addrs)
	}

	if len(addrSlice) > 0 {
		for _, addr := range addrSlice {
			ips = append(ips, iputil.AddrToList(addr)...)
		}
	}
	return
}
