# syntax=docker/dockerfile:1

# Build the app from source
FROM golang:1.23-alpine3.21 AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /grpc-cocroach-microservice

# Run the tests in the container
FROM build-stage AS run-test-stage
RUN go test -v ./...

# Deploy the app binary into a lean image
FROM gcr.io/distroless/base-debian11 AS build-release-stage

WORKDIR /

COPY --from=build-stage /grpc-cocroach-microservice /grpc-cocroach-microservice

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/grpc-cocroach-microservice"]