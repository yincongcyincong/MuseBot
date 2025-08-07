#!/bin/bash
docker build -t musebot .
docker tag musebot jackyin0822/musebot:${1}
docker push  jackyin0822/musebot:${1}
docker tag musebot jackyin0822/musebot:latest
docker push  jackyin0822/musebot:latest

