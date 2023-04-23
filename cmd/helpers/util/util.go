package util

import (
	"io"
	"log"
	"os"
)

var errorkey = "ERROR:Ampersand-Cli:cli/helpers/utils/Copy: "

func Copy(sourcefile, localcopy string) bool {

	srcFile, err := os.Open(sourcefile)
	if err != nil {
		log.Fatal(errorkey, err)
		return false
	}
	defer srcFile.Close()

	destFile, err := os.Create(localcopy) // creates if file doesn't exist

	if err != nil {
		log.Fatal(errorkey, err)
		return false
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		log.Fatal(errorkey, err)
		return false
	}

	err = destFile.Sync()
	if err != nil {
		log.Fatal(errorkey, err)
		return false
	}
	return true
}
