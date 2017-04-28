# go-oryx-lib

[![CircleCI](https://circleci.com/gh/ossrs/go-oryx-lib/tree/master.svg?style=svg&circle-token=6c8eac51700e7c8a4b64b714b3ce5775b518fd15)](https://circleci.com/gh/ossrs/go-oryx-lib/tree/master)
[![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/ossrs/go-oryx?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge)

The public library for [go-oryx](https://github.com/ossrs/go-oryx).

## Packages

The core library including:

- [x] [logger](logger/example_test.go): Connection-Oriented logger for server.
- [x] [json](json/example_test.go): Json+ supports c and c++ style comments.
- [x] [options](options/example_test.go): Frequently used service options with config file.
- [x] [http](http/example_test.go): For http response with error, jsonp and std reponse.
- [x] [asprocess](asprocess/example_test.go): The associate-process, for SRS/BMS to work with external process.
- [x] [kxps](kxps/example_test.go): The k-some-ps, for example, kbps, krps.
- [x] [https](https/example_test.go): For https server over [lego/acme](https://github.com/xenolf/lego/tree/master/acme) of [letsencrypt](https://letsencrypt.org/).
- [x] [gmoryx](gmoryx/README.md): A [gomobile](https://github.com/golang/mobile) API for go-oryx-lib.
- [ ] [rtmp](rtmp/example_test.go): The RTMP protocol stack, for oryx.
- [ ] [flv](flv/example_test.go): The FLV muxer, for oryx.

Other audio/video libraries:

- [x] [go-speex](https://github.com/winlinvip/go-speex): A go binding for [speex](https://speex.org/).
- [x] [go-fdkaac](https://github.com/winlinvip/go-fdkaac): A go binding for [fdk-aac](https://github.com/mstorsjo/fdk-aac).
- [x] [go-aresample](https://github.com/winlinvip/go-aresample): Resample the audio PCM.

## Depends

Only depends on golang standard library.

Winlin 2016
