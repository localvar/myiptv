package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Programme struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
	Title string    `json:"title"`
	Desc  string    `json:"desc"`
}

var epgLock sync.Mutex
var lastEPGUpdateTime time.Time
var epgs map[string][]Programme

// doUpdateEGP fetches EPG data from the configured EPG URL
func doUpdateEPG() error {
	cfg := getConfig()
	resp, err := http.Get(cfg.EPGURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	newEPGs := make(map[string][]Programme)
	channelGroupForEach(func(group *ChannelGroup) {
		for _, ch := range group.Channels {
			newEPGs[ch.Name] = nil
		}
	})
	id2name := make(map[string]string)

	now := time.Now()
	today := Date(now)

	decoder := xml.NewDecoder(resp.Body)
	for {
		var st xml.StartElement
		if t, err := decoder.Token(); err == io.EOF {
			break
		} else if err != nil {
			slog.Warn(
				"failed to get token",
				slog.String("error", err.Error()),
			)
			continue
		} else if tt, ok := t.(xml.StartElement); !ok {
			continue
		} else {
			st = tt
		}

		if st.Name.Local == "channel" {
			var ch struct {
				ID   string `xml:"id,attr"`
				Name string `xml:"display-name"`
			}
			if err = decoder.DecodeElement(&ch, &st); err != nil {
				slog.Warn(
					"failed to decode channel",
					slog.String("error", err.Error()),
				)
				continue
			}
			if _, ok := newEPGs[ch.Name]; ok {
				id2name[ch.ID] = ch.Name
			}
			continue
		}

		if st.Name.Local == "programme" {
			var p struct {
				Channel string `xml:"channel,attr"`
				Start   string `xml:"start,attr"`
				Stop    string `xml:"stop,attr"`
				Title   string `xml:"title"`
				Desc    string `xml:"desc"`
			}

			if err = decoder.DecodeElement(&p, &st); err != nil {
				slog.Warn(
					"failed to decode programme",
					slog.String("error", err.Error()),
				)
				continue
			}

			chName := id2name[p.Channel]
			progs, ok := newEPGs[chName]
			if !ok {
				continue
			}

			start, err := time.Parse("20060102150405 -0700", p.Start)
			if err != nil {
				slog.Warn(
					"cannot parse programme start time",
					slog.String("startTime", p.Start),
					slog.String("error", err.Error()),
				)
				continue
			}

			end, err := time.Parse("20060102150405 -0700", p.Stop)
			if err != nil {
				slog.Warn(
					"cannot parse programme stop time",
					slog.String("stopTime", p.Stop),
					slog.String("error", err.Error()),
				)
				continue
			}

			if end.Before(today) {
				continue
			}

			newEPGs[chName] = append(progs, Programme{
				Start: start,
				End:   end,
				Title: p.Title,
				Desc:  p.Desc,
			})
		}
	}

	epgs = newEPGs
	lastEPGUpdateTime = now
	return nil
}

func updateEPG(force bool) error {
	epgLock.Lock()
	defer epgLock.Unlock()

	if !force && time.Since(lastEPGUpdateTime) < 3*time.Hour {
		return nil
	}

	err := doUpdateEPG()
	if err == nil {
		slog.Info("EPG has been updated")
	} else {
		slog.Error("failed to update EPG", slog.String("error", err.Error()))
	}
	return err
}

func iptvGetEPG(w http.ResponseWriter, r *http.Request) {
	ch := r.URL.Query().Get("ch")
	if ch == "" {
		http.Error(w, "missing channel name", http.StatusBadRequest)
		return
	}
	date := strings.ReplaceAll(r.URL.Query().Get("date"), "-", "")
	if date == "" {
		http.Error(w, "missing date", http.StatusBadRequest)
		return
	}
	start, err := time.ParseInLocation("20060102", date, time.Local)
	if err != nil {
		http.Error(w, "invalid date", http.StatusBadRequest)
		return
	}
	end := start.AddDate(0, 0, 1)

	updateEPG(false)

	epgLock.Lock()
	allProgs := epgs[ch]
	epgLock.Unlock()

	// TODO: format := r.URL.Query().Get("fmt")

	type progJSONHelper struct {
		Start string `json:"start"`
		End   string `json:"end"`
		Title string `json:"title"`
		Desc  string `json:"desc"`
	}

	var progs []progJSONHelper
	for _, p := range allProgs {
		if p.Start.Before(end) && p.End.After(start) {
			progs = append(progs, progJSONHelper{
				Start: p.Start.Format("15:04"),
				End:   p.End.Format("15:04"),
				Title: p.Title,
				Desc:  p.Desc,
			})
		}
	}
	if len(progs) == 0 {
		http.Error(w, "no programme data", http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	fmt.Fprintf(w, `{"channel_name":"%s","date":"%s","epg_data":`, ch, date)

	json.NewEncoder(w).Encode(progs)
	w.Write([]byte("}"))
}

func apiGetEPG(w http.ResponseWriter, r *http.Request) {
	ch := r.PathValue("channel")
	if ch == "" {
		http.Error(w, "missing channel name", http.StatusBadRequest)
		return
	}

	var start time.Time
	if d := strings.ReplaceAll(r.URL.Query().Get("date"), "-", ""); d != "" {
		tm, err := time.ParseInLocation("20060102", d, time.Local)
		if err != nil {
			http.Error(w, "invalid date", http.StatusBadRequest)
			return
		}
		start = tm
	}

	updateEPG(false)

	epgLock.Lock()
	progs := epgs[ch]
	epgLock.Unlock()

	if !start.IsZero() {
		end := start.AddDate(0, 0, 1)

		var fprogs []Programme
		for _, p := range progs {
			if p.Start.Before(end) && p.End.After(start) {
				fprogs = append(fprogs, p)
			}
		}
		progs = fprogs
	}

	if len(progs) == 0 {
		http.Error(w, "no programme data", http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	json.NewEncoder(w).Encode(progs)
}

func apiUpdateEPG(w http.ResponseWriter, r *http.Request) {
	err := updateEPG(true)
	if err != nil {
		msg := "failed to update EPG: " + err.Error()
		http.Error(w, msg, http.StatusInternalServerError)
	}
}
