package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

const (
	CLR_W = ""
	CLR_R = "\x1b[31;1m"
	CLR_G = "\x1b[32;1m"
	CLR_B = "\x1b[34;1m"
	CLR_Y = "\x1b[33;1m"
)

var exitCode int

// Parse date by std date string
func ParseDate(dateStr string) time.Time {
	date, err := time.Parse(fmt.Sprintf("%s -0700", time.DateTime), dateStr)
	if err != nil {
		date, err = time.ParseInLocation(time.DateTime, dateStr, time.Now().Location())
		if err != nil {
			log.Fatalln(err)
		}
	}
	return date
}

// Check file if exist
func Exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

// Check file if is directory
func IsDir(path string) bool {
	file, err := os.Stat(path)
	if err != nil {
		return false
	}
	return file.IsDir()
}

// Copy folder and file
// Refer to https://www.socketloop.com/tutorials/golang-copy-directory-including-sub-directories-files
func CopyFile(source string, dest string) {
	sourcefile, err := os.Open(source)
	if err != nil {
		log.Fatalln(err)
	}
	destfile, err := os.Create(dest)
	if err != nil {
		log.Fatalln(err)
	}
	defer destfile.Close()
	defer wg.Done()
	_, err = io.Copy(destfile, sourcefile)
	if err != nil {
		log.Fatalln(err)
	}
	sourceinfo, err := os.Stat(source)
	if err != nil {
		log.Fatalln(err)
	}
	err = os.Chmod(dest, sourceinfo.Mode())
	if err != nil {
		log.Fatalln(err)
	}
	sourcefile.Close()
}

func CopyDir(source string, dest string) {
	sourceinfo, err := os.Stat(source)
	if err != nil {
		log.Fatalln(err)
	}
	err = os.MkdirAll(dest, sourceinfo.Mode())
	if err != nil {
		log.Fatalln(err)
	}
	directory, _ := os.Open(source)
	defer directory.Close()
	defer wg.Done()
	objects, err := directory.Readdir(-1)
	if err != nil {
		log.Fatalln(err)
	}
	for _, obj := range objects {
		sourcefilepointer := source + "/" + obj.Name()
		destinationfilepointer := dest + "/" + obj.Name()
		if obj.IsDir() {
			wg.Add(1)
			CopyDir(sourcefilepointer, destinationfilepointer)
		} else {
			wg.Add(1)
			go CopyFile(sourcefilepointer, destinationfilepointer)
		}
	}
}
