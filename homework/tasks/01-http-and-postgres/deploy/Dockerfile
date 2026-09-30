FROM golang:1.24-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    go build \
    -trimpath \
    -buildvcs=false \
    -ldflags="-s -w" \
    -o /out/trip-service \
    ./cmd/trip-service

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /out/trip-service /trip-service

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/trip-service"]