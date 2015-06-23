FROM scratch

COPY reaper /reaper

ENTRYPOINT ["/reaper"]
CMD ["-h"]