package main

type Config struct {
	Title          string `toml:"title"`
	RootDir        string `toml:"root_dir"`
	Addr           string `toml:"addr"`
	LivereloadAddr string `toml:"livereload_addr"`
}
