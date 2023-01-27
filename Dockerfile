FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app/

CMD ["./vpn-wg"]