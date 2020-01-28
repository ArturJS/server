FROM debian:stretch-slim

RUN apt-get update && \
    apt-get --no-install-recommends --no-install-suggests --yes --quiet install \
        apt-transport-https bash-completion ca-certificates curl && \
    apt-get clean && apt-get --yes --quiet autoremove --purge

COPY makeless-server /home/makeless-server

RUN curl -L "https://github.com/docker/compose/releases/download/1.25.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
RUN chmod +x /usr/local/bin/docker-compose
RUN chmod +x /home/makeless-server

EXPOSE 8080/tcp
WORKDIR "/home"
CMD TZ=UTC MAX_SIZE=${MAX_SIZE} TOKEN=${TOKEN} CERT_FILE=${CERT_FILE} CERT_KEY=${CERT_KEY} ./makeless-server