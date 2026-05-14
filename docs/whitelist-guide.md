# Руководство по белым спискам (Split Tunneling)

## Что такое белые списки?

Белые списки (Split Tunneling) — это технология, которая позволяет направлять трафик к разным сайтам по разным маршрутам:

- **Российские сайты** → напрямую (без VPN) — быстрее и экономит трафик
- **Заблокированные сайты** → через VPN — обход блокировок

## Преимущества

✅ **Скорость** — российские сайты открываются быстрее  
✅ **Экономия** — меньше нагрузка на VPN сервер  
✅ **Удобство** — не нужно постоянно включать/выключать VPN  
✅ **Безопасность** — банки и госуслуги работают напрямую  

## Как это работает

```
┌─────────────────────────────────────────────┐
│           Ваш компьютер                     │
│                                             │
│  ┌──────────────────────────────────────┐  │
│  │   VPN-olcRTC Client                  │  │
│  │   (с белыми списками)                │  │
│  └──────────┬───────────────────┬───────┘  │
│             │                   │           │
└─────────────┼───────────────────┼───────────┘
              │                   │
              │                   │
    ┌─────────▼─────────┐  ┌──────▼──────────┐
    │  Прямое           │  │  Через VPN      │
    │  соединение       │  │  сервер         │
    └─────────┬─────────┘  └──────┬──────────┘
              │                   │
    ┌─────────▼─────────┐  ┌──────▼──────────┐
    │ Российские сайты: │  │ Заблокированные:│
    │ • Яндекс          │  │ • YouTube       │
    │ • Госуслуги       │  │ • Twitter       │
    │ • Сбербанк        │  │ • Instagram     │
    │ • VK              │  │ • Facebook      │
    └───────────────────┘  └─────────────────┘
```

## Автоматическая настройка

VPN-olcRTC автоматически настраивает белые списки при установке:

```bash
curl -fsSL https://raw.githubusercontent.com/sjgagahvabw/Vpn-olcrtc/main/easy-install.sh | sudo bash
```

Скрипт автоматически:
1. Скачивает актуальные списки российских IP/доменов
2. Настраивает маршрутизацию
3. Создает правила для автообновления

## Ручная настройка

### Для Linux/macOS клиента

Создайте файл `whitelist.txt`:

```bash
# Российские домены (идут напрямую)
yandex.ru
vk.com
mail.ru
gosuslugi.ru
sberbank.ru
vtb.ru
tinkoff.ru
ozon.ru
wildberries.ru
avito.ru
```

Запустите клиент с белым списком:

```bash
./olcrtc \
  -mode cnc \
  -transport vless \
  -carrier wbstream \
  -id "ваш-сервер:443" \
  -client-id "ваш-uuid" \
  -key "ваш-ключ" \
  -socks-port 1080 \
  -whitelist whitelist.txt \
  -data ./data
```

### Для Windows клиента

Создайте файл `whitelist.txt` в папке с клиентом и добавьте домены (по одному на строку).

Запустите клиент через `start-client.bat` — белый список применится автоматически.

### Для Android (V2rayNG)

1. Откройте V2rayNG
2. Настройки → Routing Settings
3. Выберите "Bypass mainland China"
4. Или создайте Custom Rules:

```
domain:yandex.ru
domain:vk.com
domain:gosuslugi.ru
geoip:ru
```

### Для iOS (Shadowrocket)

1. Откройте Shadowrocket
2. Настройки → Routing
3. Выберите "Bypass China"
4. Или создайте Custom Rules в Config → Edit Config

## Готовые списки

### Российские домены (топ-100)

```
# Поисковики и почта
yandex.ru
mail.ru
rambler.ru

# Соцсети
vk.com
ok.ru
dzen.ru

# Госуслуги
gosuslugi.ru
mos.ru
nalog.gov.ru
pfr.gov.ru

# Банки
sberbank.ru
vtb.ru
alfabank.ru
tinkoff.ru
raiffeisen.ru

# E-commerce
ozon.ru
wildberries.ru
avito.ru
youla.ru
aliexpress.ru

# Новости
rbc.ru
tass.ru
ria.ru
interfax.ru
lenta.ru

# Видео и музыка
rutube.ru
zvuk.com
yandex.music

# Другое
2gis.ru
kinopoisk.ru
habr.com
```

### Заблокированные сайты (идут через VPN)

```
# Соцсети
facebook.com
instagram.com
twitter.com
x.com
threads.net

# Видео
youtube.com
youtu.be
twitch.tv

# Мессенджеры
telegram.org
discord.com
whatsapp.com

# Новости
bbc.com
cnn.com
meduza.io
novayagazeta.eu

# Другое
linkedin.com
medium.com
reddit.com
```

## Автообновление списков

VPN-olcRTC автоматически обновляет белые списки каждые 24 часа из источников:

- [antifilter.download](https://antifilter.download) — актуальные списки для России
- [zapret-info](https://github.com/zapret-info/z-i) — реестр заблокированных сайтов
- [russia-v2ray-rules](https://github.com/runetfreedom/russia-v2ray-rules-dat) — правила маршрутизации

### Ручное обновление

```bash
# На сервере
systemctl restart vpn-olcrtc

# На клиенте
./olcrtc-update-whitelist.sh
```

## Проверка работы

### Проверка что сайт идет напрямую

```bash
# Отключите VPN и проверьте IP
curl ifconfig.me

# Включите VPN и проверьте снова
curl ifconfig.me

# Проверьте российский сайт (должен показать ваш реальный IP)
curl --proxy socks5://127.0.0.1:1080 -s https://yandex.ru/internet | grep "IP"

# Проверьте заблокированный сайт (должен показать IP VPN сервера)
curl --proxy socks5://127.0.0.1:1080 ifconfig.me
```

### Проверка маршрутов (Linux)

```bash
# Показать все маршруты
ip route show

# Проверить маршрут к конкретному сайту
ip route get $(dig +short yandex.ru | head -1)
ip route get $(dig +short youtube.com | head -1)
```

## Настройка браузера

### Firefox

1. Настройки → Сеть → Параметры подключения
2. Выберите "Ручная настройка прокси"
3. SOCKS Host: `127.0.0.1`, Port: `1080`
4. Выберите "SOCKS v5"
5. Включите "Использовать прокси для DNS"

### Chrome/Edge

Используйте расширение [Proxy SwitchyOmega](https://chrome.google.com/webstore/detail/proxy-switchyomega/padekgcemlokbadohgkifijomclgjgif):

1. Установите расширение
2. Создайте профиль "VPN-olcRTC"
3. Protocol: SOCKS5, Server: 127.0.0.1, Port: 1080
4. Создайте Auto Switch профиль с правилами из whitelist.txt

## Troubleshooting

### Все сайты идут через VPN

**Проблема:** Белый список не работает, все сайты идут через VPN.

**Решение:**
```bash
# Проверьте что whitelist.txt существует
ls -la whitelist.txt

# Проверьте формат файла (должен быть Unix LF, не Windows CRLF)
file whitelist.txt

# Конвертируйте если нужно
dos2unix whitelist.txt

# Перезапустите клиент
pkill olcrtc
./start-client.sh
```

### Российские сайты не открываются

**Проблема:** Сайты из белого списка не открываются.

**Решение:**
```bash
# Проверьте DNS
nslookup yandex.ru

# Попробуйте другой DNS
echo "nameserver 8.8.8.8" > /etc/resolv.conf

# Проверьте что домен правильно написан в whitelist.txt
cat whitelist.txt | grep yandex
```

### Заблокированные сайты не открываются через VPN

**Проблема:** YouTube/Twitter не открываются даже через VPN.

**Решение:**
```bash
# Проверьте что VPN работает
curl --proxy socks5://127.0.0.1:1080 ifconfig.me

# Проверьте логи клиента
tail -f logs/client.log

# Проверьте что сервер работает
systemctl status vpn-olcrtc
```

## Дополнительные источники

- [antifilter.download](https://antifilter.download) — списки для обхода блокировок
- [zapret-info](https://github.com/zapret-info/z-i) — реестр заблокированных сайтов РФ
- [GoodbyeDPI](https://github.com/ValdikSS/GoodbyeDPI) — обход DPI без VPN
- [zapret](https://github.com/bol-van/zapret) — продвинутый обход блокировок

## Поддержка

Если у вас проблемы с белыми списками:

1. Проверьте [Issues на GitHub](https://github.com/sjgagahvabw/Vpn-olcrtc/issues)
2. Создайте новый Issue с описанием проблемы
3. Напишите в [Telegram](https://t.me/openlibrecommunity)

---

**Обновлено:** 2026-05-14
