FROM golang:1.25-alpine

WORKDIR /app

RUN apk add --no-cache git curl

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o e2e-test .

CMD ["./e2e-test"]