# --- build stage ---
FROM golang:1.22-alpine AS build
WORKDIR /src
RUN apk add --no-cache ca-certificates
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app ./cmd/app

# --- runtime stage ---
FROM gcr.io/distroless/base-debian12
WORKDIR /
COPY --from=build /app /app
EXPOSE 8081
USER nonroot:nonroot
ENTRYPOINT ["/app"]
