FROM golang:1.26.6-trixie AS builder
ENV ROOT=/build
RUN mkdir ${ROOT}
WORKDIR ${ROOT}

RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=cache,target=/root/.cache/go-build,sharing=locked \
    --mount=type=bind,source=go.sum,target=go.sum \
    --mount=type=bind,source=go.mod,target=go.mod \
    go mod download -x

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=cache,target=/root/.cache/go-build,sharing=locked \
    CGO_ENABLED=0 GOOS=linux go build -o exporter $ROOT
FROM gcr.io/distroless/cc-debian13:nonroot@sha256:08476c244888bfb425bae1ae4dd7c17f093e972aab8426883d97f3c6f280f7c1
WORKDIR /app

COPY --from=builder /build/exporter ./
EXPOSE 8080

CMD ["./exporter"]
