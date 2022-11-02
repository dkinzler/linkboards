FROM golang:latest AS builder

WORKDIR /linkboards
COPY ./ ./

RUN go build -o /go/bin/linkboards ./cmd/api

FROM gcr.io/distroless/base

COPY --from=builder /go/bin/linkboards /

ENTRYPOINT ["/linkboards"]

