# GMS

Golang Markdown Server

FYI: [mattn/mkup: Portable Markdown Previewer]( https://github.com/mattn/mkup )

## NOTE
* asset fileが存在しない場合はbinaryにbindされているファイルから読み込む
* favicon is downloaded from [favicon\.ico \(32×32\)]( https://golang.org/favicon.ico )

## Install
```
go get -u github.com/umaumax/gms
```

## Usage
### Config File Load Priority
1. `.gms.config.tml`
1. `~/.gms.config.tml`

### Run
```
gms
```

----

e.g. config file
```
# config file
title = "GMS"
root_dir = "~/md"
# HTTP service address (e.g., ':8000')
addr = ":5021"
# livereload HTTP service address (e.g., ':35279')
livereload_addr = ":0"
```
