FROM golang:1.9
RUN go get -v -u github.com/emsihyo/go.im/server
EXPOSE 9000 8000
ENV GODEBUG gctrace=1
ENTRYPOINT ["server"]

