FROM golang as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o k8s-healthcheck .

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/k8s-healthcheck /app/k8s-healthcheck

CMD [ "/app/k8s-healthcheck" ]
