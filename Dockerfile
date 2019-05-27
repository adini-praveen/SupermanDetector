FROM golang:alpine as first_stage

RUN apk add --no-cache git
RUN apk add --no-cache sqlite-libs sqlite-dev
RUN apk add --no-cache build-base

RUN echo "GOPATH:  $GOPATH" && echo "GOROOT: $GOROOT" && echo "PATH: $PATH"

RUN mkdir -p $GOPATH/src/SupermanDetector/vendor

ADD main.go $GOPATH/src/SupermanDetector/

COPY vendor/ $GOPATH/src/SupermanDetector/vendor/

RUN ls $GOPATH/src/SupermanDetector/vendor

WORKDIR $GOPATH/src/SupermanDetector/

RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o SupermanDetector $GOPATH/src/SupermanDetector/main.go


FROM golang:alpine

COPY --from=first_stage $GOPATH/src/SupermanDetector/SupermanDetector .

#COPY databases/ ./databases/

RUN ls

ENTRYPOINT ["./SupermanDetector"]