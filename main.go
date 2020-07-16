package main

import(
"./pkg"
"fmt"
"time"
)

/*
	Main Node Application
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

	// Main Purpose
	for {
		showGreeting()
		// Testing 
		time.Sleep(25*time.Second)
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
