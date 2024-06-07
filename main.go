package main

import (
	"context"
	"embed"
	"encoding/json"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

//go:embed webui/dist
var website embed.FS

var shutdown func(restart bool)

func run() {
	cfg := getConfig()
	srv := &http.Server{Addr: cfg.ServerAddr}

	shutdown = func(restart bool) {
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		srv.Shutdown(ctx)
		slog.Info("relay server stopped")
		if restart {
			run()
		}
	}

	go func(srv *http.Server) {
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			slog.Error(
				"relay server closed unexpectly",
				slog.String("error", err.Error()),
			)
			os.Exit(1)
		}
	}(srv)

	slog.Info(
		"relay server started",
		slog.String("httpAddress", cfg.ServerAddr),
		slog.String("multicastInterface", cfg.MulticastInterface),
	)
}

func main() {
	// if running as a service, disable log timestamp as the service manager
	// will add it
	if os.Getppid() == 1 {
		log.SetFlags(0)
	}

	loadConfig()
	initDDNS()

	// the website
	dist, _ := fs.Sub(website, "webui/dist")
	http.Handle("GET /", http.FileServerFS(dist))

	// APIs to manage the relay server
	http.HandleFunc("POST /api/restart", apiRestart)
	http.HandleFunc("GET /api/interfaces-and-ips", apiListInterfacesAndIPs)

	http.HandleFunc("GET /api/config", apigetConfig)
	http.HandleFunc("PUT /api/config", apiUpdateConfig)

	http.HandleFunc("GET /api/channel-groups", apiListChannelGroups)
	http.HandleFunc("PUT /api/channel-groups", apiUpdateChannelGroups)

	http.HandleFunc("GET /api/epg/{channel}", apiGetEPG)
	http.HandleFunc("POST /api/epg", apiUpdateEPG)

	http.HandleFunc("GET /api/relays", apiListRelays)
	http.HandleFunc("DELETE /api/relays/{addr}", apiCloseRelayConnection)
	http.HandleFunc("DELETE /api/relays/{addr}/{client}", apiCloseRelayClient)

	// for IPTV clients
	http.HandleFunc("GET /iptv/relay/{addr}", iptvRelay)
	http.HandleFunc("GET /iptv/channels", iptvListChannels)
	http.HandleFunc("GET /iptv/epg", iptvGetEPG)

	// run and wait `Ctrl-C` or `Term` to exit
	run()
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	<-signals

	shutdown(false)
}

// apiRestart restarts the relay server
func apiRestart(w http.ResponseWriter, r *http.Request) {
	shutdown(true)
}

// apiListInterfacesAndIPs lists all network interfaces and their IPs
func apiListInterfacesAndIPs(w http.ResponseWriter, r *http.Request) {
	_ = r

	m := GetInterfacesAndIPs()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m)
}
