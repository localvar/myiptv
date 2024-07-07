package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net"
	"net/http"
	"net/netip"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// extractPayload extracts the payload from the RTP packet
func extractPayload(pkt []byte) []byte {
	const (
		SigMPEGTS = 0x47
		MPEGAudio = 0x0E
		MPEGVideo = 0x20
		MPEGTS    = 0x21
	)

	// not a RTP packet
	if pkt[0] == SigMPEGTS {
		return pkt
	}

	// version
	if ((pkt[0] & 0xC0) >> 6) != 0x02 {
		slog.Debug("invalid RTP version")
		return nil
	}

	// minimal RTP header size if 12 bytes
	hdrLen := 12
	// add length of CSRC
	hdrLen += 4 * int(pkt[0]&0x0F)
	// add fixed length of extention headers
	if pkt[0]&0x10 != 0 {
		hdrLen += 4
	}

	// packet too short
	if len(pkt) < hdrLen {
		slog.Debug("RTP packet too short")
		return nil
	}

	// payload type
	switch pkt[1] & 0x7F {
	case MPEGAudio, MPEGVideo:
		// add profile-based length: adopted from vlc 0.8.6 code
		hdrLen += 4
	case MPEGTS:
	default:
		slog.Debug("unknown payload type")
		return nil
	}

	// add length of extention headers
	if pkt[0]&0x10 != 0 {
		hdrLen += 4 * ((int(pkt[14]) << 8) + int(pkt[14+1]))
	}

	// unknown payload
	if pkt[hdrLen] != SigMPEGTS {
		slog.Debug("unknown payload")
		return nil
	}

	// remove padding
	if pkt[0]&0x20 != 0 {
		padLen := int(pkt[len(pkt)-1])
		if hdrLen+padLen > len(pkt) {
			slog.Debug("invalid padding")
			return nil
		}
		pkt = pkt[:len(pkt)-padLen]
	}

	// remove header
	return pkt[hdrLen:]
}

// relayBuffer is a buffer for relaying multicast IPTV channel
type relayBuffer struct {
	ref atomic.Int32
	buf []byte
}

func newRelayBuffer() *relayBuffer {
	rb := relayBufPool.Get().(*relayBuffer)
	rb.ref.Store(1)
	rb.buf = rb.buf[:0]
	return rb
}

func (rb *relayBuffer) AddRef() {
	rb.ref.Add(1)
}

func (rb *relayBuffer) Release() {
	if rb.ref.Add(-1) == 0 {
		relayBufPool.Put(rb)
	}
}

// relayBufPool is the buffer pool for relay, to improve performance
var relayBufPool = sync.Pool{
	New: func() any {
		rb := relayBuffer{}
		rb.buf = make([]byte, 0, getConfig().McastPacketSize)
		return &rb
	},
}

type relayClient struct {
	addr      string
	ch        chan *relayBuffer
	createdAt time.Time
	cancel    context.CancelFunc
}

func (rc *relayClient) send(rb *relayBuffer) {
	rb.AddRef()
	select {
	case rc.ch <- rb:
	default:
		rb.Release()
	}
}

type mcastConn struct {
	addr       string
	conn       *net.UDPConn
	createdAt  time.Time
	clientLock sync.Mutex
	clients    []*relayClient
	ctx        context.Context
	cancel     context.CancelFunc
}

var mcastConns = sync.Map{}

func (mc *mcastConn) getClients() []*relayClient {
	mc.clientLock.Lock()
	defer mc.clientLock.Unlock()
	result := make([]*relayClient, len(mc.clients))
	copy(result, mc.clients)
	return result
}

func (mc *mcastConn) addClient(rc *relayClient) {
	mc.clientLock.Lock()
	defer mc.clientLock.Unlock()
	for i, c := range mc.clients {
		if c == nil {
			mc.clients[i] = rc
			return
		}
	}
	mc.clients = append(mc.clients, rc)
}

func (mc *mcastConn) closeClient(addr string) {
	mc.clientLock.Lock()
	defer mc.clientLock.Unlock()
	for _, rc := range mc.clients {
		if rc != nil && rc.addr == addr {
			rc.cancel()
			break
		}
	}
}

func (mc *mcastConn) removeClient(addr string) {
	mc.clientLock.Lock()
	defer mc.clientLock.Unlock()
	for i, rc := range mc.clients {
		if rc != nil && rc.addr == addr {
			mc.clients[i] = nil
			break
		}
	}
}

func (mc *mcastConn) sendToClients(rb *relayBuffer) int {
	count := 0
	mc.clientLock.Lock()
	defer mc.clientLock.Unlock()
	for _, rc := range mc.clients {
		if rc != nil {
			rc.send(rb)
			count++
		}
	}
	return count
}

func (mc *mcastConn) receive(bufSize int, timeout time.Duration) {
	rbuf := make([]byte, bufSize)
	wbuf := newRelayBuffer()

	mc.conn.SetReadDeadline(time.Now().Add(timeout))

LOOP:
	for {
		n, err := mc.conn.Read(rbuf)
		if err != nil {
			slog.Error(
				"failed to read from multicast connection",
				slog.String("address", mc.addr),
				slog.String("error", err.Error()),
			)
			break
		}
		if n == 0 {
			continue
		}

		p := extractPayload(rbuf[:n])

		if len(wbuf.buf)+len(p) > cap(wbuf.buf) {
			if mc.sendToClients(wbuf) == 0 {
				// all clients are gone
				break
			}
			wbuf.Release()
			wbuf = newRelayBuffer()
			mc.conn.SetReadDeadline(time.Now().Add(timeout))
		}

		wbuf.buf = append(wbuf.buf, p...)

		select {
		case <-mc.ctx.Done():
			break LOOP
		default:
		}
	}

	mcastConns.Delete(mc.addr)
	mc.cancel()
	wbuf.Release()
	mc.conn.Close()
	slog.Info("multicast connection closed", slog.String("address", mc.addr))
}

// mcastConnect establishes a connection according to the request
func mcastConnect(w http.ResponseWriter, r *http.Request) (*mcastConn, bool) {
	addr := r.PathValue("addr")
	if v, _ := mcastConns.Load(addr); v != nil {
		return v.(*mcastConn), false
	}

	ap, err := netip.ParseAddrPort(addr)
	if err != nil {
		slog.Error(
			"failed to parse multicast address",
			slog.String("address", addr),
			slog.String("error", err.Error()),
		)
		http.Error(w, "invalid multicast address", http.StatusBadRequest)
		return nil, false
	}
	udpAddr := net.UDPAddrFromAddrPort(ap)

	cfg := getConfig()
	iface, err := net.InterfaceByName(cfg.McastIface)
	if err != nil {
		slog.Error(
			"failed to get network interface",
			slog.String("interface", cfg.McastIface),
			slog.String("error", err.Error()),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil, false
	}

	conn, err := net.ListenMulticastUDP("udp", iface, udpAddr)
	if err != nil {
		slog.Error(
			"failed to listen to multicast address",
			slog.String("address", addr),
			slog.String("error", err.Error()),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil, false
	}

	ctx, cancel := context.WithCancel(context.Background())
	mc := &mcastConn{
		addr:      addr,
		conn:      conn,
		createdAt: time.Now(),
		ctx:       ctx,
		cancel:    cancel,
	}

	// reuse existing connection if any
	if v, ok := mcastConns.LoadOrStore(addr, mc); ok {
		conn.Close()
		cancel()
		return v.(*mcastConn), false
	}

	slog.Info("multicast connection established", slog.String("address", addr))
	return mc, true
}

// iptvRelayrelays a multicast IPTV channel to HTTP
func iptvRelay(w http.ResponseWriter, r *http.Request) {
	mc, created := mcastConnect(w, r)
	if mc == nil {
		return
	}

	ch := make(chan *relayBuffer, 16)
	ctx, cancel := context.WithCancel(mc.ctx)
	mc.addClient(&relayClient{
		addr:      r.RemoteAddr,
		ch:        ch,
		createdAt: time.Now(),
		cancel:    cancel,
	})
	slog.Info(
		"relay client added",
		slog.String("multicastAddress", mc.addr),
		slog.String("clientAddress", r.RemoteAddr),
	)

	if created {
		// in this case, must call addClient before receive, or the receive
		// goroutine may exit immediately because there is no client
		cfg := getConfig()
		go mc.receive(cfg.McastPacketSize, time.Second)
	}

	w.Header().Set("Content-Type", "video/MP2T")
RELAY_LOOP:
	for {
		select {
		case rb := <-ch:
			_, err := w.Write(rb.buf)
			rb.Release()
			if err != nil {
				errstr := err.Error()
				if !strings.HasSuffix(errstr, " write: broken pipe") {
					slog.Error(
						"failed to write to client",
						slog.String("multicastAddress", mc.addr),
						slog.String("error", errstr),
					)
				}
				break RELAY_LOOP
			}
		case <-ctx.Done():
			break RELAY_LOOP
		}
	}

	mc.removeClient(r.RemoteAddr)
	cancel()
	slog.Info(
		"relay client removed",
		slog.String("multicastAddress", mc.addr),
		slog.String("clientAddress", r.RemoteAddr),
	)
}

// apiListRelays lists all connections and clients
func apiListRelays(w http.ResponseWriter, r *http.Request) {
	type Client struct {
		Addr      string    `json:"addr"`
		CreatedAt time.Time `json:"createdAt"`
	}

	type Conn struct {
		Addr      string    `json:"addr"`
		CreatedAt time.Time `json:"createdAt"`
		Clients   []Client  `json:"clients"`
	}

	result := make([]Conn, 0, 8)
	mcastConns.Range(func(k, v any) bool {
		mc := v.(*mcastConn)
		conn := Conn{
			Addr:      mc.addr,
			CreatedAt: mc.createdAt,
			Clients:   make([]Client, 0, 4),
		}

		for _, rc := range mc.getClients() {
			if rc == nil {
				continue
			}
			conn.Clients = append(conn.Clients, Client{
				Addr:      rc.addr,
				CreatedAt: rc.createdAt,
			})
		}

		result = append(result, conn)
		return true
	})

	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	json.NewEncoder(w).Encode(result)
}

// apiCloseRelayConnection closes a relay connection (multicast connection),
// this will also close all of its clients
func apiCloseRelayConnection(w http.ResponseWriter, r *http.Request) {
	addr := r.PathValue("addr")
	if v, ok := mcastConns.Load(addr); ok {
		v.(*mcastConn).cancel()
	}
}

// apiCloseRelayClient closes a relay client of a relay connection
func apiCloseRelayClient(w http.ResponseWriter, r *http.Request) {
	addr := r.PathValue("addr")
	client := r.PathValue("client")
	if v, ok := mcastConns.Load(addr); ok {
		v.(*mcastConn).closeClient(client)
	}
}
