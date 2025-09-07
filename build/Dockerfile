# ---- Build stage ----
FROM golang:1.25 AS build
WORKDIR /src

# Cache deps
COPY go.mod go.sum* ./
RUN go mod download

# Copy sources
COPY . .

# Build static binary from ./cmd/wod-gen
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /out/wod-gen ./cmd/wod-gen

# ---- Runtime stage ----
FROM gcr.io/distroless/static:nonroot
WORKDIR /app

COPY --from=build /out/wod-gen /app/wod-gen

ENV HTTP_PORT=8080
EXPOSE 8080

USER nonroot:nonroot
ENTRYPOINT ["/app/wod-gen"]
