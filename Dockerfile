# cursor--Dockerize the recipe-resolver microservice.
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o recipe-resolver-microservice .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/recipe-resolver-microservice .
EXPOSE 3000
ENTRYPOINT ["./recipe-resolver-microservice"] 