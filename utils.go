package main

import (
	"log/slog"
	"net"
	"time"
)

// Date returns the date part of a time.Time
func Date(tm time.Time) time.Time {
	y, m, d := tm.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.Local)
}

// GetInterfacesAndIPs returns a map of network interfaces and their
// IP addresses. This function ignores all errors.
func GetInterfacesAndIPs() map[string][]string {
	m := make(map[string][]string)

	ifaces, err := net.Interfaces()
	if err != nil {
		slog.Error(
			"failed to get network interfaces",
			slog.String("error", err.Error()),
		)
		return m
	}

	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			slog.Error(
				"failed to get addresses of network interface",
				slog.String("interfaceName", iface.Name),
				slog.String("error", err.Error()),
			)
			continue
		}
		ips := make([]string, 0, len(addrs))
		for _, addr := range addrs {
			ipnet, ok := addr.(*net.IPNet)
			if !ok || ipnet.IP.IsLoopback() {
				continue
			}
			ip := ipnet.IP.To4()
			if ip == nil {
				continue
			}
			ips = append(ips, ip.String())
		}
		if len(ips) > 0 {
			m[iface.Name] = ips
		}
	}

	return m
}
