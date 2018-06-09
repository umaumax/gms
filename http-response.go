package main

import (
	"fmt"
	// 	"log"
	//	"html/template"
	// 	"io/ioutil"
	"net/http"
	// 	"os"
	// 	"path/filepath"
	// 	"strings"
	"text/template"
)

//	404
func PageNotFound(w http.ResponseWriter, r *http.Request) {
	b, err := OpenIncludingAsset("_assets/gopher-404.png.base64")
	if err != nil {
		http.Error(w, "404 page not found", 404)
		return
	}
	//	NOTE 404の画像を表示するが、実際には200 OKを返す
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(fmt.Sprintf(pageNotFoundTemplate, r.URL.Path, string(b))))
}

const pageNotFoundTemplate = `
<!DOCTYPE html>
<html>
	<head>
	<meta charset="UTF-8">
	<title>%s</title>
</head>
<body>
	<h1>404 page not found</h1>
	<img src="data:image/png;base64,%s" />"
	<p>Gopher Stickers</p>
	<p>The Go gopher was designed by Renee French. (<a href="http://reneefrench.blogspot.com/">http://reneefrench.blogspot.com/</a>)</p>
	<p>The gopher stickers was made by Takuya Ueda. (<a href="http://u.hinoichi.net">http://u.hinoichi.net</a>)</p>
	<p>Licensed under the Creative Commons 3.0 Attributions license.</p>
</body>
`

var dirTreeTemplate = template.Must(template.New("dirTreeTemplate").Parse(dirTreeTemplateHTML))

func DirTreeTemplate(w http.ResponseWriter, path string) (err error) {
	//	tpl, err := template.ParseFiles("templates/index.tpl")
	//	if err != nil {
	//		return
	//	}

	// 	var f func(path string, depth int) (s string)
	// 	f = func(path string, depth int) (s string) {
	// 		fi, err := os.Stat(path)
	// 		if err != nil {
	// 			return err.Error()
	// 		}
	// 		//	TODO double qoute?!
	// 		s += fmt.Sprintf(`<li><a href="%s">%s</a>`, "/"+path, filepath.Base(path))
	// 		if fi.IsDir() {
	// 			s += "<ul>"
	// 			//			if depth > 0 {
	// 			//				s += `<div style="display:none;">`
	// 			//			}
	// 			fis, err := ioutil.ReadDir(path)
	// 			_ = err
	// 			for _, fi := range fis {
	// 				if fi.IsDir() {
	// 					continue
	// 				}
	// 				if strings.HasPrefix(fi.Name(), ".") {
	// 					if !skipLogFlag {
	// 						log.Println("skip file", filepath.Join(path, fi.Name()), "starts with '.'")
	// 					}
	// 					continue
	// 				}
	// 				s += f(filepath.Join(path, fi.Name()), depth+1)
	// 			}
	// 			//			if depth > 0 {
	// 			//				s += `</div>`
	// 			//			}
	// 			s += "</ul>"
	// 		}
	// 		s += "</li>"
	// 		return
	// 	}
	// 	tree := f(path, 0)
	err = dirTreeTemplate.Execute(w, struct {
		Title string
		Links string
	}{
		Title: config.Title,
		// 		Links: tree,
	})
	if err != nil {
		return
	}
	return
}
