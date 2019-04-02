package main

import (
	"bytes"
	"encoding/json"
	"log"
	"strings"

	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"

	// "github.com/ethereum/go-ethereum/ethclient"
	"github.com/gorilla/websocket"
)

// upgrade original http request to ws
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Resp map[string]interface{}

type Call struct {
	Jsonrpc string   `json:"jsonrpc"`
	Method  string   `json:"method"`
	Params  []string `json:"params"`
	Id      string   `json:"id"`
}

func main() {
	client, endpoint := setupClient()
	rpcCallStructs := setupRpcCalls()
	rpcCallPulse(client, endpoint, rpcCallStructs)
	fmt.Println(rpcCallStructs)
}

/*
	setupClient will boot either the gui or command line pre config option based on the flag that is passed in to run
	this application. --cmd || --gui respectively. If --cmd flag is passed, the user will be prompted to enter a series
	of http endpoints with rpc enabled clients. This will allow a user to register one or many nodes with our network
	stats client.
 */
func setupClient() (http.Client, string) {
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
	endpoint := setupClientEndpoints(setupOption)

	return http.Client{}, endpoint
}

// additional option to listen in on multiple endpoint for various clients.
// this will allow other developers to setup their own cluster of nodes with custom UI
func setupClientEndpoints (setupOption string) string {
	var endpoint string
	switch setupOption {
	case "gui":
		fmt.Println("Setting up GUI")
	case "cmd":
		fmt.Println("Setting up command line")
		endpoint = parseCmdLineEndpoint()
		if endpoint == "" {
			fmt.Println("[error in setupClientEndpoint] : invalid endpoint response")
		}
		fmt.Println("endpoints in setupClient func:", endpoint)
	default:
		panic("unrecognized escape character")
	}

	return endpoint
}

/**
	parseCmdLineEndpoints will set the endpoint for the json rpc connection.
	TODO: make continuous prompt/currently hardcoded to accept only one value...
 */
func parseCmdLineEndpoint() string {
	var response string
	fmt.Println("---                                  PROMPT                                      ---")
	fmt.Println("---    Enter 'remote endpoint' for node setup                                    ---")
	fmt.Println("---    Enter 'local' for default setup for http://localhost:8545                 ---")
	_, err := fmt.Scanln(&response)
	if err != nil {
		log.Fatal(err)
		return ""
	} else if response == "local" {
		return "http://localhost:8545"
	}
	return response
}


func setupRpcCalls() []Call {
	// list of rpc strings we will be using
	rpcStrings := []string{
		`{"jsonrpc":"2.0","method":"net_peerCount","params":[],"id":"74"}`,
		`{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":"1"}`,
	}
	// parse these strings into structs
	rpcStructs := initializeCallStructs(rpcStrings)
	return rpcStructs
}

// will be called once in setup
func initializeCallStructs (rpcStrings []string) []Call {
	callSlice := []Call{}
	for i := 0; i <len(rpcStrings); i++ {
		c := new(Call)
		err := json.Unmarshal([]byte(rpcStrings[i]), c)

		if err != nil {
			fmt.Println(err)
		}
		callSlice = append(callSlice, *c)
		fmt.Println(c)
	}
	return callSlice
}

/**
	Will iterate through call structs & execute their rpc methods
	TODO: Refactor to go routines and channels
 */
func rpcCallPulse(client http.Client, url string, calls []Call) {
	// loop through calls & execute each rpc method
	for i := 0; i < len(calls); i++ {
		executeRpcCall(client, url, calls[i])
	}

}

func executeRpcCall(client http.Client, url string, call Call) (string, error) {
	paramString := parseArrayToString(call.Params)
	fmt.Println(call.Method, call.Jsonrpc, call.Params)
	jsonStr := `{"jsonrpc":"`+ call.Jsonrpc +`","method":"`+ call.Method +`","params":[`+ paramString +`],"id":`+ call.Id +`}`
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
		fmt.Println(string(body))
	}

	return string(body), nil
}

/**
	parseArrayToString will take a param array and return string delimited by comma for json rpc param string
 */
func parseArrayToString (params []string) string {
	if len(params) == 0 {
		return ""
	}
	justString := strings.Join(params,",")
	return justString
}


