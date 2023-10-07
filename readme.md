# go-wg-manager - менеджер Wireguard Server

---

## Для чего нужен

**gwg** - утилита командной строки для автоматического конфигурирования  и администрирования wireguard-сервера.
Поддерживает такие фунции как:

1) Автоматическая настройка конфигурации wireguard server;
2) Автоматическое изменение конфигурации сервера при добавлении пользователя;
3) Автоматическое изменение конфигурации сервера при удалении пользователя;
4) Просмотр состояния сервера через стандартную утилиту wg show;
5) Просмотр подробной статистики на основе стандартной утилиты wg show dump. (дорабатывается)

## Поддерживаемые платформы

- Любой дистрибутив linux на основе Debian.

## Установка

- Скачать архив с [желаемой](https://github.com/PavelMilanov/go-wg-manager/tags) версией:

```bash
curl -o https://github.com/PavelMilanov/go-wg-manager/releases/tag/latest/gwg.tar.gz
```

- Распаковать архив:

```bash
tar -xvzf gwg.tar.gz
```

- Запустить скрипт первичной настройки окружения gwg-manager и установки gwg
   ( **В конце установки будет предложено перезапустить сессию пользоватeля!** ):

```bash
./gwg-utils.sh install
```

- После перезапуска сессии пользователя запустить установку сервера Wireguard.

```bash
gwg install
```
