FROM mariadb:10.2
ADD mdbc /mdbc
CMD ["/mdbc", "init"]

