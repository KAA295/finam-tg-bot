FROM golang:1.23.0-alpine AS build

WORKDIR /build
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o cmd/app/bin/main cmd/app/main.go

FROM alpine

WORKDIR app

COPY --from=build /build/cmd/app/bin/main ./main
COPY --from=build /build/config/docker.yml ./config/docker.yml

CMD ["./main", "--config=./config/docker.yml"]
