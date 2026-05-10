// Package vless provides VLESS Reality server implementation.
package vless

import (
	"context"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/openlibrecommunity/olcrtc/internal/logger"
)

// ServerConfig holds VLESS server configuration
type ServerConfig struct {
	ListenAddr    string            // Listen address (e.g., ":443")
	UUID          string            // Server UUID for authentication
	TLSCertFile   string            // TLS certificate file path
	TLSKeyFile    string            // TLS key file path
	DestAddr      string            // Destination address for forwarding
	AllowedUUIDs  map[string]bool   // Map of allowed client UUIDs
	RealityConfig *RealityConfig    // Reality-specific configuration
}

// RealityConfig holds Reality protocol configuration
type RealityConfig struct {
	PrivateKey  string   // Reality private key
	ShortIDs    []string // Reality short IDs
	ServerNames []string // Allowed server names
	Dest        string   // Reality destination (fallback)
}

// Server implements VLESS Reality server
type Server struct {
	cfg      ServerConfig
	listener net.Listener
	mu       sync.RWMutex
	clients  map[string]*ClientConn
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

// ClientConn represents a client connection
type ClientConn struct {
	conn     net.Conn
	tlsConn  *tls.Conn
	uuid     uuid.UUID
	destConn net.Conn
	mu       sync.Mutex
	closed   bool
}

// NewServer creates a new VLESS Reality server
func NewServer(cfg ServerConfig) (*Server, error) {
	if err := validateServerConfig(cfg); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	s := &Server{
		cfg:     cfg,
		clients: make(map[string]*ClientConn),
		ctx:     ctx,
		cancel:  cancel,
	}

	return s, nil
}

// validateServerConfig validates server configuration
func validateServerConfig(cfg ServerConfig) error {
	if cfg.ListenAddr == "" {
		return fmt.Errorf("listen address required")
	}
	if cfg.UUID == "" {
		return fmt.Errorf("UUID required")
	}
	if _, err := uuid.Parse(cfg.UUID); err != nil {
		return fmt.Errorf("invalid UUID format: %w", err)
	}
	if cfg.TLSCertFile == "" || cfg.TLSKeyFile == "" {
		return fmt.Errorf("TLS certificate and key required")
	}
	return nil
}

// Start starts the VLESS server
func (s *Server) Start() error {
	// Load TLS certificate
	cert, err := tls.LoadX509KeyPair(s.cfg.TLSCertFile, s.cfg.TLSKeyFile)
	if err != nil {
		return fmt.Errorf("load tls cert: %w", err)
	}

	// Setup TLS config
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
		MaxVersion:   tls.VersionTLS13,
	}

	// Create listener
	listener, err := tls.Listen("tcp", s.cfg.ListenAddr, tlsConfig)
	if err != nil {
		return fmt.Errorf("listen failed: %w", err)
	}

	s.listener = listener
	logger.Infof("VLESS Reality server listening on %s", s.cfg.ListenAddr)

	// Start accept loop
	s.wg.Add(1)
	go s.acceptLoop()

	return nil
}

// acceptLoop accepts incoming connections
func (s *Server) acceptLoop() {
	defer s.wg.Done()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.ctx.Done():
				return
			default:
				logger.Warnf("Accept error: %v", err)
				continue
			}
		}

		s.wg.Add(1)
		go s.handleConnection(conn)
	}
}

// handleConnection handles a client connection
func (s *Server) handleConnection(conn net.Conn) {
	defer s.wg.Done()
	defer conn.Close()

	// Cast to TLS connection
	tlsConn, ok := conn.(*tls.Conn)
	if !ok {
		logger.Warnf("Not a TLS connection")
		return
	}

	// Read VLESS handshake
	clientUUID, cmd, addr, port, err := s.readHandshake(tlsConn)
	if err != nil {
		logger.Warnf("Handshake failed: %v", err)
		return
	}

	// Validate UUID
	if !s.isUUIDAllowed(clientUUID) {
		logger.Warnf("Unauthorized UUID: %s", clientUUID.String())
		return
	}

	logger.Infof("Client connected: %s, cmd=%d, target=%s:%d", clientUUID.String(), cmd, addr, port)

	// Send response
	if err := s.sendResponse(tlsConn); err != nil {
		logger.Warnf("Send response failed: %v", err)
		return
	}

	// Handle command
	switch cmd {
	case cmdTCP:
		s.handleTCP(tlsConn, addr, port, clientUUID)
	case cmdUDP:
		logger.Warnf("UDP not implemented yet")
	default:
		logger.Warnf("Unknown command: %d", cmd)
	}
}

// readHandshake reads and parses VLESS handshake
func (s *Server) readHandshake(conn *tls.Conn) (uuid.UUID, byte, string, uint16, error) {
	// Read version
	version := make([]byte, 1)
	if _, err := io.ReadFull(conn, version); err != nil {
		return uuid.UUID{}, 0, "", 0, fmt.Errorf("read version: %w", err)
	}

	if version[0] != vlessVersion {
		return uuid.UUID{}, 0, "", 0, fmt.Errorf("unsupported version: %d", version[0])
	}

	// Read UUID
	uuidBytes := make([]byte, 16)
	if _, err := io.ReadFull(conn, uuidBytes); err != nil {
		return uuid.UUID{}, 0, "", 0, fmt.Errorf("read uuid: %w", err)
	}

	clientUUID, err := uuid.FromBytes(uuidBytes)
	if err != nil {
		return uuid.UUID{}, 0, "", 0, fmt.Errorf("parse uuid: %w", err)
	}

	// Read addons length
	addonsLen := make([]byte, 1)
	if _, err := io.ReadFull(conn, addonsLen); err != nil {
		return uuid.UUID{}, 0, "", 0, fmt.Errorf("read addons len: %w", err)
	}

	// Skip addons
	if addonsLen[0] > 0 {
		addons := make([]byte, addonsLen[0])
		if _, err := io.ReadFull(conn, addons); err != nil {
			return uuid.UUID{}, 0, "", 0, fmt.Errorf("read addons: %w", err)
		}
	}

	// Read command
	cmd := make([]byte, 1)
	if _, err := io.ReadFull(conn, cmd); err != nil {
		return uuid.UUID{}, 0, "", 0, fmt.Errorf("read cmd: %w", err)
	}

	// Read port
	portBytes := make([]byte, 2)
	if _, err := io.ReadFull(conn, portBytes); err != nil {
		return uuid.UUID{}, 0, "", 0, fmt.Errorf("read port: %w", err)
	}
	port := binary.BigEndian.Uint16(portBytes)

	// Read address type
	atyp := make([]byte, 1)
	if _, err := io.ReadFull(conn, atyp); err != nil {
		return uuid.UUID{}, 0, "", 0, fmt.Errorf("read atyp: %w", err)
	}

	// Read address
	var addr string
	switch atyp[0] {
	case atypIPv4:
		ipBytes := make([]byte, 4)
		if _, err := io.ReadFull(conn, ipBytes); err != nil {
			return uuid.UUID{}, 0, "", 0, fmt.Errorf("read ipv4: %w", err)
		}
		addr = net.IP(ipBytes).String()

	case atypDomain:
		lenByte := make([]byte, 1)
		if _, err := io.ReadFull(conn, lenByte); err != nil {
			return uuid.UUID{}, 0, "", 0, fmt.Errorf("read domain len: %w", err)
		}
		domainBytes := make([]byte, lenByte[0])
		if _, err := io.ReadFull(conn, domainBytes); err != nil {
			return uuid.UUID{}, 0, "", 0, fmt.Errorf("read domain: %w", err)
		}
		addr = string(domainBytes)

	case atypIPv6:
		ipBytes := make([]byte, 16)
		if _, err := io.ReadFull(conn, ipBytes); err != nil {
			return uuid.UUID{}, 0, "", 0, fmt.Errorf("read ipv6: %w", err)
		}
		addr = net.IP(ipBytes).String()

	default:
		return uuid.UUID{}, 0, "", 0, fmt.Errorf("unsupported address type: %d", atyp[0])
	}

	return clientUUID, cmd[0], addr, port, nil
}

// sendResponse sends VLESS response
func (s *Server) sendResponse(conn *tls.Conn) error {
	// Response format: [version:1][addons_len:1][addons]
	resp := []byte{vlessVersion, 0} // version + no addons
	if _, err := conn.Write(resp); err != nil {
		return fmt.Errorf("write response: %w", err)
	}
	return nil
}

// handleTCP handles TCP proxy
func (s *Server) handleTCP(clientConn *tls.Conn, addr string, port uint16, clientUUID uuid.UUID) {
	// Use configured destination or the requested address
	destAddr := s.cfg.DestAddr
	if destAddr == "" {
		destAddr = fmt.Sprintf("%s:%d", addr, port)
	}

	// Connect to destination
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	destConn, err := dialer.Dial("tcp", destAddr)
	if err != nil {
		logger.Warnf("Dial destination failed: %v", err)
		return
	}
	defer destConn.Close()

	logger.Infof("Proxying %s -> %s", clientUUID.String(), destAddr)

	// Bidirectional copy
	errCh := make(chan error, 2)

	go func() {
		_, err := io.Copy(destConn, clientConn)
		errCh <- err
	}()

	go func() {
		_, err := io.Copy(clientConn, destConn)
		errCh <- err
	}()

	// Wait for either direction to finish
	<-errCh
}

// isUUIDAllowed checks if UUID is allowed
func (s *Server) isUUIDAllowed(clientUUID uuid.UUID) bool {
	// If no allowed UUIDs configured, allow all
	if len(s.cfg.AllowedUUIDs) == 0 {
		return true
	}

	return s.cfg.AllowedUUIDs[clientUUID.String()]
}

// Stop stops the server
func (s *Server) Stop() error {
	s.cancel()

	if s.listener != nil {
		if err := s.listener.Close(); err != nil {
			logger.Warnf("Close listener error: %v", err)
		}
	}

	// Close all client connections
	s.mu.Lock()
	for _, client := range s.clients {
		client.Close()
	}
	s.mu.Unlock()

	// Wait for all goroutines
	s.wg.Wait()

	logger.Infof("VLESS server stopped")
	return nil
}

// Close closes a client connection
func (c *ClientConn) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true

	if c.destConn != nil {
		c.destConn.Close()
	}

	if c.tlsConn != nil {
		c.tlsConn.Close()
	}

	if c.conn != nil {
		c.conn.Close()
	}

	return nil
}
