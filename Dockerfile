FROM golang:1.25.6 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /plata-go-assignment

FROM alpine:3.22 AS runtime

WORKDIR /app

COPY --from=build /plata-go-assignment /plata-go-assignment

ENV PORT=3000
EXPOSE 3000

ENTRYPOINT ["/plata-go-assignment"]
