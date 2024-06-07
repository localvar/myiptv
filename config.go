package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
)

// Channel is IPTV channel in configuration
type Channel struct {
	// name of the channel, such as 'CCTV-1'
	Name string `json:"name"`

	// display name of the channel, such as 'CCTV-1综合'
	DisplayName string `json:"displayName,omitempty"`

	// logo is the logo of the channel
	Logo string `json:"logo,omitempty"`

	// hiden channels are not shown in the channel list
	Hide bool `json:"hide,omitempty"`

	// sources of the channel, if a source does NOT begin with 'http',
	// MyIPTV regards it as a multicast address.
	Sources []string `json:"sources,omitempty"`
}

// ChannelGroup is a group of IPTV channels
type ChannelGroup struct {
	// name of the channel group, such as 'CCTV' or '央视'
	Name string `json:"name"`

	// channels in the group
	Channels []Channel `json:"channels,omitempty"`
}

// Config defines MyIPTV configuration
type Config struct {
	// HTTP server address, including IP address and port
	ServerAddr string `json:"serverAddr,omitempty"`

	// name of the multicast interface
	MulticastInterface string `json:"multicastInterface,omitempty"`

	// EPG URL, default is 'http://epg.51zmt.top:8000/e.xml'
	EPGURL string `json:"epgURL,omitempty"`

	// MCastBufferSize is the buffer size for one multicast packet,
	// default is 2048
	MCastBufferSize int `json:"mCastPacketSize,omitempty"`

	// WriteBufferSize is the buffer size, which is used to buffer the
	// multicast packets before writing to clients, default is 131072
	WriteBufferSize int `json:"writeBufferSize,omitempty"`
}

// Clourflare DDNS configuration
type DDNSConfig struct {
	// RecordName is the domain name to update, for example: 'blog.localvar.cn'
	RecordName string `json:"recordName"`

	// ZoneID is the zone ID of Cloudflare
	ZoneID string `json:"zoneID"`

	// APIKey is the Cloudflare APIKey
	APIKey string `json:"apiKey"`

	// WANIPProviders is a list of URLs to get the WAN IP address, for example:
	//
	//     ["http://ipv4.icanhazip.com", "https://whatismyip.akamai.com"]
	//
	// Note if the provider returns an IPV4 address, the program will update
	// the A record of the domain name, if the provider returns an IPV6 address,
	// the program will update the AAAA record.
	WANIPProviders []string `json:"wanIPProviders"`

	// DNSServers is an optional list of DNS servers to resolve the domain name,
	// the servers must include the port number (typically 53), if set, the
	// program will use these DNS servers to speed up the resolution, for
	// example: ["beth.ns.cloudflare.com:53", "rudy.ns.cloudflare.com:53"]
	DNSServers []string `json:"dnsServers,omitempty"`
}

// findBestIP finds the best IP address from the http server
func findBestIP(m map[string][]string) string {
	first := ""
	for _, ips := range m {
		for _, ip := range ips {
			// we use the first IP by default
			if first == "" {
				first = ip
			}
			// but if there's a 192.168.x.x IP, we use it
			if strings.HasPrefix(ip, "192.168.") {
				return ip
			}
		}
	}
	return first
}

// populateDefault populates default values to the configuration.
func (cfg *Config) populateDefault() {
	if cfg.EPGURL == "" {
		cfg.EPGURL = "http://epg.51zmt.top:8000/e.xml"
	}

	if cfg.MCastBufferSize == 0 {
		cfg.MCastBufferSize = 2048
	}

	if cfg.WriteBufferSize == 0 {
		cfg.WriteBufferSize = 131072
	}

	if cfg.ServerAddr != "" && cfg.MulticastInterface != "" {
		return
	}

	// next, try to find the best network interface and IP address
	// for the multicast interface and the http server

	m := GetInterfacesAndIPs()

	// If no network interface is found and server address is not configured
	// use '0.0.0.0:7709'. Note this may not work, and even if the http server
	// could be started, the generated IPTV relay addresses may not be
	// accessible from other devices.
	if len(m) == 0 {
		if cfg.ServerAddr == "" {
			cfg.ServerAddr = "0.0.0.0:7709"
		}
		return
	}

	if len(m) == 1 {
		if cfg.ServerAddr == "" {
			cfg.ServerAddr = findBestIP(m) + ":7709"
		} else if cfg.MulticastInterface == "" {
			for iface := range m {
				cfg.MulticastInterface = iface
			}
		}
		return
	}

	// if multicast interface is configured, remove it so that the http server
	// won't use it
	if cfg.MulticastInterface != "" {
		delete(m, cfg.MulticastInterface)
	}

	// try configure the http server address
	if cfg.ServerAddr == "" {
		cfg.ServerAddr = findBestIP(m) + ":7709"
	}

	// try configure the multicast interface
	if cfg.MulticastInterface == "" {
	IFACE_LOOP:
		for iface, ips := range m {
			for _, ip := range ips {
				// remove the interface if it has the http server address
				if strings.HasPrefix(cfg.ServerAddr, ip+":") {
					delete(m, iface)
					break IFACE_LOOP
				}
			}
		}
		for iface := range m {
			cfg.MulticastInterface = iface
			break
		}
	}
}

var (
	// configPath is the path of the configuration file, it may be updated
	// in function 'loadConfig'.
	configPath = "myiptv.json"

	// ddnsConfig is the Cloudflare DDNS configuration,
	// it won't be changed after the program starts.
	ddnsConfig *DDNSConfig = nil

	// config is the current configuration, it's an atomic value which
	// points to a Config object.
	config atomic.Value

	channelGroups []ChannelGroup

	// though we use atomic.Value to store the configuration, we still need
	// the lock, because we need to update the configuration file.
	configLock sync.Mutex
)

// getConfig returns the current configuration.
func getConfig() *Config {
	return config.Load().(*Config)
}

// getDDNSConfig returns the DDNS configuration.
func getDDNSConfig() *DDNSConfig {
	return ddnsConfig
}

// allConfig is a helper structure to help load and save configuration,
// because we use a single file to store both configuration and channel
// groups.
type allConfig struct {
	DDNS          *DDNSConfig    `json:"ddns,omitempty"`
	Config        *Config        `json:"config"`
	ChannelGroups []ChannelGroup `json:"channelGroups,omitempty"`
}

// loadConfig loads configuration from 'myiptv.json', it ignores all errors,
// default values are used if there's an error, or any item is missing.
func loadConfig() {
	// try to find the configuration file in the current directory, if not
	// found, try to find it in the directory of the executable.
	if fi, err := os.Stat(configPath); err != nil || fi.IsDir() {
		if exePath, err := os.Executable(); err == nil {
			path := filepath.Join(filepath.Dir(exePath), configPath)
			if fi, err = os.Stat(path); err == nil && !fi.IsDir() {
				configPath = path
			}
		} else {
			slog.Error(
				"failed to get executable path",
				slog.String("error", err.Error()),
			)
		}
	}

	var allCfg allConfig
	if f, err := os.Open(configPath); err == nil {
		if err = json.NewDecoder(f).Decode(&allCfg); err != nil {
			slog.Error(
				"failed to decode configuration file",
				slog.String("error", err.Error()),
			)
		}
		f.Close()
	} else {
		slog.Error(
			"failed to open configuration file",
			slog.String("error", err.Error()),
		)
	}

	cfg := allCfg.Config
	if cfg == nil {
		cfg = new(Config)
	}
	cfg.populateDefault()
	config.Store(cfg)

	// this function is only called at program startup, no need to lock
	channelGroups = allCfg.ChannelGroups
	ddnsConfig = allCfg.DDNS
}

// saveConfig saves the configuration to 'configPath'.
func saveConfig(cfg *Config, chGrps []ChannelGroup) error {
	allCfg := allConfig{
		DDNS:          ddnsConfig,
		Config:        cfg,
		ChannelGroups: chGrps,
	}

	data, err := json.Marshal(&allCfg)
	if err != nil {
		slog.Error(
			"failed to marshal config",
			slog.String("error", err.Error()),
		)
		return err
	}

	err = os.WriteFile(configPath, data, 0666)
	if err != nil {
		slog.Error(
			"failed to write config to file",
			slog.String("error", err.Error()),
		)
		return err
	}

	return nil
}

// apigetConfig returns the current configuration
func apigetConfig(w http.ResponseWriter, r *http.Request) {
	cfg := getConfig()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cfg)
}

// apiUpdateConfig updates and saves the configuration
func apiUpdateConfig(w http.ResponseWriter, r *http.Request) {
	var cfg Config
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		slog.Error(
			"failed to decode request body",
			slog.String("error", err.Error()),
		)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	configLock.Lock()
	defer configLock.Unlock()

	if err := saveConfig(&cfg, channelGroups); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// call populateDefault after saving the configuration, because we don't
	// want to save the populated default values.
	cfg.populateDefault()
	config.Store(&cfg)
}

// channelGroupForEach iterates all channel groups and calls the function.
func channelGroupForEach(fn func(*ChannelGroup)) {
	configLock.Lock()
	defer configLock.Unlock()

	for i := range channelGroups {
		fn(&channelGroups[i])
	}
}

// apiListChannelGroups lists all channel groups
func apiListChannelGroups(w http.ResponseWriter, r *http.Request) {
	_ = r

	configLock.Lock()
	defer configLock.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(channelGroups)
}

// apiUpdateChannelGroup updates the channel group
func apiUpdateChannelGroups(w http.ResponseWriter, r *http.Request) {
	var chGrps []ChannelGroup

	if err := json.NewDecoder(r.Body).Decode(&chGrps); err != nil {
		slog.Error(
			"failed to decode request body",
			slog.String("error", err.Error()),
		)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	configLock.Lock()
	defer configLock.Unlock()

	if err := saveConfig(getConfig(), chGrps); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	channelGroups = chGrps
}
