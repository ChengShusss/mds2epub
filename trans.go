package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/go-shiori/go-epub"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
)

var (
	e *epub.Epub
)

var (
	secMap   map[string]string = map[string]string{}
	basePath string
)

func Trans() {
	// Create a new EPUB
	var err error
	e, err = epub.NewEpub("My title")
	if err != nil {
		log.Println(err)
	}

	// Set the author
	e.SetAuthor("Hingle McCringleberry")

	// Add a section
	// section1Body := `<h1>Section 1</h1>
	// <p>This is a paragraph.</p>`
	// _, err = e.AddSection(section1Body, "Section 1", "", "")
	// if err != nil {
	// 	log.Println(err)
	// 	os.Exit(1)
	// }

	basePath = "testdata/Ren-master - 副本"
	err = walkDir("")
	if err != nil {
		fmt.Printf("Failed to walk, err: %v\n", err)
		os.Exit(1)
	}

	// Write the EPUB
	err = e.Write("testdata/My EPUB.epub")
	if err != nil {
		// handle error
		fmt.Printf("Falied to create epub, err: %v\n", err)
	}

}

func walkDir(relPath string) error {
	curPath := path.Join(basePath, relPath)
	entries, err := os.ReadDir(curPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		secName := getSecNameFromPath(path.Join(relPath, entry.Name()))

		if entry.IsDir() {
			if _, ok := secMap[secName]; secName != "" && !ok {
				// Add a section
				sectionBody := fmt.Sprintf("<h1>%s</h1>", secName)
				sectionFile, err := e.AddSection(sectionBody, secName, "", "")
				if err != nil {
					log.Println(err)
					os.Exit(1)
				}
				// newFile = path.Join("xhtml", sectionFile)
				secMap[secName] = sectionFile
				fmt.Printf("Build Section, %v, file: %s\n", secName, sectionFile)

			}

			newPath := path.Join(relPath, entry.Name())
			fmt.Printf("Current: %s, newPath: %s\n", curPath, newPath)
			err := walkDir(newPath)
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

			_, err = e.AddSubSection(secMap[secName], string(output), extractName(entry.Name()), "", "")
			fmt.Printf("Add subsection %s after %s, FilePath: %v\n", extractName(entry.Name()), secMap[secName], path.Join(curPath, entry.Name()))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func getSecNameFromPath(input string) string {
	split := strings.Split(input, "/")
	return split[0]
}

func extractName(origin string) string {
	r, err := regexp.Compile(`\d+_(.*)\.md`)
	if err != nil {
		fmt.Printf("Failed to compile regexp, err: %Rv\n", err)
		return origin
	}

	res := r.FindStringSubmatch(origin)
	if len(res) == 0 {
		return origin
	}

	return res[1]
}
