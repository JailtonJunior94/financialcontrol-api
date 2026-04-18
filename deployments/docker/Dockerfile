FROM golang:1.25-alpine AS builder
WORKDIR /app

COPY . .

RUN go clean --modcache
RUN GOOS=linux go build -o main ./cmd/financialcontrol-api

FROM alpine:3.22
WORKDIR /app

RUN apk --no-cache add tzdata

ENV TZ=America/Sao_Paulo
ENV PORT=4000

COPY --from=builder /app/main .
COPY configs/config.Development.yaml ./configs/config.Development.yaml
COPY configs/config.Staging.yaml ./configs/config.Staging.yaml
COPY configs/config.Production.yaml ./configs/config.Production.yaml

EXPOSE 4000
CMD [ "/app/main" ]
