FROM alpine:3.18.0

RUN apk update && apk add ffmpeg go

WORKDIR /app
COPY . .
RUN go build -o /app/main main.go

CMD ["/app/main"]
