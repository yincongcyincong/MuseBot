#!/bin/bash
rm -rf ./output
rm -rf ./release
mkdir ./output
mkdir ./release
set CGO_ENABLED=0
#linux
GOOS=linux GOARCH=386 go build && mv telegram-deepseek-bot output/telegram-deepseek-bot && tar zcfv "release/telegram-deepseek-bot-linux-386.tar.gz" ./output/telegram-deepseek-bot
rm -rf ./output/*
GOOS=linux GOARCH=amd64 go build && mv telegram-deepseek-bot output/telegram-deepseek-bot && tar zcfv "release/telegram-deepseek-bot-linux-amd64.tar.gz" ./output/telegram-deepseek-bot
rm -rf ./output/*

#darwin
GOOS=darwin GOARCH=amd64 go build && mv telegram-deepseek-bot output/telegram-deepseek-bot && tar zcfv "release/telegram-deepseek-bot-darwin-amd64.tar.gz" ./output/telegram-deepseek-bot
rm -rf ./output/*
GOOS=darwin GOARCH=arm64 go build && mv telegram-deepseek-bot output/telegram-deepseek-bot && tar zcfv "release/telegram-deepseek-bot-darwin-arm64.tar.gz" ./output/telegram-deepseek-bot
#windows
rm -rf telegram-deepseek-bot

GOOS=windows GOARCH=386 go build && mv telegram-deepseek-bot.exe output/telegram-deepseek-bot.exe && tar zcfv "release/telegram-deepseek-bot-windows-386.tar.gz" ./output/telegram-deepseek-bot.exe
rm -rf ./output/*
GOOS=windows GOARCH=amd64 go build && mv telegram-deepseek-bot.exe output/telegram-deepseek-bot.exe && tar zcfv "release/telegram-deepseek-bot-windows-amd64.tar.gz" ./output/telegram-deepseek-bot.exe

rm -rf ./output
rm -rf telegram-deepseek-bot telegram-deepseek-bot.exe
