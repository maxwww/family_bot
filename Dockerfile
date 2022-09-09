FROM golang:latest as builder
WORKDIR /app
COPY ./ ./
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app cmd/bot/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
RUN apk add --no-cache tzdata
WORKDIR /root
COPY --from=builder /app/app .
CMD ["./app"]
