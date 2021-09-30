# Produces a small container with a statically-linked arm64 binary
FROM golang:alpine as builder

# Updated CA-certs and time zone data are required for HTTPS requests and
# git is needed to install the package dependencies. Alpine 3.0 removed these.
RUN apk update && apk add --no-cache git ca-certificates tzdata && update-ca-certificates
WORKDIR $GOPATH/src/sean-public/gofind
COPY . .

# Create a non-root user we can copy to the final image
ENV USER=appuser
ENV UID=10001
RUN adduser --disabled-password --gecos "" --home "/nonexistent" \
    --shell "/sbin/nologin" --no-create-home --uid "${UID}" "${USER}"

# Fetch dependencies but don't build them yet
RUN go get -d -v

# Build everything at once, static and targeted for the final environment
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -a -o /go/bin/gofind \
    -ldflags='-w -s -extldflags "-static"' .

# Start over with a blank image and move only what's necessary
FROM scratch
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
COPY --from=builder /go/bin/gofind /go/bin/gofind

USER appuser:appuser
EXPOSE 8080
ENTRYPOINT ["/go/bin/gofind"]
