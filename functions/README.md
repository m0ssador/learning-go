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

# Домашнее задание 3

В сервисе Ledger бюджеты хранятся в map[string]Budget и проверяются в AddTransaction до записи транзакции.

Budget содержит **Category**, **Limit** и **Period**. Лимит действует на всю категорию: **Period** сохраняется, а сумма сравнивается с Limit без разбиения по месяцу или году. **SetBudget** добавляет или обновляет запись. **LoadBudgets** читает JSON-массив, при ошибке чтения возвращает read budgets: ..., при ошибке разбора — parse budgets: ..., и только после проверки всех элементов вызывает **SetBudget**.

Если для категории бюджет задан и текущая сумма транзакций вместе с новой превышает **Limit**, **AddTransaction** возвращает budget exceeded и ничего не записывает. Сумма, равная лимиту, проходит. Категория без бюджета не ограничивается.

![alt text](../docs/functions/image-3.png)