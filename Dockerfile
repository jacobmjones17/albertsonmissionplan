# Build front-end assets into backend/web/dist (see frontend/vite.config.ts).
FROM node:22-alpine AS frontend
WORKDIR /src/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

FROM golang:1.25-alpine AS backend
WORKDIR /src/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
COPY --from=frontend /src/backend/web/dist ./web/dist
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /wardmission ./cmd/main.go

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=backend /wardmission ./wardmission
# DATABASE_URL must be set at runtime (e.g. Render Postgres internal URL).
EXPOSE 8080
ENTRYPOINT ["./wardmission"]
