FROM golang:1.9
RUN go get -v -u github.com/emsihyo/go.im/client
EXPOSE 9529
ENV GODEBUG gctrace=1
ENTRYPOINT ["client"]

