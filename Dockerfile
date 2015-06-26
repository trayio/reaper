FROM alpine:3.2

RUN apk --update add ca-certificates

COPY reaper /reaper

ENTRYPOINT ["/reaper"]
CMD ["-h"]