# learning-go

Gateway отдаёт HTTP API поверх Ledger. Бюджеты и расходы лежат в PostgreSQL. Сводка отчёта и список бюджетов кэшируются в Redis на 30 секунд.

## Слои

Зависимости направлены к домену.

`functions/ledger/internal/domain` хранит `Budget`, `Transaction` и их `Validate`. Там же интерфейсы репозиториев: бюджет сохраняется, читается по категории и списком; расход добавляется, читается списком, суммируется по категории и собирается в сводку за период. Этот пакет не импортирует SQL и HTTP.

`functions/ledger/internal/repository/pg` выполняет эти операции в PostgreSQL. `internal/repository/cached` кэширует список бюджетов и сводку в Redis на 30 секунд и удаляет ключ `budgets:all` после сохранения бюджета.

`functions/ledger/service` проверяет сущность и правило бюджета. Если для категории есть лимит и текущая сумма вместе с новой транзакцией больше лимита, `AddTransaction` возвращает `budget exceeded` и запись не создаётся. К базе сервис обращается только через интерфейсы домена.

`app.New(ctx)` в `functions/ledger/app` читает `DATABASE_URL` или `DB_*`, открывает PostgreSQL, вызывает `PingContext`, подключает Redis, собирает репозитории и возвращает сервис, `closeFn` и ошибку. Миграции фабрика не запускает.

Пакеты `service` и `app` лежат вне `internal`: Gateway — отдельный модуль, а Go не даёт импортировать чужой `internal`.

Gateway в `main` вызывает `app.New`. При ошибке подключения пишет её в лог и завершается. После успеха стоит `defer closeFn()`. `api.New` получает сервис, обработчики вызывают его методы. DTO, пути и коды ответов прежние. Исходники Gateway импортируют `ledger/app` и `ledger/service` и не импортируют `database/sql`, `pgx` и `repository`.

## База и Redis

База по умолчанию — `cashapp`. Строка подключения берётся из `DATABASE_URL`. Если её нет, Ledger собирает её из `DB_HOST` (`localhost`), `DB_PORT` (`5432`), `DB_USER` (`postgres`), `DB_PASS` (`postgres`) и `DB_NAME` (`cashapp`), с `sslmode=disable`.

Redis: `REDIS_ADDR` (`localhost:6379`), `REDIS_DB` (`0`), `REDIS_PASSWORD` (пусто).

Миграции Goose лежат в `functions/ledger/migrations`. Команды запускаются из `functions/ledger`:

```powershell
goose -dir ./migrations postgres $env:DATABASE_URL up
goose -dir ./migrations postgres $env:DATABASE_URL down
```

Пул соединений Ledger: `SetMaxOpenConns(10)`, `SetMaxIdleConns(5)`, `SetConnMaxLifetime(30 * time.Minute)`.

На этой машине порт 5432 занят локальным PostgreSQL, проверочный контейнер слушал `127.0.0.1:5433`. Redis — `localhost:6379`.

## Запуск

Сначала PostgreSQL и Redis, затем `goose ... up`, затем Gateway:

```powershell
cd functions/gateway
$env:DATABASE_URL = "postgres://postgres:postgres@127.0.0.1:5432/cashapp?sslmode=disable"
$env:REDIS_ADDR = "localhost:6379"
go run .
```

Сервер слушает `:8080`. Ответы `/api` приходят с заголовком `Content-Type: application/json; charset=utf-8`. Ошибка имеет вид `{"error":"..."}`:

- `400` — неверный JSON, данные не прошли `Validate` или у отчёта неверные даты;
- `409` — сумма по категории превышает лимит, тело `{"error":"budget exceeded"}`;
- `500` — прочая внутренняя ошибка, в том числе сбой базы.

`GET /ping` по-прежнему отвечает текстом `pong`.

Отдельный сценарий Ledger, из каталога `functions/ledger`:

```powershell
go run ./cmd/ledger
```

Он тоже подключается к PostgreSQL и Redis и читает `budgets.json`.

## Эндпоинты

`POST /api/transactions` создаёт транзакцию. Тело: `amount`, `category`, `description`, `date` в формате `YYYY-MM-DD`. Успех: `201` и созданная транзакция, включая `id`.

`GET /api/transactions` возвращает массив транзакций и `200`. Если записей нет, тело равно `[]`.

`POST /api/budgets` создаёт или обновляет бюджет категории. Тело: `category`, `limit`. Успех: `201` и бюджет.

`GET /api/budgets` возвращает массив бюджетов и `200`. Если записей нет, тело равно `[]`.

`GET /api/reports/summary?from=YYYY-MM-DD&to=YYYY-MM-DD` возвращает массив `{"category","total"}` и `200`. Повтор того же запроса в течение 30 секунд читает Redis. В журнале Ledger пишется `cache hit` или `cache miss`.

Каждый запрос пишется в консоль сервера: метод, путь и длительность.

## Примеры

Задать бюджет категории «еда»:

```shell
curl -X POST http://localhost:8080/api/budgets \
-H "Content-Type: application/json" \
-d '{"category":"еда","limit":5000}'
```

Создать транзакцию:

```shell
curl -X POST http://localhost:8080/api/transactions \
-H "Content-Type: application/json" \
-d '{"amount":450,"category":"еда","description":"ланч","date":"2025-09-10"}'
```

Список транзакций:

```shell
curl http://localhost:8080/api/transactions
```

Список бюджетов:

```shell
curl http://localhost:8080/api/budgets
```

Сводка за сентябрь. Второй вызов в течение 30 секунд берёт тот же JSON из Redis:

```shell
curl "http://localhost:8080/api/reports/summary?from=2025-09-01&to=2025-09-30"
curl "http://localhost:8080/api/reports/summary?from=2025-09-01&to=2025-09-30"
```

В PowerShell те же запросы вызываются через `curl.exe`. Продолжение строки — обратная кавычка, а не `\`. На этой машине проверка шла на `127.0.0.1:8080`: адрес `0.0.0.0:8080` занимал Docker.

## Тесты

Команды запускаются из каталога сервиса, `functions/ledger` или `functions/gateway`. Для проверок базы и Redis нужны `DATABASE_URL` и `REDIS_ADDR`. Без `DATABASE_URL` эти проверки пропускаются.

```powershell
go test ./... -cover -coverprofile=cover.out
go tool cover -func=cover.out
go tool cover -html=cover.out -o cover.html
```

С подключенной базой `go test ./... -cover` даёт 38.0% для пакета `ledger` и 58.4% для `gateway/internal/api`. `go tool cover -func` по модулю: Ledger 35.1% (в сумму входят `cmd/ledger`, `internal/db` и `internal/cache` без тестов), Gateway 62.8%.
