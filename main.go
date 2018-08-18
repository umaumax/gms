package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	// "gopkg.in/fsnotify.v1"
	"github.com/fsnotify/fsnotify"

	"github.com/omeid/livereload"
	"github.com/russross/blackfriday"
)

func OpenIncludingAsset(name string) (data []byte, err error) {
	data, err = ioutil.ReadFile(name)
	if err == nil {
		log.Println("load local file:", name)
	} else {
		data, err = Asset(name)
	}
	return
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()
	initConfig()

	//	NOTE 実はDirでなくともよい
	fi, err := os.Stat(config.RootDir)
	if err != nil {
		log.Fatalln(err)
	}
	dir := config.RootDir
	walkRoot := "."
	if fi.IsDir() {
		//	NOTE この操作ok?!
		walkRoot = dir
	} else {
		dir = filepath.Dir(config.RootDir)
		walkRoot = filepath.Base(config.RootDir)
	}
	//	カレントディレクトリを変更
	// 	os.Chdir(dir)
	// 	cwd := "."
	cwd := dir

	//	livereload.js server
	lrs := livereload.New("mkup")
	defer lrs.Close()
	//	NOTE config.LivereloadAddrが":0"の場合に割り当てた結果を代入するためのチャネル
	livereloadAddrChan := make(chan string)
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/livereload.js", func(w http.ResponseWriter, r *http.Request) {
			b, err := OpenIncludingAsset("_assets/livereload.js")
			if err != nil {
				PageNotFound(w, r)
				//				http.Error(w, "404 page not found", 404)
				return
			}
			w.Header().Set("Content-Type", "application/javascript")
			w.Write(b)
			return
		})
		mux.Handle("/", lrs)
		l, err := net.Listen("tcp", config.LivereloadAddr)
		if err != nil {
			log.Fatalln(err)
		}
		//	NOTE e.g) [::]:56597 -> :56597
		splits := strings.Split(l.Addr().String(), ":")
		addr := ":" + splits[len(splits)-1]
		livereloadAddrChan <- addr
		//		s := &http.Server{Handler: mux}
		//		log.Fatalln(s.Serve(l))
		log.Fatalln(http.Serve(l, mux))
		//		log.Fatal(http.ListenAndServe(addr, mux))
	}()
	livereloadAddr := <-livereloadAddrChan
	config.LivereloadAddr = livereloadAddr

	//	file watcher
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalln("fsnotify.NewWatcher", err)
	}

	log.Println("watch dir root : ", walkRoot)

	symbolicFilePathMap := map[string]string{}
	//	TODO 基本的にマークダウンファイルのみ監視するように
	//	NOTE 同一ディレクトリをfsw.Addしてもno error
	checkWatch := func(path string) (err error) {
		fi, err := os.Stat(path)
		if err != nil {
			return
		}

		//	. or _ 始まりのディレクトリは無視する
		if fi.Name() != "." &&
			(strings.HasPrefix(fi.Name(), ".") || strings.HasPrefix(fi.Name(), "_")) {
			if !skipLogFlag {
				log.Println("skip dir (prefix . or _) : ", fi.Name())
			}
			if fi.IsDir() {
				return filepath.SkipDir
			}
			return
		}

		if !fi.IsDir() {
			// NOTE: シンボリックリンクファイルの場合は指し示す先のファイルもしくはディレクトリを監視しなければ，変更が検知できない
			// NOTE: vimの場合はtemporaryファイルを作成し，renameしているので，ファイルのみの監視ではなく，ディレクトリの監視が望ましい
			if ret, err := IsSymlink(path); err == nil && ret {
				realPath, _ := os.Readlink(path)
				log.Println("watch symbolic link file", realPath, "from", path)
				symbolicFilePathMap[realPath] = path
				fsw.Add(filepath.Dir(realPath))
				return nil
			}
			return
		}
		log.Println("watch dir", path)
		fsw.Add(path)
		return
	}

	accessMap := NewAccessMap()
	accessMap.AutoDelete(func(s string) {
		err := fsw.Remove(s)
		if err != nil {
			log.Println("dir watch remove error:", err)
		}
	})

	go func() {
		fullPath := func(path string) (ret string) {
			ret = path
			if !filepath.IsAbs(path) {
				ret = filepath.Join(cwd, path)
			}
			return
		}

		//	監視ディレクトリを追加
		if false {
			err := filepath.Walk(walkRoot, func(path string, info os.FileInfo, err error) error {
				if info == nil {
					return err
				}
				return checkWatch(fullPath(path))
			})
			fmt.Println("filepath walk error:", err)
		}

		for {
			select {
			case event := <-fsw.Events:
				log.Println("event", event)
				eventPath := filepath.Clean(event.Name)
				//	NOTE backup fileは無視する仕様
				if strings.HasSuffix(eventPath, "~") {
					if !skipLogFlag {
						log.Println("skip backup file (suffix ~) : ", eventPath)
					}
					break
				}

				//	event.Name example
				//	"/???/???/hoge/piyo/main.go": REMOVE
				//	"/???/???/hoge/piyo/main.go": RENAME
				//	"/???/???/hoge/piyo/snippet": CHMOD
				//	"/???/???/hoge/piyo/snippet": CREATE
				//	"/???/???/hoge/piyo/snippet": WRITE
				//	"/???/???/hoge/piyo/snippet": REMOVE|RENAME
				op := event.Op
				is := func(flag fsnotify.Op) bool {
					return op&flag == flag
				}
				// removeFlag := op & fsnotify.Remove
				//				renameFlag := op & fsnotify.Rename
				//				chmodFlag := op & fsnotify.Chmod
				//				createFlag := op & fsnotify.Create
				//				writeFlag := op & fsnotify.Write
				// if removeFlag == 0 {
				//	NOTE CREATE or WRITEを監視すれば十分だと思われる
				if is(fsnotify.Create) || is(fsnotify.Write) {
					if filepath.IsAbs(eventPath) {
						if symbolicPath, ok := symbolicFilePathMap[eventPath]; ok {
							//	NOTE ファイルの場合には監視対象から外れてしまうので再度追加
							fsw.Add(eventPath)
							log.Printf("%s's event was found by watching symbolic file at %s\n", eventPath, symbolicPath)
							eventPath = symbolicPath
						}
					}
					if !filepath.IsAbs(eventPath) {
						eventPath, err = filepath.Rel(cwd, eventPath)
						if err != nil {
							log.Println(err)
							break
						}
					}
					log.Println(eventPath)
					if err == nil {
						relPath := "/" + filepath.ToSlash(eventPath)

						//	NOTE 監視情報更新(new dir or new symbolic file)
						if false {
							err := checkWatch(eventPath)
							if err != nil {
								log.Println(err)
								break
							}
						}
						fi, err := os.Stat(eventPath)
						if err != nil {
							log.Println(err)
							break
						}
						if !fi.IsDir() && filepath.Ext(fi.Name()) == ".md" {
							log.Println("reload", relPath)
							lrs.Reload(relPath, true)
						}
					}
				}
				//	NOTE Remove
				if is(fsnotify.Remove) {
					if fullPath, err := filepath.Rel(cwd, event.Name); err == nil {
						relPath := "/" + filepath.ToSlash(fullPath)
						err := fsw.Remove(fullPath)
						if err != nil {
							log.Println(err)
							break
						}
						log.Println("delete watcher", relPath)
					}
				}
				//	panic(fmt.Errorf("fsnotify.Event %s not found", event.Op))
			case err := <-fsw.Errors:
				if err != nil {
					log.Println(err)
				}
			}
		}
	}()

	//	file server
	fs := http.FileServer(http.Dir(cwd))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Path
		name = strings.TrimPrefix(name, "/")
		// NOTE:指定したディレクトリのファイルツリーの取得
		if strings.HasPrefix(name, "_api/") {
			name = strings.TrimPrefix(name, "_api/")
			name = filepath.Clean(name)
			fpath := filepath.Join(cwd, name)
			log.Println("api", name)

			var data interface{} = []string{}

			fi, err := os.Stat(fpath)
			if err != nil {
				log.Printf("os.Stat:[%s]:%s\n", name, err)
			} else if fi.IsDir() {
				var names []string
				symbolicLinks := make(map[string]string)
				var fWalkFunc filepath.WalkFunc
				var appendFile = func(path string) {
					// NOTE: only md files
					if !strings.HasSuffix(path, ".md") {
						if !skipLogFlag {
							log.Println("skip file", path, "not end with '.md'")
						}
						return
					}
					names = append(names, path)
				}

				fWalkFunc = func(path string, fi os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if fi.Name() != "." && strings.HasPrefix(fi.Name(), ".") {
						if !skipLogFlag {
							log.Println("skip", path, "starts with '.'")
						}
						if fi.IsDir() {
							return filepath.SkipDir
						}
						return nil
					}
					if strings.HasPrefix(fi.Name(), "_") {
						if !skipLogFlag {
							log.Println("skip", path, "starts with '_'")
						}
						if fi.IsDir() {
							return filepath.SkipDir
						}
						return nil
					}
					if !fi.IsDir() {
						if fi.Mode()&os.ModeSymlink == os.ModeSymlink {
							realPath, err := os.Readlink(path)
							if err != nil {
								log.Println("read symbolic link err:", err)
								return nil
							}
							log.Println("traverse symbolic link:", realPath)
							if _, ok := symbolicLinks[realPath]; ok {
								log.Println("[warning] loop symbolic link")
								return nil
							}
							// NOTE: to avoid "/." e.g. "xxx/."
							realPath = filepath.Clean(realPath)
							symbolicLinks[realPath] = path
							if fi, err := os.Stat(realPath); err != nil && !fi.IsDir() {
								appendFile(realPath)
								return nil
							}
							// NOTE: dir
							err = filepath.Walk(realPath, fWalkFunc)
							return nil
						}
						appendFile(path)
					}
					return nil
				}
				err = filepath.Walk(fpath, fWalkFunc)
				for i, _ := range names {
					// NOTE: rewrite symboliclink
					// NOTE: シンボリックリンクがあるパス上に複数ある場合には短い方が優先されている
					for realPath, path := range symbolicLinks {
						if strings.HasPrefix(names[i], realPath) {
							names[i] = strings.Replace(names[i], realPath, path, 1)
							break
						}
					}
					// NOTE: delete cwd path
					if strings.HasPrefix(names[i], cwd+"/") {
						names[i] = strings.Replace(names[i], cwd+"/", "", 1)
						continue
					}
				}
				data = struct {
					Names []string `json:"names"`
				}{
					names,
				}
				if err != nil {
					log.Println("filepath.Walk", err)
				}
			}

			err = WriteJson(w, r, data)
			if err != nil {
				http.Error(w, "InternalServerError", http.StatusInternalServerError)
				return
			}
			return
		}
		if name == "favicon.ico" {
			name = "_assets/favicon.ico"
		}
		if strings.HasPrefix(name, "_assets/") {
			b, err := OpenIncludingAsset(name)
			if err != nil {
				PageNotFound(w, r)
				//				http.Error(w, "404 page not found", 404)
				return
			}

			w.Header().Set("Content-Type", mime.TypeByExtension(filepath.Ext(name)))
			w.Write(b)
			return
		}
		// use Clean() prevent directory traversal attack
		//	https://golang.org/pkg/path/filepath/#Clean
		name = filepath.Clean(name)
		fpath := filepath.Join(cwd, name)
		fi, err := os.Stat(fpath)
		if err != nil {
			log.Println("os.Stat", err)
			PageNotFound(w, r)
			return
		}
		if fi.IsDir() {
			// NOTE: 厳密にはtreeを構築するためのindex.htmlにあたるファイルを返す(実際には，_apiで取得する)
			err = DirTreeTemplate(w, fpath)
			if err != nil {
				log.Println("template", err)
				//	NOTE 厳密には404ではない?!
				PageNotFound(w, r)
				return
			}
			return
		}

		ext := filepath.Ext(name)
		switch ext {
		//	NOTE 実行結果を表示
		case ".go":
			b, err := goexec("." + fpath)
			content := string(b)
			if err != nil {
				content += err.Error()
			}
			w.Write([]byte(fmt.Sprintf(goexecTemplateHTML, name, config.LivereloadAddr, content)))
			return
			//	NOTE マークダウン生成
		case ".txt", ".md", ".mkd", ".markdown":
			b, err := ioutil.ReadFile(fpath)
			if err != nil {
				if os.IsNotExist(err) {
					PageNotFound(w, r)
					//					http.Error(w, "404 page not found", 404)
					return
				}
				http.Error(w, err.Error(), 500)
				return
			}

			//	NOTE アクセスしたファイルのディレクトリの監視
			//	TODO シンボリックリンクだった場合は?!
			watchDir := filepath.Dir(fpath)
			accessMap.Append(watchDir)
			log.Println("watch dir add", watchDir, "by accessing", name)
			if err := checkWatch(watchDir); err != nil {
				log.Println("watch dir add error:", err)
			}
			if ret, err := IsSymlink(fpath); err == nil && ret {
				if err := checkWatch(fpath); err != nil {
					log.Println("watch symbolic file add error:", err)
				}
			}

			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			renderer := blackfriday.HtmlRenderer(0, "", "")
			b = blackfriday.Markdown(b, renderer, extensions)
			w.Write([]byte(fmt.Sprintf(markdownTemplateHTML, name, config.LivereloadAddr, string(b))))
			//	NOTE 通常のファイル
		default:
			fs.ServeHTTP(w, r)
			return
		}
	})

	server := &http.Server{
		Addr: config.Addr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL.RequestURI())
			http.DefaultServeMux.ServeHTTP(w, r)
		}),
	}

	fmt.Fprintln(os.Stderr, "Lisening at "+config.Addr)
	fmt.Fprintln(os.Stderr, "Lisening at "+config.LivereloadAddr+" livereload")
	log.Fatal(server.ListenAndServe())
}
