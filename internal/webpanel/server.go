// Package webpanel provides web-based management interface for olcRTC.
package webpanel

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/openlibrecommunity/olcrtc/internal/logger"
)

// Server represents the web panel HTTP server
type Server struct {
	addr       string
	httpServer *http.Server
	mu         sync.RWMutex
	configs    map[string]*TunnelConfig
	stats      map[string]*ConnectionStats
	ctx        context.Context
	cancel     context.CancelFunc
}

// TunnelConfig represents a tunnel configuration
type TunnelConfig struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Mode          string    `json:"mode"` // "srv" or "cnc"
	Transport     string    `json:"transport"` // "vless", "datachannel", etc.
	Enabled       bool      `json:"enabled"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	
	// VLESS specific
	VLESSConfig   *VLESSConfig `json:"vless_config,omitempty"`
	
	// WebRTC specific
	WebRTCConfig  *WebRTCConfig `json:"webrtc_config,omitempty"`
	
	// Common
	ClientID      string `json:"client_id"`
	KeyHex        string `json:"key_hex"`
	SOCKSHost     string `json:"socks_host,omitempty"`
	SOCKSPort     int    `json:"socks_port,omitempty"`
	SOCKSUser     string `json:"socks_user,omitempty"`
	SOCKSPass     string `json:"socks_pass,omitempty"`
	DNSServer     string `json:"dns_server,omitempty"`
}

// VLESSConfig holds VLESS Reality configuration
type VLESSConfig struct {
	ServerAddr    string `json:"server_addr"`
	UUID          string `json:"uuid"`
	Flow          string `json:"flow"`
	ServerName    string `json:"server_name"`
	PublicKey     string `json:"public_key"`
	ShortID       string `json:"short_id"`
	SpiderX       string `json:"spider_x"`
	Fingerprint   string `json:"fingerprint"`
	AllowInsecure bool   `json:"allow_insecure"`
}

// WebRTCConfig holds WebRTC configuration
type WebRTCConfig struct {
	Carrier         string `json:"carrier"`
	RoomID          string `json:"room_id"`
	VideoWidth      int    `json:"video_width,omitempty"`
	VideoHeight     int    `json:"video_height,omitempty"`
	VideoFPS        int    `json:"video_fps,omitempty"`
	VideoBitrate    string `json:"video_bitrate,omitempty"`
}

// ConnectionStats represents connection statistics
type ConnectionStats struct {
	TunnelID      string    `json:"tunnel_id"`
	Status        string    `json:"status"` // "connected", "disconnected", "error"
	BytesSent     uint64    `json:"bytes_sent"`
	BytesReceived uint64    `json:"bytes_received"`
	Connections   int       `json:"connections"`
	LastSeen      time.Time `json:"last_seen"`
	Uptime        int64     `json:"uptime"` // seconds
	StartedAt     time.Time `json:"started_at"`
}

// NewServer creates a new web panel server
func NewServer(addr string) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	
	s := &Server{
		addr:    addr,
		configs: make(map[string]*TunnelConfig),
		stats:   make(map[string]*ConnectionStats),
		ctx:     ctx,
		cancel:  cancel,
	}
	
	mux := http.NewServeMux()
	
	// API endpoints
	mux.HandleFunc("/api/configs", s.handleConfigs)
	mux.HandleFunc("/api/configs/", s.handleConfigByID)
	mux.HandleFunc("/api/stats", s.handleStats)
	mux.HandleFunc("/api/stats/", s.handleStatsByID)
	mux.HandleFunc("/api/tunnels/start", s.handleStartTunnel)
	mux.HandleFunc("/api/tunnels/stop", s.handleStopTunnel)
	mux.HandleFunc("/api/generate-config", s.handleGenerateConfig)
	mux.HandleFunc("/api/export", s.handleExportConfig)
	mux.HandleFunc("/api/import", s.handleImportConfig)
	mux.HandleFunc("/api/subscription", s.handleSubscription)
	
	// Static files (web UI)
	mux.HandleFunc("/", s.handleIndex)
	
	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      s.corsMiddleware(mux),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	
	return s
}

// Start starts the web panel server
func (s *Server) Start() error {
	logger.Infof("Starting web panel on %s", s.addr)
	
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Errorf("Web panel server error: %v", err)
		}
	}()
	
	return nil
}

// Stop stops the web panel server
func (s *Server) Stop() error {
	s.cancel()
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown error: %w", err)
	}
	
	logger.Infof("Web panel stopped")
	return nil
}

// corsMiddleware adds CORS headers
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// handleConfigs handles GET /api/configs and POST /api/configs
func (s *Server) handleConfigs(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.getConfigs(w, r)
	case http.MethodPost:
		s.createConfig(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// getConfigs returns all configurations
func (s *Server) getConfigs(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	configs := make([]*TunnelConfig, 0, len(s.configs))
	for _, cfg := range s.configs {
		configs = append(configs, cfg)
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"configs": configs,
	})
}

// createConfig creates a new configuration
func (s *Server) createConfig(w http.ResponseWriter, r *http.Request) {
	var cfg TunnelConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}
	
	// Generate ID if not provided
	if cfg.ID == "" {
		cfg.ID = fmt.Sprintf("tunnel-%d", time.Now().Unix())
	}
	
	cfg.CreatedAt = time.Now()
	cfg.UpdatedAt = time.Now()
	
	s.mu.Lock()
	s.configs[cfg.ID] = &cfg
	s.mu.Unlock()
	
	logger.Infof("Created tunnel config: %s", cfg.ID)
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(cfg)
}

// handleConfigByID handles GET/PUT/DELETE /api/configs/{id}
func (s *Server) handleConfigByID(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/api/configs/"):]
	
	switch r.Method {
	case http.MethodGet:
		s.getConfig(w, r, id)
	case http.MethodPut:
		s.updateConfig(w, r, id)
	case http.MethodDelete:
		s.deleteConfig(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// getConfig returns a specific configuration
func (s *Server) getConfig(w http.ResponseWriter, r *http.Request, id string) {
	s.mu.RLock()
	cfg, exists := s.configs[id]
	s.mu.RUnlock()
	
	if !exists {
		http.Error(w, "Config not found", http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cfg)
}

// updateConfig updates a configuration
func (s *Server) updateConfig(w http.ResponseWriter, r *http.Request, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	cfg, exists := s.configs[id]
	if !exists {
		http.Error(w, "Config not found", http.StatusNotFound)
		return
	}
	
	var updates TunnelConfig
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}
	
	// Update fields
	updates.ID = id
	updates.CreatedAt = cfg.CreatedAt
	updates.UpdatedAt = time.Now()
	
	s.configs[id] = &updates
	
	logger.Infof("Updated tunnel config: %s", id)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updates)
}

// deleteConfig deletes a configuration
func (s *Server) deleteConfig(w http.ResponseWriter, r *http.Request, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if _, exists := s.configs[id]; !exists {
		http.Error(w, "Config not found", http.StatusNotFound)
		return
	}
	
	delete(s.configs, id)
	delete(s.stats, id)
	
	logger.Infof("Deleted tunnel config: %s", id)
	
	w.WriteHeader(http.StatusNoContent)
}

// handleStats handles GET /api/stats
func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	stats := make([]*ConnectionStats, 0, len(s.stats))
	for _, stat := range s.stats {
		stats = append(stats, stat)
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"stats": stats,
	})
}

// handleStatsByID handles GET /api/stats/{id}
func (s *Server) handleStatsByID(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/api/stats/"):]
	
	s.mu.RLock()
	stat, exists := s.stats[id]
	s.mu.RUnlock()
	
	if !exists {
		http.Error(w, "Stats not found", http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stat)
}

// handleStartTunnel handles POST /api/tunnels/start
func (s *Server) handleStartTunnel(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TunnelID string `json:"tunnel_id"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}
	
	s.mu.RLock()
	cfg, exists := s.configs[req.TunnelID]
	s.mu.RUnlock()
	
	if !exists {
		http.Error(w, "Tunnel not found", http.StatusNotFound)
		return
	}
	
	// TODO: Actually start the tunnel
	logger.Infof("Starting tunnel: %s", req.TunnelID)
	
	// Update stats
	s.mu.Lock()
	s.stats[req.TunnelID] = &ConnectionStats{
		TunnelID:  req.TunnelID,
		Status:    "connected",
		StartedAt: time.Now(),
		LastSeen:  time.Now(),
	}
	s.mu.Unlock()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "started",
		"tunnel":  cfg,
	})
}

// handleStopTunnel handles POST /api/tunnels/stop
func (s *Server) handleStopTunnel(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TunnelID string `json:"tunnel_id"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}
	
	// TODO: Actually stop the tunnel
	logger.Infof("Stopping tunnel: %s", req.TunnelID)
	
	// Update stats
	s.mu.Lock()
	if stat, exists := s.stats[req.TunnelID]; exists {
		stat.Status = "disconnected"
		stat.LastSeen = time.Now()
	}
	s.mu.Unlock()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "stopped",
	})
}

// handleIndex serves the web UI
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	
	// Serve embedded HTML
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(indexHTML))
}

// UpdateStats updates connection statistics
func (s *Server) UpdateStats(tunnelID string, stats *ConnectionStats) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	stats.TunnelID = tunnelID
	stats.LastSeen = time.Now()
	s.stats[tunnelID] = stats
}

const indexHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>olcRTC Web Panel</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #0f172a; color: #e2e8f0; }
        .container { max-width: 1200px; margin: 0 auto; padding: 20px; }
        h1 { margin-bottom: 30px; color: #60a5fa; }
        .card { background: #1e293b; border-radius: 8px; padding: 20px; margin-bottom: 20px; box-shadow: 0 4px 6px rgba(0,0,0,0.3); }
        .btn { padding: 10px 20px; border: none; border-radius: 6px; cursor: pointer; font-size: 14px; transition: all 0.2s; }
        .btn-primary { background: #3b82f6; color: white; }
        .btn-primary:hover { background: #2563eb; }
        .btn-success { background: #10b981; color: white; }
        .btn-danger { background: #ef4444; color: white; }
        .tunnel-list { display: grid; gap: 15px; }
        .tunnel-item { background: #334155; padding: 15px; border-radius: 6px; display: flex; justify-content: space-between; align-items: center; }
        .tunnel-info h3 { color: #60a5fa; margin-bottom: 5px; }
        .tunnel-info p { color: #94a3b8; font-size: 14px; }
        .status { display: inline-block; padding: 4px 12px; border-radius: 12px; font-size: 12px; font-weight: 600; }
        .status-connected { background: #10b981; color: white; }
        .status-disconnected { background: #6b7280; color: white; }
        .form-group { margin-bottom: 15px; }
        .form-group label { display: block; margin-bottom: 5px; color: #94a3b8; }
        .form-group input, .form-group select { width: 100%; padding: 10px; border: 1px solid #475569; background: #1e293b; color: #e2e8f0; border-radius: 6px; }
        .actions { display: flex; gap: 10px; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🚀 olcRTC Web Panel</h1>
        
        <div class="card">
            <h2>Tunnels</h2>
            <div id="tunnels" class="tunnel-list"></div>
            <button class="btn btn-primary" onclick="showAddTunnel()" style="margin-top: 15px;">+ Add Tunnel</button>
        </div>
        
        <div id="addTunnelForm" class="card" style="display: none;">
            <h2>Add New Tunnel</h2>
            <form onsubmit="addTunnel(event)">
                <div class="form-group">
                    <label>Name</label>
                    <input type="text" id="name" required>
                </div>
                <div class="form-group">
                    <label>Transport</label>
                    <select id="transport" onchange="toggleTransportConfig()">
                        <option value="vless">VLESS Reality</option>
                        <option value="datachannel">WebRTC DataChannel</option>
                        <option value="videochannel">WebRTC VideoChannel</option>
                    </select>
                </div>
                <div id="vlessConfig">
                    <div class="form-group">
                        <label>Server Address</label>
                        <input type="text" id="serverAddr" placeholder="example.com:443">
                    </div>
                    <div class="form-group">
                        <label>UUID</label>
                        <input type="text" id="uuid" placeholder="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx">
                    </div>
                    <div class="form-group">
                        <label>Server Name (SNI)</label>
                        <input type="text" id="serverName" placeholder="example.com">
                    </div>
                </div>
                <div class="actions">
                    <button type="submit" class="btn btn-success">Create</button>
                    <button type="button" class="btn btn-danger" onclick="hideAddTunnel()">Cancel</button>
                </div>
            </form>
        </div>
    </div>
    
    <script>
        async function loadTunnels() {
            const res = await fetch('/api/configs');
            const data = await res.json();
            const container = document.getElementById('tunnels');
            container.innerHTML = data.configs.map(t => \`
                <div class="tunnel-item">
                    <div class="tunnel-info">
                        <h3>\${t.name}</h3>
                        <p>Transport: \${t.transport} | ID: \${t.id}</p>
                    </div>
                    <div class="actions">
                        <span class="status status-\${t.enabled ? 'connected' : 'disconnected'}">\${t.enabled ? 'Active' : 'Inactive'}</span>
                        <button class="btn btn-success" onclick="startTunnel('\${t.id}')">Start</button>
                        <button class="btn btn-danger" onclick="stopTunnel('\${t.id}')">Stop</button>
                    </div>
                </div>
            \`).join('');
        }
        
        function showAddTunnel() {
            document.getElementById('addTunnelForm').style.display = 'block';
        }
        
        function hideAddTunnel() {
            document.getElementById('addTunnelForm').style.display = 'none';
        }
        
        function toggleTransportConfig() {
            const transport = document.getElementById('transport').value;
            document.getElementById('vlessConfig').style.display = transport === 'vless' ? 'block' : 'none';
        }
        
        async function addTunnel(e) {
            e.preventDefault();
            const transport = document.getElementById('transport').value;
            const config = {
                name: document.getElementById('name').value,
                transport: transport,
                mode: 'cnc',
                enabled: false
            };
            
            if (transport === 'vless') {
                config.vless_config = {
                    server_addr: document.getElementById('serverAddr').value,
                    uuid: document.getElementById('uuid').value,
                    server_name: document.getElementById('serverName').value,
                    flow: 'xtls-rprx-vision',
                    fingerprint: 'chrome'
                };
            }
            
            await fetch('/api/configs', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(config)
            });
            
            hideAddTunnel();
            loadTunnels();
        }
        
        async function startTunnel(id) {
            await fetch('/api/tunnels/start', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ tunnel_id: id })
            });
            loadTunnels();
        }
        
        async function stopTunnel(id) {
            await fetch('/api/tunnels/stop', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ tunnel_id: id })
            });
            loadTunnels();
        }
        
        loadTunnels();
        setInterval(loadTunnels, 5000);
    </script>
</body>
</html>`
