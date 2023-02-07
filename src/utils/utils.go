package utils

import (
	"archive/zip"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

func ParamsToURL(params map[string]string) string {
	var url string
	for key, value := range params {
		url += "&" + key + "=" + value
	}
	return url
}

func FindFolder(parent string, search string) string { // we dont check for directory now
	for {
		files, err := ioutil.ReadDir(parent)
		if err != nil {
			log.Fatal(err)
		}

		for _, f := range files {
			if f.Name() != search {
				continue
			}

			return filepath.Join(parent, f.Name())
		}
	}
}

func ZipFolder(src string, dst string) error { // add splitting into multiple files
	archive, err := os.Create(dst)
	if err != nil {
		log.Fatal(err)
	}
	defer archive.Close()

	zipWriter := zip.NewWriter(archive)
	defer zipWriter.Close()

	filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		header, _ := zip.FileInfoHeader(info)
		header.Name, _ = filepath.Rel(src, path)
		header.Method = zip.Deflate

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		_, err = io.Copy(writer, file)
		return err
	})

	return nil
}
