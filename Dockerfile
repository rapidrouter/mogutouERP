# ====================
# Stage 1: Build the Vue frontend
# ====================
FROM node:18-alpine AS frontend-builder

WORKDIR /app/web

# Install dependencies first (better layer caching)
COPY web/package.json web/package-lock.json* ./
RUN npm install --legacy-peer-deps

# Copy the rest of the frontend source and build
COPY web/ ./
# webpack 4 / uglifyjs need the legacy OpenSSL provider on Node 17+
ENV NODE_OPTIONS=--openssl-legacy-provider
RUN npm run build

# ====================
# Stage 2: Build the Go backend
# ====================
FROM golang:1.22-alpine AS backend-builder

RUN apk add --no-cache git gcc musl-dev

WORKDIR /app

# Copy module definition and source code
COPY go.mod go.sum* ./
COPY *.go ./
COPY api/ ./api/
COPY conf/ ./conf/
COPY middleware/ ./middleware/
COPY models/ ./models/
COPY pkg/ ./pkg/
COPY utils/ ./utils/

# Copy the built frontend dist into repo root (go:embed expects these here)
COPY --from=frontend-builder /app/web/dist/index.html ./index.html
COPY --from=frontend-builder /app/web/dist/log.png ./log.png
COPY --from=frontend-builder /app/web/dist/static ./static

# Resolve module graph (generates go.sum), then build a static binary
RUN go mod tidy && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o mogutou main.go router.go

# ====================
# Stage 3: Minimal runtime image
# ====================
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata && \
    adduser -D -h /app appuser

WORKDIR /app

# Copy the binary
COPY --from=backend-builder /app/mogutou /app/mogutou

# Copy default config (overridable via volume mount)
COPY conf/ /etc/conf/

# Directory for SQLite data persistence
RUN mkdir -p /app/data && chown -R appuser:appuser /app

USER appuser

ENV MOGUTOU_NO_BROWSER=1
ENV MOGUTOU_SQLITE_PATH=/app/data/mgt.db

EXPOSE 1988

ENTRYPOINT ["/app/mogutou"]