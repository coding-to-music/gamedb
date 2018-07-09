FROM scratch
COPY steam-authority /steam-authority
COPY templates /templates
COPY node_modules /node_modules
COPY assets /assets
EXPOSE 8085
CMD ["/steam-authority"]





FROM golang:1.10.3
WORKDIR /go/src/github.com/steam-authority/steam-authority/
RUN go get -d -v golang.org/x/net/html
COPY app.go .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=0 /go/src/github.com/alexellis/href-counter/app .
CMD ["./app"]
