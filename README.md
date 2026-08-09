# file-drop

Self-destructing file and text sharing application built with Go, PostgreSQL, and React.

## Features

- **File & Text Drops**: Upload binary files or text snippets.
- **Protected Text View**: Recipient mode hides text payload initially until clicked.
- **Expiration & Self-Destruction**: Expiration options (`10m`, `1h`, `1d`, `7d`, `3d`) + Burn After Download auto-cleanup.
- **Upload Progress**: Real-time upload progress, speed calculation, and upload cancellation.
- **Security & Limits**: Configurable max upload size, IP rate limiting, security headers, and graceful server shutdown.

## Tech Stack

- **Backend**: Go (Chi v5, pgx/v5)
- **Database**: PostgreSQL 16
- **Frontend**: React 19, TypeScript, Tailwind CSS, Base UI

## Environment Variables

- `PORT` (default: `8000`)
- `STORAGE_PATH` (default: `./blobs`)
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`
- `MAX_UPLOAD_SIZE` (default: `100` MB)
- `RATE_LIMIT_UPLOADS` (default: `10` req/min/IP)

## Quick Start

Run with Docker Compose:

```bash
docker compose up --build -d
```

Or run Go server locally:

```bash
go run cmd/server/main.go
```

## API

- `POST /api/upload` - Upload file or text drop
- `GET /api/f/:id` - Fetch drop metadata or payload (`?download=true`)
- `DELETE /api/f/:id` - Delete drop
- `GET /api/config` - Fetch server config
- `GET /healthz` - Health check

## License

MIT
