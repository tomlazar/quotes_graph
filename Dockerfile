FROM golang:1.12-alpine3.9

# add our cgo dependencies
RUN apk add --no-cache ca-certificates cmake make g++ openssl-dev git curl pkgconfig
# clone seabolt-1.7.0 source code
RUN git clone https://github.com/neo4j-drivers/seabolt.git /seabolt 
# invoke cmake build and install artifacts - default location is /usr/local
WORKDIR /seabolt/build
# CMAKE_INSTALL_LIBDIR=lib is a hack where we override default lib64 to lib to workaround a defect
# in our generated pkg-config file 
RUN cmake -D CMAKE_BUILD_TYPE=Release -D CMAKE_INSTALL_LIBDIR=lib .. && cmake --build . --target install

WORKDIR $GOPATH/src/github.com/tomlazar/quotes_graph

COPY . .
RUN go install

CMD ["quotes_graph"]