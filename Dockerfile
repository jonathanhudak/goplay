# Do your stuff, build a static binary
FROM golang:alpine as builder

ENV GOPATH /go/

# Install dep
RUN apk add --no-cache git mercurial \
  && go get -u github.com/golang/dep/cmd/dep


WORKDIR $GOPATH/src/golog
COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure --vendor-only

RUN apk del git mercurial

COPY . .

RUN CGO_ENABLED=0 go build -o /app main.go

# Copy binary from builder to itch
FROM jarlefosen/itch
COPY --from=builder /app /app


ENTRYPOINT ["/app"]
