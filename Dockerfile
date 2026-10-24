# Fiery Gambit practice target. Builds the wmd-trigger device into a container
# so you get a real IP on a network to recon — closest thing to the range.
FROM golang:1.24-bookworm AS build
WORKDIR /src
COPY target/trigger.go .
RUN go build -o wmd-trigger trigger.go

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends bash procps netcat-openbsd \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /opt/trigger
COPY target/safety.sh target/detonate ./
COPY --from=build /src/wmd-trigger ./
COPY docker-entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x safety.sh detonate wmd-trigger /usr/local/bin/entrypoint.sh
# LEVEL / MINUTES / PORT / BOMB_ID are set per-container in docker-compose.yml
ENV LEVEL=random MINUTES=20 PORT=9803 BOMB_ID=device
EXPOSE 9803
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
