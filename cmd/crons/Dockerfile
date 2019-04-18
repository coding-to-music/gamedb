# Build image
FROM golang:1.12-alpine AS build-env
WORKDIR /root/
COPY ../../ ./
RUN apk update && apk add git
RUN CGO_ENABLED=0 GOOS=linux go build -a

# Runtime image
FROM alpine:3.8 AS runtime-env
WORKDIR /root/
COPY --from=build-env /root/crons ./
COPY health-check.sh ./health-check.sh
RUN ["chmod", "+x", "health-check.sh"]
RUN touch ./google-auth.json
RUN apk update && apk add ca-certificates curl bash
CMD ["./crons"]
HEALTHCHECK --interval=60s --timeout=10s --start-period=30s --retries=2 CMD ./health-check.sh
