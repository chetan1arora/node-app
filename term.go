package main

import(
"io"
"os"
)

var pipeWrite = "pipe/write"
var pipeRead = "pipe/read" 
func main(){
	// Opening pipe to Write  
	writeFile,err := os.OpenFile(pipeWrite, os.O_WRONLY, os.ModeNamedPipe)
	if err != nil {
		panic(err)
	}
	// Opening pipe to Read
	readFile, err := os.OpenFile(pipeRead, os.O_RDONLY, os.ModeNamedPipe) 
	if err != nil {
		panic(err)
	}
	// Hooking remote output to stdout
	go io.Copy(os.Stdout,readFile)
	// Hooking local Input to remote
	io.Copy(writeFile, os.Stdin)
	return
}