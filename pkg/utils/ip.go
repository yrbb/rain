package utils

import (
	"encoding/binary"
	"errors"
	"net"
	"strings"
)

func IP(prefix ...string) (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	var pArr []string
	if len(prefix) > 0 {
		pArr = strings.Split(prefix[0], ",")
	}

	host := ""

	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if !ok || ipNet.IP.To4() == nil {
			continue
		}

		ip := ipNet.IP.String()

		if len(pArr) == 0 && !ipNet.IP.IsLoopback() {
			host = ip
			break
		}

		for _, p := range pArr {
			if strings.HasPrefix(ip, p) {
				host = ip
				break
			}
		}

		if host != "" {
			break
		}
	}

	return host, nil
}

func IP2Long(ipAddr string) (uint32, error) {
	ip := net.ParseIP(ipAddr)
	if ip == nil {
		return 0, errors.New("invalid ipAddr format")
	}

	return binary.BigEndian.Uint32(ip.To4()), nil
}

func Long2IP(long uint32) string {
	ipByte := make([]byte, 4)
	binary.BigEndian.PutUint32(ipByte, long)

	return net.IP(ipByte).String()
}
