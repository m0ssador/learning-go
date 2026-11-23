# learning-go

Gateway отдаёт HTTP API поверх Ledger. Бюджеты и расходы лежат в PostgreSQL. Список бюджетов кэшируется в Redis на 30 секунд. Сводка за период считается параллельно, отдельной горутиной на категорию.

## Слои

Зависимости направлены к домену.

`functions/ledger/internal/domain` хранит `Budget`, `Transaction` и их `Validate`. Там же интерфейсы репозиториев: бюджет сохраняется, читается по категории и списком; расход добавляется, читается списком, суммируется по категории и собирается в сводку за период. Этот пакет не импортирует SQL и HTTP.

`functions/ledger/internal/repository/pg` выполняет эти операции в PostgreSQL. `internal/repository/cached` кэширует список бюджетов в Redis на 30 секунд и удаляет ключ `budgets:all` после сохранения бюджета.

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

- `400` — неверный JSON, данные не прошли `Validate`, у отчёта неверные даты или у пакетного импорта неверный `workers`;
- `409` — сумма по категории превышает лимит, тело `{"error":"budget exceeded"}`;
- `504` — запрос не уложился в таймаут, тело `{"error":"request timeout"}`;
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

`POST /api/transactions/bulk` принимает JSON-массив транзакций в том же формате, что и одиночный `POST /api/transactions`: `amount`, `category`, `description`, `date`. Ledger разбирает массив конвейером. Задания идут в канал `jobs`, их читают N воркеров. Каждый воркер вызывает `AddTransaction`: проверка полей, лимит категории и запись в базу. Итог каждой позиции уходит в канал `results`. Канал `jobs` закрывается после отправки всех заданий. Канал `results` закрывается, когда `sync.WaitGroup` дождётся всех воркеров. Отдельная горутина-коллектор читает `results` и собирает сводку. Счётчики `accepted` и `rejected` увеличиваются через `atomic`. Список ошибок дописывает только коллектор. Число воркеров задаётся query-параметром `workers`. Если параметра нет, берётся `DefaultImportWorkers` — 4. `workers` меньше 1 или не число дают `400` и `{"error":"invalid workers"}`.

Успех импорта — `200`, даже если часть элементов отклонена:

```json
{"accepted":1,"rejected":2,"errors":[{"index":1,"error":"transaction amount must not be zero"},{"index":2,"error":"budget exceeded"}]}
```

`index` — позиция в исходном массиве, с нуля. Отклонённая транзакция в базу не пишется. Если запрос оборван по таймауту или клиент отменил его, ответ тот же, что у остальных маршрутов `/api`: `504` и `{"error":"request timeout"}`. Воркеры на `ctx.Done()` прекращают брать новые задания.

`POST /api/budgets` создаёт или обновляет бюджет категории. Тело: `category`, `limit`. Успех: `201` и бюджет.

`GET /api/budgets` возвращает массив бюджетов и `200`. Если записей нет, тело равно `[]`.

`GET /api/reports/summary?from=YYYY-MM-DD&to=YYYY-MM-DD` возвращает `200` и объект сумм по категориям, например `{"еда":450,"транспорт":89}`. Даты включительно. Если расходов за период нет, тело равно `{}`. Ledger берёт категории за период и считает сумму каждой в отдельной горутине. Пока расчёт идёт, раз в 400 мс в журнал пишется `идёт расчёт отчёта`. По окончании пишется `расчёт отчёта завершён`, при отмене контекста — `расчёт отчёта остановлен`.

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

Пакетный импорт, два воркера. Вторая транзакция с нулевой суммой отклоняется, третья не проходит лимит:

```shell
curl -X POST "http://localhost:8080/api/transactions/bulk?workers=2" \
-H "Content-Type: application/json" \
-d '[{"amount":100,"category":"еда","description":"хлеб","date":"2025-09-10"},{"amount":0,"category":"еда","description":"пусто","date":"2025-09-10"},{"amount":6000,"category":"еда","description":"сверх","date":"2025-09-11"}]'
```

Список транзакций:

```shell
curl http://localhost:8080/api/transactions
```

Список бюджетов:

```shell
curl http://localhost:8080/api/budgets
```

Сводка за сентябрь:

```shell
curl "http://localhost:8080/api/reports/summary?from=2025-09-01&to=2025-09-30"
```

Пример ответа, если за сентябрь есть только обед на 450:

```json
{"еда":450}
```

## Таймауты запросов

На все маршруты `/api` вешается middleware `RequestTimeout`. По умолчанию лимит `2s`, константа `DefaultRequestTimeout`. Другое значение задаётся переменной `REQUEST_TIMEOUT` в формате Go (`50ms`, `1s`, `2s`). Пустое или неразобранное значение оставляет 2 секунды. При старте Gateway пишет выбранный лимит в журнал.

Middleware кладёт `context.WithTimeout` в запрос. Ledger передаёт этот контекст в `QueryContext` и `ExecContext`. Если время вышло или клиент оборвал соединение, API отвечает `504` и `{"error":"request timeout"}`. Тикер расчёта при этом останавливается.

В PowerShell те же запросы вызываются через `curl.exe`. Продолжение строки — обратная кавычка, а не `\`. На этой машине проверка шла на `127.0.0.1:8080`: адрес `0.0.0.0:8080` занимал Docker.

## Тесты

Команды запускаются из каталога сервиса, `functions/ledger` или `functions/gateway`. Для проверок базы и Redis нужны `DATABASE_URL` и `REDIS_ADDR`. Без `DATABASE_URL` эти проверки пропускаются.

```powershell
go test ./... -race
go test ./... -cover -coverprofile=cover.out
go tool cover -func=cover.out
go tool cover -html=cover.out -o cover.html
```

С подключенной базой `go test ./... -cover` даёт 38.0% для пакета `ledger` и 58.4% для `gateway/internal/api`. `go tool cover -func` по модулю: Ledger 35.1% (в сумму входят `cmd/ledger`, `internal/db` и `internal/cache` без тестов), Gateway 62.8%.
