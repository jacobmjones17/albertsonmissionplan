# Deploying Albertson Ward Mission Plan

The production artifact is a single Go binary that serves the React SPA and JSON API from one process. Data lives in **PostgreSQL** (`DATABASE_URL`).

**Production domain:** `albertsonwardplan.com`  
**Hosting target:** [Render](https://render.com/) — **free Web Service** + **free Postgres** via **`render.yaml`**.

---

## Free-tier caveats (Render)

- **Postgres:** Free databases **expire 30 days after creation** unless you upgrade ([Render docs](https://render.com/docs/free)). Plan ahead for upgrade or export before expiry.
- **Web:** Free instances **spin down after ~15 minutes idle**; first request after idle may take ~1 minute ([docs](https://render.com/docs/free)).
- **SMTP:** Free Web Services **cannot send outbound mail on ports 25 / 465 / 587**. Password-reset and leader notification emails via classic SMTP often **will not work on Free**; use a paid tier or an HTTPS email API later.

---

## What you need to do (checklist)

1. **Push this repository** to GitHub/GitLab with **`render.yaml`** at the repo root.
2. Render Dashboard → **New → Blueprint** → select repo → branch **`main`**.
3. Provide **`PUBLIC_BASE_URL`** when prompted — start with **`https://<your-service-name>.onrender.com`** (shown after deploy), then change to **`https://albertsonwardplan.com`** once DNS works.
4. After deploy: **Custom domains** on the Web Service → add **`albertsonwardplan.com`** → copy DNS instructions into **Namecheap Advanced DNS**.
5. When production HTTPS works, set **`PUBLIC_BASE_URL=https://albertsonwardplan.com`** (no slash) on the Web Service and redeploy.

`DATABASE_URL` is wired automatically from the **Render Postgres** instance defined in `render.yaml`.

---

## Local development with Postgres

Run PostgreSQL locally (Docker example):

```bash
docker run --name wardmission-pg -e POSTGRES_USER=wardmission -e POSTGRES_PASSWORD=wardmission \
  -e POSTGRES_DB=wardmission -p 5432:5432 -d postgres:16-alpine
```

In `.env` at repo root:

```bash
DATABASE_URL=postgres://wardmission:wardmission@localhost:5432/wardmission?sslmode=disable
SESSION_SECRET=change-me-to-a-long-random-string-at-least-16-chars
COOKIE_SECURE=false
PUBLIC_BASE_URL=http://localhost:8080
```

Then:

```bash
cd backend && go run ./cmd/main.go
```

Frontend against this backend (optional): `cd frontend && npm run dev` with `DEV_TRUST_VITE=true` if needed.

Bootstrap an admin without the UI:

```bash
cd backend && DATABASE_URL='postgres://wardmission:wardmission@127.0.0.1:5432/wardmission?sslmode=disable' \
  go run ./cmd/seedadmin -email you@example.org -password 'your-password'
```

---

## Build locally (binary only)

```bash
cd frontend && npm ci && npm run build
cd ../backend && go build -o wardmission ./cmd/main.go
```

Run with **`DATABASE_URL`** set.

## Docker (single container)

The image expects **`DATABASE_URL`** at runtime (points at any reachable Postgres):

```bash
docker build -t wardmission:latest .
docker run --rm -p 8080:8080 \
  -e DATABASE_URL='postgres://USER:PASS@HOST:5432/DBNAME?sslmode=require' \
  -e SESSION_SECRET="$(openssl rand -hex 24)" \
  -e PUBLIC_BASE_URL=https://albertsonwardplan.com \
  -e COOKIE_SECURE=true \
  wardmission:latest
```

---

## Environment variables

| Variable | Notes |
|----------|--------|
| `DATABASE_URL` | **Required.** PostgreSQL URL (Render sets this from Postgres) |
| `SESSION_SECRET` | Required random string ≥16 chars |
| `PUBLIC_BASE_URL` | Canonical HTTPS URL **without** trailing slash |
| `COOKIE_SECURE` | `true` behind HTTPS |
| `PORT` | Listen port (default `:8080`; Render sets `PORT`) |
| `MAIL_*` | See `.env.example` |

---

## Namecheap + Render DNS

Match **`PUBLIC_BASE_URL`** to the hostname visitors use. Email links use **`PUBLIC_BASE_URL`**.

---

## After deploy

1. Visit **`/admin`**, register the first leader (auto-approved), or use **`seedadmin`** against production **`DATABASE_URL`** (run locally with prod URL — be careful).
2. Plan Postgres lifecycle on Free (upgrade before 30‑day expiry).
3. Confirm **`PUBLIC_BASE_URL`** matches the live URL.
