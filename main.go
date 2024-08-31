package main

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"
)

//go:embed all:webui/dist
var website embed.FS

// set by the build script
var Version string

var shutdown func(restart bool)

func run() {
	cfg := getConfig()
	srv := &http.Server{Addr: cfg.ServerAddr}

	shutdown = func(restart bool) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		srv.Shutdown(ctx)
		cancel()
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
		slog.String("multicastInterface", cfg.McastIface),
	)
}

func main() {
	// if running as a service, disable log timestamp as the service manager
	// will add it
	if os.Getppid() == 1 {
		log.SetFlags(0)
	}

	slog.Info("MyIPTV", slog.String("version", Version))

	loadConfig()
	initDDNS()

	// the website
	dist, _ := fs.Sub(website, "webui/dist")
	http.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		upath := r.URL.Path
		if !strings.HasPrefix(upath, "/") {
			upath = "/" + upath
			r.URL.Path = upath
		}
		upath = path.Clean(upath)

		if f, err := http.FS(dist).Open(upath); err == nil {
			f.Close()
		} else if errors.Is(err, fs.ErrNotExist) {
			upath = "/"
		}

		http.ServeFileFS(w, r, dist, upath)
	})

	// a simple file server for the 'file' directory as an add-on feature
	http.Handle("GET /file/", http.FileServer(http.Dir("")))

	// APIs to manage the relay server
	http.HandleFunc("POST /api/restart", apiRestart)
	http.HandleFunc("GET /api/interfaces-and-ips", apiListInterfacesAndIPs)

	http.HandleFunc("GET /api/config", apiGetConfig)
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
