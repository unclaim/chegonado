# Инструкции по сборке

Для сборки и запуска проекта вам потребуются:
* Go 1.24.3+
* Docker и Docker Compose

## Локальная сборка

Для сборки исполняемого файла используйте `Makefile`:

```sh
make build
```

Исполняемый файл будет создан в `bin/server`.

## Сборка Docker-образа

Для сборки Docker-образа используйте Docker Compose:

```sh
docker-compose build
```
