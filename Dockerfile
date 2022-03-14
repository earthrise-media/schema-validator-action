FROM golang AS build-env
ARG DATE
ARG VERSION
ARG REVISION
WORKDIR /go/src/app
ADD . /go/src/app

RUN go get -d -v ./...

RUN go build -o /go/bin/schema-validator

FROM gcr.io/distroless/base

ARG DATE
ARG VERSION
ARG REVISION

COPY --from=build-env /go/bin/schema-validator /
CMD ["/schema-validator"]

LABEL org.opencontainers.image.created=$DATE
LABEL org.opencontainers.image.url="https://github.com/earthrise-media/trace-schemas"
LABEL org.opencontainers.image.source="https://github.com/earthrise-media/trace-schemas"
LABEL org.opencontainers.image.version=$VERSION
LABEL org.opencontainers.image.revision=$REVISION
LABEL org.opencontainers.image.vendor="Earthrise Media"
LABEL org.opencontainers.image.title="schema-validator"
LABEL org.opencontainers.image.description="This is a json schema validator"
LABEL org.opencontainers.image.authors="tingold"
