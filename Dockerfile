FROM instrumentisto/glide:0.13.0 as builder
WORKDIR /go/src/github.com/KalypsoCloud/jolokia_exporter/

COPY glide.yaml glide.lock ./
RUN glide install --strip-vendor

COPY . ./
RUN CGO_ENABLED=0 go build -a -ldflags '-s' -installsuffix cgo -o bin/jolokia_exporter .


FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
<<<<<<< Updated upstream
COPY --from=builder /go/src/github.com/KalypsoCloud/jolokia_exporter/jolokia_exporter/bin .
RUN chmod +x jolokia_exporter
=======
<<<<<<< Updated upstream
COPY --from=builder /go/src/github.com/KalypsoCloud/jolokia_exporter/jolokia_exporter .
=======
COPY --from=builder /go/src/github.com/KalypsoCloud/jolokia_exporter/bin/jolokia_exporter .
RUN chmod +x jolokia_exporter
>>>>>>> Stashed changes
>>>>>>> Stashed changes
ENTRYPOINT ["./jolokia_exporter"]
