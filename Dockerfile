FROM golang:1.27.2-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /threadline .

FROM alpine:3.23
RUN apk add --no-cache ca-certificates && adduser -D -u 10001 app
WORKDIR /app
COPY --from=build /threadline /app/threadline
COPY public /app/public
USER app
EXPOSE 8080
CMD ["/app/threadline"]
