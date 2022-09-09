FROM golang:1.16-alpine AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./
RUN go build .

CMD ["./lidt-checker"]


FROM alpine:latest AS prod

WORKDIR /app
COPY --from=build /app/lidt-checker ./

CMD ["./lidt-checker"]

