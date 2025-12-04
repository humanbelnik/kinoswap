# FROM golang:1.21-alpine AS builder

# WORKDIR /app
# COPY go.mod go.sum ./
# RUN go mod download

# COPY . .
# RUN CGO_ENABLED=0 GOOS=linux go build -o /mock-s3-server ./cmd/mock-s3-server

# FROM alpine:latest
# RUN apk --no-cache add ca-certificates
# WORKDIR /root/
# COPY --from=builder /mock-s3-server .
# EXPOSE 9090
# CMD ["./mock-s3-server"]

FROM golang:latest

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o myapp .

EXPOSE 9090

CMD ["./myapp"]