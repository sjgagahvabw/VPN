# 🚀 Быстрый старт VPN-olcRTC

## Для тебя (на сервере)

```bash
# 1. Подключись к VPS по SSH
ssh root@your-server-ip

# 2. Запусти установку одной командой
curl -fsSL https://raw.githubusercontent.com/sjgagahvabw/VPN/master/easy-install.sh | sudo bash

# 3. Готово! Скопируй VLESS URL из вывода
```

## Для клиентов

### Android
1. Установи [V2rayNG](https://github.com/2dust/v2rayNG/releases)
2. Нажми "+" → "Import config from Clipboard"
3. Вставь VLESS URL
4. Подключись

### iOS
1. Установи [Shadowrocket](https://apps.apple.com/app/shadowrocket/id932747118)
2. Нажми "+" → "Type" → "Subscribe"
3. Вставь VLESS URL
4. Подключись

### Windows/macOS/Linux
1. Открой `http://your-server-ip:8080`
2. Скачай клиент для своей ОС
3. Запусти

## Что получилось

✅ **Установка одной командой** — просто скопировал и запустил  
✅ **VLESS Reality** — высокая скорость, низкая задержка  
✅ **Маскировка под WB** — использует `stream.wb.ru`  
✅ **Белые списки** — российские сайты напрямую, остальное через VPN  
✅ **Веб-панель** — управление через браузер  
✅ **Безопасность** — исправлены все уязвимости  

## Управление

```bash
# Статус
systemctl status vpn-olcrtc

# Перезапуск
systemctl restart vpn-olcrtc

# Логи
journalctl -u vpn-olcrtc -f

# Информация о подключении
cat /opt/vpn-olcrtc/connection-info.txt
```

## Ссылки

- **GitHub**: https://github.com/sjgagahvabw/VPN
- **Документация**: https://github.com/sjgagahvabw/VPN/blob/master/docs/whitelist-guide.md
- **Telegram**: [@openlibrecommunity](https://t.me/openlibrecommunity)

---

**Готово!** Теперь у тебя есть простой VPN который работает через VLESS + маскируется под Wildberries 🎉
