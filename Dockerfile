FROM golang:1.10.3 AS build-env
RUN apk --no-cache add ca-certificates
RUN mkdir -p /go/src/github.com/steam-authority/steam-authority/
WORKDIR /go/src/github.com/steam-authority/steam-authority/
COPY . /go/src/github.com/steam-authority/steam-authority/
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=build-env /go/src/github.com/steam-authority/steam-authority/steam-authority .
COPY templates /templates
COPY node_modules /node_modules
COPY assets /assets
EXPOSE 80:8081
CMD ["./steam-authority"]
