FROM golang:1.13

COPY --from=jotadrilo/watchman /usr/local/bin/watchman* /usr/local/bin/
COPY --from=jotadrilo/watchman /usr/local/var/run/watchman /usr/local/var/run/watchman

CMD go get -v && go test -v -race ./...
