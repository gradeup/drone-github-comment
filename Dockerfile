FROM alpine
ADD gitdrone /bin/
RUN apk -Uuv add ca-certificates
ENTRYPOINT /bin/gitdrone