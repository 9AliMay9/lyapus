FROM golang:1.26.6-bookworm AS build

WORKDIR /src

ENV GOTOOLCHAIN=local

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ ./cmd/
COPY internal/ ./internal/

RUN CGO_ENABLED=0 GOOS=linux go build \
  -mod=readonly \
  -trimpath \
  -buildvcs=false \
  -o /out/apiserver \
  ./cmd/apiserver

FROM scratch AS runtime

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /out/apiserver /apiserver

USER 65532:65532

ENV LYAPUS_HTTP_ADDR=0.0.0.0:8080

EXPOSE 8080

ENTRYPOINT ["/apiserver"]
