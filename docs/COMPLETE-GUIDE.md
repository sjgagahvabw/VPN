# olcRTC - Полное руководство по установке и использованию

## 🚀 Быстрая установка на сервер

### Автоматическая установка (рекомендуется)

Выполните одну команду на вашем сервере:

```bash
curl -fsSL https://raw.githubusercontent.com/openlibrecommunity/olcrtc/master/install.sh | sudo bash
```

Или скачайте и запустите скрипт:

```bash
wget https://raw.githubusercontent.com/openlibrecommunity/olcrtc/master/install.sh
chmod +x install.sh
sudo ./install.sh
```

Скрипт автоматически:
- ✅ Установит все зависимости (Go, Mage, Nginx)
- ✅ Клонирует и соберёт olcRTC
- ✅ Сгенерирует UUID и ключи шифрования
- ✅ Настроит SSL сертификаты (Let's Encrypt или самоподписанные)
- ✅ Создаст systemd сервис
- ✅ Настроит файрвол
- ✅ Запустит веб-панель управления

### После установки

1. **Откройте веб-панель** в браузере:
   ```
   http://ваш-домен.com
   или
   http://IP-адрес-сервера:8080
   ```

2. **Получите конфигурацию для клиентов**:
   ```bash
   /opt/olcrtc/generate-client-config.sh
   ```

3. **Проверьте статус сервера**:
   ```bash
   systemctl status olcrtc-server
   ```

## 📱 Подключение клиентов

### Через веб-панель (самый простой способ)

1. Откройте веб-панель в браузере
2. Нажмите на кнопку **"📱 Get Config"** у нужного туннеля
3. Выберите вашу платформу:
   - **Windows** - скачает .bat файл
   - **macOS** - скачает .sh файл
   - **Linux** - скачает .sh файл
   - **Android** - покажет VLESS URL для импорта
   - **iOS** - покажет VLESS URL для импорта

### Windows

1. Скачайте olcrtc.exe с [releases](https://github.com/openlibrecommunity/olcrtc/releases)
2. Скачайте конфигурационный файл через веб-панель
3. Запустите .bat файл двойным кликом
4. Настройте браузер на использование SOCKS5 прокси: `127.0.0.1:1080`

### macOS / Linux

1. Скачайте olcrtc бинарник
2. Скачайте конфигурационный файл через веб-панель
3. Сделайте скрипт исполняемым:
   ```bash
   chmod +x olcrtc-*.sh
   ```
4. Запустите:
   ```bash
   ./olcrtc-linux.sh
   ```

### Android

#### Вариант 1: V2rayNG (рекомендуется)

1. Установите [V2rayNG](https://play.google.com/store/apps/details?id=com.v2ray.ang) из Google Play
2. В веб-панели нажмите **"📱 Get Config"** → **"VLESS URL"**
3. Скопируйте URL
4. В V2rayNG: нажмите **"+"** → **"Import config from clipboard"**
5. Подключитесь

#### Вариант 2: QR код

1. В веб-панели нажмите **"📱 Get Config"** → **"QR Code"**
2. В V2rayNG: нажмите **"+"** → **"Scan QR code"**
3. Отсканируйте QR код

#### Вариант 3: Subscription URL

1. В веб-панели нажмите **"📋 Subscription"**
2. Скопируйте Subscription URL
3. В V2rayNG: **"⋮"** → **"Subscription setting"** → **"+"**
4. Вставьте URL и сохраните
5. Обновите подписку

### iOS

#### Вариант 1: Shadowrocket

1. Установите [Shadowrocket](https://apps.apple.com/app/shadowrocket/id932747118) из App Store
2. В веб-панели нажмите **"📱 Get Config"** → **"VLESS URL"**
3. Скопируйте URL
4. Откройте Shadowrocket
5. Нажмите **"+"** → **"Type"** → **"VLESS"**
6. Вставьте URL или отсканируйте QR код

#### Вариант 2: Subscription URL

1. В веб-панели нажмите **"📋 Subscription"**
2. Скопируйте Subscription URL
3. В Shadowrocket: **"Home"** → **"+"** → **"Subscribe"**
4. Вставьте URL

## 🎛️ Использование веб-панели

### Создание туннеля

1. Нажмите **"+ Add Tunnel"**
2. Заполните поля:
   - **Name**: Имя туннеля (например, "My VPN")
   - **Transport**: Выберите "VLESS Reality" для лучшей производительности
   - **Server Address**: Адрес вашего сервера (например, `vpn.example.com:443`)
   - **UUID**: UUID сервера (получите из `/opt/olcrtc/generate-client-config.sh`)
   - **Server Name (SNI)**: Доменное имя (например, `vpn.example.com`)
   - **Encryption Key**: Ключ шифрования (получите из конфигурации сервера)
3. Нажмите **"Create"**

### Управление туннелями

- **▶️ Start** - Запустить туннель
- **⏹️ Stop** - Остановить туннель
- **📱 Get Config** - Получить конфигурацию для клиентов
- **🗑️ Delete** - Удалить туннель

### Экспорт конфигураций

Веб-панель поддерживает экспорт в различных форматах:

- **VLESS URL** - для импорта в мобильные приложения
- **QR Code** - для быстрого сканирования
- **JSON** - для программного использования
- **Shell Script** - для автоматизации на десктопах

### Subscription URL

Subscription URL позволяет автоматически импортировать все туннели в клиентские приложения:

1. Нажмите **"📋 Subscription"** в веб-панели
2. Скопируйте URL
3. Добавьте его в ваше клиентское приложение
4. Все туннели будут автоматически синхронизированы

## 🔧 Ручная настройка

### Клиент (Linux/macOS)

```bash
./olcrtc \
  -mode cnc \
  -link direct \
  -transport vless \
  -carrier telemost \
  -id "vpn.example.com:443" \
  -client-id "ваш-uuid" \
  -dns "vpn.example.com" \
  -socks-host "127.0.0.1" \
  -socks-port 1080 \
  -key "ваш-ключ-шифрования" \
  -data ./data
```

### Сервер

```bash
./olcrtc \
  -mode srv \
  -link direct \
  -transport vless \
  -carrier telemost \
  -id ":443" \
  -client-id "ваш-uuid" \
  -key "ваш-ключ-шифрования" \
  -data ./data \
  -webpanel ":8080"
```

## 🌐 Настройка браузера

### Chrome / Edge

1. Установите расширение [Proxy SwitchyOmega](https://chrome.google.com/webstore/detail/proxy-switchyomega/padekgcemlokbadohgkifijomclgjgif)
2. Создайте новый профиль:
   - Protocol: SOCKS5
   - Server: 127.0.0.1
   - Port: 1080
3. Переключитесь на созданный профиль

### Firefox

1. Настройки → Основные → Параметры сети → Настроить
2. Выберите "Ручная настройка прокси"
3. SOCKS Host: `127.0.0.1`, Port: `1080`
4. Выберите "SOCKS v5"
5. Отметьте "Использовать прокси DNS при использовании SOCKS v5"

### Safari (macOS)

1. Системные настройки → Сеть → Дополнительно
2. Вкладка "Прокси"
3. Отметьте "SOCKS Proxy"
4. Сервер: `127.0.0.1:1080`

## 📊 Мониторинг

### Просмотр логов

```bash
# Логи сервера
journalctl -u olcrtc-server -f

# Логи в файле
tail -f /opt/olcrtc/logs/server.log
```

### Проверка статуса

```bash
# Статус сервиса
systemctl status olcrtc-server

# Проверка портов
netstat -tuln | grep -E '443|8080|1080'

# Тест подключения
curl -x socks5://127.0.0.1:1080 https://ifconfig.me
```

### Статистика в веб-панели

Веб-панель показывает:
- Статус туннелей (активен/неактивен)
- Количество подключений
- Переданные данные
- Время работы (uptime)

## 🔒 Безопасность

### Рекомендации

1. **Используйте сильные ключи**:
   ```bash
   openssl rand -hex 32
   ```

2. **Регулярно меняйте UUID**:
   ```bash
   uuidgen
   ```

3. **Настройте файрвол**:
   ```bash
   # Разрешить только необходимые порты
   ufw allow 80/tcp
   ufw allow 443/tcp
   ufw deny 8080/tcp  # Закрыть веб-панель извне
   ```

4. **Используйте SSH туннель для веб-панели**:
   ```bash
   ssh -L 8080:localhost:8080 user@your-server.com
   ```

5. **Настройте автообновление сертификатов**:
   ```bash
   # Certbot автоматически настроит cron job
   certbot renew --dry-run
   ```

## 🛠️ Устранение неполадок

### Сервер не запускается

```bash
# Проверьте логи
journalctl -u olcrtc-server -n 50

# Проверьте конфигурацию
cat /etc/olcrtc/config.env

# Проверьте права
ls -la /opt/olcrtc
```

### Не удаётся подключиться

```bash
# Проверьте, что сервер слушает порт
netstat -tuln | grep 443

# Проверьте файрвол
ufw status

# Проверьте DNS
nslookup ваш-домен.com

# Тест TLS
openssl s_client -connect ваш-домен.com:443
```

### Низкая скорость

1. Попробуйте другой транспорт (VLESS обычно быстрее)
2. Проверьте загрузку сервера: `htop`
3. Проверьте сетевую задержку: `ping ваш-сервер.com`
4. Используйте сервер ближе к вашему местоположению

### Веб-панель недоступна

```bash
# Проверьте, что сервис запущен
systemctl status olcrtc-server

# Проверьте порт
netstat -tuln | grep 8080

# Проверьте Nginx
systemctl status nginx
nginx -t
```

## 📚 Дополнительные ресурсы

- [VLESS Reality Guide](docs/vless-guide.md) - Подробное руководство по VLESS
- [Configuration Examples](docs/configuration-examples.md) - Примеры конфигураций
- [Build and Test](docs/build-and-test.md) - Сборка из исходников

## 💬 Поддержка

- **GitHub Issues**: https://github.com/openlibrecommunity/olcrtc/issues
- **Telegram**: [@openlibrecommunity](https://t.me/openlibrecommunity)
- **Email**: zarazaex@tuta.io

## 🎉 Готово!

Теперь у вас есть полностью настроенный VPN сервер с:
- ✅ VLESS Reality для обхода блокировок
- ✅ Веб-панель для удобного управления
- ✅ Поддержка всех платформ (Windows, macOS, Linux, Android, iOS)
- ✅ Автоматическая генерация конфигураций
- ✅ Subscription URL для синхронизации

Наслаждайтесь свободным интернетом! 🌍
