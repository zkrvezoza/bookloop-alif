# BookLoop

Онлайн-библиотека — REST API на Go. Финальный проект Alif Academy.

## Запуск

\`\`\`bash
docker compose up --build
\`\`\`

После старта: `GET http://localhost:8080/health` → `{"status":"ok"}`

## Роли

- `user` (читатель) — каталог, взять книгу, очередь, свои книги, возврат
- `librarian` (библиотекарь) — CRUD каталога, все выдачи, статистика, возврат/потеря, сроки

Роль выбирается при регистрации (`POST /auth/register`).

## Аутентификация

1. `POST /auth/register` — `{login, password, role}`
2. `POST /auth/login` — `{login, password}` → `{access_token, refresh_token}`
3. Запросы: заголовок `Authorization: Bearer <access_token>`
4. `POST /auth/refresh` — `{refresh_token}` → новая пара (ротация)
5. `POST /auth/logout` — `{refresh_token}` → отзыв

## Основные эндпоинты

| Метод | Путь | Роль | Описание |
|---|---|---|---|
| GET | /api/v1/books | любой | каталог, поиск, пагинация |
| POST | /api/v1/books/:id/loan | user | взять книгу |
| GET | /api/v1/loans | user | свои выдачи |
| PUT | /api/v1/loans/:id/return | user | вернуть книгу |
| POST | /api/v1/manage/books | librarian | добавить книгу |
| GET | /api/v1/manage/loans | librarian | все выдачи |
| GET | /api/v1/manage/stats | librarian | статистика |

## Тесты

\`\`\`bash
go test ./... -cover
\`\`\`

## Переменные окружения

См. `docker-compose.yml`: `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `JWT_SECRET`.