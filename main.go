package main

import(
"./pkg"
"syscall"
"fmt"
"os"
)

/*
 Greeting
*/
func showGreeting(){
	fmt.Println("|----------------------------------------------------|")
	fmt.Println("*------------------____---_______---__---------___---*")
	fmt.Println("|----/|---/-------/----------/-----/--\\-------/---\\--|")
	fmt.Println("*---/-|--/-------/___-------/---------|-------|---|--*")
	fmt.Println("|--/--|-/-------/----------/---------/----_---|---|--|")
	fmt.Println("*-/---|/-------/____------/---------/__--|_|--\\___/--*")
	fmt.Println("|----------------------------------------------------|")
	fmt.Println("")
	fmt.Println("")
	fmt.Println("Choose the operation:")
	fmt.Println("[1] Get/Search File")
	fmt.Println("[2] Upload file to network")
	fmt.Println("[3] Show info")
	fmt.Println("[4] Graceful Exit")
	return
}
/*
 Handle Error
*/
func handleErr(err error){
	if err != nil {
		panic(err)
	}
}

var pipeWrite = "pipe/read"
var pipeRead = "pipe/write"

func waitForTerminal(){
	os.RemoveAll("pipe")
	err := syscall.Mkdir("pipe",0777)
	handleErr(err)
	
	// Making Pipe for transport
	syscall.Mknod(pipeRead, syscall.S_IFIFO|0666, 0)
	syscall.Mknod(pipeWrite, syscall.S_IFIFO|0666, 0)
	// Opening pipe to Read
	readFile, err := os.OpenFile(pipeRead, os.O_RDONLY, os.ModeNamedPipe)
	handleErr(err)
	readFd := int(readFile.Fd())
	
	// Opening pipe to Write   
	writeFile,err := os.OpenFile(pipeWrite, os.O_WRONLY, os.ModeNamedPipe)
	handleErr(err)
	writeFd := int(writeFile.Fd())

	// Replace Stdin and Stdout with pipe processes
	err = syscall.Close(0) // STDIN
	err = syscall.Close(1) // STDOUT
	handleErr(err)
	err = syscall.Dup2(readFd,0)
	err = syscall.Dup2(writeFd,1)
	handleErr(err)
	return
}

/*
	Main Node Application
*/
func main(){
	var myNode base.Node

	// Initialize Node
	myNode.CheckConnectivity()

	// Discover Friends
	myNode.DiscoverFriends()

	// Start Listening
	go myNode.ListenerLoop()

	// Join Society
	go myNode.JoinSociety()

	// Wait for terminal
	waitForTerminal()

	// Client Functions
	for {
		showGreeting()
		var op int
		fmt.Scanln(&op)
		switch op {
		case 1: // Search feature
			myNode.SearchProc()
		case 2: // Upload feature
			myNode.UploadProc()
		case 3: // info feature
			myNode.ShowInfo()
		case 4: // Graceful Exit
			// myNode.GracefulExit()
			break
		}
	}
	return
}
