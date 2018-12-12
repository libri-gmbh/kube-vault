FROM scalify/glide:0.13.0 as builder
WORKDIR /go/src/github.com/libri-gmbh/kube-vault-sidecar/

COPY glide.yaml glide.lock ./
RUN glide install --strip-vendor

COPY . ./
RUN CGO_ENABLED=0 go build -a -ldflags '-s' -installsuffix cgo -o bin/kube-vault-sidecar .


FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /go/src/github.com/libri-gmbh/kube-vault-sidecar/bin/kube-vault-sidecar .
RUN chmod +x sidecar
ENTRYPOINT ["./kube-vault-sidecar"]
