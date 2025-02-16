### youtube-telegram-bot Dockerfile ###
FROM golang:alpine
RUN mkdir /youtube-telegram-bot
ADD . /youtube-telegram-bot
WORKDIR /youtube-telegram-bot
RUN go build -o youtube-telegram-bot .
LABEL Name=youtube-telegram-bot Version=1.0.0
COPY config.json /youtube-telegram-bot/config.json
CMD [ "./youtube-telegram-bot" ]