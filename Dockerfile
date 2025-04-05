FROM golang:latest

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN apt-get update && apt-get install -y gcc libc6-dev

ENV CGO_ENABLED=1

RUN go build -o ./app ./cmd/app
RUN go build -o ./migrator ./cmd/migrator
RUN mkdir -p ./config ./storage

COPY ./config/local.yaml ./config/local.yaml

EXPOSE 8080

CMD ["/bin/sh", "-c", "./migrator --storage=./storage/url_profile.db --migration-path=./migrations && ./app --config=./config/local.yaml"]
