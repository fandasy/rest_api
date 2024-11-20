Привет читатель, это мой пэт-проект, реализация REST API с использованием фреймворка Gin

Функционал
---

- Создание изображений из ASCII символов из ссылок на изображение
- Перенаправление на изображение по его id
- Кеширование данных при помощи redis для поиска по id
- Контроллер запросов

Поддерживаемые форматы png, jpg, jpeg, webp

YAML
---
```
env: "local"

storage_path: "user=postgres dbname=dbname password=password host=localhost port=5432 sslmode=disable"

redis:
  addr: "localhost:6379"
  password: ""
  db: 0
  ttl: 5m

# Настройки форматирования изображений
image_settings:
  maxWidth: 5000
  maxHeight: 5000
  chars: "@%#*+=:~-. "  | символы использующиеся в создании изображений (Тёмный - Светлый)

http_server:
  address: "localhost:8082"
  timeout: 4s
  idle_timeout: 60s

req_limit:
  max_num_req: 5 |  максимальное кол-во запросов в интервал time_slice
  time_slice: 1s |  Промежуток времени
  ban_time: 60s  |  Время бана
```

В зависимости от env запускаются типы логирования:
```
local - text, уровень Debug, вывод в консоль
dev   - json, уровень Debug, вывод в консоль
prod  - json, уровень Info,  вывод в консоль
```

Запуск
---

Для запуска rest api существует два варианта:
- Запустить в консоли с флагами: -config
- Использовать переменные среды: CONFIG_PATH

```
start name.exe -config local.yaml
```
