FROM alpine:3.18.0

RUN apk update && apk add ffmpeg go

WORKDIR /app
COPY . /app
RUN go build -o /main main.go

CMD ["/app/main"]
