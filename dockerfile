# Build
FROM golang:1.24.0-alpine AS build-env
RUN apk add build-base
WORKDIR /app
COPY . /app
RUN go mod download
RUN go build .
RUN go install github.com/projectdiscovery/alterx/cmd/alterx@v0.0.4
RUN go install github.com/projectdiscovery/dnsx/cmd/dnsx@v1.2.2
RUN go install github.com/projectdiscovery/subfinder/v2/cmd/subfinder@v2.7.0

# Release
FROM alpine:3.20.3
COPY --from=build-env /app/asmm8 /usr/local/bin/
COPY --from=build-env /go/bin/alterx /usr/local/bin/
COPY --from=build-env /go/bin/dnsx /usr/local/bin/
COPY --from=build-env /go/bin/subfinder /usr/local/bin/
CMD ["asmm8","launch"]