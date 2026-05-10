# Примеры конфигураций olcRTC

## VLESS Reality конфигурации

### Базовая конфигурация клиента

```json
{
  "name": "My VLESS Tunnel",
  "transport": "vless",
  "mode": "cnc",
  "enabled": true,
  "vless_config": {
    "server_addr": "example.com:443",
    "uuid": "12345678-1234-1234-1234-123456789abc",
    "flow": "xtls-rprx-vision",
    "server_name": "example.com",
    "fingerprint": "chrome",
    "allow_insecure": false
  },
  "client_id": "my-client-001",
  "key_hex": "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
  "socks_host": "127.0.0.1",
  "socks_port": 1080,
  "dns_server": "1.1.1.1:53"
}
```

### Конфигурация с аутентификацией SOCKS5

```json
{
  "name": "Secure VLESS Tunnel",
  "transport": "vless",
  "mode": "cnc",
  "enabled": true,
  "vless_config": {
    "server_addr": "vpn.example.com:443",
    "uuid": "87654321-4321-4321-4321-cba987654321",
    "flow": "xtls-rprx-vision",
    "server_name": "vpn.example.com",
    "fingerprint": "firefox",
    "allow_insecure": false
  },
  "client_id": "secure-client",
  "key_hex": "fedcba9876543210fedcba9876543210fedcba9876543210fedcba9876543210",
  "socks_host": "127.0.0.1",
  "socks_port": 1080,
  "socks_user": "myuser",
  "socks_pass": "mypassword",
  "dns_server": "8.8.8.8:53"
}
```

### Конфигурация сервера

```json
{
  "name": "VLESS Server",
  "transport": "vless",
  "mode": "srv",
  "enabled": true,
  "vless_config": {
    "server_addr": ":443",
    "uuid": "12345678-1234-1234-1234-123456789abc",
    "flow": "xtls-rprx-vision",
    "fingerprint": "chrome"
  },
  "client_id": "my-client-001",
  "key_hex": "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
  "dns_server": "1.1.1.1:53"
}
```

## WebRTC конфигурации

### DataChannel транспорт

```json
{
  "name": "WebRTC DataChannel",
  "transport": "datachannel",
  "mode": "cnc",
  "enabled": true,
  "webrtc_config": {
    "carrier": "telemost",
    "room_id": "my-room-123"
  },
  "client_id": "webrtc-client",
  "key_hex": "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789",
  "socks_host": "127.0.0.1",
  "socks_port": 1081
}
```

### VideoChannel транспорт

```json
{
  "name": "WebRTC VideoChannel",
  "transport": "videochannel",
  "mode": "cnc",
  "enabled": true,
  "webrtc_config": {
    "carrier": "jazz",
    "room_id": "video-room-456",
    "video_width": 1920,
    "video_height": 1080,
    "video_fps": 30,
    "video_bitrate": "2M"
  },
  "client_id": "video-client",
  "key_hex": "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
  "socks_host": "127.0.0.1",
  "socks_port": 1082
}
```

## Командная строка примеры

### VLESS Reality

#### Простой клиент
```bash
./olcrtc \
  -mode cnc \
  -link direct \
  -transport vless \
  -carrier telemost \
  -id "vpn.example.com:443" \
  -client-id "12345678-1234-1234-1234-123456789abc" \
  -dns "vpn.example.com" \
  -socks-host "127.0.0.1" \
  -socks-port 1080 \
  -key "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef" \
  -data ./data
```

#### Клиент с debug логами
```bash
./olcrtc \
  -mode cnc \
  -link direct \
  -transport vless \
  -carrier telemost \
  -id "vpn.example.com:443" \
  -client-id "12345678-1234-1234-1234-123456789abc" \
  -dns "vpn.example.com" \
  -socks-host "127.0.0.1" \
  -socks-port 1080 \
  -key "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef" \
  -data ./data \
  -debug
```

#### Клиент с SOCKS5 аутентификацией
```bash
./olcrtc \
  -mode cnc \
  -link direct \
  -transport vless \
  -carrier telemost \
  -id "vpn.example.com:443" \
  -client-id "12345678-1234-1234-1234-123456789abc" \
  -dns "vpn.example.com" \
  -socks-host "127.0.0.1" \
  -socks-port 1080 \
  -socks-user "myuser" \
  -socks-pass "mypassword" \
  -key "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef" \
  -data ./data
```

#### Сервер
```bash
./olcrtc \
  -mode srv \
  -link direct \
  -transport vless \
  -carrier telemost \
  -id ":443" \
  -client-id "12345678-1234-1234-1234-123456789abc" \
  -key "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef" \
  -data ./data
```

#### Сервер с веб-панелью
```bash
./olcrtc \
  -mode srv \
  -link direct \
  -transport vless \
  -carrier telemost \
  -id ":443" \
  -client-id "12345678-1234-1234-1234-123456789abc" \
  -key "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef" \
  -data ./data \
  -webpanel ":8080"
```

### WebRTC DataChannel

#### Клиент
```bash
./olcrtc \
  -mode cnc \
  -link direct \
  -transport datachannel \
  -carrier telemost \
  -id "room-123" \
  -client-id "webrtc-client" \
  -socks-host "127.0.0.1" \
  -socks-port 1080 \
  -key "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789" \
  -data ./data
```

#### Сервер
```bash
./olcrtc \
  -mode srv \
  -link direct \
  -transport datachannel \
  -carrier telemost \
  -id "room-123" \
  -client-id "webrtc-client" \
  -key "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789" \
  -data ./data
```

### WebRTC VideoChannel

#### Клиент
```bash
./olcrtc \
  -mode cnc \
  -link direct \
  -transport videochannel \
  -carrier jazz \
  -id "video-room-456" \
  -client-id "video-client" \
  -socks-host "127.0.0.1" \
  -socks-port 1080 \
  -key "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef" \
  -video-w 1920 \
  -video-h 1080 \
  -video-fps 30 \
  -video-bitrate "2M" \
  -video-hw "none" \
  -video-codec "qrcode" \
  -data ./data
```

#### Сервер
```bash
./olcrtc \
  -mode srv \
  -link direct \
  -transport videochannel \
  -carrier jazz \
  -id "video-room-456" \
  -client-id "video-client" \
  -key "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef" \
  -video-w 1920 \
  -video-h 1080 \
  -video-fps 30 \
  -video-bitrate "2M" \
  -video-hw "none" \
  -video-codec "qrcode" \
  -data ./data
```

## Docker Compose примеры

### VLESS Reality стек

```yaml
version: '3.8'

services:
  olcrtc-server:
    image: olcrtc:latest
    container_name: olcrtc-vless-server
    restart: unless-stopped
    ports:
      - "443:443"
      - "8080:8080"
    volumes:
      - ./data:/data
      - ./certs:/certs
    command: >
      -mode srv
      -link direct
      -transport vless
      -carrier telemost
      -id ":443"
      -client-id "12345678-1234-1234-1234-123456789abc"
      -key "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
      -data /data
      -webpanel ":8080"
    networks:
      - olcrtc-net

  olcrtc-client:
    image: olcrtc:latest
    container_name: olcrtc-vless-client
    restart: unless-stopped
    ports:
      - "1080:1080"
    volumes:
      - ./data:/data
    command: >
      -mode cnc
      -link direct
      -transport vless
      -carrier telemost
      -id "olcrtc-server:443"
      -client-id "12345678-1234-1234-1234-123456789abc"
      -dns "olcrtc-server"
      -socks-host "0.0.0.0"
      -socks-port 1080
      -key "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
      -data /data
    depends_on:
      - olcrtc-server
    networks:
      - olcrtc-net

networks:
  olcrtc-net:
    driver: bridge
```

### Множественные туннели

```yaml
version: '3.8'

services:
  # VLESS туннель
  vless-client:
    image: olcrtc:latest
    container_name: olcrtc-vless
    restart: unless-stopped
    ports:
      - "1080:1080"
    volumes:
      - ./data:/data
    command: >
      -mode cnc
      -transport vless
      -id "vpn.example.com:443"
      -client-id "vless-uuid"
      -socks-host "0.0.0.0"
      -socks-port 1080
      -key "vless-key"
      -data /data

  # WebRTC DataChannel туннель
  webrtc-client:
    image: olcrtc:latest
    container_name: olcrtc-webrtc
    restart: unless-stopped
    ports:
      - "1081:1081"
    volumes:
      - ./data:/data
    command: >
      -mode cnc
      -transport datachannel
      -carrier telemost
      -id "room-123"
      -client-id "webrtc-uuid"
      -socks-host "0.0.0.0"
      -socks-port 1081
      -key "webrtc-key"
      -data /data
```

## Systemd сервисы

### VLESS клиент сервис

`/etc/systemd/system/olcrtc-vless-client.service`:

```ini
[Unit]
Description=olcRTC VLESS Client
After=network.target

[Service]
Type=simple
User=olcrtc
WorkingDirectory=/opt/olcrtc
ExecStart=/opt/olcrtc/olcrtc \
  -mode cnc \
  -link direct \
  -transport vless \
  -carrier telemost \
  -id "vpn.example.com:443" \
  -client-id "12345678-1234-1234-1234-123456789abc" \
  -dns "vpn.example.com" \
  -socks-host "127.0.0.1" \
  -socks-port 1080 \
  -key "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef" \
  -data /opt/olcrtc/data
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

### VLESS сервер сервис

`/etc/systemd/system/olcrtc-vless-server.service`:

```ini
[Unit]
Description=olcRTC VLESS Server
After=network.target

[Service]
Type=simple
User=olcrtc
WorkingDirectory=/opt/olcrtc
ExecStart=/opt/olcrtc/olcrtc \
  -mode srv \
  -link direct \
  -transport vless \
  -carrier telemost \
  -id ":443" \
  -client-id "12345678-1234-1234-1234-123456789abc" \
  -key "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef" \
  -data /opt/olcrtc/data \
  -webpanel ":8080"
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

## Скрипты автоматизации

### Скрипт запуска клиента

`start-client.sh`:

```bash
#!/bin/bash

# Конфигурация
SERVER="vpn.example.com:443"
UUID="12345678-1234-1234-1234-123456789abc"
KEY="0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
SOCKS_PORT=1080

# Проверка, что olcrtc не запущен
if pgrep -x "olcrtc" > /dev/null; then
    echo "olcRTC уже запущен"
    exit 1
fi

# Запуск
./olcrtc \
  -mode cnc \
  -link direct \
  -transport vless \
  -carrier telemost \
  -id "$SERVER" \
  -client-id "$UUID" \
  -dns "$(echo $SERVER | cut -d: -f1)" \
  -socks-host "127.0.0.1" \
  -socks-port "$SOCKS_PORT" \
  -key "$KEY" \
  -data ./data &

echo "olcRTC запущен на порту $SOCKS_PORT"
```

### Скрипт остановки

`stop-client.sh`:

```bash
#!/bin/bash

pkill -SIGTERM olcrtc
echo "olcRTC остановлен"
```

### Скрипт проверки статуса

`check-status.sh`:

```bash
#!/bin/bash

SOCKS_PORT=1080

if pgrep -x "olcrtc" > /dev/null; then
    echo "✓ olcRTC запущен"
    
    # Проверка SOCKS5 прокси
    if curl -x socks5://127.0.0.1:$SOCKS_PORT -s https://ifconfig.me > /dev/null 2>&1; then
        echo "✓ SOCKS5 прокси работает"
        IP=$(curl -x socks5://127.0.0.1:$SOCKS_PORT -s https://ifconfig.me)
        echo "  IP: $IP"
    else
        echo "✗ SOCKS5 прокси не отвечает"
    fi
else
    echo "✗ olcRTC не запущен"
fi
```

## Генераторы конфигураций

### Генератор UUID и ключа

`generate-config.sh`:

```bash
#!/bin/bash

echo "Генерация конфигурации olcRTC..."
echo

# Генерация UUID
UUID=$(uuidgen | tr '[:upper:]' '[:lower:]')
echo "UUID: $UUID"

# Генерация ключа
KEY=$(openssl rand -hex 32)
echo "Key: $KEY"

echo
echo "Пример команды клиента:"
echo "./olcrtc -mode cnc -transport vless -id \"server:443\" -client-id \"$UUID\" -key \"$KEY\" -socks-port 1080 -data ./data"

echo
echo "Пример команды сервера:"
echo "./olcrtc -mode srv -transport vless -id \":443\" -client-id \"$UUID\" -key \"$KEY\" -data ./data"
```

Сделайте скрипт исполняемым:
```bash
chmod +x generate-config.sh
./generate-config.sh
```

## Заметки

- Всегда используйте одинаковый UUID и ключ на клиенте и сервере
- Для production используйте сильные случайные ключи
- Регулярно меняйте UUID и ключи для безопасности
- Используйте TLS сертификаты от доверенных CA для production
- Настройте файрвол для ограничения доступа к серверу
