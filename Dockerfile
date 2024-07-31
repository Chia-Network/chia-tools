FROM golang:1 as builder

COPY . /app
WORKDIR /app

RUN make build

FROM gcr.io/distroless/static-debian12

COPY --from=builder /app/bin/chia-tools /chia-tools

ENTRYPOINT ["/chia-tools"]
