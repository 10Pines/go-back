FROM golang:alpine
RUN mkdir /app
ADD . /app/
WORKDIR /app
RUN ./build
ENTRYPOINT ["./go-back"]