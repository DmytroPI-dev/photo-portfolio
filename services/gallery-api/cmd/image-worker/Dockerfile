# Build an ARM64 Go binary. The final image is built with --platform linux/arm64
# by the release command documented in the worker README.
FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS build

ARG TARGETOS
ARG TARGETARCH

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -ldflags="-s -w" -o /out/bootstrap ./cmd/image-worker

# Amazon Linux 2023 does not currently publish libvips, while Alpine does.
# Lambda container images can use a custom base when the executable implements
# the Runtime API, which aws-lambda-go's lambda.Start does.
FROM alpine:3.21

# Alpine packages the libvips shared library and its `vips`/`vipsheader` CLIs
# separately. The worker needs both the encoder and the command-line tools.
RUN apk add --no-cache ca-certificates vips vips-tools

COPY --from=build /out/bootstrap /bootstrap

ENTRYPOINT ["/bootstrap"]
