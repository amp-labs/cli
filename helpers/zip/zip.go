package zip

import (
  "archive/zip"
  "fmt"
  "io"
  "log"
  "os"
) 

const errorkey = "ERROR:Ampersand-Cli:cli/helpers/zip: "

func Zip(filename, zipPath string) bool {
  fmt.Println("Beginning to zip amp.yaml.....")
  
  // Open the file to be zipped
  file, err := os.Open(filename)
  if err != nil {
    log.Fatal(errorkey,err)
    return false
  }
  defer file.Close()
  
  // Create the output zip file
  zipFile, err := os.Create(zipPath)
  if err != nil {
    log.Fatal(errorkey,err)
    return false
  }
  defer zipFile.Close()
  
  // Create a new zip archive
  archive := zip.NewWriter(zipFile)
  defer archive.Close()
  
  // Add the file to the zip archive
  fileInfo, err := file.Stat()
  if err != nil {
    log.Fatal(errorkey,err)
    return false
  }
  
  header, err := zip.FileInfoHeader(fileInfo)
  if err != nil {
    log.Fatal(errorkey,err)
    return false
  }
  
  header.Name = file.Name()
  
  writer, err := archive.CreateHeader(header)
  if err != nil {
    log.Fatal(errorkey,err)
    return false
  }
  
  if _, err = io.Copy(writer, file); err != nil {
    log.Fatal(errorkey,err)
    return false
  }
  
  fmt.Println("amp.yaml File zipped successfully!.....")
  return true
}
