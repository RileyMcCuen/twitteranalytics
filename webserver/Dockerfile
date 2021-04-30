FROM alpine:3.7
# https://stackoverflow.com/questions/34729748/installed-go-binary-not-found-in-path-on-alpine-linux-docker
# The below line helps to deal with issues caused by compiling binary with glibc which does note exist in alpine
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2
COPY ./build/twitteranalytics ./twitteranalytics
COPY ./static ./static
# CMD ["./twitteranalytics"]
