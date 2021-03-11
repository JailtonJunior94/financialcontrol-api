FROM golang:1.16-alpine

WORKDIR /go/src/

COPY . .

RUN go clean --modcache
RUN GOOS=linux go build main.go

EXPOSE 3000
ENTRYPOINT ["./main"]