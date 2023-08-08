FROM golang:alpine AS build

WORKDIR /app

COPY go.mod go.sum main.go ./
RUN go mod download

COPY cmd/ cmd/
COPY vmedis/ vmedis/

RUN GOOS=linux GOARCH=arm GOARM=5 go build -o /vmedis-proxy

FROM gcr.io/distroless/base-debian11 AS release

WORKDIR /

COPY --from=build /vmedis-proxy /vmedis-proxy

USER nonroot:nonroot

ENTRYPOINT ["/vmedis-proxy"]
