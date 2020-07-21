package main

import(
"./pkg"
"syscall"
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
	fmt.Println("[3] Graceful Exit")
	fmt.Println("[4] Show info")
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

var pipeWrite = "/go/pipe/read"
var pipeRead = "/go/pipe/write"

func waitForTerminal(){
	err := os.Mkdir("/go/pipe",0666)	
	handleErr(err)
	err = syscall.Mkfifo(pipeWrite, 0666)
	handleErr(err)
	err = syscall.Mkfifo(pipeRead, 0666)
	handleErr(err)
	// to open pipe to write    
	readFd, err := syscall.Open(pipeRead, syscall.O_RDONLY| syscall.O_CREAT, 755)
	handleErr(err)
	//to open pipe to read    
	writeFd,err = syscall.Open(pipeWrite, syscall.O_WRONLY| syscall.O_CREAT, 755)
	handleErr(err)
	// Wait for client terminal before closing process fds
	temp := make([]byte,10)
	_, err = syscall.Read(readFd,temp)
	handleErr(err)
	// Replace process fds with remote processes
	err = syscall.Close(0) // STDIN
	handleErr(err)
	err = syscall.Close(1) // STDOUT
	handleErr(err)
	err = syscall.Dup2(readFd,0)
	handleErr(err)
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

	// Main Purpose
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
