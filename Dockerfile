# Build image
FROM golang:1.12-alpine AS build-env
WORKDIR /root/
COPY . ./
RUN apk update && apk add git
RUN CGO_ENABLED=0 GOOS=linux go build -a

# Runtime image
FROM alpine:3.8 AS runtime-env
WORKDIR /root/
COPY --from=build-env /root/website ./
COPY package.json ./package.json
COPY templates ./templates
COPY assets ./assets
RUN touch ./google-auth.json
RUN apk update && apk add ca-certificates curl bash
CMD ["./website"]
