FROM golang:1.17 as builder
WORKDIR /app
COPY . .
RUN make build

FROM alpine
WORKDIR /app
COPY --from=builder /app/notification /app
COPY --from=builder  /app/alert /app
COPY --from=builder  /app/docs /app
#RUN chmod +x /notification
#RUN ./notification
#CMD ["/notification"]
