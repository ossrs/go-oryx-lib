# go-oryx-lib

[![CircleCI](https://circleci.com/gh/ossrs/go-oryx-lib/tree/master.svg?style=svg&circle-token=6c8eac51700e7c8a4b64b714b3ce5775b518fd15)](https://circleci.com/gh/ossrs/go-oryx-lib/tree/master)

The public library for [go-oryx](https://github.com/ossrs/go-oryx).

## Packages

The core library including:

- [x] [logger](logger/example_test.go): Connection-Oriented logger for application server.
- [x] [json](json/example_test.go): Json+ supports c and c++ style comments.
- [x] [options](options/example_test.go): Frequently used service options with config file.
- [x] [http](http/example_test.go): For http response with error, jsonp and std reponse.
- [x] [asprocess](asprocess/example_test.go): The associate-process, for SRS/BMS to work with external process.
- [x] [kxps](kxps/example_test.go): The k-some-ps, for example, kbps, krps.
- [x] [https](https/example_test.go): For https server over [lego/acme](https://github.com/xenolf/lego/tree/master/acme) of [letsencrypt](https://letsencrypt.org/).
- [x] [gmoryx](gmoryx/README.md): A [gomobile](https://github.com/golang/mobile) API for go-oryx-lib.
- [x] [flv](flv/example_test.go): The FLV muxer and demuxer, for oryx.
- [x] [errors](errors/example_test.go): Fork from [pkg/errors](https://github.com/pkg/errors).
- [x] [aac](aac/example_test.go): The AAC utilities, for oryx.
- [ ] [avc](avc/example_test.go): The AVC utilities, for oryx.
- [ ] [rtmp](rtmp/example_test.go): The RTMP protocol stack, for oryx.

> Remark: For library, please never use `logger`, use `errors` instead.

Other audio/video libraries:

- [x] [go-speex](https://github.com/winlinvip/go-speex): A go binding for [speex](https://speex.org/).
- [x] [go-fdkaac](https://github.com/winlinvip/go-fdkaac): A go binding for [fdk-aac](https://github.com/mstorsjo/fdk-aac).
- [x] [go-aresample](https://github.com/winlinvip/go-aresample): Resample the audio PCM.

## License

This library just depends on golang standard library,
we do this by copying the code of other libraries,
while all the licenses are liberal:

1. [go-oryx-lib](LICENSE) uses [MIT License](https://github.com/ossrs/go-oryx-lib/blob/master/LICENSE).
1. [pkg/errors](errors/LICENSE) uses [BSD 2-clause "Simplified" License](https://github.com/pkg/errors/blob/master/LICENSE).
1. [acme](https/acme/LICENSE) uses [MIT License](https://github.com/xenolf/lego/blob/master/LICENSE).
1. [jose](https/jose/LICENSE) uses [Apache License 2.0](https://github.com/square/go-jose/blob/v1.1.0/LICENSE).
1. [letsencrypt](https/letsencrypt/LICENSE) uses [BSD 3-clause "New" or "Revised" License](https://github.com/rsc/letsencrypt/blob/master/LICENSE).

Winlin 2016
