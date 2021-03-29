FROM  golang:alpine3.13
WORKDIR /v2ui
RUN apk add build-base
COPY . .
RUN go build;\
    chmod +x v2ui
CMD ./v2ui
