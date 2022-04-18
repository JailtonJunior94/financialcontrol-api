FROM golang:1.18-alpine3.14 AS builder
WORKDIR /app

COPY . .

RUN go clean --modcache
RUN GOOS=linux go build -o main main.go

FROM alpine:3.14
WORKDIR /app

RUN apk --no-cache add tzdata

ENV TZ=America/Sao_Paulo
ENV PORT=4000

COPY --from=builder /app/main .
COPY config.Development.yaml .
COPY config.Staging.yaml .
COPY config.Production.yaml .

EXPOSE 4000
CMD [ "/app/main" ]