FROM scratch
WORKDIR /app
COPY ./bin/sms sms
ENTRYPOINT [ "/app/sms" ]
