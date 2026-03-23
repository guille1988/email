FROM golang:1.25-alpine

RUN apk add --no-cache git make build-base
RUN go install github.com/air-verse/air@latest

WORKDIR /email

COPY go.mod go.sum ./
RUN go mod download

CMD ["air", "-c", ".air.toml"]