# ---------- STAGE 1: Build ----------
FROM golang:1.25-alpine AS builder

WORKDIR /src

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

# Copy and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy full source
COPY . .

# Build binary
RUN go build -trimpath -ldflags="-s -w" -o /app/server ./cmd

# ---------- STAGE 2: Run ----------
# TODO: use distroless image
FROM alpine:latest

WORKDIR /app

# Create nonroot user and group
RUN addgroup -g 65532 -S nonroot && \
    adduser -u 65532 -S nonroot -G nonroot

COPY --from=builder /app/server /app/server

USER nonroot:nonroot
EXPOSE 8080

ENTRYPOINT ["/app/server"]