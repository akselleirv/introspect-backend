FROM golang:latest

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN  CGO_ENABLED=0 GOOS=linux go build -a -o main .

FROM alpine:latest
EXPOSE 8080
COPY --from=0 /app/main .
COPY ./questions.json .
ENTRYPOINT ["./main"]
