#!/bin/bash

# olcRTC Auto-Deploy Script
# Автоматическая установка и настройка olcRTC с VLESS Reality и веб-панелью

set -e

echo "=================================="
echo "olcRTC Auto-Deploy Script"
echo "=================================="
echo ""

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Функция для вывода сообщений
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Проверка root прав
if [ "$EUID" -ne 0 ]; then 
    log_error "Пожалуйста, запустите скрипт с правами root (sudo)"
    exit 1
fi

# Определение ОС
if [ -f /etc/os-release ]; then
    . /etc/os-release
    OS=$ID
else
    log_error "Не удалось определить операционную систему"
    exit 1
fi

log_info "Обнаружена ОС: $OS"

# Установка зависимостей
log_info "Установка зависимостей..."

case $OS in
    ubuntu|debian)
        apt-get update
        apt-get install -y git curl wget openssl uuidgen nginx certbot python3-certbot-nginx
        ;;
    centos|rhel|fedora)
        yum install -y git curl wget openssl util-linux nginx certbot python3-certbot-nginx
        ;;
    *)
        log_error "Неподдерживаемая ОС: $OS"
        exit 1
        ;;
esac

# Установка Go
log_info "Установка Go..."
GO_VERSION="1.25.0"
if ! command -v go &> /dev/null; then
    wget https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz
    tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz
    rm go${GO_VERSION}.linux-amd64.tar.gz
    echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
    export PATH=$PATH:/usr/local/go/bin
    log_info "Go установлен"
else
    log_info "Go уже установлен"
fi

# Установка Mage
log_info "Установка Mage..."
if ! command -v mage &> /dev/null; then
    /usr/local/go/bin/go install github.com/magefile/mage@latest
    export PATH=$PATH:$(go env GOPATH)/bin
    log_info "Mage установлен"
else
    log_info "Mage уже установлен"
fi

# Создание пользователя olcrtc
log_info "Создание пользователя olcrtc..."
if ! id -u olcrtc &> /dev/null; then
    useradd -r -s /bin/false olcrtc
    log_info "Пользователь olcrtc создан"
else
    log_info "Пользователь olcrtc уже существует"
fi

# Клонирование репозитория
log_info "Клонирование репозитория olcRTC..."
INSTALL_DIR="/opt/olcrtc"
if [ -d "$INSTALL_DIR" ]; then
    log_warn "Директория $INSTALL_DIR уже существует. Обновление..."
    cd $INSTALL_DIR
    git pull
else
    git clone https://github.com/openlibrecommunity/olcrtc.git $INSTALL_DIR
    cd $INSTALL_DIR
fi

# Сборка проекта
log_info "Сборка olcRTC..."
export PATH=$PATH:/usr/local/go/bin:$(go env GOPATH)/bin
mage build

# Создание директорий
log_info "Создание директорий..."
mkdir -p $INSTALL_DIR/data
mkdir -p $INSTALL_DIR/configs
mkdir -p $INSTALL_DIR/logs
mkdir -p /etc/olcrtc

# Генерация конфигурации
log_info "Генерация конфигурации..."

# Запрос домена
read -p "Введите ваш домен (например, vpn.example.com): " DOMAIN
if [ -z "$DOMAIN" ]; then
    log_error "Домен не может быть пустым"
    exit 1
fi

# Генерация UUID и ключа
UUID=$(uuidgen | tr '[:upper:]' '[:lower:]')
KEY=$(openssl rand -hex 32)

log_info "Сгенерированные данные:"
echo "  UUID: $UUID"
echo "  Key: $KEY"
echo "  Domain: $DOMAIN"

# Сохранение конфигурации
cat > /etc/olcrtc/config.env <<EOF
# olcRTC Configuration
DOMAIN=$DOMAIN
UUID=$UUID
KEY=$KEY
VLESS_PORT=443
WEB_PORT=8080
SOCKS_PORT=1080
EOF

log_info "Конфигурация сохранена в /etc/olcrtc/config.env"

# Получение SSL сертификата
log_info "Получение SSL сертификата от Let's Encrypt..."
read -p "Получить SSL сертификат? (y/n): " GET_SSL

if [ "$GET_SSL" = "y" ]; then
    certbot certonly --standalone -d $DOMAIN --non-interactive --agree-tos --register-unsafely-without-email
    
    if [ $? -eq 0 ]; then
        log_info "SSL сертификат получен"
        CERT_PATH="/etc/letsencrypt/live/$DOMAIN/fullchain.pem"
        KEY_PATH="/etc/letsencrypt/live/$DOMAIN/privkey.pem"
    else
        log_warn "Не удалось получить SSL сертификат. Будет использоваться самоподписанный сертификат."
        # Создание самоподписанного сертификата
        mkdir -p /etc/olcrtc/certs
        openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
            -keyout /etc/olcrtc/certs/key.pem \
            -out /etc/olcrtc/certs/cert.pem \
            -subj "/CN=$DOMAIN"
        CERT_PATH="/etc/olcrtc/certs/cert.pem"
        KEY_PATH="/etc/olcrtc/certs/key.pem"
    fi
else
    log_info "Создание самоподписанного сертификата..."
    mkdir -p /etc/olcrtc/certs
    openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
        -keyout /etc/olcrtc/certs/key.pem \
        -out /etc/olcrtc/certs/cert.pem \
        -subj "/CN=$DOMAIN"
    CERT_PATH="/etc/olcrtc/certs/cert.pem"
    KEY_PATH="/etc/olcrtc/certs/key.pem"
fi

# Создание systemd сервиса для сервера
log_info "Создание systemd сервиса..."

cat > /etc/systemd/system/olcrtc-server.service <<EOF
[Unit]
Description=olcRTC VLESS Server with Web Panel
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=$INSTALL_DIR
EnvironmentFile=/etc/olcrtc/config.env
ExecStart=$INSTALL_DIR/olcrtc \\
  -mode srv \\
  -link direct \\
  -transport vless \\
  -carrier telemost \\
  -id ":\${VLESS_PORT}" \\
  -client-id "\${UUID}" \\
  -key "\${KEY}" \\
  -data $INSTALL_DIR/data \\
  -webpanel ":\${WEB_PORT}" \\
  -tls-cert "$CERT_PATH" \\
  -tls-key "$KEY_PATH"
Restart=always
RestartSec=10
StandardOutput=append:$INSTALL_DIR/logs/server.log
StandardError=append:$INSTALL_DIR/logs/server-error.log

[Install]
WantedBy=multi-user.target
EOF

# Настройка Nginx для веб-панели
log_info "Настройка Nginx..."

cat > /etc/nginx/sites-available/olcrtc <<EOF
server {
    listen 80;
    server_name $DOMAIN;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }
}
EOF

ln -sf /etc/nginx/sites-available/olcrtc /etc/nginx/sites-enabled/
nginx -t && systemctl reload nginx

# Настройка файрвола
log_info "Настройка файрвола..."
if command -v ufw &> /dev/null; then
    ufw allow 80/tcp
    ufw allow 443/tcp
    ufw allow 8080/tcp
    ufw --force enable
elif command -v firewall-cmd &> /dev/null; then
    firewall-cmd --permanent --add-port=80/tcp
    firewall-cmd --permanent --add-port=443/tcp
    firewall-cmd --permanent --add-port=8080/tcp
    firewall-cmd --reload
fi

# Установка прав
log_info "Установка прав..."
chown -R olcrtc:olcrtc $INSTALL_DIR/data
chown -R olcrtc:olcrtc $INSTALL_DIR/configs
chown -R olcrtc:olcrtc $INSTALL_DIR/logs
chmod +x $INSTALL_DIR/olcrtc

# Запуск сервиса
log_info "Запуск olcRTC сервера..."
systemctl daemon-reload
systemctl enable olcrtc-server
systemctl start olcrtc-server

# Проверка статуса
sleep 3
if systemctl is-active --quiet olcrtc-server; then
    log_info "✓ olcRTC сервер успешно запущен!"
else
    log_error "✗ Не удалось запустить olcRTC сервер"
    log_info "Проверьте логи: journalctl -u olcrtc-server -f"
    exit 1
fi

# Создание скрипта для генерации клиентских конфигураций
log_info "Создание скрипта для генерации клиентских конфигураций..."

cat > $INSTALL_DIR/generate-client-config.sh <<'SCRIPT_EOF'
#!/bin/bash

source /etc/olcrtc/config.env

echo "=================================="
echo "olcRTC Client Configuration"
echo "=================================="
echo ""
echo "Конфигурация для подключения к серверу:"
echo ""
echo "Server: $DOMAIN:$VLESS_PORT"
echo "UUID: $UUID"
echo "Key: $KEY"
echo ""
echo "--- Команда для Linux/macOS клиента ---"
echo ""
cat <<EOF
./olcrtc \\
  -mode cnc \\
  -link direct \\
  -transport vless \\
  -carrier telemost \\
  -id "$DOMAIN:$VLESS_PORT" \\
  -client-id "$UUID" \\
  -dns "$DOMAIN" \\
  -socks-host "127.0.0.1" \\
  -socks-port 1080 \\
  -key "$KEY" \\
  -data ./data
EOF
echo ""
echo "--- VLESS URL для импорта ---"
echo ""
echo "vless://$UUID@$DOMAIN:$VLESS_PORT?encryption=none&flow=xtls-rprx-vision&security=tls&sni=$DOMAIN&fp=chrome&type=tcp&headerType=none#olcRTC-$DOMAIN"
echo ""
echo "--- JSON конфигурация ---"
echo ""
cat <<EOF
{
  "name": "olcRTC-$DOMAIN",
  "transport": "vless",
  "mode": "cnc",
  "vless_config": {
    "server_addr": "$DOMAIN:$VLESS_PORT",
    "uuid": "$UUID",
    "flow": "xtls-rprx-vision",
    "server_name": "$DOMAIN",
    "fingerprint": "chrome",
    "allow_insecure": false
  },
  "key_hex": "$KEY",
  "socks_host": "127.0.0.1",
  "socks_port": 1080
}
EOF
echo ""
echo "=================================="
echo "Веб-панель доступна по адресу:"
echo "http://$DOMAIN"
echo "или"
echo "http://$(curl -s ifconfig.me):8080"
echo "=================================="
SCRIPT_EOF

chmod +x $INSTALL_DIR/generate-client-config.sh

# Вывод итоговой информации
echo ""
echo "=================================="
log_info "Установка завершена успешно!"
echo "=================================="
echo ""
echo "Информация о сервере:"
echo "  Домен: $DOMAIN"
echo "  VLESS порт: 443"
echo "  Веб-панель: http://$DOMAIN или http://$(curl -s ifconfig.me):8080"
echo ""
echo "Для получения клиентской конфигурации выполните:"
echo "  $INSTALL_DIR/generate-client-config.sh"
echo ""
echo "Управление сервисом:"
echo "  Статус: systemctl status olcrtc-server"
echo "  Логи: journalctl -u olcrtc-server -f"
echo "  Перезапуск: systemctl restart olcrtc-server"
echo "  Остановка: systemctl stop olcrtc-server"
echo ""
echo "Конфигурация сохранена в: /etc/olcrtc/config.env"
echo ""
log_info "Откройте веб-панель в браузере для дальнейшей настройки!"
echo "=================================="
