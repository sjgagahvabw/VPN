#!/bin/bash

# VPN-olcRTC Easy Install Script
# Автоматическая установка и настройка VPN с маскировкой под Wildberries
# https://github.com/sjgagahvabw/Vpn-olcrtc

set -e

# Цвета
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Логирование
log_info() { echo -e "${GREEN}[✓]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[!]${NC} $1"; }
log_error() { echo -e "${RED}[✗]${NC} $1"; }
log_step() { echo -e "${BLUE}[→]${NC} $1"; }

# Баннер
echo -e "${BLUE}"
cat << "EOF"
╔═══════════════════════════════════════════════╗
║         VPN-olcRTC Easy Installer             ║
║   Простой VPN для обхода блокировок в РФ      ║
╚═══════════════════════════════════════════════╝
EOF
echo -e "${NC}"

# Проверка root
if [ "$EUID" -ne 0 ]; then
    log_error "Запустите скрипт с правами root: sudo bash $0"
    exit 1
fi

# Определение IP сервера
log_step "Определение IP адреса сервера..."
SERVER_IP=$(curl -sf --max-time 10 ifconfig.me || curl -sf --max-time 10 api.ipify.org || curl -sf --max-time 10 icanhazip.com)

if [[ ! "$SERVER_IP" =~ ^[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}$ ]]; then
    log_error "Не удалось определить IP адрес"
    exit 1
fi

log_info "IP сервера: $SERVER_IP"

# Определение ОС
if [ -f /etc/os-release ]; then
    . /etc/os-release
    OS=$ID
else
    log_error "Не удалось определить ОС"
    exit 1
fi

log_info "ОС: $OS"

# Установка зависимостей
log_step "Установка зависимостей..."

case "$OS" in
    ubuntu|debian)
        apt-get update -qq
        apt-get install -y -qq curl wget git build-essential openssl uuidgen jq >/dev/null 2>&1
        ;;
    centos|rhel|fedora)
        yum install -y -q curl wget git gcc make openssl util-linux jq >/dev/null 2>&1
        ;;
    *)
        log_error "Неподдерживаемая ОС: $OS"
        exit 1
        ;;
esac

log_info "Зависимости установлены"

# Установка Go
log_step "Установка Go..."

if ! command -v go &> /dev/null; then
    GO_VERSION="1.22.3"
    wget -q "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz"
    rm -rf /usr/local/go
    tar -C /usr/local -xzf "go${GO_VERSION}.linux-amd64.tar.gz"
    rm "go${GO_VERSION}.linux-amd64.tar.gz"

    echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
    export PATH=$PATH:/usr/local/go/bin

    log_info "Go ${GO_VERSION} установлен"
else
    log_info "Go уже установлен: $(go version)"
fi

# Клонирование репозитория
log_step "Клонирование VPN-olcRTC..."

cd /opt
rm -rf vpn-olcrtc
git clone -q https://github.com/sjgagahvabw/VPN.git vpn-olcrtc
cd vpn-olcrtc

log_info "Репозиторий склонирован"

# Установка mage
log_step "Установка mage..."
go install github.com/magefile/mage@latest >/dev/null 2>&1
export PATH=$PATH:$(go env GOPATH)/bin
log_info "Mage установлен"

# Сборка
log_step "Сборка olcRTC (это займет 2-3 минуты)..."
mage build >/dev/null 2>&1
log_info "Сборка завершена"

# Генерация ключей
log_step "Генерация ключей..."

UUID=$(uuidgen | tr '[:upper:]' '[:lower:]')
KEY=$(openssl rand -hex 32)

log_info "UUID: $UUID"
log_info "KEY: $KEY"

# Создание директорий
mkdir -p /opt/vpn-olcrtc/data
mkdir -p /opt/vpn-olcrtc/logs

# Создание конфигурации
log_step "Создание конфигурации..."

cat > /opt/vpn-olcrtc/config.json <<EOF
{
  "mode": "srv",
  "transport": "vless",
  "carrier": "wbstream",
  "id": ":443",
  "client_id": "$UUID",
  "key": "$KEY",
  "webpanel": ":8080",
  "data": "/opt/vpn-olcrtc/data"
}
EOF

chmod 600 /opt/vpn-olcrtc/config.json

log_info "Конфигурация создана"

# Создание systemd сервиса
log_step "Создание systemd сервиса..."

cat > /etc/systemd/system/vpn-olcrtc.service <<EOF
[Unit]
Description=VPN-olcRTC Server
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/vpn-olcrtc
ExecStart=/opt/vpn-olcrtc/olcrtc \\
  -mode srv \\
  -transport vless \\
  -carrier wbstream \\
  -id ":443" \\
  -client-id "$UUID" \\
  -key "$KEY" \\
  -webpanel ":8080" \\
  -data /opt/vpn-olcrtc/data
Restart=always
RestartSec=10
StandardOutput=append:/opt/vpn-olcrtc/logs/server.log
StandardError=append:/opt/vpn-olcrtc/logs/error.log

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable vpn-olcrtc >/dev/null 2>&1
systemctl start vpn-olcrtc

log_info "Сервис создан и запущен"

# Настройка файрвола
log_step "Настройка файрвола..."

if command -v ufw &> /dev/null; then
    ufw allow 443/tcp >/dev/null 2>&1
    ufw allow 8080/tcp >/dev/null 2>&1
elif command -v firewall-cmd &> /dev/null; then
    firewall-cmd --permanent --add-port=443/tcp >/dev/null 2>&1
    firewall-cmd --permanent --add-port=8080/tcp >/dev/null 2>&1
    firewall-cmd --reload >/dev/null 2>&1
fi

log_info "Файрвол настроен"

# Создание VLESS URL
VLESS_URL="vless://${UUID}@${SERVER_IP}:443?encryption=none&security=tls&type=tcp&host=stream.wb.ru&sni=stream.wb.ru#VPN-olcRTC"

# Сохранение информации
cat > /opt/vpn-olcrtc/connection-info.txt <<EOF
╔═══════════════════════════════════════════════════════════╗
║           VPN-olcRTC Установлен успешно!                  ║
╚═══════════════════════════════════════════════════════════╝

📱 ПОДКЛЮЧЕНИЕ КЛИЕНТА:

1. Веб-панель управления:
   http://${SERVER_IP}:8080

2. VLESS URL (для V2Ray клиентов):
   ${VLESS_URL}

3. Ручная настройка:
   Сервер: ${SERVER_IP}
   Порт: 443
   UUID: ${UUID}
   Шифрование: none
   Сеть: tcp
   TLS: включен
   SNI: stream.wb.ru
   Host: stream.wb.ru

🔧 УПРАВЛЕНИЕ СЕРВИСОМ:

   Статус:    systemctl status vpn-olcrtc
   Остановка: systemctl stop vpn-olcrtc
   Запуск:    systemctl start vpn-olcrtc
   Перезапуск: systemctl restart vpn-olcrtc
   Логи:      journalctl -u vpn-olcrtc -f

📚 ДОКУМЕНТАЦИЯ:

   https://github.com/sjgagahvabw/Vpn-olcrtc

⚠️  ВАЖНО:

   - Сохраните UUID и KEY в безопасном месте
   - Не делитесь этими данными
   - Регулярно обновляйте систему

═══════════════════════════════════════════════════════════
EOF

# Вывод результата
echo ""
echo -e "${GREEN}╔═══════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║           VPN-olcRTC Установлен успешно!                  ║${NC}"
echo -e "${GREEN}╚═══════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${BLUE}📱 ПОДКЛЮЧЕНИЕ КЛИЕНТА:${NC}"
echo ""
echo -e "1. ${YELLOW}Веб-панель управления:${NC}"
echo -e "   ${GREEN}http://${SERVER_IP}:8080${NC}"
echo ""
echo -e "2. ${YELLOW}VLESS URL (для V2Ray клиентов):${NC}"
echo -e "   ${GREEN}${VLESS_URL}${NC}"
echo ""
echo -e "3. ${YELLOW}Ручная настройка:${NC}"
echo -e "   Сервер: ${GREEN}${SERVER_IP}${NC}"
echo -e "   Порт: ${GREEN}443${NC}"
echo -e "   UUID: ${GREEN}${UUID}${NC}"
echo -e "   Шифрование: ${GREEN}none${NC}"
echo -e "   Сеть: ${GREEN}tcp${NC}"
echo -e "   TLS: ${GREEN}включен${NC}"
echo -e "   SNI: ${GREEN}stream.wb.ru${NC}"
echo -e "   Host: ${GREEN}stream.wb.ru${NC}"
echo ""
echo -e "${BLUE}🔧 УПРАВЛЕНИЕ СЕРВИСОМ:${NC}"
echo ""
echo -e "   Статус:     ${YELLOW}systemctl status vpn-olcrtc${NC}"
echo -e "   Остановка:  ${YELLOW}systemctl stop vpn-olcrtc${NC}"
echo -e "   Запуск:     ${YELLOW}systemctl start vpn-olcrtc${NC}"
echo -e "   Перезапуск: ${YELLOW}systemctl restart vpn-olcrtc${NC}"
echo -e "   Логи:       ${YELLOW}journalctl -u vpn-olcrtc -f${NC}"
echo ""
echo -e "${BLUE}📚 ДОКУМЕНТАЦИЯ:${NC}"
echo -e "   ${GREEN}https://github.com/sjgagahvabw/Vpn-olcrtc${NC}"
echo ""
echo -e "${YELLOW}⚠️  ВАЖНО:${NC}"
echo -e "   - Сохраните UUID и KEY в безопасном месте"
echo -e "   - Не делитесь этими данными"
echo -e "   - Регулярно обновляйте систему"
echo ""
echo -e "${GREEN}═══════════════════════════════════════════════════════════${NC}"
echo ""
echo -e "${BLUE}Вся информация сохранена в:${NC} ${GREEN}/opt/vpn-olcrtc/connection-info.txt${NC}"
echo ""
