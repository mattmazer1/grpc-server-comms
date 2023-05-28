FROM golang:1.18-alpine as BUILD

WORKDIR /chatserver

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o exec ./server/main.go

FROM alpine:latest

WORKDIR /newserver

COPY --from=BUILD /chatserver/exec /newserver/

EXPOSE 9000 

CMD ./exec