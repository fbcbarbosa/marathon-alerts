FROM alpine:3.3
ADD https://github.com/ashwanthkumar/marathon-alerts/releases/download/v0.3.5/marathon-alerts-linux-amd64 marathon-alerts
RUN apk --no-cache add ca-certificates && \
    chmod 755 marathon-alerts
ENTRYPOINT ["./marathon-alerts"]
