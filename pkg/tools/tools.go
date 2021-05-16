package tools

import (
	"fmt"
	"net"
	"strings"
)

// GetNetIPList ...
func GetNetIPList(ipListStr string) (ipList []net.IP, err error) {
	ipList = make([]net.IP, 0)
	ipStrs := strings.Split(ipListStr, ",")
	for _, ipStr := range ipStrs {
		netIP := net.ParseIP(ipStr)
		if netIP == nil {
			err = fmt.Errorf("IP is invalid: ip=%v", ipStr)
			return
		}
		ipList = append(ipList, netIP)
	}
	return
}
