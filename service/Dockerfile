FROM golang:1.24.1

WORKDIR /app

COPY ./go.* ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/service/main ./cmd/service/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/cli/main ./cmd/cli/main.go