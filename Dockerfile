FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:alpine AS build

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

COPY go.mod go.sum main.go ./
RUN go mod download

COPY cmd/ cmd/
COPY vmedis/ vmedis/

RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /vmedis-proxy

FROM --platform=${BUILDPLATFORM:-linux/amd64} gcr.io/distroless/base-debian11 AS release

WORKDIR /

COPY --from=build /vmedis-proxy /vmedis-proxy

USER nonroot:nonroot

EXPOSE 8080

CMD ["/vmedis-proxy", "serve"]
