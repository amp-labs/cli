package util

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"path"
	"fmt"
)

var errorkey = "ERROR:Ampersand-Cli:cli/helpers/utils/Copy: "

func CopyDir(srcDir string) (destination string, copied bool) {
	var tempDir = os.TempDir()
	var folderName = "/amp"
	var destFolder  = filepath.ToSlash(filepath.Join(tempDir,folderName))

	var err error
	var srcinfo os.FileInfo
	var fds []os.FileInfo

	if srcinfo, err = os.Stat(srcDir); err != nil {
		log.Fatal(errorkey, err)
		return "",false
	}

	if err = os.MkdirAll(destFolder, srcinfo.Mode()); err != nil {
		log.Fatal(errorkey, err)
		return "",false
	}

	if fds, err = ioutil.ReadDir(srcDir); err != nil {
		log.Fatal(errorkey, err)
		return "",false
	}

	for _, fd := range fds {
		srcfp := path.Join(srcDir, fd.Name())
		dstfp := path.Join(destFolder, fd.Name())

		if err = copyFile(srcfp, dstfp); err != nil {
			fmt.Println("abc")
			fmt.Println(err)
		}
	}

	return destFolder,true
}

func copyFile(sourcefile, localcopy string) error  {

	srcFile, err := os.Open(sourcefile)
	if err != nil {
		log.Fatal(errorkey, err)
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(localcopy) // creates if file doesn't exist

	if err != nil {
		log.Fatal(errorkey, err)
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		log.Fatal(errorkey, err)
		return err
	}

	err = destFile.Sync()
	if err != nil {
		log.Fatal(errorkey, err)
		return err
	}
	return nil
}
