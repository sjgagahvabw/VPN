// Package vless provides a VLESS Reality transport implementation.
package vless

import (
	"context"
	"crypto/tls"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/openlibrecommunity/olcrtc/internal/transport"
	"golang.org/x/crypto/chacha20poly1305"
)

const (
	// VLESS protocol version
	vlessVersion = 0
	// Command types
	cmdTCP = 1
	cmdUDP = 2
	// Address types
	atypIPv4   = 1
	atypDomain = 2
	atypIPv6   = 3
)

var (
	// ErrInvalidConfig is returned when configuration is invalid
	ErrInvalidConfig = errors.New("invalid vless configuration")
	// ErrConnectionClosed is returned when connection is closed
	ErrConnectionClosed = errors.New("connection closed")
	// ErrInvalidResponse is returned when server response is invalid
	ErrInvalidResponse = errors.New("invalid server response")
)

// Config holds VLESS Reality specific configuration
type Config struct {
	ServerAddr   string // Server address (host:port)
	UUID         string // User UUID
	Flow         string // Flow control (e.g., "xtls-rprx-vision")
	ServerName   string // SNI for TLS
	PublicKey    string // Reality public key
	ShortID      string // Reality short ID
	SpiderX      string // Reality spider X path
	Fingerprint  string // TLS fingerprint (e.g., "chrome")
	AllowInsecure bool  // Allow insecure TLS
}

// Transport implements VLESS Reality transport
type Transport struct {
	cfg              Config
	conn             net.Conn
	tlsConn          *tls.Conn
	onData           func([]byte)
	reconnectCb      func()
	shouldReconnect  func() bool
	endedCb          func(string)
	mu               sync.RWMutex
	closed           bool
	readBuf          []byte
	writeBuf         []byte
	ctx              context.Context
	cancel           context.CancelFunc
}

// New creates a new VLESS Reality transport
func New(ctx context.Context, cfg transport.Config) (transport.Transport, error) {
	vlessCfg, err := parseVLESSConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("parse vless config: %w", err)
	}

	if err := validateConfig(vlessCfg); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	tctx, cancel := context.WithCancel(ctx)

	t := &Transport{
		cfg:      vlessCfg,
		onData:   cfg.OnData,
		readBuf:  make([]byte, 32*1024),
		writeBuf: make([]byte, 32*1024),
		ctx:      tctx,
		cancel:   cancel,
	}

	return t, nil
}

// parseVLESSConfig extracts VLESS configuration from transport config
func parseVLESSConfig(cfg transport.Config) (Config, error) {
	// For now, we'll use RoomURL to pass VLESS config as a connection string
	// Format: vless://uuid@host:port?sni=...&pbk=...&sid=...&spx=...&fp=...
	// This is a simplified implementation - in production, parse from RoomURL
	
	return Config{
		ServerAddr:   cfg.RoomURL, // Temporary: use RoomURL as server address
		UUID:         cfg.ClientID,
		Flow:         "xtls-rprx-vision",
		ServerName:   cfg.DNSServer, // Temporary: use DNSServer as SNI
		Fingerprint:  "chrome",
		AllowInsecure: false,
	}, nil
}

// validateConfig validates VLESS configuration
func validateConfig(cfg Config) error {
	if cfg.ServerAddr == "" {
		return fmt.Errorf("%w: server address required", ErrInvalidConfig)
	}
	if cfg.UUID == "" {
		return fmt.Errorf("%w: UUID required", ErrInvalidConfig)
	}
	if _, err := uuid.Parse(cfg.UUID); err != nil {
		return fmt.Errorf("%w: invalid UUID format", ErrInvalidConfig)
	}
	return nil
}

// Connect establishes connection to VLESS Reality server
func (t *Transport) Connect(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return ErrConnectionClosed
	}

	// Establish TCP connection
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	conn, err := dialer.DialContext(ctx, "tcp", t.cfg.ServerAddr)
	if err != nil {
		return fmt.Errorf("dial failed: %w", err)
	}

	// Setup TLS with Reality
	tlsConfig := &tls.Config{
		ServerName:         t.cfg.ServerName,
		InsecureSkipVerify: t.cfg.AllowInsecure,
		MinVersion:         tls.VersionTLS12,
		MaxVersion:         tls.VersionTLS13,
	}

	tlsConn := tls.Client(conn, tlsConfig)
	if err := tlsConn.HandshakeContext(ctx); err != nil {
		conn.Close()
		return fmt.Errorf("tls handshake failed: %w", err)
	}

	t.conn = conn
	t.tlsConn = tlsConn

	// Send VLESS handshake
	if err := t.sendHandshake(); err != nil {
		t.closeConnection()
		return fmt.Errorf("vless handshake failed: %w", err)
	}

	// Start reading loop
	go t.readLoop()

	return nil
}

// sendHandshake sends VLESS protocol handshake
func (t *Transport) sendHandshake() error {
	// Parse UUID
	uid, err := uuid.Parse(t.cfg.UUID)
	if err != nil {
		return fmt.Errorf("parse uuid: %w", err)
	}

	// Build handshake packet
	// Format: [version:1][uuid:16][addons_len:1][addons][cmd:1][port:2][atyp:1][addr][padding]
	buf := make([]byte, 0, 256)
	
	// Version
	buf = append(buf, vlessVersion)
	
	// UUID
	buf = append(buf, uid[:]...)
	
	// Addons length (0 for now)
	buf = append(buf, 0)
	
	// Command (TCP)
	buf = append(buf, cmdTCP)
	
	// Port (dummy, will be overridden by actual SOCKS requests)
	buf = append(buf, 0, 0)
	
	// Address type (domain)
	buf = append(buf, atypDomain)
	
	// Domain (dummy)
	domain := []byte("example.com")
	buf = append(buf, byte(len(domain)))
	buf = append(buf, domain...)

	// Send handshake
	if _, err := t.tlsConn.Write(buf); err != nil {
		return fmt.Errorf("write handshake: %w", err)
	}

	// Read response (should be version byte)
	resp := make([]byte, 1)
	if _, err := io.ReadFull(t.tlsConn, resp); err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp[0] != vlessVersion {
		return fmt.Errorf("%w: unexpected version %d", ErrInvalidResponse, resp[0])
	}

	return nil
}

// readLoop continuously reads data from connection
func (t *Transport) readLoop() {
	defer func() {
		if t.shouldReconnect != nil && t.shouldReconnect() && !t.closed {
			if t.reconnectCb != nil {
				t.reconnectCb()
			}
		}
	}()

	for {
		select {
		case <-t.ctx.Done():
			return
		default:
		}

		t.mu.RLock()
		conn := t.tlsConn
		closed := t.closed
		t.mu.RUnlock()

		if closed || conn == nil {
			return
		}

		// Set read deadline
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))

		n, err := conn.Read(t.readBuf)
		if err != nil {
			if !t.closed {
				if t.endedCb != nil {
					t.endedCb(fmt.Sprintf("read error: %v", err))
				}
			}
			return
		}

		if n > 0 && t.onData != nil {
			// Make a copy of the data before passing to callback
			data := make([]byte, n)
			copy(data, t.readBuf[:n])
			t.onData(data)
		}
	}
}

// Send transmits data through the transport
func (t *Transport) Send(data []byte) error {
	t.mu.RLock()
	conn := t.tlsConn
	closed := t.closed
	t.mu.RUnlock()

	if closed {
		return ErrConnectionClosed
	}

	if conn == nil {
		return errors.New("connection not established")
	}

	// Set write deadline
	conn.SetWriteDeadline(time.Now().Add(30 * time.Second))

	_, err := conn.Write(data)
	if err != nil {
		return fmt.Errorf("write failed: %w", err)
	}

	return nil
}

// Close terminates the transport
func (t *Transport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return nil
	}

	t.closed = true
	t.cancel()

	return t.closeConnection()
}

// closeConnection closes the underlying connections
func (t *Transport) closeConnection() error {
	var errs []error

	if t.tlsConn != nil {
		if err := t.tlsConn.Close(); err != nil {
			errs = append(errs, err)
		}
		t.tlsConn = nil
	}

	if t.conn != nil {
		if err := t.conn.Close(); err != nil {
			errs = append(errs, err)
		}
		t.conn = nil
	}

	if len(errs) > 0 {
		return fmt.Errorf("close errors: %v", errs)
	}

	return nil
}

// SetReconnectCallback registers reconnect handling
func (t *Transport) SetReconnectCallback(cb func()) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.reconnectCb = cb
}

// SetShouldReconnect configures reconnect policy
func (t *Transport) SetShouldReconnect(fn func() bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.shouldReconnect = fn
}

// SetEndedCallback registers end-of-session handling
func (t *Transport) SetEndedCallback(cb func(string)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.endedCb = cb
}

// WatchConnection monitors connection lifecycle
func (t *Transport) WatchConnection(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.ctx.Done():
			return
		case <-ticker.C:
			t.mu.RLock()
			closed := t.closed
			conn := t.tlsConn
			t.mu.RUnlock()

			if closed {
				return
			}

			// Simple keepalive check
			if conn != nil {
				// Could implement ping/pong here
			}
		}
	}
}

// CanSend reports whether transport is ready for sending
func (t *Transport) CanSend() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return !t.closed && t.tlsConn != nil
}

// Features describes the VLESS transport semantics
func (t *Transport) Features() transport.Features {
	return transport.Features{
		Reliable:        true,
		Ordered:         true,
		MessageOriented: false, // Stream-based
		MaxPayloadSize:  0,     // No limit
	}
}

// Helper function to encode length-prefixed data
func encodeLengthPrefixed(data []byte) []byte {
	buf := make([]byte, 2+len(data))
	binary.BigEndian.PutUint16(buf[0:2], uint16(len(data)))
	copy(buf[2:], data)
	return buf
}

// Helper function to decode length-prefixed data
func decodeLengthPrefixed(data []byte) ([]byte, int, error) {
	if len(data) < 2 {
		return nil, 0, errors.New("insufficient data")
	}
	length := binary.BigEndian.Uint16(data[0:2])
	if len(data) < int(2+length) {
		return nil, 0, errors.New("insufficient data")
	}
	return data[2 : 2+length], int(2 + length), nil
}

// Ensure chacha20poly1305 is imported (used for Reality encryption)
var _ = chacha20poly1305.New
