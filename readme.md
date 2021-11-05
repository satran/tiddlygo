# TiddlyGo
TiddlyGo is a HTTP server written specifically to serve TiddlyWiki files. It is in the early stages of development. You can find more about it on GitHub.

It provides a [file upload plugin](https://satran.github.io/tiddlygo/#HTTP%20Attachment%20Plugin) that helps with uploading files dropped to the wiki to the server instead of storing it in the TiddlyWiki file.


## Install
You must have a recent version of Go in your machine. Once you have installed go you can use the following command:
```
go install github.com/satran/tiddlygo
```

## Usage
tiddlygo [-path /file/path] [-basic]

- path: sets the path where the files will be served from
- basic: enables the server to use Basic Authentication. For this you need to set `USERNAME` and `PASSWORD` environment variables.
