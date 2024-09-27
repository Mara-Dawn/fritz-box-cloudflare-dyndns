FROM registry.semaphoreci.com/golang:1.23.1 as builder

ENV APP_HOME /app

WORKDIR "$APP_HOME"
COPY . .

RUN go mod download
RUN go mod verify
RUN go build -o app

FROM registry.semaphoreci.com/golang:1.21.1

ENV APP_HOME /app
RUN mkdir -p "$APP_HOME"
WORKDIR "$APP_HOME"

COPY --from=builder "$APP_HOME"/app $APP_HOME
