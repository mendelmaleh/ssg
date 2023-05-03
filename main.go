package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

var (
	mdtohtml chan (string)
	wg       sync.WaitGroup
)

func walkfunc(path string, info fs.FileInfo, err error) error {
	if err != nil {
		return err
	}

	if info.IsDir() {
		return nil
	}

	fmt.Println(path)

	wg.Add(1)
	if filepath.Ext(path) == ".md" {
		mdtohtml <- path
	}

	return nil
}

func main() {
	if err := func() error {
		// templates
		tmpl := template.Must(template.ParseFiles("base.tmpl"))

		// markdown
		mdtohtml = make(chan string)
		go func() {
			for path := range mdtohtml {
				md, err := os.ReadFile(path)
				if err != nil {
					log.Fatal(err)
				}

				p := parser.NewWithExtensions(parser.CommonExtensions)
				r := html.NewRenderer(html.RendererOptions{
					Flags: html.CommonFlags,
				})

				tree := p.Parse(md)
				// ast.Print(os.Stdout, tree)
				doc := markdown.Render(tree, r)

				// replace "src" with "out" and ".md" with ".html"
				path = "out" + path[3:len(path)-3] + ".html"

				// TODO: mkdirall
				f, err := os.Create(path)
				if err != nil {
					log.Fatal(err)
				}

				tmpl.ExecuteTemplate(f, "base", template.HTML(
					bytes.ReplaceAll(doc, []byte("\n"), []byte("\n\t\t")),
				))

				wg.Done()
			}
		}()

		if err := filepath.Walk("src", walkfunc); err != nil {
			return err
		}

		wg.Wait()

		return nil
	}(); err != nil {
		log.Fatal(err)
	}
}
