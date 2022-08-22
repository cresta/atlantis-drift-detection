FROM golang:1.19.0 as build

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go test -v ./...

RUN CGO_ENABLED=0 GOOS=linux go build -a -tags netgo -ldflags '-w' -o /helm-autoupdate ./cmd/atlantis-drift-detection/main.go

FROM ubuntu:22:10
COPY --from=build /atlantis-drift-detection /atlantis-drift-detection

ENTRYPOINT ["/atlantis-drift-detection"]
