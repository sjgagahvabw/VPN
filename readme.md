<div align="center">

<img src="https://github.com/openlibrecommunity/material/blob/master/olcrtc.png" width="250" height="250">

![License](https://img.shields.io/badge/license-WTFPL-0D1117?style=flat-square&logo=open-source-initiative&logoColor=green&labelColor=0D1117)
![Golang](https://img.shields.io/badge/-Golang-0D1117?style=flat-square&logo=go&logoColor=00A7D0)

</div>


## About
olcRTC - across the sea

Project that allows users to bypass blocking by parasitizing and tunneling on unblocked and whitelisted services in Russia, use legal webRTC services

**NEW:** Now with VLESS Reality support and Web Management Panel!

## Features

- 🚀 **Multiple Transport Protocols**
  - VLESS Reality - High-speed, low-latency protocol with Reality technology
  - WebRTC DataChannel - Reliable data channel transport
  - WebRTC VideoChannel - Video-based steganography transport
  - WebRTC SEI Channel - SEI metadata transport
  - WebRTC VP8 Channel - VP8 codec transport

- 🎛️ **Web Management Panel**
  - Easy-to-use web interface for managing tunnels
  - Real-time statistics and monitoring
  - Create, configure, and control tunnels with one click
  - Support for both VLESS and WebRTC configurations

- 🔒 **Security**
  - End-to-end encryption with AES-256
  - TLS 1.3 support for VLESS
  - Reality technology for advanced DPI bypass
  - UUID-based authentication

- 🌐 **Flexible Deployment**
  - Client (cnc) and Server (srv) modes
  - SOCKS5 proxy support
  - Custom DNS resolver
  - Multiple carrier support (Telemost, Jazz, WBStream)

## Status

Beta
<br>
See all info in [issues](https://github.com/openlibrecommunity/olcrtc/issues)
<br>
Issues? contact us at [@openlibrecommunity](https://t.me/openlibrecommunity)
<br>
Or wait for the release or at least a release
<br>
Community android client: [alananisimov/olcbox](https://github.com/alananisimov/olcbox)

## Documentation

**🎯 [Complete Installation & Usage Guide](docs/COMPLETE-GUIDE.md)** - Start here!

### Guides

- [VLESS Reality Integration Guide](docs/vless-guide.md) - Detailed VLESS setup
- [Configuration Examples](docs/configuration-examples.md) - Ready-to-use configs
- [Build and Test Guide](docs/build-and-test.md) - Build from source

### Original Docs

- [For noobs](docs/fast.md)
- [Manual](docs/manual.md)
- [Setting matrix](docs/settings.md)
- [Client URI format](docs/uri.md)
- [Client subscription format](docs/sub.md)

## Quick Start

### 🚀 One-Command Server Installation

Install and configure olcRTC server with web panel in one command:

```bash
curl -fsSL https://raw.githubusercontent.com/openlibrecommunity/olcrtc/master/install.sh | sudo bash
```

This will:
- ✅ Install all dependencies
- ✅ Build olcRTC from source
- ✅ Generate UUID and encryption keys
- ✅ Setup SSL certificates
- ✅ Configure systemd service
- ✅ Start web management panel

After installation, open the web panel at `http://your-domain.com` or `http://your-ip:8080`

### 📱 Client Setup (All Platforms)

1. **Open web panel** in your browser
2. **Click "📱 Get Config"** on any tunnel
3. **Select your platform**:
   - Windows → Download .bat file and run
   - macOS/Linux → Download .sh file and run
   - Android → Import VLESS URL in V2rayNG
   - iOS → Import VLESS URL in Shadowrocket

### 🎯 Manual Setup

**VLESS Reality Server:**
```bash
./olcrtc \
  -mode srv \
  -link direct \
  -transport vless \
  -carrier telemost \
  -id ":443" \
  -client-id "your-uuid" \
  -key "your-32-byte-hex-key" \
  -data ./data \
  -webpanel ":8080"
```

**VLESS Reality Client:**
```bash
./olcrtc \
  -mode cnc \
  -link direct \
  -transport vless \
  -carrier telemost \
  -id "your-server.com:443" \
  -client-id "your-uuid" \
  -dns "your-server.com" \
  -socks-host "127.0.0.1" \
  -socks-port 1080 \
  -key "your-32-byte-hex-key" \
  -data ./data
```

**WebRTC DataChannel (Classic mode):**
```bash
# Server
./olcrtc -mode srv -transport datachannel -carrier telemost -id "room-id" -client-id "id" -key "key" -data ./data

# Client
./olcrtc -mode cnc -transport datachannel -carrier telemost -id "room-id" -client-id "id" -socks-port 1080 -key "key" -data ./data
```



## Build

```bash
# install mage first
go install github.com/magefile/mage@latest

# build cli + ui
mage build

# build cli only
mage buildCLI

# build cli with b codec, clones b repo, builds libb.so, compiles with -tags b
mage buildCLIB

# cross-compile for linux / windows / darwin
mage cross

# android aar via gomobile
mage mobile

# container image
mage podman
mage docker

# lint / test / clean
mage lint
mage test
mage clean

```

<div align="center">

---


Telegram: [zarazaex](https://t.me/zarazaexe)
<br>
Email: [zarazaex@tuta.io](mailto:zarazaex@tuta.io)
<br>
Site: [zarazaex.xyz](https://zarazaex.xyz)
<br>
Made for: [olcNG](https://github.com/zarazaex69/olcng)


</div>
