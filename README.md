# learning-go

Gateway отдаёт HTTP API поверх Ledger. Транзакции и бюджеты хранятся в памяти процесса.

```powershell
cd functions/gateway
go run .
```

Сервер слушает `http://localhost:8080`. Ответы `/api` приходят с заголовком `Content-Type: application/json; charset=utf-8`. Ошибка имеет вид `{"error":"..."}`:

- `400` — неверный JSON или данные не прошли `Validate`;
- `409` — сумма по категории превышает лимит, тело `{"error":"budget exceeded"}`;
- `500` — прочая внутренняя ошибка.

`GET /ping` по-прежнему отвечает текстом `pong`.

## Эндпоинты

`POST /api/transactions` создаёт транзакцию. Тело: `amount`, `category`, `description`, `date`. Успех: `201` и созданная транзакция, включая `id`.

`GET /api/transactions` возвращает массив транзакций и `200`. Если записей нет, тело равно `[]`.

`POST /api/budgets` создаёт или обновляет бюджет категории. Тело: `category`, `limit`. Успех: `201` и бюджет.

`GET /api/budgets` возвращает массив бюджетов и `200`. Если записей нет, тело равно `[]`.

Каждый запрос пишется в консоль сервера: метод, путь и длительность.

## Примеры

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

Задать бюджет категории «еда»:

```shell
curl -X POST http://localhost:8080/api/budgets \
-H "Content-Type: application/json" \
-d '{"category":"еда","limit":5000}'
```

Список бюджетов:

```shell
curl http://localhost:8080/api/budgets
```

В PowerShell те же запросы вызываются через `curl.exe`. Продолжение строки — обратная кавычка, а не `\`.
