FROM golang:1.23.1

WORKDIR /app

COPY . .
COPY ./app.env /app/app.env

RUN go mod tidy

CMD ["go", "run", "./cmd/main.go"]