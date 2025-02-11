FROM golang:1.23.5 AS builder

WORKDIR /build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app .

FROM scratch
COPY --from=builder /build/app /app
ENTRYPOINT ["/app"]
