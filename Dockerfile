FROM docker.io/golang:1.22.1
LABEL authors="Knick"

RUN useradd -u 1001 -m iamuser

COPY ./app /app

RUN chmod 755 /app

USER 1001

ENTRYPOINT ["/app"]