FROM golang:1.27.1-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o server .

FROM alpine:3.20
WORKDIR /app
COPY --from=build /app/server .
EXPOSE 8082
CMD ["./server"]