FROM instrumentisto/glide:0.13.0 as builder
WORKDIR /go/src/github.com/KalypsoCloud/jolokia_exporter/

COPY glide.yaml glide.lock ./
RUN glide install --strip-vendor

COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o jolokia_exporter .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /go/src/github.com/KalypsoCloud/jolokia_exporter/jolokia_exporter .
ENTRYPOINT ["./jolokia_exporter"]
