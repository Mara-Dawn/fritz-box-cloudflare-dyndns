FROM golang:1.23.1

WORKDIR /app

ENV WEB_PORT=${WEB_PORT}

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./

RUN  go build -o ./app

CMD [ "./app" ]
