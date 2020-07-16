package base

import(
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"
	"strings"
	"math/rand"
	"crypto/sha1"
	"encoding/gob"
	"encoding/json"
	"bytes"
)

// Node data structure
type Node struct{
	id int `json:"id"`
	friends map[int]string `json:"friends"`
	ready bool `json:"ready"`
	Version int `json:"version"`
	CreatedAt time.Time `json:"createdAt"`
	ip string `json:"ip"`
	subnet int `json:"subnet"`
}

// Configuration
// Remember to make a folder /files

var (
	CLOSE_FRIENDS = 4
	DEVICE_BITS = 4
	PORT = "9999"
	INTERFACE_NAME = "eth0"
	local="files/"
	MULTICAST_IP = []byte{233,0,0,0}
	MULTICAST_PORT = "10101"
	DASHBOARD_ADDRESS = "http:///"
)

/*
	Checking connectivity
	Fetching IP + Subnet
*/

func (node *Node) CheckConnectivity() {
	ifcs,err := net.Interfaces()
	if err != nil{
		fmt.Println("Fetching interfaces error")
		panic(err)
	}
	for _,v := range ifcs {
		fmt.Println(v.Name)
	}

	myIfc,err := net.InterfaceByName(INTERFACE_NAME)
	if err!= nil{
		fmt.Println("Interface name error")
		panic(err)
	}
	// Probably the first address+subnet
	addrs,err := myIfc.Addrs()
	if err != nil {
		fmt.Println("Interface addr error")
		panic(err)
	}
	if len(addrs) == 0 {
		fmt.Println("No addrs found")
		panic(err)
	}
	l := strings.Split(addrs[0].String(),"/")
	node.ip = l[0]
	subnet, err := strconv.ParseInt(l[1],10,32) 
	if err != nil {
		fmt.Println("Subnet Parsing error")
		panic(err)
	}
	node.subnet = int(subnet)
	return
}

/*
	Discover Friends
*/

func (node *Node) DiscoverFriends() {
	addr := net.ParseIP(node.ip)
	if len(addr)== 16 {
		addr = addr[12:]
		fmt.Println(addr)
	}
	subnet := node.subnet
	// if num == -1 {
		// num = 32- subnet
	// }
	lim := (1<<(32-subnet) - 1)
	// Testing
	lim = 16
	p := 3
	id := make([]byte,100)

	// ImPORTant Feature for checking availibity in Gossip network
	for x:=1; x < lim; x++ {
		addr[p] ^=  byte(x)
		y := fmt.Sprintf("%v.%v.%v.%v",addr[0],addr[1],addr[2],addr[3])
		addr[p] ^= byte(x)
		destAddr,err := net.ResolveTCPAddr("tcp",net.JoinHostPort(y,PORT))
		if err != nil {
			panic(fmt.Sprintf("%v\n",err))
		}
		conn,err := net.DialTCP("tcp", nil, destAddr)
		if err != nil {
			fmt.Println(err) // No Need to print
			continue
		}
		// Connected
		fmt.Println("Connected")
		fmt.Fprintf(conn,"id\n")
		status,err := conn.Read(id)
		if status == 1 {
			node.friends[int(id[0])] = y
		}
		if len(node.friends) == CLOSE_FRIENDS {
			break
		}
		// Testing
		break
		// Changing byte order
		if x == 255 {
			x = 1
			p--
			lim = lim>>8
		}
	}
}

/*
  Find Remote id using BFS Queue based impletation
*/

func (node *Node) FindFriend(id int) (string){
	
	visited := make(map[string]bool)
	var friendNode *Node;
	
	// Direct Check for neighbours
	if x,found := node.friends[id]; found == true {
			return x
	}
	
	// Channel based queue with non-blocking select condition
	queue := make(chan string)
	for _,x := range node.friends {
		queue <- x
	}

	for {
		select {
		case s := <-queue:
			if _,t := visited[s]; t == true {
				continue
			}
			visited[s] = true
			// Connect to friend
			destAddr,err := net.ResolveTCPAddr("tcp",net.JoinHostPort(s,PORT))
			if err != nil {
				fmt.Println(err)
			}
			conn,err := net.DialTCP("tcp",nil,destAddr)
			if err != nil {
				fmt.Println(err) // Not necessary to print this error
				continue
			}
			fmt.Fprintf(conn,"info\n")
			// Read the data into another map
			dec := gob.NewDecoder(conn)
			err = dec.Decode(friendNode)
			if err != nil {
				fmt.Println(err)
				panic(fmt.Sprintf("%v\n",err))
			}
			fof := friendNode.friends
			if x,found := fof[id]; found==true {
				node.friends[id] = x
				return x
			}
			for _,s := range fof {
				queue <- s
			}
		default:
			fmt.Printf("friend not found")
			break
		}
	}
	return ""
}

/*
	Uploading files or directory
*/

func (node *Node) UploadFile(path string) error{
	f,err := os.Open(path)
	if err != nil {
		return err
	}
	st,err := f.Stat()
	if err != nil {
		return err
	}
	if st.IsDir() == true {
		filenames,err := f.Readdirnames(0)
		if err == nil {
			return err
		}
		for _,n := range filenames {
			err := node.UploadFile(path+n)
			if err != nil {
				fmt.Println(err,fmt.Sprintf("...in uploading %v\n",path+n))
			}
		}
		return nil
	}
	// Read file and upload
	fileName := st.Name()
	h := sha1.New()
	h.Write([]byte(fileName))
	hash := h.Sum(nil)

	destId := int(hash[0]>>(8-DEVICE_BITS))

	if destId == node.id {
		conn,err := os.Open(fmt.Sprintf("%v/%v"))
		if err != nil{
			return err
		}
		PutFile(path, conn)
	} else {
		// Send to the multicast group
		ip := MULTICAST_IP
		ip[3] ^= byte(destId)
		udpAddr,err := net.ResolveUDPAddr("udp",net.JoinHostPort(string(ip),PORT))
		if err!= nil {
			return err
		}
		conn,err := net.DialUDP("udp",nil,udpAddr)
		if err != nil {
			return err
		}
		fmt.Fprintf(conn,"put %v:%v\n",st.Name(),st.Size())	
		PutFile(path, conn)
	}
	return nil
}

/*
	Send a file to destination
*/

func PutFile(path string, conn io.Writer) {
	f,err := os.Open(path)
	if err != nil { 
		fmt.Println(err)
		return
	}
	_,err = io.Copy(conn,f)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Successful: %v Put request\n",path)
	return
}


/*
	Upload Procedure 
*/

func (node *Node) UploadProc(){
	var path string
	fmt.Println("Enter file or directory path to upload...")
	fmt.Scanln(&path)
	err := node.UploadFile(path)
	if err != nil {
		fmt.Println(err,fmt.Sprintf("...in uploading %v\n",path))
	}
	return
}

/*
	Assigning id to myself
*/

func (node *Node) Enlightenment() {
	if node.id > -1 {
		fmt.Println("Already Enlightened")
	}
	seed := rand.NewSource(time.Now().UnixNano())
	r := rand.New(seed)
	high := 1<<DEVICE_BITS
	tmp := 0
	for {
		tmp = r.Intn(high)
		_,found := node.friends[tmp]
		if found == false {
			break
		}
	}
	node.id = tmp
	return
}

/*	
	Servicing loop would contains all types of request-response
	Main Server Thread
*/

func (node *Node) ListenerLoop() {

	listenAddr,err := net.ResolveTCPAddr("tcp",net.JoinHostPort(node.ip,PORT))
	if err!= nil {
		panic(err)
	}
	ln, err := net.ListenTCP("tcp", listenAddr)
	if err != nil {
		panic(err)
	}
	for {
		conn, err := ln.AcceptTCP() 
		if err != nil {
			fmt.Println(err)
		}
		go node.ServiceRequest(conn)
	}
	return
}

/* 
	Service Requests
	net.IPConn is a structure that implements Conn and PacketConn interfaces
*/
func (node *Node) ServiceRequest(conn *net.TCPConn) {
	defer (*conn).Close()
	req := make([]byte, 100)
	reqCount,err := (*conn).Read(req)
	if err!= nil {
		fmt.Println(err)
	}
	if reqCount > 0 {
		reqArgs := strings.Split(string(req)," ")
		switch reqArgs[0] {
		case "id":
			fmt.Printf("%v\n",node.id)
		case "info":
			// Serialize all node data
			enc := gob.NewEncoder(conn)
			err := enc.Encode(*node)
			if err != nil {
				fmt.Println(err)
				return
			}
		case "put":
			if len(reqArgs)!= 3 {
				fmt.Println("Put request not enough arguments")
				return
			}
			fileName := reqArgs[1]
			// No Null Checking
			fileSize,err := strconv.ParseInt(reqArgs[2],10,32)
			if err!= nil {
				fmt.Println(err)
				return
			}
			GetFile(conn,fileName, int(fileSize), false)

		case "get":
			if len(reqArgs) != 2 {
				fmt.Println("Get Request requires 2 arguments")
				return
			}
			fileName := reqArgs[1]
			// If file exists
			_,err := os.Open(fileName)
			if err!= nil {
				fmt.Fprintf(conn,"Error:%v\n",err)
				return
			}
			PutFile(fileName,conn)
		}
	}
	return
}

/*
	Get File from Remote conn
	Store if print is false
	else Print
*/
func GetFile(conn io.Reader,name string, size int, print bool) {
	f := os.Stdout
	var err error
	if print == false {
		f,err = os.Create(fmt.Sprintf("%v/%v",local,name))
		if err!= nil {
			fmt.Println(err)
		}
	}
	wLen,err := io.Copy(f, conn)	
	if err != nil {
		fmt.Println(err)
		return
	}
	if int(wLen) != size {
		fmt.Println("Get returned with insufficient data")
		return
	}
	return
}


/*
 Search/Get feature
*/

func (node *Node)SearchProc() {
	var fileName string
	fmt.Println("Search File:")
	fmt.Scanln(&fileName)
	// Hashing file name to find its location
	h := sha1.New()
	h.Write([]byte(fileName))
	hash := h.Sum(nil)

	// Making conn as reader interface
	destId := int(hash[0]>>(8-DEVICE_BITS))
	if destId == node.id {
		conn,err := os.Open(fmt.Sprintf("%v/%v",local,fileName))
		if err != nil {
			fmt.Println(err)
			return
		}
		st,err := conn.Stat()
		if err!= nil {
			fmt.Println(err)
			return
		}
		fileSize := int(st.Size())
		GetFile(conn, fileName, fileSize,true)

	} else {
		destIP := node.FindFriend(destId)
		destAddr,err := net.ResolveTCPAddr("tcp",net.JoinHostPort(destIP,PORT))
		if err!= nil {
			fmt.Println(err)
			return
		}
		conn,err := net.DialTCP("tcp",nil,destAddr)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer conn.Close()
		fmt.Fprintf(conn,"get %v\n",fileName)

		res := make([]byte,100)
		_,err = conn.Read(res)
		if err!= nil {
			fmt.Println(err)
			return
		}
		args := strings.Split(string(res)," ")
		if(len(args)!= 3 || args[0] != "put"){
			fmt.Println(string(res))
			return
		}
		fileSize,err := strconv.ParseInt(args[2],10,32)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("%v: %v bytes\n",fileName, fileSize)
		GetFile(conn, fileName, int(fileSize),true)
	}
	return
}


/*
	Connect to the Multicast Group
*/
func (node *Node) JoinSociety() {
	if node.id < 0 {
		node.Enlightenment()
	}
	ip := MULTICAST_IP
	ip[3] ^= byte(node.id)
	udpAddr,err := net.ResolveUDPAddr("udp",net.JoinHostPort(string(ip),MULTICAST_PORT))
	if err!= nil {
		panic(err)
	}
	myIfc,err := net.InterfaceByName(INTERFACE_NAME)
	if err!= nil{
		panic(err)
	}
	// Listen for data
	conn,err := net.ListenMulticastUDP("udp", myIfc, udpAddr)
	if err!= nil{
		panic(err)
	}
	bufSize := 5000
	buf := make([]byte,bufSize)	
	// Only one file can be sent in a multicast group
	for {
		_,err = conn.Read(buf) // This should be blocking
		if err != nil {
			panic(err)
		}
		req := strings.Split(string(buf)," ") 
		if req[0] != "put" {
			continue
		}
		fileName := req[1]
		fileSize,err := strconv.ParseInt(req[2],10,32)
		if err!= nil {
			fmt.Println(err)
			return
		}
		GetFile(conn,fileName,int(fileSize),false)
	}
}


/*
 	Show info to user
*/
func (node *Node) ShowInfo() {

	fmt.Println("Node Info")

	fmt.Printf("ip:%v\n",node.ip)
	if node.id >= 0 {
		fmt.Printf("Assigned:%v\n",node.id)
	} else{
		fmt.Println("Unassigned!!")
	}
	return
}

// /*
// 	Graceful Exit
// */

// func (node *Node) GracefulExit() {
// 	return

// }


/*
 Send ping to Dashboard
*/


func (node *Node) SendToDashboard() {
	buf,err := json.Marshal(*node)
	if err!= nil {
		fmt.Println(err)
		return
	}
	r := bytes.NewReader(buf)
    resp, err := http.Post(DASHBOARD_ADDRESS ,"application/json",r)
    if err != nil {
    	fmt.Println(err)
    	return
    }
    fmt.Printf("Logged Message:%v\n", resp.Status)
    return
}

