# Build image
FROM golang:1.10-alpine AS build-env
WORKDIR /go/src/github.com/gamedb/website/
COPY . /go/src/github.com/gamedb/website/
RUN apk update && apk add curl git openssh
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
RUN dep ensure
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo

# Runtime image
FROM alpine:3.8 AS runtime-env
WORKDIR /root/
COPY package.json ./package.json
COPY --from=build-env /go/src/github.com/gamedb/website/website ./
COPY templates ./templates
COPY assets ./assets
COPY site.webmanifest ./site.webmanifest
COPY robots.txt ./robots.txt
COPY browserconfig.xml ./browserconfig.xml
RUN touch ./google-auth.json
RUN apk update && apk add ca-certificates curl bash
CMD ["./website"]
