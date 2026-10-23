FROM golang:1.26 AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/in3words-server .

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /out/in3words-server /usr/local/bin/in3words-server
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/in3words-server"]
CMD ["serve"]
