FROM golang:1.24.0 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o gBalancer .

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/gBalancer .

RUN chmod +x ./gBalancer

CMD ["./gBalancer"]