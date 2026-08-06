FROM node:24-alpine AS frontend-builder
WORKDIR /app/web

RUN corepack enable && corepack prepare pnpm@latest --activate

COPY web/package.json web/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

COPY web/ ./
RUN pnpm build

FROM golang:1.26.5-alpine AS backend-builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY --from=frontend-builder /app/web/dist ./web/dist
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

FROM alpine:latest
RUN apk add --no-cache ca-certificates

COPY --from=backend-builder /app/server /server

EXPOSE 8080

EXPOSE ${PORT}
ENTRYPOINT ["/server"]
