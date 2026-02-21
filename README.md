# FCY-Payment-Processor
A multi-currency payment processor.

## Run with Docker Compose

1. Install Docker + Docker Compose and ensure Docker daemon is running.
2. Clone the repo and `cd` into project root.
3. If you do not want to change any values, launch directly with defaults:

```bash
docker compose up --build
```

4. If you want custom values, update `docker-compose.yml` first (see section below), then launch:

```bash
docker compose up --build
```

The command starts:
- `db` (PostgreSQL 16)
- `app` (API on `http://localhost:8080`)

The app runs migrations on startup and ensures default rates/internal transient accounts exist.

## What to change before running on another machine

Edit `docker-compose.yml`.

### 1) Database credentials and DSN
Under `services.db.environment`, set:
- `POSTGRES_DB`
- `POSTGRES_USER`
- `POSTGRES_PASSWORD`

Then under `services.app.environment`, update `DATABASE_DSN` to match the same DB values:

```text
DATABASE_DSN=Host=db;Port=5432;Database=<POSTGRES_DB>;Username=<POSTGRES_USER>;Password=<POSTGRES_PASSWORD>;Timeout=30;CommandTimeout=30
```

### 2) Channel authentication values
Under `services.app.environment`, change, if you want to use a different Basic Auth credentials:
- `CHANNEL_ID`
- `CHANNEL_KEY`

Use values appropriate for the target environment.

### 3) Exposed ports (if needed)
If `8080` or `5432` is occupied, change:
- `services.app.ports` (left side host port)
- `services.db.ports` (left side host port)

Example:
- `9000:8080` exposes API on `http://localhost:9000`

### 4) Optional business/config values
Under `services.app.environment`, adjust if you want:
- `GREY_BANK_CODE`
- `CHARGE_PERCENT`, `VAT_PERCENT`, `CHARGE_MIN_AMOUNT`, `CHARGE_MAX_AMOUNT`
- Internal/external GL account numbers

## Useful commands

Start in background:

```bash
docker compose up --build -d
```

View logs:

```bash
docker compose logs -f app
```

Stop:

```bash
docker compose down
```
