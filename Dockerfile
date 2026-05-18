# Stage 1: Build Frontend (SvelteKit)
FROM oven/bun:1 AS frontend-builder
WORKDIR /app
# Copy package.json and install dependencies
COPY frontend/package.json frontend/bun.lockb ./
RUN bun install --frozen-lockfile
# Copy frontend source and build
COPY frontend/ ./
RUN bun run build

# Stage 2: Build Backend (Go)
FROM golang:1.22-alpine AS backend-builder
WORKDIR /app/backend

# Install gcc and musl-dev to compile go-sqlite3 (CGO is required by modernc.org/sqlite? No, modernc/sqlite is pure Go! CGO_ENABLED=0 works!)
# Wait, let's just enable CGO_ENABLED=0 to be safe and portable.
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

# Copy go mod files and download deps
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy the rest of the backend source
COPY backend/ ./

# Create the dist directory where Go embed expects it
RUN mkdir -p web/dist

# Copy built frontend from Stage 1 into the embedded directory
COPY --from=frontend-builder /app/build/ ./web/dist/

# Build the binary
RUN go build -ldflags="-w -s" -o /app/bin/server cmd/server/main.go

# Stage 3: Minimal Production Image
FROM alpine:3.19
WORKDIR /app

# Install CA certificates for external API requests (Google/Bing) and tzdata for timezones
RUN apk --no-cache add ca-certificates tzdata

# Create data directory for SQLite persistence
RUN mkdir -p /app/data

# Copy the standalone Go binary
COPY --from=backend-builder /app/bin/server /usr/local/bin/server

# Expose port
EXPOSE 8080

# Run the server
CMD ["server"]
