FROM busybox

ENV PROTOCOL http
ENV PORT 8080

COPY gogetter /gogetter

EXPOSE 8080

ENTRYPOINT ["/gogetter"]
