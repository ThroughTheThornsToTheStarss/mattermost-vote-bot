FROM golang:1.21

WORKDIR /app
COPY . .

RUN apt-get update && apt-get install -y libssl-dev pkg-config
RUN go mod tidy
RUN go build -o vote-bot ./cmd/bot

CMD ["sh", "-c", "sleep 15 && ./vote-bot"]
