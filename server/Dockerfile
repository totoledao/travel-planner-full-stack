FROM golang:1.22.5

WORKDIR /server

COPY  go.mod go.sum ./

RUN go mod download && go mod verify

COPY . .

RUN go build -o server/bin/server .

EXPOSE 8080
ENTRYPOINT [ "server/bin/server" ]