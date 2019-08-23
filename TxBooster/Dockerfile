FROM ubuntu:18.04
MAINTAINER Wonkyu Lee <kurzweil@nuritelecom.com>
WORKDIR /usr/local/booster
COPY main /usr/local/booster
COPY configs /usr/local/booster/configs
EXPOSE 8080
COPY crypto-config /Users/kurzweil/hf2/exchange-bc-production-qa/crypto-config
CMD ./main