ARG BUILDPLATFORM="linux/amd64"
FROM --platform=$BUILDPLATFORM golang:1.18-alpine as build

ARG TARGETPLATFORM
ARG BUILDPLATFORM

RUN apk add --no-cache git

COPY ./ /go/src/github.com/meyskens/m-planner

WORKDIR /go/src/github.com/meyskens/m-planner

RUN export GOARM=6 && \
    export GOARCH=amd64 && \
    if [ "$TARGETPLATFORM" == "linux/arm64" ]; then export GOARCH=arm64; fi && \
    if [ "$TARGETPLATFORM" == "linux/arm" ]; then export GOARCH=arm; fi && \
    go build -ldflags "-X main.revision=$(git rev-parse --short HEAD)" ./cmd/planner/

FROM alpine:3.13

RUN apk add --no-cache ca-certificates tzdata

RUN mkdir -p /go/src/github.com/meyskens/m-planner
WORKDIR /go/src/github.com/meyskens/m-planner

COPY --from=build /go/src/github.com/meyskens/m-planner/planner /usr/local/bin/

ENTRYPOINT [ "/usr/local/bin/planner" ]
CMD [ "bot" ]
