FROM golang:alpine as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o ./server main.go

FROM alpine

COPY --from=builder /app/server  /app/server

EXPOSE 8080

CMD ["/app/server"]
