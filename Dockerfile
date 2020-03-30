FROM golang:1-alpine as builder
WORKDIR /go/src/github.com/libri-gmbh/kube-vault/

COPY . ./
RUN export CGO_ENABLED=0 && \
    go get && \
    go build -a -ldflags '-s' -installsuffix cgo -o bin/kube-vault .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /go/src/github.com/libri-gmbh/kube-vault/bin/kube-vault .
RUN chmod +x ./kube-vault
ENTRYPOINT ["./kube-vault"]
