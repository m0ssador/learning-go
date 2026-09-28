# Домашнее задание 2

Проект состоит из двух отдельных Go-модулей.

## Gateway

HTTP-шлюз. `GET /ping` отвечает `pong` со статусом 200.

```powershell
cd functions/gateway
go run .
```

Прослушиватель `http://localhost:8080`.

![alt text](../docs/functions/image.png)

Проверка:

```powershell
curl.exe http://localhost:8080/ping
```

Ожидаемый ответ: `pong`.

![alt text](../docs/functions/image-1.png)

## Ledger

Бизнес-логика (финансовые транзакции).

```powershell
cd functions/ledger
go run .
```

![alt text](../docs/functions/image-2.png)
