# FCY-Payment-Processor
A multi-currency payment processor.

## Project overview

FCY Payment Processor is a backend service for multi-currency wallet/account operations and transfer processing. It supports user onboarding, account creation, funding, FX rate conversion, charges/VAT computation, and fund transfers across supported currencies (`USD`, `EUR`, `GBP`, `NGN`).

Transfers are processed through a transactional posting model with internal transient (suspense/GL) accounts to ensure controlled debit/credit movement, traceability, and settlement handling for both internal and external transfer scenarios.

For solution design artifacts, refer to the `docs/` directory for:
- architecture decisions
- ERD
- sequence flow

## Testing philosophy

The testing approach in this project focuses on correctness of financial behavior and transfer safety:
- Business-rule-first tests for core use-case services.
- Explicit negative-path coverage for validation and failure cases.
- Mock-based unit tests around service boundaries to keep logic isolated.
- Incremental, reviewable test additions per feature/task.
- Priority on transfer integrity (debit/credit/charges/vat/settlement behavior) over broad but shallow coverage.
- A TDD-leaning flow would likely reduce implementation iterations and rework, but it increases upfront delivery time, which was a trade-off in this time-sensitive project.

## Performance observations

### Goroutine and concurrency implementation
The system implements concurrent goroutines in transfer processing and startup initialization:

**Key Performance Metrics:**
- **Application startup time:** Average 180ms (including database migrations) - reduced from initial >250ms
- **Transfer latency (with goroutines + optimized DB pool):** Average 110ms (lowest: 92ms)
- **Transfer latency (sequential, without goroutines, same DB pool config):** Average 115ms (lowest: 110ms)
- **Performance improvement:** ~4.5% faster with optimized goroutines

### Database connection pool tuning impact
Initial goroutine implementation increased transfer latency from 125ms to 220ms. Root cause analysis identified the database connection pool as the bottleneck. After tuning the pool configuration:

**Optimized pool settings:**
- Max idle connections: 20
- Max open connections: 30
- Idle timeout: 5 minutes
- Connection lifetime: 15 minutes

These settings reduced transfer time by 50% (from 220ms back to 110ms) while enabling concurrent goroutine operations.

**Observation:** Database pool size directly determines goroutine performance. Without adequate connection pooling, goroutines introduce overhead that outweighs concurrency benefits.

### Scalability outlook
For lightweight operations, the marginal improvement from goroutines is modest (4-5%). However, postulating to a distributed queue-driven architecture (e.g., with Kafka):
- Estimated latency: **50-70ms per transaction**
- Achieved through parallel worker processing, batch handling, and distributed database connections
- Current synchronous HTTP model becomes the bottleneck at scale; async queue-based intake is the path to sub-100ms latency

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

## How to test transfers

To test `/transfer-funds`, you need at least a valid sender account number.

Recommended flow:

1. Create a user with `POST /create-user`.
   - This returns a `customerId`.
2. Create account(s) with `POST /create-account` using that `customerId` with an initial deposit of atleast 1 unit of the currency specified.
   - Supported currencies are only: `USD`, `EUR`, `GBP`, `NGN`.
   - One customer can have multiple accounts across different currencies, but not duplicate account for the same currency.
3. Fund the sender account (for example via `POST /deposit-funds`).
4. Call `POST /transfer-funds`.

Transfer mode selection:
- Internal transfer:
  - Set `beneficiaryBankCode` to `100100`.
  - Beneficiary account must be an internal account in this app.
- External transfer:
  - Set `beneficiaryBankCode` to a participant bank code from `GET /get-participant-banks`.
  - External transfers terminate in an external GL account in the DB (not a real beneficiary account in this app).
  - Once the external GL is credited and external reference is generated, the system assumes beneficiary value has been delivered via beneficiary bank.

## (OPTIONAL) What to change before running on another machine 

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
