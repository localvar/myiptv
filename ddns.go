package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/netip"
	"strings"
	"time"
)

func getWANIP() (netip.Addr, error) {
	var err error

	for _, url := range getDDNSConfig().WANIPProviders {
		resp, e := http.Get(url)
		if e != nil {
			err = e
			continue
		}

		data, e := io.ReadAll(resp.Body)
		resp.Body.Close()
		if e != nil {
			err = e
			continue
		}

		data = bytes.Trim(data, "\r\n \t")
		addr, e := netip.ParseAddr(string(data))
		if e != nil {
			err = e
			continue
		}

		return addr, nil
	}
	return netip.Addr{}, err
}

func getDNSIPs(ip6 bool) ([]netip.Addr, error) {
	cfg := getDDNSConfig()

	r := net.Resolver{PreferGo: true}
	if len(cfg.DNSServers) > 0 {
		r.Dial = func(ctx context.Context, n, a string) (net.Conn, error) {
			var (
				err  error
				conn net.Conn
			)
			for _, server := range cfg.DNSServers {
				d := net.Dialer{Timeout: 5 * time.Second}
				conn, err = d.DialContext(ctx, "udp", server)
				if err == nil {
					break
				}
			}
			return conn, err
		}
	}

	network := "ip4"
	if ip6 {
		network = "ip6"
	}

	ips, err := r.LookupIP(context.Background(), network, cfg.RecordName)
	if err != nil {
		return nil, err
	}

	result := make([]netip.Addr, len(ips))
	for i, ip := range ips {
		result[i] = netip.MustParseAddr(ip.String())
	}

	return result, nil
}

func getRecordID() (string, error) {
	cfg := getDDNSConfig()

	// curl "https://api.cloudflare.com/client/v4/zones/$CFZONE_ID/dns_records?name=$CFRECORD_NAME" -H "Authorization: Bearer $CFKEY"
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records?name=%s", cfg.ZoneID, cfg.RecordName)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Success bool `json:"success"`
		Result  []struct {
			ID string `json:"id"`
		} `json:"result"`
		Errors []struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"errors"`
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", err
	}
	if !result.Success {
		return "", errors.New(result.Errors[0].Message)
	}

	return result.Result[0].ID, nil
}

func updateDNS() {
	cfg := getDDNSConfig()

	wanIP, err := getWANIP()
	if err != nil {
		slog.Error("failed to get WAN IP", slog.String("error", err.Error()))
		return
	}

	dnsIPs, err := getDNSIPs(wanIP.Is6())
	if err != nil {
		slog.Error("failed to get DNS IPs", slog.String("error", err.Error()))
		return
	}

	for _, ip := range dnsIPs {
		if ip == wanIP {
			return
		}
	}

	slog.Info(
		"IP address changed, begin updating DNS record",
		slog.String("domain", cfg.RecordName),
		slog.String("new ip", wanIP.String()),
	)

	recordID, err := getRecordID()
	if err != nil {
		slog.Error("failed to get DNS record ID", slog.String("error", err.Error()))
		return
	}

	typ := "A"
	if wanIP.Is6() {
		typ = "AAAA"
	}
	body := fmt.Sprintf(`{"content":"%s","name":"%s","proxied":false,"type":"%s","ttl":1}`, wanIP, cfg.RecordName, typ)

	// "https://api.cloudflare.com/client/v4/zones/$CFZONE_ID/dns_records/$CFRECORD_ID"
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s", cfg.ZoneID, recordID)
	req, err := http.NewRequest(http.MethodPatch, url, strings.NewReader(body))
	if err != nil {
		slog.Debug("failed to create HTTP request", slog.String("error", err.Error()))
		return
	}

	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if resp != nil {
		resp.Body.Close()
	}
	if err == nil {
		slog.Info("DNS record updated successfully")
	} else {
		slog.Error("failed to update DNS record", slog.String("error", err.Error()))
	}
}

func initDDNS() {
	if cfg := getDDNSConfig(); cfg == nil || cfg.RecordName == "" {
		return
	}

	go func() {
		for {
			updateDNS()
			time.Sleep(5 * time.Minute)
		}
	}()
}
