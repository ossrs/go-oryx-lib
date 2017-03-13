# GMORYX

The GMORYX(GOMOBILE ORYX) is a API adapter to use [go-oryx-lib](https://github.com/ossrs/go-oryx-lib)
in Android or iOS.

- [x] AndroidExample, The example for Android.

## AndroidExample

Please build out the library `gmoryx.aar`:

```
cd $GOPATH/src/github.com/ossrs/go-oryx-lib/gmoryx &&
mkdir -p AndroidExample/app/libs &&
gomobile bind -target=android -o AndroidExample/app/libs/gmoryx.aar
```

For setup enviroment for gomobile, read [blog post](http://blog.csdn.net/win_lin/article/details/60956485).

Open this project in AndroidStudio, run in Android phone, which will start a web server:

![GMOryx on Android](https://cloud.githubusercontent.com/assets/2777660/23847853/4abcce20-080f-11e7-83e3-3e12cae4dda3.png)

Access the web server:

![Firefox Client](https://cloud.githubusercontent.com/assets/2777660/23847860/52d54010-080f-11e7-8c97-4f8901aa4b35.png)

Winlin, 2017


