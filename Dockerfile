# builder
FROM golang:alpine as builder

RUN mkdir -p /go/src/github.com/dwin/apm-go-playground/

WORKDIR /go/src/github.com/dwin/apm-go-playground/

COPY . .
##RUN rm -rf testhelper
#RUN find . -name "test_*" | grep -v vendor | xargs rm
#RUN go test ./...

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -v -a -installsuffix cgo -o http-server \
    /go/src/github.com/dwin/apm-go-playground/app/main.go

# actual container
FROM alpine:3.9
RUN apk --update add ca-certificates bash curl

RUN mkdir -p /app

WORKDIR /app

COPY --from=builder /go/src/github.com/dwin/apm-go-playground/http-server .

EXPOSE 9000/tcp

HEALTHCHECK --interval=30s --timeout=1s --start-period=5s \
    CMD curl -f http://localhost:9000/status || exit 1

CMD ["./http-server"]