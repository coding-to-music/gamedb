FROM golang:1.10.3 AS build-env
RUN mkdir -p /go/src/github.com/steam-authority/steam-authority/
WORKDIR /go/src/github.com/steam-authority/steam-authority/
COPY . /go/src/github.com/steam-authority/steam-authority/
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo

FROM alpine:3.8
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
WORKDIR /root/
COPY --from=build-env /go/src/github.com/steam-authority/steam-authority/steam-authority .
COPY templates /templates
COPY node_modules /node_modules
COPY assets /assets
EXPOSE 80:8081
CMD ["./steam-authority"]
