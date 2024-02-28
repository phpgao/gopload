FROM golang:alpine AS builder
WORKDIR /go/src/github.com/phpgao/gopload/
COPY ./* ./
RUN CGO_ENABLED=0 go build

FROM alpine:3
RUN apk --no-cache add ca-certificates
WORKDIR /usr/local/bin/
COPY --from=builder /go/src/github.com/phpgao/gopload/main ./gopload
CMD ["gopload"]