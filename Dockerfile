FROM golang:1.24.5 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/app ./cmd/service


FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=build /out/app /app/app
COPY web/ /app/web/
EXPOSE 8081
USER nonroot:nonroot
ENTRYPOINT ["/app/app"]
