 FROM golang:1.15 as builder

WORKDIR /workspace

COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download

COPY cmd/ cmd/
COPY controllers/ controllers/

RUN CGO_ENABLED=0 GOOS=linux GO111MODULE=on go build -a -o scw-clerk ./cmd/

FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/scw-clerk .
USER nonroot:nonroot

ENTRYPOINT ["/scw-clerk"]