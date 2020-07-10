FROM golang:alpine
RUN mkdir /app
ADD . /app/
WORKDIR /app
RUN go build -o go-back cmd/go-back/main.go
CMD ["./go-back"]