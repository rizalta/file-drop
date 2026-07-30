FROM golang:1.26.5-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

FROM alpine:latest
RUN apk add --no-cache ca-certificates

ENV PORT=8080
COPY --from=builder /app/server /server

EXPOSE ${PORT}
ENTRYPOINT ["/server"]
