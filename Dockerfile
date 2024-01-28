FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:latest AS build

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

RUN go env -w GOMODCACHE=/root/.cache/go-build

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/root/.cache/go-build go mod download

COPY . ./

RUN --mount=type=cache,target=/root/.cache/go-build GOOS=${TARGETOS} GOARCH=${TARGETARCH} CGO_ENABLED=0 go build -o /vmedis-proxy

FROM --platform=${BUILDPLATFORM:-linux/amd64} gcr.io/distroless/static-debian11 AS release

WORKDIR /

COPY --from=build /vmedis-proxy /vmedis-proxy

USER nonroot:nonroot

EXPOSE 8080

CMD ["/vmedis-proxy", "serve"]
