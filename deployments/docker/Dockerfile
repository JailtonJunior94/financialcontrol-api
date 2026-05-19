FROM golang:1.25-alpine AS builder
WORKDIR /app

COPY . .

ARG GIT_SHA=dev
RUN go clean --modcache
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-X main.version=${GIT_SHA}" -o main ./cmd/financialcontrol-api

FROM alpine:3.22
WORKDIR /app

RUN apk --no-cache add tzdata

ENV TZ=America/Sao_Paulo
ENV PORT=4000

COPY --from=builder /app/main .
COPY configs/config.development.yaml ./configs/config.development.yaml
COPY configs/config.staging.yaml ./configs/config.staging.yaml
COPY configs/config.production.yaml ./configs/config.production.yaml

EXPOSE 4000
CMD [ "/app/main" ]
