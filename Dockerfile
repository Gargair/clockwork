# Stage 1: Build Go server (statically linked)
FROM golang:1.25-alpine AS server-builder

WORKDIR /build

# Copy dependency files first for better layer caching
COPY server/go.mod server/go.sum ./
RUN go mod download

# Copy server source code
COPY server/ ./

# Build statically linked binary
# CGO_ENABLED=0: Disable CGO for pure Go binary
# GOOS=linux GOARCH=amd64: Target Linux x86_64
# -ldflags '-w -s': Strip debug info to reduce binary size
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags '-w -s' \
    -o /app/server \
    ./cmd/server

# Build healthcheck binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags '-w -s' \
    -o /app/healthcheck \
    ./cmd/healthcheck

# Stage 2: Build client SPA
FROM node:25-alpine AS client-builder

WORKDIR /build

# Copy dependency files first for better layer caching
COPY client/package.json client/package-lock.json ./
RUN npm ci

# Copy client source code
COPY client/ ./

# Build client SPA
RUN npm run build

# Stage 3: Final runtime image (minimal)
FROM gcr.io/distroless/static:nonroot

# Copy statically linked Go binary from Stage 1
COPY --from=server-builder /app/server /app/server

# Copy healthcheck binary from Stage 1
COPY --from=server-builder /app/healthcheck /app/healthcheck

# Copy built client assets from Stage 2
COPY --from=client-builder /build/dist /app/static

# Copy migration files
COPY server/migrations /app/migrations

# Set working directory
WORKDIR /app

# Set environment variables defined by build stage
ENV STATIC_DIR=/app/static
ENV PORT=8080
ENV ENV=production
ENV MIGRATIONS_DIR=/app/migrations

EXPOSE ${PORT}
ENTRYPOINT ["/app/server"]
