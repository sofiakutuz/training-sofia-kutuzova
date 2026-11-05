# Funguje mi to. Petr
FROM golang:1.27-alpine AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -o /out/objednavky .

FROM alpine:3.24
RUN adduser -D -u 10001 objednavky
RUN  mkdir /data
RUN chown 10001 /data
COPY --from=build /out/objednavky /usr/local/bin/objednavky
COPY start.sh /usr/local/bin/start.sh
USER 10001
EXPOSE 8080
HEALTHCHECK --interval=10s --timeout=3s --start-period=5s \
  CMD wget -qO- http://127.0.0.1:8080/healthz || exit 1
ENTRYPOINT ["/usr/local/bin/start.sh"]