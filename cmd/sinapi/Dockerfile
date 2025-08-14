# --- Stage 1: The "Builder" ---
FROM golang:1.24-alpine AS builder

WORKDIR /src

# Copy the module files from the root directory
COPY ../../go.mod ../../go.sum ./
RUN go mod download

# Copy the rest of the source code from the root directory
COPY ../../ .

# Build the specific service
RUN CGO_ENABLED=0 go build -o /src/sin-api ./cmd/sinapi


# --- Stage 2: The "Final Image" ---
FROM alpine:latest
COPY --from=builder /src/sin-api /app/sin-api
CMD ["/app/sin-api"]