package main

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/gomarkdown/markdown"
)

type Page struct {
	Name    string
	Content template.HTML
}

var (
	tmpl = template.Must(template.ParseFiles("base.tmpl"))

	srcdir string
	outdir string
)

func walkfunc(path string, info fs.FileInfo, err error) error {
	if err != nil || info.IsDir() {
		return err
	}

	// prepare path
	rp, err := filepath.Rel(srcdir, path)
	if err != nil {
		return err
	}

	dest := filepath.Join(outdir, rp)
	if err := os.MkdirAll(filepath.Dir(dest), os.ModePerm); err != nil {
		return err
	}

	// process file
	if filepath.Ext(path) == ".md" {
		doc, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		doc = markdown.ToHTML(doc, nil, nil)
		dest = dest[:len(dest)-2] + "html"

		f, err := os.Create(dest)
		if err != nil {
			return err
		}

		tmpl.ExecuteTemplate(f, "base", Page{Name: filepath.Base(path), Content: template.HTML(
			bytes.ReplaceAll(doc, []byte("\n"), []byte("\n\t\t")),
		)})
	} else {
		src, err := os.Open(path)
		if err != nil {
			return err
		}

		dst, err := os.Create(dest)
		if err != nil {
			return err
		}

		_, err = io.Copy(dst, src)
	}

	fmt.Println(path, "->", dest)

	return err
}

func init() {
	flag.StringVar(&srcdir, "src", "src", "source directory")
	flag.StringVar(&outdir, "out", "site", "out directory")
	flag.Parse()
}

func main() {
	if err := filepath.Walk(srcdir, walkfunc); err != nil {
		log.Fatal(err)
	}
}
