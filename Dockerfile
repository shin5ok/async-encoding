FROM alpine:3.18.0

RUN apk update && \
        apk add ffmpeg go

COPY . .
RUN go build -o /main main.go

CMD ["/main"]
