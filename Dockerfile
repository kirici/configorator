# Stage 1
FROM golang:1.21.5-alpine3.18 AS BUILD

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

WORKDIR /src

# Cache dependencies
COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy actual source
COPY . .

# Optional vulncheck
# RUN go install -v golang.org/x/vuln/cmd/govulncheck@latest

RUN CGO_ENABLED=$CGO_ENABLED GOOS=$GOOS GOARCH=$GOARCH go build -ldflags='-s -w -extldflags "-static"' -o ./out/app .

# Stage 2
FROM alpine:3.18

RUN apk update \
  && apk -U upgrade \
  && apk add --no-cache ca-certificates bash gcc \
  && update-ca-certificates --fresh \
  && rm -rf /var/cache/apk/*

RUN addgroup runner_group && adduser -S runner -u 1000 -G runner_group

WORKDIR /src

COPY --chown=runner:runner_group --from=BUILD /src/out/app /src/app

RUN chmod +x app

ENTRYPOINT ["/src/app"]