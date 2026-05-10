// Package webpanel provides enhanced web-based management interface for olcRTC.
package webpanel

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/openlibrecommunity/olcrtc/internal/logger"
)

// handleGenerateConfig generates client configurations for various platforms
func (s *Server) handleGenerateConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		TunnelID string `json:"tunnel_id"`
		Platform string `json:"platform"` // "windows", "macos", "linux", "android", "ios", "url"
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

	var config string
	var contentType string

	switch req.Platform {
	case "url":
		config = s.generateVLESSURL(cfg)
		contentType = "text/plain"
	case "json":
		configJSON, _ := json.MarshalIndent(cfg, "", "  ")
		config = string(configJSON)
		contentType = "application/json"
	case "qrcode":
		config = s.generateQRCodeData(cfg)
		contentType = "text/plain"
	case "windows", "macos", "linux":
		config = s.generateShellScript(cfg, req.Platform)
		contentType = "text/plain"
	case "android", "ios":
		config = s.generateVLESSURL(cfg)
		contentType = "text/plain"
	default:
		http.Error(w, "Unsupported platform", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.Write([]byte(config))
}

// generateVLESSURL generates VLESS URL for import
func (s *Server) generateVLESSURL(cfg *TunnelConfig) string {
	if cfg.VLESSConfig == nil {
		return ""
	}

	v := cfg.VLESSConfig
	params := url.Values{}
	params.Add("encryption", "none")
	params.Add("flow", v.Flow)
	params.Add("security", "tls")
	params.Add("sni", v.ServerName)
	params.Add("fp", v.Fingerprint)
	params.Add("type", "tcp")
	params.Add("headerType", "none")

	if v.PublicKey != "" {
		params.Add("pbk", v.PublicKey)
	}
	if v.ShortID != "" {
		params.Add("sid", v.ShortID)
	}
	if v.SpiderX != "" {
		params.Add("spx", v.SpiderX)
	}

	// Parse server address
	host := v.ServerAddr
	if strings.Contains(host, ":") {
		parts := strings.Split(host, ":")
		host = parts[0]
	}

	return fmt.Sprintf("vless://%s@%s?%s#%s",
		v.UUID,
		v.ServerAddr,
		params.Encode(),
		url.QueryEscape(cfg.Name),
	)
}

// generateQRCodeData generates data for QR code
func (s *Server) generateQRCodeData(cfg *TunnelConfig) string {
	return s.generateVLESSURL(cfg)
}

// generateShellScript generates shell script for client setup
func (s *Server) generateShellScript(cfg *TunnelConfig, platform string) string {
	var script strings.Builder

	if platform == "windows" {
		script.WriteString("@echo off\r\n")
		script.WriteString("REM olcRTC Client Configuration\r\n")
		script.WriteString("REM Generated for Windows\r\n\r\n")
		script.WriteString("olcrtc.exe ^\r\n")
	} else {
		script.WriteString("#!/bin/bash\n")
		script.WriteString("# olcRTC Client Configuration\n")
		script.WriteString(fmt.Sprintf("# Generated for %s\n\n", platform))
		script.WriteString("./olcrtc \\\n")
	}

	lineSep := " \\\n"
	if platform == "windows" {
		lineSep = " ^\r\n"
	}

	script.WriteString(fmt.Sprintf("  -mode %s%s", cfg.Mode, lineSep))
	script.WriteString(fmt.Sprintf("  -link direct%s", lineSep))
	script.WriteString(fmt.Sprintf("  -transport %s%s", cfg.Transport, lineSep))
	script.WriteString(fmt.Sprintf("  -carrier telemost%s", lineSep))

	if cfg.VLESSConfig != nil {
		script.WriteString(fmt.Sprintf("  -id \"%s\"%s", cfg.VLESSConfig.ServerAddr, lineSep))
		script.WriteString(fmt.Sprintf("  -client-id \"%s\"%s", cfg.VLESSConfig.UUID, lineSep))
		script.WriteString(fmt.Sprintf("  -dns \"%s\"%s", cfg.VLESSConfig.ServerName, lineSep))
	}

	if cfg.SOCKSHost != "" {
		script.WriteString(fmt.Sprintf("  -socks-host \"%s\"%s", cfg.SOCKSHost, lineSep))
	}
	if cfg.SOCKSPort > 0 {
		script.WriteString(fmt.Sprintf("  -socks-port %d%s", cfg.SOCKSPort, lineSep))
	}
	if cfg.SOCKSUser != "" {
		script.WriteString(fmt.Sprintf("  -socks-user \"%s\"%s", cfg.SOCKSUser, lineSep))
	}
	if cfg.SOCKSPass != "" {
		script.WriteString(fmt.Sprintf("  -socks-pass \"%s\"%s", cfg.SOCKSPass, lineSep))
	}

	script.WriteString(fmt.Sprintf("  -key \"%s\"%s", cfg.KeyHex, lineSep))
	script.WriteString("  -data ./data\n")

	return script.String()
}

// handleExportConfig exports configuration in various formats
func (s *Server) handleExportConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	format := r.URL.Query().Get("format")
	tunnelID := r.URL.Query().Get("tunnel_id")

	if tunnelID == "" {
		http.Error(w, "tunnel_id required", http.StatusBadRequest)
		return
	}

	s.mu.RLock()
	cfg, exists := s.configs[tunnelID]
	s.mu.RUnlock()

	if !exists {
		http.Error(w, "Tunnel not found", http.StatusNotFound)
		return
	}

	switch format {
	case "json":
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.json", cfg.ID))
		json.NewEncoder(w).Encode(cfg)

	case "url":
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.txt", cfg.ID))
		w.Write([]byte(s.generateVLESSURL(cfg)))

	case "qr":
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(s.generateQRCodeData(cfg)))

	case "script":
		platform := r.URL.Query().Get("platform")
		if platform == "" {
			platform = "linux"
		}
		w.Header().Set("Content-Type", "text/plain")
		ext := ".sh"
		if platform == "windows" {
			ext = ".bat"
		}
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s%s", cfg.ID, ext))
		w.Write([]byte(s.generateShellScript(cfg, platform)))

	default:
		http.Error(w, "Unsupported format", http.StatusBadRequest)
	}
}

// handleImportConfig imports configuration from various formats
func (s *Server) handleImportConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	format := r.URL.Query().Get("format")
	
	switch format {
	case "json":
		var cfg TunnelConfig
		if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
			http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
			return
		}

		if cfg.ID == "" {
			cfg.ID = fmt.Sprintf("tunnel-%d", len(s.configs)+1)
		}

		s.mu.Lock()
		s.configs[cfg.ID] = &cfg
		s.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cfg)

	case "url":
		var req struct {
			URL string `json:"url"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
			return
		}

		cfg, err := s.parseVLESSURL(req.URL)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid VLESS URL: %v", err), http.StatusBadRequest)
			return
		}

		s.mu.Lock()
		s.configs[cfg.ID] = cfg
		s.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cfg)

	default:
		http.Error(w, "Unsupported format", http.StatusBadRequest)
	}
}

// parseVLESSURL parses VLESS URL into TunnelConfig
func (s *Server) parseVLESSURL(vlessURL string) (*TunnelConfig, error) {
	if !strings.HasPrefix(vlessURL, "vless://") {
		return nil, fmt.Errorf("invalid VLESS URL scheme")
	}

	// Remove scheme
	vlessURL = strings.TrimPrefix(vlessURL, "vless://")

	// Split by @
	parts := strings.SplitN(vlessURL, "@", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid VLESS URL format")
	}

	uuid := parts[0]

	// Split address and params
	addrParts := strings.SplitN(parts[1], "?", 2)
	serverAddr := addrParts[0]

	// Parse fragment (name)
	name := "Imported Tunnel"
	if strings.Contains(parts[1], "#") {
		nameParts := strings.SplitN(parts[1], "#", 2)
		if len(nameParts) == 2 {
			name, _ = url.QueryUnescape(nameParts[1])
		}
	}

	// Parse query parameters
	params := url.Values{}
	if len(addrParts) == 2 {
		queryPart := strings.Split(addrParts[1], "#")[0]
		params, _ = url.ParseQuery(queryPart)
	}

	cfg := &TunnelConfig{
		ID:        fmt.Sprintf("imported-%d", len(s.configs)+1),
		Name:      name,
		Transport: "vless",
		Mode:      "cnc",
		Enabled:   false,
		VLESSConfig: &VLESSConfig{
			ServerAddr:  serverAddr,
			UUID:        uuid,
			Flow:        params.Get("flow"),
			ServerName:  params.Get("sni"),
			PublicKey:   params.Get("pbk"),
			ShortID:     params.Get("sid"),
			SpiderX:     params.Get("spx"),
			Fingerprint: params.Get("fp"),
		},
		SOCKSHost: "127.0.0.1",
		SOCKSPort: 1080,
	}

	return cfg, nil
}

// generateSubscription generates subscription URL for multiple configs
func (s *Server) generateSubscription() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var urls []string
	for _, cfg := range s.configs {
		if cfg.Transport == "vless" && cfg.VLESSConfig != nil {
			urls = append(urls, s.generateVLESSURL(cfg))
		}
	}

	// Join with newlines and base64 encode
	subscription := strings.Join(urls, "\n")
	encoded := base64.StdEncoding.EncodeToString([]byte(subscription))

	return encoded
}

// handleSubscription handles subscription endpoint
func (s *Server) handleSubscription(w http.ResponseWriter, r *http.Request) {
	subscription := s.generateSubscription()
	
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Subscription-Userinfo", fmt.Sprintf("upload=0; download=0; total=0; expire=0"))
	w.Header().Set("Profile-Update-Interval", "24")
	w.Header().Set("Profile-Title", "olcRTC")
	
	w.Write([]byte(subscription))
}
