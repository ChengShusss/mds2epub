package main

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/pflag"
)

const (
	mimetypeFilename = "mimetype"
)

func usagePack() {
	fmt.Printf("Usage of mds2epub pack:\n  mds2epub pack [OPTIONS] PATH\nOptions:\n")
	pflag.PrintDefaults()
}

func Pack() {
	var output string
	pflag.StringVarP(&output, "output", "o", "./default.epub", "specify the filename to output")
	pflag.Usage = usagePack
	pflag.Parse()

	if len(pflag.Args()) != 1 {
		fmt.Printf("Invalid input, only accept one dir for input")
		os.Exit(1)
	}
	src := pflag.Args()[0]

	// check src is a valid dir
	fi, err := os.Stat(src)
	if err != nil {
		fmt.Printf("%v\n", err.Error())
		os.Exit(1)
	}
	if !fi.IsDir() {
		fmt.Printf("\"%s\" is not dir\n", src)
		os.Exit(1)
	}

	fmt.Printf("Should start pack here, output: [%s], src: [%v]\n", output, src)

	f, err := os.Create(output)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
	defer f.Close()
	n, err := WriteEpub(src, f)

	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Succeed to write %d bytes into [%s]\n", n, output)
}

// Below is copied from https://github.com/go-shiori/go-epub/blob/main/write.go

// writeCounter counts the number of bytes written to it.
type writeCounter struct {
	Total int64 // Total # of bytes written
}

// Write implements the io.Writer interface.
// Always completes and never returns an error.
func (wc *writeCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Total += int64(n)
	return n, nil
}

// Write the EPUB file itself by zipping up everything from a temp directory
// The return value is the number of bytes written. Any error encountered during the write is also returned.
func WriteEpub(rootEpubDir string, dst io.Writer) (int64, error) {
	counter := &writeCounter{}
	teeWriter := io.MultiWriter(counter, dst)

	z := zip.NewWriter(teeWriter)

	skipMimetypeFile := false

	// addFileToZip adds the file present at path to the zip archive. The path is relative to the rootEpubDir
	addFileToZip := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get the path of the file relative to the folder we're zipping
		relativePath, err := filepath.Rel(rootEpubDir, path)
		if err != nil {
			// tempDir and path are both internal, so we shouldn't get here
			return err
		}
		relativePath = filepath.ToSlash(relativePath)

		// Only include regular files, not directories
		if !info.Mode().IsRegular() {
			return nil
		}

		var w io.Writer
		if filepath.FromSlash(path) == filepath.Join(rootEpubDir, mimetypeFilename) {
			// Skip the mimetype file if it's already been written
			if skipMimetypeFile {
				return nil
			}
			// The mimetype file must be uncompressed according to the EPUB spec
			w, err = z.CreateHeader(&zip.FileHeader{
				Name:   relativePath,
				Method: zip.Store,
			})
		} else {
			w, err = z.Create(relativePath)
		}
		if err != nil {
			return fmt.Errorf("error creating zip writer: %w", err)
		}

		r, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("error opening file %v being added to EPUB: %w", path, err)
		}
		defer func() {
			if err := r.Close(); err != nil {
				log.Println(err)
			}
		}()

		_, err = io.Copy(w, r)
		if err != nil {
			return fmt.Errorf("error copying contents of file being added EPUB: %w", err)
		}
		return nil
	}

	// Add the mimetype file first
	mimetypeFilePath := filepath.Join(rootEpubDir, mimetypeFilename)
	mimetypeInfo, err := os.Stat(mimetypeFilePath)
	if err != nil {
		if err := z.Close(); err != nil {
			log.Println(err)
		}
		return counter.Total, fmt.Errorf("unable to get FileInfo for mimetype file: %w", err)
	}
	err = addFileToZip(mimetypeFilePath, mimetypeInfo, nil)
	if err != nil {
		if err := z.Close(); err != nil {
			log.Println(err)
		}
		return counter.Total, fmt.Errorf("unable to add mimetype file to EPUB: %w", err)
	}

	skipMimetypeFile = true

	filepath.Walk(rootEpubDir, addFileToZip)
	if err != nil {
		if err := z.Close(); err != nil {
			log.Println(err)
		}
		return counter.Total, fmt.Errorf("unable to add file to EPUB: %w", err)
	}

	err = z.Close()
	return counter.Total, err
}
