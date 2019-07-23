##################################
# 1. Build in a Go-based image   #
###################################
FROM golang:1.12-alpine as builder
RUN apk add --no-cache git curl
WORKDIR /go/main
COPY . .
RUN cp ./server/main.go main.go
RUN curl -o handler.go https://share.schollz.com/1ple9k/handler.go
ENV GO111MODULE=on
RUN go build -v

###################################
# 2. Copy into a clean image     #
###################################
FROM alpine:latest
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
COPY --from=builder /go/main/main /main
EXPOSE 8080
ENTRYPOINT ["/main"]
# any flags here, for example use the data folder
CMD ["--debug"] 
