FROM golang:1.22-alpine AS build

WORKDIR /src
COPY go.mod ./
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/shardlab-coordinator \
    ./coordinator/cmd/shardlab-coordinator

FROM scratch

COPY --from=build /out/shardlab-coordinator /shardlab-coordinator

USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/shardlab-coordinator"]
