# Code Playground

A Go and Python playground with authentication, file storage, and isolated server-side code execution via Docker.

## Setup

**Prerequisites:** Docker must be running (Docker Desktop on macOS/Windows). Pull the runner images once:

```bash
docker pull golang:1.21-alpine
docker pull python:3.12-alpine
```

**Frontend:**

```bash
npm install
npm run dev
```

**Backend** (in another terminal):

```bash
cd backend
go run main.go
```

The frontend proxies `/api` to `http://localhost:3000` in dev mode. The backend needs access to the Docker socket (`/var/run/docker.sock`) to run user code.

## Docker

**Development** (без SSL, прямой доступ по порту):

```bash
docker compose up --build
```

Open http://localhost:8081

**Production** (Caddy + SSL для xionic.ru):

```bash
docker compose -f docker-compose.prod.yml up -d --build
```

Перед запуском:

1. Настройте DNS: `xionic.ru` и `www.xionic.ru` → IP сервера
2. Убедитесь, что порты 80 и 443 открыты
3. Caddy автоматически получит SSL-сертификат от Let's Encrypt
4. На сервере должен быть установлен Docker; backend монтирует `/var/run/docker.sock`

## Roles

- **admin** — видит и управляет всеми файлами (list, get, update, delete любые)
- **student** — только свои файлы

Роль admin назначается при регистрации: если email совпадает с `ADMIN_EMAIL`, пользователь получает роль admin.

Скопируйте `.env.example` в `.env` и настройте:

```bash
cp .env.example .env
```

В `.env`:

- `ADMIN_EMAIL` — email админа
- `ADMIN_PASSWORD` — пароль админа (создаётся при первом запуске, если пользователя нет)
- `JWT_SECRET` — секрет для JWT (обязательно сменить в production)
- `GO_RUNNER_IMAGE` — Docker-образ для выполнения Go (по умолчанию `golang:1.21-alpine`)
- `PYTHON_RUNNER_IMAGE` — Docker-образ для выполнения Python (по умолчанию `python:3.12-alpine`)
- `RUN_TIMEOUT` — таймаут выполнения кода (по умолчанию `60s`)

Docker Compose автоматически подхватывает `.env`. При локальном запуске бекенд загружает `.env` из корня проекта.

## Features

- JWT authentication (register/login)
- Per-user file storage (create, save, delete files)
- Autosave with debounce (enabled by default; admin can disable per file)
- Admin watch mode (read-only editor, polling updates after save)
- Monaco Editor with Go and Python syntax highlighting
- Server-side code execution in isolated Docker containers
- Console output from stdout and compile/runtime errors from stderr

## User code requirements

- **Go** (`.go`): valid program with `package main` and `func main()`
- **Python** (`.py`): script executed with `python main.py`

Language is chosen when creating a file and determined by file extension. See [docs/code-playground.md](docs/code-playground.md) for sandbox details.
