#!/bin/bash
docker build -t telegram-deepseek-bot .
docker tag telegram-deepseek-bot jackyin0822/telegram-deepseek-bot:${1}
docker push  jackyin0822/telegram-deepseek-bot:${1}
docker tag telegram-deepseek-bot jackyin0822/telegram-deepseek-bot:latest
docker push  jackyin0822/telegram-deepseek-bot:latest

