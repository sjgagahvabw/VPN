# VLESS Reality Integration Guide

## Overview

olcRTC теперь поддерживает VLESS Reality протокол в дополнение к WebRTC туннелированию. Это позволяет использовать более надёжный и быстрый метод обхода блокировок.

## Что нового

### VLESS Reality Transport

VLESS Reality - это современный протокол обхода блокировок, который:
- Использует TLS 1.3 для маскировки трафика
- Поддерживает Reality технологию для обхода DPI
- Обеспечивает высокую скорость и низкую задержку
- Совместим с существующей инфраструктурой VLESS

### Веб-панель управления

Новая веб-панель позволяет:
- Управлять туннелями через удобный интерфейс
- Создавать и настраивать VLESS и WebRTC туннели
- Мониторить статистику подключений
- Запускать и останавливать туннели одним кликом

## Быстрый старт

### 1. Запуск с VLESS Reality

#### Клиент (cnc mode)

```bash
./olcrtc \
  -mode cnc \
  -link direct \
  -transport vless \
  -carrier telemost \
  -id "your-server-address:443" \
  -client-id "your-uuid" \
  -dns "example.com" \
  -socks-host "127.0.0.1" \
  -socks-port 1080 \
  -key "your-32-byte-hex-key" \
  -data ./data
```

#### Сервер (srv mode)

```bash
./olcrtc \
  -mode srv \
  -link direct \
  -transport vless \
  -carrier telemost \
  -id ":443" \
  -client-id "your-uuid" \
  -key "your-32-byte-hex-key" \
  -data ./data
```

### 2. Запуск веб-панели

Добавьте флаг `-webpanel` для запуска веб-интерфейса:

```bash
./olcrtc \
  -mode cnc \
  -webpanel ":8080" \
  ... (остальные параметры)
```

Откройте браузер и перейдите на `http://localhost:8080`

## Конфигурация VLESS Reality

### Параметры клиента

- `server_addr` - адрес сервера (например, `example.com:443`)
- `uuid` - UUID пользователя для аутентификации
- `flow` - режим flow control (рекомендуется `xtls-rprx-vision`)
- `server_name` - SNI для TLS (например, `example.com`)
- `public_key` - публичный ключ Reality (опционально)
- `short_id` - короткий ID Reality (опционально)
- `fingerprint` - отпечаток TLS (например, `chrome`, `firefox`, `safari`)

### Пример конфигурации

```json
{
  "name": "My VLESS Tunnel",
  "transport": "vless",
  "mode": "cnc",
  "vless_config": {
    "server_addr": "example.com:443",
    "uuid": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
    "flow": "xtls-rprx-vision",
    "server_name": "example.com",
    "fingerprint": "chrome",
    "allow_insecure": false
  },
  "socks_host": "127.0.0.1",
  "socks_port": 1080
}
```

## Сравнение транспортов

| Характеристика | VLESS Reality | WebRTC DataChannel | WebRTC VideoChannel |
|----------------|---------------|-------------------|---------------------|
| Скорость | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐ |
| Стабильность | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐ |
| Обход DPI | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| Задержка | Низкая | Средняя | Высокая |
| Сложность настройки | Средняя | Низкая | Высокая |

## Веб-панель API

### Endpoints

#### GET /api/configs
Получить список всех конфигураций туннелей

```bash
curl http://localhost:8080/api/configs
```

#### POST /api/configs
Создать новую конфигурацию

```bash
curl -X POST http://localhost:8080/api/configs \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My Tunnel",
    "transport": "vless",
    "mode": "cnc",
    "vless_config": {
      "server_addr": "example.com:443",
      "uuid": "your-uuid",
      "server_name": "example.com"
    }
  }'
```

#### GET /api/configs/{id}
Получить конкретную конфигурацию

```bash
curl http://localhost:8080/api/configs/tunnel-123
```

#### PUT /api/configs/{id}
Обновить конфигурацию

```bash
curl -X PUT http://localhost:8080/api/configs/tunnel-123 \
  -H "Content-Type: application/json" \
  -d '{...}'
```

#### DELETE /api/configs/{id}
Удалить конфигурацию

```bash
curl -X DELETE http://localhost:8080/api/configs/tunnel-123
```

#### POST /api/tunnels/start
Запустить туннель

```bash
curl -X POST http://localhost:8080/api/tunnels/start \
  -H "Content-Type: application/json" \
  -d '{"tunnel_id": "tunnel-123"}'
```

#### POST /api/tunnels/stop
Остановить туннель

```bash
curl -X POST http://localhost:8080/api/tunnels/stop \
  -H "Content-Type: application/json" \
  -d '{"tunnel_id": "tunnel-123"}'
```

#### GET /api/stats
Получить статистику всех туннелей

```bash
curl http://localhost:8080/api/stats
```

#### GET /api/stats/{id}
Получить статистику конкретного туннеля

```bash
curl http://localhost:8080/api/stats/tunnel-123
```

## Генерация UUID

Для генерации UUID можно использовать:

```bash
# Linux/macOS
uuidgen

# Python
python3 -c "import uuid; print(uuid.uuid4())"

# Online
# https://www.uuidgenerator.net/
```

## Генерация ключа шифрования

```bash
# Генерация 32-байтного ключа в hex формате
openssl rand -hex 32
```

## Настройка сервера VLESS Reality

### Требования

- Сервер с публичным IP
- Доменное имя (для SNI)
- TLS сертификат (Let's Encrypt или самоподписанный)

### Пример настройки с Let's Encrypt

```bash
# Установка certbot
sudo apt install certbot

# Получение сертификата
sudo certbot certonly --standalone -d your-domain.com

# Сертификаты будут в:
# /etc/letsencrypt/live/your-domain.com/fullchain.pem
# /etc/letsencrypt/live/your-domain.com/privkey.pem
```

### Запуск сервера

```bash
./olcrtc \
  -mode srv \
  -link direct \
  -transport vless \
  -carrier telemost \
  -id ":443" \
  -client-id "your-uuid" \
  -key "your-hex-key" \
  -tls-cert "/etc/letsencrypt/live/your-domain.com/fullchain.pem" \
  -tls-key "/etc/letsencrypt/live/your-domain.com/privkey.pem" \
  -data ./data
```

## Troubleshooting

### Проблема: Не удаётся подключиться к серверу

**Решение:**
1. Проверьте, что сервер запущен и доступен
2. Убедитесь, что порт 443 открыт в файрволе
3. Проверьте правильность UUID и ключа шифрования
4. Проверьте логи: `./olcrtc -debug ...`

### Проблема: TLS handshake failed

**Решение:**
1. Проверьте правильность SNI (server_name)
2. Убедитесь, что сертификат действителен
3. Попробуйте другой fingerprint (chrome, firefox, safari)

### Проблема: Низкая скорость

**Решение:**
1. Попробуйте использовать VLESS вместо WebRTC
2. Проверьте загрузку сервера
3. Используйте ближайший к вам сервер

## Миграция с WebRTC на VLESS

Если вы используете WebRTC туннели, вы можете легко мигрировать на VLESS:

1. Создайте новую конфигурацию с транспортом `vless`
2. Настройте VLESS параметры
3. Запустите новый туннель
4. Протестируйте подключение
5. Остановите старый WebRTC туннель

## Безопасность

### Рекомендации

1. **Используйте сильные ключи**: Генерируйте случайные 32-байтные ключи
2. **Регулярно меняйте UUID**: Меняйте UUID каждые 1-3 месяца
3. **Используйте TLS 1.3**: Убедитесь, что используется TLS 1.3
4. **Защитите веб-панель**: Используйте аутентификацию для веб-панели
5. **Мониторьте логи**: Регулярно проверяйте логи на подозрительную активность

### Аутентификация веб-панели

В будущих версиях будет добавлена аутентификация. Пока рекомендуется:
- Запускать веб-панель только на localhost
- Использовать SSH туннель для удалённого доступа
- Настроить nginx с basic auth перед веб-панелью

## Производительность

### Оптимизация

1. **Используйте VLESS для максимальной скорости**
2. **Настройте MTU**: Оптимальное значение 1400-1500
3. **Увеличьте буферы**: Для высокоскоростных соединений
4. **Используйте BBR congestion control** (Linux):
   ```bash
   echo "net.core.default_qdisc=fq" >> /etc/sysctl.conf
   echo "net.ipv4.tcp_congestion_control=bbr" >> /etc/sysctl.conf
   sysctl -p
   ```

## Примеры использования

### Пример 1: Простой VLESS туннель

```bash
# Сервер
./olcrtc -mode srv -transport vless -id ":443" -client-id "uuid" -key "key"

# Клиент
./olcrtc -mode cnc -transport vless -id "server.com:443" -client-id "uuid" -key "key" -socks-port 1080
```

### Пример 2: Гибридная конфигурация

Используйте VLESS как основной транспорт и WebRTC как fallback:

```bash
# Основной туннель (VLESS)
./olcrtc -mode cnc -transport vless -id "server.com:443" -socks-port 1080 ...

# Резервный туннель (WebRTC)
./olcrtc -mode cnc -transport datachannel -carrier telemost -socks-port 1081 ...
```

### Пример 3: Использование с браузером

```bash
# Запустите туннель
./olcrtc -mode cnc -transport vless -socks-port 1080 ...

# Настройте браузер на использование SOCKS5 прокси:
# Host: 127.0.0.1
# Port: 1080
```

## Поддержка

- GitHub Issues: https://github.com/openlibrecommunity/olcrtc/issues
- Telegram: [@openlibrecommunity](https://t.me/openlibrecommunity)
- Email: zarazaex@tuta.io

## Лицензия

WTFPL - Do What The Fuck You Want To Public License

## Благодарности

- Проект VLESS/Xray за протокол
- Проект olcNG за вдохновение
- Сообщество за тестирование и обратную связь
