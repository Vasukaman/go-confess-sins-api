# --- Stage 1: The "Builder" ---
FROM golang:1.24-alpine AS builder
# ... (identical to the builder stage above)
WORKDIR /src
COPY ../../go.mod ../../go.sum ./
RUN go mod download
COPY ../../ .
RUN CGO_ENABLED=0 go build -o /src/web-server ./cmd/web

# --- Stage 2: The "Final Image" ---
FROM alpine:latest
WORKDIR /app
# Copy the static web assets from the root
COPY --from=builder /src/web .
# Copy the compiled executable
COPY --from=builder /src/web-server .
CMD ["/app/web-server"]