# Albertson Ward — frontend (Vite + React)

- **Dev:** from repo root, start the Go API on `:8080` (`go run ./backend/cmd` or `cd backend && go run ./cmd`), then `npm run dev` here. Vite proxies `/api` and `/auth` to the backend.
- **Production build:** `npm run build` writes static files to `../backend/web/dist` (embedded by the Go binary when you `go build` the server).

## Optional home photos

Place images in `public/` (Vite serves them at the site root):

- `ward-hero.jpg` — wide banner for the home hero (e.g. Christ portrait). If the file is missing, a neutral gradient is used.
- `ward-missionaries.jpg` — photo for the dark split section (e.g. missionaries in the field). If missing, a gradient panel is used.

Use images you have rights to publish (ward photography, Church media per guidelines, or licensed stock).

## If `npm install` fails with `ENOTEMPTY` / rename errors

That usually means a half-written `node_modules` (interrupted install or two installs at once). From this directory:

```bash
rm -rf node_modules package-lock.json .vite
npm install
```

Close other terminals or IDE tasks that might be touching `node_modules` while installing.

## `EBADENGINE` warnings

This repo pins **Vite 5** and drops ESLint from the default install so **Node 18+** and **Node 20.16** work. Upgrading to **Node 20.19+** or **22 LTS** is still recommended for other global tools.
