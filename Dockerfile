FROM alpine

COPY ./dist/fibonacci_linux_amd64 .
RUN chmod 755 /fibonacci_linux_amd64
RUN ln -s /fibonacci_linux_amd64 /bin/fibonacci

CMD ["fibonacci"]

