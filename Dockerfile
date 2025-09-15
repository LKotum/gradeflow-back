FROM golang:1.22 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/api ./cmd/api

FROM gcr.io/distroless/base-debian12
COPY --from=build /out/api /api
EXPOSE 8080
USER 65532:65532
ENTRYPOINT ["/api"]
