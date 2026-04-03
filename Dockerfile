FROM golang:1.22-bookworm AS builder

ARG VERSION=dev

WORKDIR /workspace
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# inject the VERSION into the main.Version variable
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-X main.Version=${VERSION}" \
    -a -o manager ./cmd/manager

FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/manager .
USER 65532:65532
ENTRYPOINT ["/manager"]