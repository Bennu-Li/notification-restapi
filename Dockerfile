FROM golang:1.17 as builder
WORKDIR /app
COPY . .
RUN make build

FROM alpine
WORKDIR /app
RUN mkdir -p /app/database
RUN mkdir -p /app/alert
RUN mkdir -p /app/docs
COPY --from=builder /app/notification /app
COPY --from=builder  /app/alert/* /app/alert
COPY --from=builder  /app/database/* /app/database
COPY --from=builder  /app/docs/* /app/docs
