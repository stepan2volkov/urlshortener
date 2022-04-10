FROM golang:latest AS builder

WORKDIR /app
COPY . .
RUN go mod download
RUN make install-tools
RUN make build

FROM scratch
WORKDIR /app
COPY --from=builder /app/build/urlshortener .
COPY --from=builder /app/web /app/web
EXPOSE 8000
CMD ["./urlshortener"]