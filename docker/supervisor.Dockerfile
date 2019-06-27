FROM alpine:3.7
COPY ./bin/function /app/function
ARG listen_addr
ENV LISTEN_ADDR=$listen_addr
CMD /app/function -laddr=${LISTEN_ADDR}
