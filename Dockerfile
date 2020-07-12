FROM golang:alpine
WORKDIR /app
COPY . /app/
RUN ./build
ENTRYPOINT ["./go-back"]