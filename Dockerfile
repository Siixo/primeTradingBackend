# STAGE 1: Build
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Enable vendor mode (no network needed during build)
# We copy the vendor directory created on the host
COPY go.mod go.sum ./
COPY vendor/ ./vendor/

# Copier tout le reste du code
COPY . .

# Build le binaire à partir du dossier cmd/server using the vendored dependencies
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -ldflags="-w -s" -a -installsuffix cgo -o /go/bin/app ./cmd/server

# STAGE 2: Minimal image
FROM scratch

# Copier binaire
COPY --from=builder /app/assets ./assets
COPY --from=builder /go/bin/app /go/bin/app
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

# Non-root user (optionnel)
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
USER nobody:nobody

ENTRYPOINT ["/go/bin/app"]
