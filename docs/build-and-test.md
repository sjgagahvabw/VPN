# Руководство по сборке и тестированию olcRTC с VLESS Reality

## Требования

- Go 1.25.0 или выше
- Mage (система сборки)
- Git

## Установка зависимостей

### 1. Установка Go

**Linux:**
```bash
wget https://go.dev/dl/go1.25.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.25.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
```

**macOS:**
```bash
brew install go
```

**Windows:**
Скачайте установщик с https://go.dev/dl/

### 2. Установка Mage

```bash
go install github.com/magefile/mage@latest
```

## Сборка проекта

### Базовая сборка

```bash
cd olcrtc
mage build
```

Это создаст исполняемый файл `olcrtc` в корне проекта.

### Сборка только CLI

```bash
mage buildCLI
```

### Кросс-компиляция

Для сборки под разные платформы:

```bash
mage cross
```

Это создаст бинарники для:
- Linux (amd64, arm64)
- Windows (amd64)
- macOS (amd64, arm64)

### Сборка для Android

```bash
mage mobile
```

Создаст AAR файл для использования в Android приложениях.

## Тестирование

### Запуск тестов

```bash
mage test
```

### Запуск линтера

```bash
mage lint
```

## Тестирование VLESS транспорта

### Подготовка

1. Сгенерируйте UUID:
```bash
# Linux/macOS
uuidgen

# Или используйте Python
python3 -c "import uuid; print(uuid.uuid4())"
```

2. Сгенерируйте ключ шифрования:
```bash
openssl rand -hex 32
```

3. Создайте директорию для данных:
```bash
mkdir -p data
```

### Локальное тестирование

#### Терминал 1 - Сервер

```bash
./olcrtc \
  -mode srv \
  -link direct \
  -transport vless \
  -carrier telemost \
  -id ":8443" \
  -client-id "ваш-uuid" \
  -key "ваш-hex-ключ" \
  -data ./data \
  -debug
```

#### Терминал 2 - Клиент

```bash
./olcrtc \
  -mode cnc \
  -link direct \
  -transport vless \
  -carrier telemost \
  -id "localhost:8443" \
  -client-id "ваш-uuid" \
  -dns "localhost" \
  -socks-host "127.0.0.1" \
  -socks-port 1080 \
  -key "ваш-hex-ключ" \
  -data ./data \
  -debug
```

#### Терминал 3 - Тестирование подключения

```bash
# Тест через curl
curl -x socks5://127.0.0.1:1080 https://ifconfig.me

# Тест через wget
wget -e use_proxy=yes -e socks_proxy=127.0.0.1:1080 https://ifconfig.me -O -

# Тест скорости
curl -x socks5://127.0.0.1:1080 -o /dev/null https://speed.cloudflare.com/__down?bytes=100000000
```

## Тестирование веб-панели

### Запуск с веб-панелью

```bash
./olcrtc \
  -mode cnc \
  -link direct \
  -transport vless \
  -carrier telemost \
  -id "localhost:8443" \
  -client-id "ваш-uuid" \
  -socks-host "127.0.0.1" \
  -socks-port 1080 \
  -key "ваш-hex-ключ" \
  -data ./data \
  -webpanel ":8080"
```

### Тестирование API

```bash
# Получить список конфигураций
curl http://localhost:8080/api/configs

# Создать новую конфигурацию
curl -X POST http://localhost:8080/api/configs \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Tunnel",
    "transport": "vless",
    "mode": "cnc",
    "vless_config": {
      "server_addr": "localhost:8443",
      "uuid": "ваш-uuid",
      "server_name": "localhost",
      "flow": "xtls-rprx-vision",
      "fingerprint": "chrome"
    },
    "socks_host": "127.0.0.1",
    "socks_port": 1080
  }'

# Получить статистику
curl http://localhost:8080/api/stats
```

### Открыть веб-интерфейс

Откройте браузер и перейдите на:
```
http://localhost:8080
```

## Тестирование производительности

### Benchmark тест

```bash
# Запустите сервер и клиент как описано выше

# Тест пропускной способности
iperf3 -c iperf.he.net -p 5201 --socks5 127.0.0.1:1080

# Тест задержки
ping -c 10 8.8.8.8  # без прокси
# vs
curl -x socks5://127.0.0.1:1080 https://cloudflare.com/cdn-cgi/trace  # через прокси
```

### Сравнение транспортов

Создайте скрипт для сравнения:

```bash
#!/bin/bash

echo "Testing VLESS..."
time curl -x socks5://127.0.0.1:1080 -o /dev/null -s https://speed.cloudflare.com/__down?bytes=10000000

echo "Testing DataChannel..."
time curl -x socks5://127.0.0.1:1081 -o /dev/null -s https://speed.cloudflare.com/__down?bytes=10000000
```

## Отладка

### Включение debug логов

Добавьте флаг `-debug` к любой команде:

```bash
./olcrtc -mode cnc -debug ...
```

### Проверка логов

Логи выводятся в stdout. Для сохранения в файл:

```bash
./olcrtc -mode cnc ... 2>&1 | tee olcrtc.log
```

### Типичные проблемы

#### 1. "connection refused"

**Причина:** Сервер не запущен или порт заблокирован

**Решение:**
```bash
# Проверьте, что сервер запущен
ps aux | grep olcrtc

# Проверьте, что порт открыт
netstat -tuln | grep 8443
```

#### 2. "TLS handshake failed"

**Причина:** Неправильный SNI или проблемы с сертификатом

**Решение:**
- Проверьте правильность server_name
- Убедитесь, что используется правильный fingerprint
- Попробуйте allow_insecure: true для тестирования

#### 3. "UUID mismatch"

**Причина:** UUID клиента и сервера не совпадают

**Решение:**
- Убедитесь, что используете одинаковый UUID на клиенте и сервере
- Проверьте, что UUID в правильном формате

#### 4. "Key size error"

**Причина:** Ключ шифрования не 32 байта

**Решение:**
```bash
# Сгенерируйте новый ключ правильного размера
openssl rand -hex 32
```

## Очистка

```bash
# Очистка собранных файлов
mage clean

# Полная очистка включая зависимости
rm -rf vendor/
go clean -modcache
```

## Continuous Integration

### GitHub Actions пример

```yaml
name: Build and Test

on: [push, pull_request]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.25'
    
    - name: Install Mage
      run: go install github.com/magefile/mage@latest
    
    - name: Build
      run: mage build
    
    - name: Test
      run: mage test
    
    - name: Lint
      run: mage lint
```

## Docker тестирование

### Сборка Docker образа

```bash
mage docker
```

### Запуск в Docker

**Сервер:**
```bash
docker run -d \
  -p 8443:8443 \
  -v $(pwd)/data:/data \
  olcrtc:latest \
  -mode srv \
  -transport vless \
  -id ":8443" \
  -client-id "uuid" \
  -key "key" \
  -data /data
```

**Клиент:**
```bash
docker run -d \
  -p 1080:1080 \
  -v $(pwd)/data:/data \
  olcrtc:latest \
  -mode cnc \
  -transport vless \
  -id "server:8443" \
  -client-id "uuid" \
  -socks-host "0.0.0.0" \
  -socks-port 1080 \
  -key "key" \
  -data /data
```

## Производственное развёртывание

### Systemd сервис

Создайте `/etc/systemd/system/olcrtc.service`:

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
  -transport vless \
  -id ":443" \
  -client-id "your-uuid" \
  -key "your-key" \
  -data /opt/olcrtc/data
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

Активация:
```bash
sudo systemctl daemon-reload
sudo systemctl enable olcrtc
sudo systemctl start olcrtc
sudo systemctl status olcrtc
```

## Мониторинг

### Prometheus метрики

В будущих версиях будет добавлена поддержка Prometheus метрик.

### Логирование

Для централизованного логирования используйте:

```bash
./olcrtc ... 2>&1 | logger -t olcrtc
```

Или настройте rsyslog для отправки логов в центральный сервер.

## Поддержка

Если у вас возникли проблемы:

1. Проверьте [Issues](https://github.com/openlibrecommunity/olcrtc/issues)
2. Создайте новый Issue с подробным описанием проблемы
3. Присоединяйтесь к [Telegram группе](https://t.me/openlibrecommunity)

## Вклад в проект

Мы приветствуем вклад в проект! Пожалуйста:

1. Fork репозиторий
2. Создайте feature branch
3. Сделайте изменения
4. Запустите тесты
5. Создайте Pull Request

## Лицензия

WTFPL - Do What The Fuck You Want To Public License
