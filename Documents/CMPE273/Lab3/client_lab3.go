package main

import (

	//"encoding/json"
	"fmt"
	"io/ioutil"
	//"math"
	//"crypto/md5"
	"log"
	//"math"
	"net/http"
	"os"
	//"strconv"
	"hash"
	"hash/crc32"
	"sort"
	"strings"
)

var (
	crc32Table = crc32.MakeTable(crc32.Castagnoli)
)

type nodeScore struct {
	node  []byte
	score uint32
}

type byScore []nodeScore

func (s byScore) Len() int           { return len(s) }
func (s byScore) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s byScore) Less(i, j int) bool { return s[i].score < s[j].score }

type Hash struct {
	nodes  []nodeScore
	hasher hash.Hash32
}

// New creates a new Hash with the given keys
func New(nodes ...string) *Hash {
	hash := &Hash{}
	hash.hasher = crc32.New(crc32Table)
	hash.Add(nodes...)
	return hash
}

// Add takes any number of nodes and adds them to this Hash.
func (h *Hash) Add(nodes ...string) {
	for _, node := range nodes {
		h.nodes = append(h.nodes, nodeScore{[]byte(node), 0})
	}
}

// Get returns the node with the highest score for the given key. If this Hash
// has no nodes, an empty string is returned.
func (h *Hash) Get(key string) string {
	keyBytes := []byte(key)

	var maxScore uint32
	var maxNode []byte
	var score uint32

	for _, node := range h.nodes {
		score = h.hash(node.node, keyBytes)
		if score > maxScore {
			maxScore = score
			maxNode = node.node
		}
	}

	return string(maxNode)
}

// GetN returns n nodes for the given key, ordered by descending score.
func (h *Hash) GetN(n int, key string) []string {
	if len(h.nodes) == 0 || n == 0 {
		return []string{}
	}

	if n > len(h.nodes) {
		n = len(h.nodes)
	}

	keyBytes := []byte(key)
	for i := 0; i < len(h.nodes); i++ {
		h.nodes[i].score = h.hash(h.nodes[i].node, keyBytes)
	}
	sort.Sort(sort.Reverse(byScore(h.nodes)))

	nodes := make([]string, n)
	for i := 0; i < n; i++ {
		nodes[i] = string(h.nodes[i].node)
	}
	return nodes
}

func (h *Hash) hash(node, key []byte) uint32 {
	h.hasher.Reset()
	h.hasher.Write(key)
	h.hasher.Write(node)
	return h.hasher.Sum32()
}

var Ring *Hash
var myServers []string
var myData = make(map[string]string)

//Main function for Client, if Client is to be run as a Simple Standalone GO application
func main() {
	myServers = []string{
		"127.0.0.1:3001",
		"127.0.0.1:3002",
		"127.0.0.1:3003",
		"127.0.0.1:3004",
		"127.0.0.1:3005"}

	Ring = New(myServers...)
	argument2 := os.Args[2]
	fmt.Printf("Argument2: %s", argument2)

	//Generate seed-data to be distributed across the servers
	GenMappings(argument2)

	//Distribute keys
	for key, value := range myData {
		fmt.Println("Distributing/Saving this <key-valye> pair : <", key, "-", value, ">")
		PutOperation(key, value)
	}

	//Try to retrive few keys
	fmt.Println("Retrieving FEW keys now!")
	GetOperation("2")
	GetOperation("4")
	GetOperation("1")

	//Try to retrive few keys
	fmt.Println("Retrieving ALL keys now!")

	GetAllOperation("127.0.0.1:3001")
	GetAllOperation("127.0.0.1:3002")
	GetAllOperation("127.0.0.1:3003")
	GetAllOperation("127.0.0.1:3004")
	GetAllOperation("127.0.0.1:3005")

}

func GenMappings(values string) {

	f := func(c rune) bool {
		return c == ',' || c == '-' || c == '>'
	}
	// Separate into fields with func.
	fields := strings.FieldsFunc(values, f)
	fmt.Println(fields)
	fmt.Println(len(fields))
	for i := 0; i < len(fields); i = i + 2 {

		myData[fields[i]] = fields[i+1]
		fmt.Printf("key : %s, value : %s\n", fields[i], myData[fields[i]])
	}
}

func PutOperation(key string, value string) {
	// Grab id
	id := key

	server := Ring.Get(id)
	fmt.Println("The <key-value> pair to be saved : <", key, "-", value, ">")
	fmt.Println("The server for this key : ", server)

	//make a corresponding PUT request to the server here
	url := "http://" + server + "/keys/" + id + "/" + value

	client := &http.Client{}
	request, err := http.NewRequest("PUT", url, strings.NewReader(""))
	response, err := client.Do(request)
	if err != nil {
		log.Fatal(err)
	} else {
		defer response.Body.Close()
		_, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Fatal(err)
		}
	}
	fmt.Println("Key saved successfully!")
	fmt.Println("------------------------------------------------------------")
}

/*
* GetOperation function - Function to support GET operation and retrieve keys from sharded the data set into three server instances
 */
func GetOperation(key string) {
	// Grab id
	id := key

	server := Ring.Get(id)
	fmt.Println("Retrieving key from this server : ", server) //127.0.0.1:3001

	//make a corresponding GET request to the server here
	url := "http://" + server + "/keys/" + id

	response, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	} else {
		defer response.Body.Close()
		contents, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("The Response is : %s\n", string(contents))
	}

	fmt.Println("Key retrieved successfully!")
	fmt.Println("------------------------------------------------------------")

}

/*
* GetAllOperation function - Function to support GET ALL operation and retrieve ALL keys from any particular server instances
 */
func GetAllOperation(server string) {
	fmt.Println("Retrieving ALL keys from this server : ", server)

	//make a corresponding GET request to the server here
	url := "http://" + server + "/keys"

	response, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	} else {
		defer response.Body.Close()
		contents, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("The Response is : %s\n", string(contents))
	}

	fmt.Println("Keys retrieved successfully!")
	fmt.Println("------------------------------------------------------------")

}
