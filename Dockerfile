##################################
# 1. Build in a Go-based image   #
###################################
FROM golang:1.12-alpine as builder
RUN apk add --no-cache git # add deps here (like make) if needed
WORKDIR /go/main
COPY . .
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
