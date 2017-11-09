FROM golang:1.9
RUN go get -v -u github.com/emsihyo/go.im/server
EXPOSE 10001 10002
ENV GODEBUG gctrace=1
ENTRYPOINT ["server"]
CMD ['--tcp 10001 --http 10002 namespace chat']

