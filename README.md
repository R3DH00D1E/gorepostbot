# VK to TG RepostBot on GOLang

Бот для репоста записей из ВК в Telegram.

Дань уважения моему хорошему [другу](github.com/dx3mod) и его [оригинальному](https://github.com/dx3mod/repostbot/tree/master) репостботу.

## Настройка

Для работы бота необходимо установить следующие переменные окружения:

- `TG_TOKEN` - токен Telegram бота
- `VK_TOKEN` - токен доступа к API ВКонтакте
- `TARGET_USER` - идентификатор пользователя ВКонтакте (например, "durov")
- `TARGET_CHAT` - ID чата Telegram для репоста
- `CACHE_FILE` - путь к файлу кэша (по умолчанию "cache.json")
- `DEBUG` - включение отладочного режима (1 или true)
- `INTERVAL` - интервал проверки новых постов (в секундах, по умолчанию 120)

## Запуск

```bash
go run main.go
```

## Сборка

```bash
go build -o vktgbot main.go
```

Многое нужно фиксить...
