package main

import(
"fmt"
"io"
"os"
)

var pipeWrite = "/go/pipe/write"
var pipeRead = "/go/pipe/read"

func readFromRemote(file *os.File){
	for {
		_,err := io.Copy(os.Stdout,file)
		if err != nil {
			panic(err)
		}
	}
	return
}

func main(){
	// to open pipe to write    
	readFile, err := os.OpenFile(pipeRead, os.O_RDWR, os.ModeNamedPipe) 
	if err != nil {
		panic(err)
	}
	//to open pipe to read    
	writeFile,err = os.OpenFile(pipeWrite, os.O_RDONLY, os.ModeNamedPipe)
	if err != nil {
		panic(err)
	}
	go readFromRemote(readFile)
	// Write to Remote
	
	// Send test byte
	test := byte("Hello")
	_, err = writeFile.Write(test)
	if err != nil {
		panic(err)
	}
	// Join stdin to Remote file
	for {
		_,err := io.Copy(writeFile,os.Stdin)
		if err != nil {
			panic(err)
		}
	}
	return
}