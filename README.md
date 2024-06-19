# go-musthave-metrics-tpl

Шаблон репозитория для трека «Сервер сбора метрик и алертинга».

## Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` — адрес вашего репозитория на GitHub без префикса `https://`) для создания модуля.

## Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m main template https://github.com/Yandex-Practicum/go-musthave-metrics-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/main .github
```

Затем добавьте полученные изменения в свой репозиторий.

## Запуск автотестов

Для успешного запуска автотестов называйте ветки `iter<number>`, где `<number>` — порядковый номер инкремента. Например, в ветке с названием `iter4` запустятся автотесты для инкрементов с первого по четвёртый.

При мёрже ветки с инкрементом в основную ветку `main` будут запускаться все автотесты.

Подробнее про локальный и автоматический запуск читайте в [README автотестов](https://github.com/Yandex-Practicum/go-autotests).

## Локальное тестирование

export TEST_CRYPT_PUB_KEY=/Users/fanishadeev/go/src/github.com/sotavant/yandex-metrics/cmd/crypto/public.pem TEST_CRYPT_PRIV_KEY=/Users/fanishadeev/go/src/github.com/sotavant/yandex-metrics/cmd/crypto/private.pem TEST_TABLE_NAME=test_table TEST_DATABASE_DSN="postgres://ya_metrics:ya_metrics@localhost:5432/ya_metrics?sslmode=disable"; go test ./...

-17849.57kB

## Запуск с флагами
go build -ldflags "-X main.buildVersion=0.1 -X 'main.buildDate=$(date +'%Y/%m/%d')' -X 'main.buildCommit=$(git rev-parse --short HEAD)'"