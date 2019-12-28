FROM golang:1.13.5 as builder
COPY . $GOPATH/src/github.com/mooneyow/dyndo/
WORKDIR $GOPATH/src/github.com/mooneyow/dyndo/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod vendor -a -installsuffix cgo -o /go/bin/dyndo

FROM scratch
COPY --from=builder /go/bin/dyndo /opt/dyndo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
WORKDIR /opt
ENTRYPOINT ["/opt/dyndo"]
