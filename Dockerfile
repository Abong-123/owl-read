# ============ STAGE 1: Build frontend (Svelte) ============
FROM node:22-alpine AS frontend-builder
WORKDIR /app/frontend

COPY frontend/package*.json ./
RUN npm ci

COPY frontend/ ./
RUN npm run build

# ============ STAGE 2: Build backend (Go) ============
FROM golang:1.27-alpine AS backend-builder
WORKDIR /app

COPY go.mod go.sum ./
COPY vendor/ ./vendor/
COPY backend/ ./backend/

RUN CGO_ENABLED=0 GOOS=linux go build \
    -mod=vendor \
    -ldflags="-s -w" \
    -o /app/server \
    ./backend/cmd/server

# ============ STAGE 3: Runtime ============
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata && \
    adduser -D -u 1000 app

WORKDIR /app

COPY --from=backend-builder /app/server ./server
COPY --from=frontend-builder /app/frontend/build ./frontend/build

# go.mod dibutuhkan findProjectRoot
COPY go.mod ./

RUN mkdir -p /app/database/cover /app/database/book && \
    chown -R app:app /app

USER app

EXPOSE 8080

ENV GIN_MODE=release
ENV PORT=8080

CMD ["./server"]
