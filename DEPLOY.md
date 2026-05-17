# Deploying Albertson Ward Mission Plan

The production artifact is a single Go binary that serves the React SPA and JSON API from one process. SQLite stores data on disk (`DATABASE_PATH`).

**Production domain:** `albertsonwardplan.com`  
**Hosting target for this repo:** [Render](https://render.com/) via **`render.yaml`** (Dockerfile build).

---

## Important: Render pricing vs SQLite

Render’s **free Web Service tier uses an ephemeral filesystem** — **SQLite would not survive redeploys/restarts reliably**, so **do not run this app on Render Free** if you care about keeping leaders, logins, and ward data.

To use **SQLite on Render**, you need a **paid instance** (this blueprint uses **`plan: starter`**) plus the **persistent disk** defined in `render.yaml` (mounted at **`/data`**). That avoids Fly.io, but **Render still bills for compute + disk** (see [Render pricing](https://render.com/pricing)).

---

## What you need to do (checklist)

1. **Push this repository** to GitHub or GitLab (Render deploys from Git).
2. In the [Render Dashboard](https://dashboard.render.com/), click **New → Blueprint**.
3. Connect the repository and approve the **`render.yaml`** in the repo root. When prompted for **`PUBLIC_BASE_URL`**, use your service’s **`https://<name>.onrender.com`** URL first (you’ll change it after the custom domain works).
4. After the first deploy succeeds, open **`https://<name>.onrender.com`**, then **`/admin`**, and register the **first leader** (auto-approved).
5. **Custom domain (`albertsonwardplan.com`):** in the Web Service → **Settings → Custom Domains**, add **`albertsonwardplan.com`**. Render shows the **DNS records** to add at **Namecheap → Advanced DNS** (often **CNAME** for `www`, and their instructions for apex).
6. When **`https://albertsonwardplan.com`** loads with TLS, update the **`PUBLIC_BASE_URL`** environment variable to **`https://albertsonwardplan.com`** (no trailing slash) and **trigger a manual deploy** or **Clear build cache & deploy** if needed so the running service picks it up.
7. **Optional mail:** add `MAIL_ENABLED`, `SMTP_*`, `MAIL_FROM`, `MAIL_FROM_NAME` in **Environment** (same as `.env.example`).

---

## Build locally

From the repository root:

```bash
cd frontend && npm ci && npm run build
cd ../backend && go build -o wardmission ./cmd/main.go
```

Run `./wardmission` from `backend/` (or set `DATABASE_PATH` to an absolute path). Copy `.env.example` to `.env` and tune variables.

## Docker

```bash
docker build -t wardmission:latest .
docker run --rm -p 8080:8080 \
  -e SESSION_SECRET="$(openssl rand -hex 24)" \
  -e PUBLIC_BASE_URL=https://albertsonwardplan.com \
  -e COOKIE_SECURE=true \
  -v ward-data:/data \
  wardmission:latest
```

Mount a volume on `/data` so `wardmission.db` survives container restarts.

## Environment variables

| Variable | Notes |
|----------|--------|
| `SESSION_SECRET` | Required; random string ≥16 chars (`render.yaml` can generate one) |
| `DATABASE_PATH` | Use **`/data/wardmission.db`** with the Render disk (set in blueprint) |
| `PUBLIC_BASE_URL` | Canonical HTTPS URL **without** trailing slash — **`https://albertsonwardplan.com`** in production |
| `COOKIE_SECURE` | `true` when served over HTTPS (set in blueprint) |
| `PORT` | Render injects `PORT`; the Go server reads it automatically |
| `MAIL_ENABLED` | `true` to enable SMTP |
| `SMTP_*`, `MAIL_FROM`, `MAIL_FROM_NAME` | See `.env.example` |
| `PASSWORD_RESET_HOURS` | Lifetime of password-reset links (default 24) |

## Hosted options

**Render** — primary blueprint (`render.yaml`): Dockerfile Web Service + persistent disk at `/data`.

**Fly.io / Railway / VPS** — same Docker image pattern; persistent disk or volume required anywhere you run SQLite.

## Namecheap + Render DNS

- Use **Advanced DNS** unless nameservers point elsewhere.
- Follow **exactly** what Render shows under **Custom Domains** for your service (targets change per workspace).
- **`PUBLIC_BASE_URL`** must match the URL users actually use (`https://albertsonwardplan.com` once cut over).

## After deploy

1. Visit **`https://albertsonwardplan.com/admin`** (or `*.onrender.com` until DNS is ready); register the first leader account (auto-approved).
2. Turn on outbound mail so pending signups and mission experiences notify leaders.
3. Confirm **`PUBLIC_BASE_URL`** matches the live URL.
