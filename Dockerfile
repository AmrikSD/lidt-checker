FROM golang:1.16-alpine AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./
RUN go build .

CMD ["./japan-tourism"]


FROM alpine:latest AS prod

WORKDIR /app
COPY --from=build /app/japan-tourism ./

CMD ["./japan-tourism"]

