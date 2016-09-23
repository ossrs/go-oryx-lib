# go-oryx-lib

[![CircleCI](https://circleci.com/gh/ossrs/go-oryx-lib/tree/master.svg?style=svg&circle-token=6c8eac51700e7c8a4b64b714b3ce5775b518fd15)](https://circleci.com/gh/ossrs/go-oryx-lib/tree/master)
[![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/ossrs/go-oryx?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge)

The public library for [go-oryx](https://github.com/ossrs/go-oryx).

## core

The core library including:

- [x] [logger](logger/example_test.go): Connection-Oriented logger for server.
- [x] [json](json/example_test.go): Json+ supports c and c++ style comments.
- [x] [options](options/example_test.go): Frequently used service options with config file.
- [x] [http](http/example_test.go): For http response with error, jsonp and std reponse.
- [x] [asprocess](asprocess/example_test.go): The associate-process, for SRS/BMS to work with external process.
- [x] [kxps](kxps/example_test.go): The k-some-ps, for example, kbps, krps.
- [x] [https](https/example_test.go): For https server over [lego/acme](https://github.com/xenolf/lego/tree/master/acme) of [letsencrypt](https://letsencrypt.org/).
- [ ] [rtmp](rtmp/example_test.go): The rtmp protocol stack, for oryx.

Winlin 2016
