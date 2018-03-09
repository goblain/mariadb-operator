FROM alpine:latest
ADD mdbc /mdbc
CMD ["/mdbc", "init"]

