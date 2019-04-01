package main

import (
	"bytes"
	//"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	//"log"
	"net/http"

	// "github.com/ethereum/go-ethereum/ethclient"
	"github.com/gorilla/websocket"
)

// upgrade original http request to ws
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// used for json format a response from an RPC call
// type Resp struct {
// 	jsonrpc string
// 	id      int
// 	result  string
// }

type Resp map[string]interface{}

// used to json format an RPC call
type Call struct {
	Jsonrpc string   `json:"jsonrpc"`
	Method  string   `json:"method"`
	Params  []string `json:"params"`
	Id      string   `json:"id"`
}

func main() {
	client := setupClient()
	fmt.Println("client", client)
}

/*
	setupClient will boot either the gui or command line pre config option based on the flag that is passed in to run
	this application. --cmd || --gui respectively. If --cmd flag is passed, the user will be prompted to enter a series
	of http endpoints with rpc enabled clients. This will allow a user to register one or many nodes with our network
	stats client.
 */
func setupClient() http.Client {
	cmdLinePtr := flag.Bool("cmd", false, "should we boot without gui")
	guiPtr := flag.Bool("gui", false, "should we boot with gui")

	flag.Parse()

	cmd := *cmdLinePtr
	gui := *guiPtr

	setupOption := "cmd"

	if cmd {
		fmt.Println("[command line setup]: ", cmd)
	} else if gui {
		fmt.Println("[gui setup]: ", gui)
		// future flag for switching cmdline testing to gui
		setupOption = "gui"
	}

	fmt.Println(setupOption)
	setupClientEndpoints(setupOption)

	return http.Client{}
}

// additional option to listen in on multiple endpoint for various clients.
// this will allow other developers to setup their own cluster of nodes with custom UI
// this will also allow a single server for multiple nodes vs many API's for different nodes on a single server.
// instead of piping a single node data within the web socket connection, you can pipe many sets of aggregated data.
func setupClientEndpoints (setupOption string) []string {
	var endpoints = []string{"http://localhost:8545"}
	switch setupOption {
	case "gui":
		fmt.Println("Setting up GUI")
	case "cmd":
		fmt.Println("Setting up command line")
		// get all node endpoints (enhanced user option) for configuration
		fmt.Println(endpoints)
	default:
		panic("unrecognized escape character")
	}

	return endpoints
}


func parseCmdLineEndpoints() {

}

// client, call struct
// attach array[string] with call struct (enable multiple local nodes to share stats with single handler
// to maximize server efficiency. Ability to run many nodes with single server api that registers to all nodes at any
// given endpoint.
// this would allow other developers to custom display their nodes as a subset.
func setupRpcCalls() {

}

// returns peer count
func getPeerCount(client http.Client, url string) (string, error) {
	// panic("get peer count function")

	jsonStr := `{"jsonrpc":"2.0","method":"net_peerCount","params":[],"id":74}`
	jsonBytes := []byte(jsonStr)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBytes))

	if err != nil {
		fmt.Println("[error] RPC post request")
		return "", errors.New("[error] RPC post request")
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("[error] RPC post request")
		return "", errors.New("[error] RPC post request")
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		fmt.Println("[error] IO read all error")
		return "", errors.New("[error] IO read all error")
	} else {
		fmt.Println("[ body - result ]")
		fmt.Println(body)
	}

	return string(body), nil
}

// returns the latest block number
func getLatestBlock(client http.Client, url string) (string, error) {
	// panic("get peer count function")

	jsonStr := `{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}`
	jsonBytes := []byte(jsonStr)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBytes))

	if err != nil {
		fmt.Println("[error] RPC post request")
		return "", errors.New("[error] RPC post request")
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("[error] RPC post request")
		return "", errors.New("[error] RPC post request")
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		fmt.Println("[error] IO read all error")
		return "", errors.New("[error] IO read all error")
	}

	return string(body), nil
}
