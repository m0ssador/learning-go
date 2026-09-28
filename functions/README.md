# Gateway и Ledger

Два отдельных Go-модуля в одном репозитории.

## Gateway

HTTP-шлюз. `GET /ping` отвечает `pong` со статусом 200.

```powershell
cd functions/gateway
go run .
```

Сервер слушает `http://localhost:8080`. Проверка:

```powershell
curl.exe http://localhost:8080/ping
```

Ожидаемый ответ: `pong`.

## Ledger

Бизнес-логика. При запуске печатает `Ledger service started`, добавляет три тестовые транзакции в память и выводит список.

```powershell
cd functions/ledger
go run .
```

gRPC-сервер пока не поднимается: модуль только стартует и показывает транзакции в консоли.
