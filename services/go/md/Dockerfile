FROM golang:1.25.1-alpine AS build

WORKDIR /src
COPY services/go/md/go.mod services/go/md/go.sum ./services/go/md/
WORKDIR /src/services/go/md
RUN go mod download

COPY services/go/md/ ./
RUN CGO_ENABLED=0 go build -o /out/md .

FROM alpine:3.20
RUN adduser -D -u 10001 app

COPY infra/seed/holidays.json /etc/papertrading/holidays.json
RUN chown app:app /etc/papertrading/holidays.json

USER app

COPY --from=build /out/md /usr/local/bin/md
EXPOSE 6011
ENTRYPOINT ["/usr/local/bin/md"]

