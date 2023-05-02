FROM golang:1.20-alpine AS builder

WORKDIR /src

# Restore dependencies
COPY audit/ .

RUN go mod tidy

# Build executable
RUN go build -o /src/audit ./

FROM scratch
WORKDIR /src
COPY --from=builder /src/audit ./
EXPOSE 8084
CMD ["/src/audit"]
