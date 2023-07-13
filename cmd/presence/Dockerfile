FROM alpine:latest

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

RUN update-ca-certificates

WORKDIR ~/app

COPY ./build/ .

EXPOSE 8080

CMD ["./server"]
