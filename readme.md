# VPN-olcRTC 🚀

**Простой VPN для обхода блокировок в России через VLESS Reality + WebRTC**

Установка одной командой. Работает сразу. Маскируется под Wildberries Stream.

## ⚡ Быстрый старт

### На сервере (VPS):

```bash
curl -fsSL https://raw.githubusercontent.com/sjgagahvabw/Vpn-olcrtc/main/easy-install.sh | sudo bash
```

Скрипт автоматически:
- ✅ Установит все зависимости
- ✅ Соберет olcRTC
- ✅ Настроит VLESS Reality с маскировкой под WB
- ✅ Создаст белые списки для российских сайтов
- ✅ Запустит веб-панель управления
- ✅ Выдаст готовую ссылку для подключения

### На клиенте:

**Windows/macOS/Linux:**
Откройте веб-панель `http://ваш-сервер:8080` и скачайте клиент для вашей ОС.

**Android:**
1. Установите [V2rayNG](https://github.com/2dust/v2rayNG/releases)
2. Импортируйте VLESS URL из веб-панели

**iOS:**
1. Установите [Shadowrocket](https://apps.apple.com/app/shadowrocket/id932747118)
2. Импортируйте VLESS URL из веб-панели

## 🎯 Особенности

### Маскировка под Wildberries
- Использует `stream.wb.ru` как carrier
- Неотличим от обычного видеозвонка
- Работает даже при глубокой инспекции пакетов

### Белые списки (Split Tunneling)
- Российские сайты → напрямую (быстрее)
- Заблокированные сайты → через VPN
- Автообновление списков каждые 24 часа

### Два режима работы

**1. VLESS Reality** (рекомендуется)
- Высокая скорость
- Низкая задержка
- Стабильное соединение
- Совместимость с V2Ray клиентами

**2. WebRTC** (максимальная скрытность)
- Туннелирование через WebRTC
- Паразитирует на легальных сервисах
- Несколько транспортов на выбор

## 📱 Поддерживаемые платформы

- ✅ Windows 10/11
- ✅ macOS 10.15+
- ✅ Linux (Ubuntu, Debian, CentOS)
- ✅ Android 5.0+
- ✅ iOS 12.0+

## 🔒 Безопасность

- AES-256 шифрование
- TLS 1.3
- Reality технология для обхода DPI
- UUID аутентификация
- Автоматическая генерация ключей

## 📊 Производительность

- Скорость: до 1 Gbps
- Задержка: < 10ms
- Потребление RAM: ~50-100MB
- Потребление CPU: ~5-10%

## 🛠️ Ручная установка

Если автоматический скрипт не подходит:

```bash
# Клонирование
git clone https://github.com/sjgagahvabw/Vpn-olcrtc.git
cd Vpn-olcrtc

# Сборка
go install github.com/magefile/mage@latest
mage build

# Запуск сервера
./olcrtc \
  -mode srv \
  -transport vless \
  -carrier wbstream \
  -id ":443" \
  -client-id "$(uuidgen)" \
  -key "$(openssl rand -hex 32)" \
  -webpanel ":8080" \
  -data ./data

# Запуск клиента
./olcrtc \
  -mode cnc \
  -transport vless \
  -carrier wbstream \
  -id "ваш-сервер.com:443" \
  -client-id "ваш-uuid" \
  -key "ваш-ключ" \
  -socks-port 1080 \
  -data ./data
```

## 📚 Документация

- [Полное руководство](docs/COMPLETE-GUIDE.md)
- [VLESS Reality гайд](docs/vless-guide.md)
- [Примеры конфигураций](docs/configuration-examples.md)
- [Белые списки](docs/whitelist-guide.md)

## 🆘 Поддержка

- **GitHub Issues**: [Сообщить о проблеме](https://github.com/sjgagahvabw/Vpn-olcrtc/issues)
- **Telegram**: [@openlibrecommunity](https://t.me/openlibrecommunity)

## 📝 Changelog

### v2.0 (2026-05-14)
- ✅ Упрощенная установка (одна команда)
- ✅ Маскировка под Wildberries Stream
- ✅ Белые списки для российских сайтов
- ✅ Исправлены уязвимости безопасности
- ✅ Улучшенная веб-панель

### v1.0 (оригинальный olcRTC)
- VLESS Reality транспорт
- WebRTC туннелирование
- Базовая веб-панель

## ⚖️ Лицензия

WTFPL - Do What The Fuck You Want To Public License

## 🙏 Благодарности

- [olcRTC](https://github.com/openlibrecommunity/olcrtc) - оригинальный проект
- [Xray-core](https://github.com/XTLS/Xray-core) - VLESS протокол
- Сообщество за тестирование

---

**Сделано для свободного интернета** 🌐
