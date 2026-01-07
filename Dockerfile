# STAGE 1: Build
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Installer git pour go mod
RUN apk add --no-cache git

# Copier go.mod et go.sum
COPY go.mod go.sum ./
RUN go mod tidy

# Copier tout le reste du code
COPY . .

# Build le binaire à partir du dossier cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -a -installsuffix cgo -o /go/bin/app ./cmd/server

# STAGE 2: Minimal image
FROM scratch

# Copier binaire
COPY --from=builder /go/bin/app /go/bin/app
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

# Non-root user (optionnel)
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
USER nobody:nobody

ENTRYPOINT ["/go/bin/app"]
