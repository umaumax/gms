# GMS

Golang Markdown Server

FYI: [mattn/mkup: Portable Markdown Previewer]( https://github.com/mattn/mkup )

## TODO
* load dynamic markdown show
  * e.g. register command curl 'http://localhost/?file=xxx'

## Usage
```
go get github.com/umaumax/gms
gms
```

## Config File
1. `.gms.config.tml`
1. `~/.gms.config.tml`

e.g.
```
# config file
title = "GMS"
root_dir = "~/md"
# HTTP service address (e.g., ':8000')
addr = ":5021"
# livereload HTTP service address (e.g., ':35279')
livereload_addr = ":0"
```
