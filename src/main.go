package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
)

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
	client, endpoint, rpcCallStructs := initNode()     // all basic setup is done here
	sequenceRpcCalls(client, endpoint, rpcCallStructs) // the non dynamic way to do this -> later refactor
}

func initNode() (http.Client, string, []Call) {
	client, endpoint := setupClient()
	rpcCallStructs := setupRpcCalls()
	return client, endpoint, rpcCallStructs
}

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

func setupClientEndpoints(setupOption string) string {
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

// rpc calls => structs
func setupRpcCalls() []Call {
	rpcStrings := []string{
		`{"jsonrpc":"2.0","method":"net_peerCount","params":[],"id":"74"}`,
		`{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":"1"}`,
		`{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["eth_blockNumber", "true"],"id":"1"}`,
	}
	rpcStructs := initializeCallStructs(rpcStrings)
	return rpcStructs
}

func initializeCallStructs(rpcStrings []string) []Call {
	callSlice := []Call{}

	for i := 0; i < len(rpcStrings); i++ {
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

// sequence the calls manually and execute all calls.
// TODO: needs to have a way to handle errors and dependent calls/input/output
func sequenceRpcCalls(client http.Client, url string, calls []Call) {
	peerCount, err := executeRpcCall(client, url, calls[0])

	if err != nil {
		fmt.Println("[error] peer count")
	}

	fmt.Println("peercount", peerCount["result"])

	blockNumber, err := executeRpcCall(client, url, calls[1])

	fmt.Println("block num", blockNumber["result"])

	if err != nil {
		fmt.Println("[error] block number")
	}

	str := fmt.Sprint(blockNumber["result"])
	newArr := []string{str, "true"}
	calls[2].Params = newArr // this will all be dynamic later -> future refactor

	block, err := executeRpcCall(client, url, calls[2])

	if err != nil {
		fmt.Println("[error] block")
	}

	fmt.Println("blockNumber", block["result"])
}

func executeRpcCall(client http.Client, url string, call Call) (Resp, error) {
	paramString := parseArrayToString(call.Params)
	jsonStr := `{"jsonrpc":"` + call.Jsonrpc + `","method":"` + call.Method + `","params":[` + paramString + `],"id":` + call.Id + `}`
	jsonBytes := []byte(jsonStr)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBytes))

	if err != nil {
		fmt.Println("[error] RPC post request")
		return nil, errors.New("[error] RPC post request")
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("[error] RPC post request")
		return nil, errors.New("[error] RPC post request")
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		fmt.Println("[error] IO read all error")
		return nil, errors.New("[error] IO read all error")
	} else {
		fmt.Println("[ body - result ]")
		fmt.Println(string(body))
	}

	r := new(Resp)
	err = json.Unmarshal([]byte(body), r)

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("response", r)

	return *r, nil
}

// parseArrayToString returns string with proper quotes. Returns empty string if there are no params.
func parseArrayToString(params []string) string {
	str := ""

	if len(params) == 0 {
		return ""
	}

	for i := 0; i < len(params); i++ {
		if strings.EqualFold("true", params[i]) != true && strings.EqualFold("false", params[i]) != true {
			str = str + strconv.Quote(params[i])
		} else {
			str = str + params[i]
		}
		if i < len(params)-1 {
			str = str + ", "
		}
	}
	return str
}
