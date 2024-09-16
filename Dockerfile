FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY src/go.mod src/go.sum ./
RUN go mod download

RUN go install github.com/pressly/goose/v3/cmd/goose@latest

COPY src .

RUN go build -o /app/bin/app ./cmd/main.go

FROM alpine:3.18 AS runner

WORKDIR /app

COPY --from=builder /app/bin/app /app/app

COPY src/internal/migrations /app/migrations

COPY --from=builder /go/bin/goose /usr/local/bin/goose

COPY /src/.env /app/.env

CMD ["sh", "-c", "goose -dir /app/migrations postgres 'user=postgres password=postgres dbname=tender_db host=db port=5432 sslmode=disable' up && ./app"]

