package main

import (
	"fmt"
	"net/http"
	"strings"
)

// writeSourceURL writes the URL of a source to the response writer
func writeSourceURL(w http.ResponseWriter, svrAddr, src string) {
	if strings.HasPrefix(strings.ToLower(src), "http") {
		fmt.Fprintln(w, src)
	} else {
		fmt.Fprintf(w, "http://%s/iptv/relay/%s\n", svrAddr, src)
	}
}

// listChannelsInM3U8 lists all IPTV channels in M3U8 format
func listChannelsInM3U8(w http.ResponseWriter, _ *http.Request) {
	cfg := getConfig()

	w.Header().Set("Content-Type", "application/x-mpegURL;charset=UTF-8")

	id := 1
	fmt.Fprintln(w, "#EXTM3U")

	channelGroupForEach(func(group *ChannelGroup) {
		for _, ch := range group.Channels {
			if ch.Hide || len(ch.Sources) == 0 {
				continue
			}

			dn := ch.DisplayName
			if dn == "" {
				dn = ch.Name
			}

			fmt.Fprintf(w,
				`#EXTINF:-1 tvg-id="%d" tvg-name="%s" tvg-logo="%s" group-title="%s",%s`,
				id,
				ch.Name,
				ch.Logo,
				group.Name,
				dn)
			fmt.Fprintln(w)

			writeSourceURL(w, cfg.ServerAddr, ch.Sources[0])
			id++
		}
	})
}

// listChannelsInText lists all IPTV channels in text format, DIYP style
func listChannelsInText(w http.ResponseWriter, _ *http.Request) {
	cfg := getConfig()

	w.Header().Set("Content-Type", "text/plain;charset=UTF-8")

	channelGroupForEach(func(group *ChannelGroup) {
		fmt.Fprintf(w, "%s,#genre#\n", group.Name)

		for _, ch := range group.Channels {
			if ch.Hide || len(ch.Sources) == 0 {
				continue
			}

			for _, src := range ch.Sources {
				fmt.Fprint(w, ch.Name, ",")
				writeSourceURL(w, cfg.ServerAddr, src)
			}
		}

		fmt.Fprintln(w)
	})
}

// iptvListChannels lists all IPTV channels in required format
func iptvListChannels(w http.ResponseWriter, r *http.Request) {
	switch strings.ToLower(r.URL.Query().Get("fmt")) {
	case "m3u", "m3u8":
		listChannelsInM3U8(w, r)
	case "", "txt", "text":
		listChannelsInText(w, r)
	default:
		http.Error(w, "supported format", http.StatusBadRequest)
	}
}
