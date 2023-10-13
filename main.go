package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"regexp"

	"github.com/go-shiori/go-epub"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
)

var (
	e *epub.Epub
)

func main() {
	// Create a new EPUB
	var err error
	e, err = epub.NewEpub("My title")
	if err != nil {
		log.Println(err)
	}

	// Set the author
	e.SetAuthor("Hingle McCringleberry")

	// Add a section
	section1Body := `<h1>Section 1</h1>
	<p>This is a paragraph.</p>`
	_, err = e.AddSection(section1Body, "Section 1", "", "")
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	err = walkDir("testdata", "")
	if err != nil {
		fmt.Printf("Failed to walk, err: %v\n", err)
		os.Exit(1)
	}

	// Write the EPUB
	err = e.Write("My EPUB.epub")
	if err != nil {
		// handle error
		fmt.Printf("Falied to create epub, err: %v\n", err)
	}

}

func walkDir(curPath, parent string) error {
	entries, err := os.ReadDir(curPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			// fmt.Printf("Walk into dir: %v\n", entry.Name())

			// if parent == "" {
			// 	section1Body := fmt.Sprintf("<h1>%s</h1>", entry.Name())
			// 	_, err = e.AddSection(section1Body, entry.Name(), parent + entry.Name() + ".html", "")
			// 	if err != nil {
			// 		return err
			// 	}
			// } else {
			// 	section1Body := fmt.Sprintf("<h1>%s</h1>", entry.Name())
			// 	_, err = e.AddSubSection(parent + ".html", )
			// 	if err != nil {
			// 		return err
			// 	}
			// }

			err := walkDir(path.Join(
				curPath, entry.Name()),
				parent+"/"+entry.Name(),
			)
			if err != nil {
				return err
			}
			continue
		}

		if path.Ext(entry.Name()) == ".md" {
			// TODO more complicated process of md file
			input, err := os.ReadFile(path.Join(curPath, entry.Name()))
			if err != nil {
				return err
			}

			output := markdown.ToHTML(input, nil, &html.Renderer{})

			// fmt.Println(string(output))
			imgRE := regexp.MustCompile(`<img.*?src="(.*?)".*?>`)
			imgs := imgRE.FindAllStringSubmatch(string(output), -1)

			fmt.Printf("Parent: %v, Curfile: %v\n", parent, path.Join(curPath, entry.Name()))
			for _, img := range imgs {
				fmt.Printf("<images: %v>\n", img)
				fmt.Println(img[1])
			}

			_, err = e.AddSection(string(output), parent+"-"+entry.Name(), "", "")
			if err != nil {
				return err
			}

		}

	}

	return nil
}
