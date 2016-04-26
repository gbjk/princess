FROM ubuntu:latest

COPY princess /usr/bin/myapp

ENTRYPOINT ["/usr/bin/myapp"]
