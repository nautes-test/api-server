FROM ubuntu:22.04

RUN  apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates netbase curl git vim openssh-client \
    && rm -rf /var/lib/apt/lists/ \
    && apt-get autoremove -y && apt-get autoclean -y 

COPY  ./bin /app
COPY  ./configs /data/conf

WORKDIR /app

EXPOSE 8000
EXPOSE 9000
VOLUME /data/conf

RUN git config --global credential.helper 'cache --timeout=10800'
CMD ["./manager", "-conf", "/data/conf"]