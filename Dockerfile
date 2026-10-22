FROM golang:1.27-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go test ./... && go vet ./...
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags='-s -w' -o /tasklet .

FROM alpine:3.23 AS runner
RUN apk add --no-cache ca-certificates tzdata && adduser -D -u 10001 tasklet
WORKDIR /app
COPY --from=builder /tasklet /app/tasklet
USER tasklet
EXPOSE 8080
CMD ["/app/tasklet"]
