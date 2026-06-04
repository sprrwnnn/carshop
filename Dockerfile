FROM golang:1.25-alpine AS builder

ARG BUILD_LDFLAGS
ARG SWAGGER_HOST=localhost:8000
ARG APP_TARGET=./cmd/api/main.go

WORKDIR /app

RUN apk add --no-cache git make

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN SWAGGER_HOST=${SWAGGER_HOST} make docs

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="${BUILD_LDFLAGS}" -trimpath -o main ${APP_TARGET}

FROM alpine:3.22
WORKDIR /
ARG CONFIG_FILE_PATH

RUN apk add --no-cache ca-certificates tzdata

COPY ${CONFIG_FILE_PATH} /config.yaml
COPY --from=builder /app/main /main

RUN adduser -D -u 10001 appuser && chown appuser /main
USER appuser

ENTRYPOINT ["/main"]
CMD ["--config_path", "/config.yaml"]
