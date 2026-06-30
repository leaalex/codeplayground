FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download 2>/dev/null || true

COPY . .
RUN CGO_ENABLED=0 go build -o /backend .

FROM alpine:3.19

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /backend .

ENV PORT=3000
ENV DB_PATH=/data/playground.db

EXPOSE 3000

CMD ["./backend"]
