FROM alpine

RUN apk update && \
    apk add --no-cache tzdata curl

WORKDIR /app

ARG TARGETOS
ARG TARGETARCH

COPY ./bld/webmon-$TARGETOS-$TARGETARCH /app/webmon

EXPOSE 8080
ENTRYPOINT ["/app/webmon"]
CMD []
