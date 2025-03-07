# VK - TG RepostBot on GOLang

Бот для автоматического репоста записей из ВКонтакте в Telegram с поддержкой вложений и обновления изменённых постов.

Дань уважения моему хорошему [другу](https://github.com/dx3mod) и его [оригинальному](https://github.com/dx3mod/repostbot/tree/master) репостботу, которого я, с позволения автора, переписал на гошку и доработа(или испоганил).

## Возможности

- Автоматический репост текста из ВК в Telegram
- Поддержка вложений (фотографии)
- Обновление изменённых постов ВК в Telegram
- Разбиение длинных сообщений на несколько частей
- Кэширование обработанных постов
- Поддержка как стандартной, так и Systemd установки

## Конфигурация

Бот использует JSON-файл конфигурации. Путь к файлу можно задать через переменную окружения `CONFIG_PATH` или использовать файл `config.json` в текущей директории.

Пример `config.json`:
```json
{
    "vk_token": "your_vk_token_here",
    "tg_token": "your_telegram_token_here",
    "chat_id": "-1001234567890",
    "poll_interval": 10,
    "target_user": "user_id_vk",
    "cache_file": "path/to/cache.json"
}
```

Параметры конфигурации:
- `vk_token` - токен доступа к API ВКонтакте
- `tg_token` - токен Telegram бота
- `chat_id` - ID чата Telegram для репоста (использовать отрицательные значения для групп)
- `poll_interval` - интервал проверки новых постов (в секундах, по умолчанию 10)
- `target_user` - идентификатор пользователя/группы ВКонтакте
- `cache_file` - путь к файлу кэша (по умолчанию "cache.json")

## Запуск

```bash
go run main.go
```

## Сборка

```bash
go build -o repostbot main.go
```
## Установка через systemd

1. Создайте файл конфигурации в домашней папке:
```bash
mkdir -p ~/repostbot
cp config.json ~/repostbot/
```

2. Поместите исполняемый файл в /usr/local/bin:
```bash
sudo cp repostbot /usr/local/bin/
sudo chmod +x /usr/local/bin/repostbot
```

3. Настройте systemd сервис:
```bash
sudo cp deploy/repostbot.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable repostbot.service
sudo systemctl start repostbot.service
```

## Автоматический деплой
Проект настроен для автоматического развёртывания через GitHub Actions. Конфигурация находится в deploy.yml.

Для автоматического деплоя требуется настроить следующие секреты в GitHub:
- `TG_TOKEN` - токен Telegram бота
- `VK_TOKEN` - токен доступа к API ВКонтакте
- `TARGET_USER` - идентификатор пользователя/группы ВКонтакте
- `CHAT_ID` - ID чата Telegram для репоста
- `POLL_INTERVAL` - интервал проверки новых постов
- `USER` - имя пользователя на сервере для установки

## Зависимости

- Стандартные библиотеки Go для работы с HTTP и JSON
- Внешних зависимостей нет

Многое нужно фиксить...
