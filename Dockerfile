FROM golang:1.18-alpine3.16 as builder
WORKDIR /var/build
COPY . .
RUN ./build

FROM alpine:3.14
WORKDIR /app
COPY --from=builder /var/build/go-back .
ENTRYPOINT ["./go-back"]
