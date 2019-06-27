FROM alpine:3.7
COPY ./bin/useless-controller /app/useless-controller
CMD /app/useless-controller
