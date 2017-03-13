# GMA

The GMA(GOMOBILE API) is a API, adapt to gomobile for Android or iOS.

- [ ] AndroidExample, The example for Android.

## AndroidExample

Please build out the library `gma.aar`:

```
cd $GOPATH/src/github.com/ossrs/go-oryx-lib/gma &&
mkdir -p AndroidExample/app/libs &&
gomobile bind -target=android -o AndroidExample/app/libs/gma.aar
```

For setup enviroment for gomobile, read [blog post](http://blog.csdn.net/win_lin/article/details/60956485).

Winlin, 2017


