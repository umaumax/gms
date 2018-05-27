package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/k0kubun/pp"
	"github.com/mitchellh/go-homedir"
)

var (
	configFilepath string
	config         Config

	_rootDir        string
	_addr           string
	_livereloadAddr string
)

var defaultConfigFilepaths = []string{
	".gms.config.tml",
	"~/.gms.config.tml",
}

func init() {
	log.SetFlags(log.Llongfile)

	flag.StringVar(&configFilepath, "c", "", "config toml file path (priority:"+strings.Join(defaultConfigFilepaths, ",")+")")

	flag.StringVar(&_rootDir, "root", "", "serve file's root dir")
	flag.StringVar(&_addr, "http", "", "HTTP service address (e.g., ':8765')")
	flag.StringVar(&_livereloadAddr, "lrhttp", "", "livereload HTTP service address (e.g., ':0 or :35279 ...')")
}

func initConfig() {
	configFilepaths := append([]string{configFilepath}, defaultConfigFilepaths...)
	//	parse "~" as homedir
	var err error
	var configFilepath string
	for i, v := range configFilepaths {
		if v == "" {
			continue
		}
		configFilepath, err = homedir.Expand(v)
		if err == nil {
			log.Println("load config file at", v)
			break
		}
		log.Println("home dir", err, "priority:", i+1)
	}
	if err != nil {
		log.Fatalln("please create " + defaultConfigFilepaths[0])
	}

	_, err = toml.DecodeFile(configFilepath, &config)
	if err != nil {
		configFilepath, err := homedir.Expand(filepath.Join("~", configFilepath))
		if err != nil {
			log.Fatalln("home dir", err)
		}
		_, err = toml.DecodeFile(configFilepath, &config)
		if err != nil {
			log.Fatalln("toml decode", err)
		}
	}

	//	NOTE flag optionがconfig fileの設定を上書き
	if _rootDir != "" {
		config.RootDir = _rootDir
	}
	config.RootDir, err = homedir.Expand(config.RootDir)
	if err != nil {
		log.Fatalln("home dir", err)
	}
	if config.RootDir == "" {
		//	Current Working Directory
		cwd, err := os.Getwd()
		if err != nil {
			log.Fatalln("os.Getwd", err)
		}
		config.RootDir = cwd
	}

	if _addr != "" {
		config.Addr = _addr
	}
	if !strings.HasPrefix(config.Addr, ":") {
		config.Addr = ":" + config.Addr
	}
	if _livereloadAddr != "" {
		config.LivereloadAddr = _livereloadAddr
	}
	if config.LivereloadAddr == "" {
		//	ポート番号自動割当
		config.LivereloadAddr = ":0"
	}
	pp.Println(config)
	log.Println("start", config.Title)
}
