package main

import (
	"io/ioutil"
	"bytes"
	"encoding/json"
	"log"
	"strings"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"strconv"
	"path/filepath"

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

type TypedRpc []struct {
	NetPeerCount struct {
		Jsonrpc string `json:"jsonrpc"`
		Method  struct {
			Function string `json:"function"`
		} `json:"method"`
		Params []interface{} `json:"params"`
		ID     string        `json:"id"`
	} `json:"net_peerCount,omitempty"`
	EthBlockNumber struct {
		Jsonrpc string `json:"jsonrpc"`
		Method  struct {
			Function string `json:"function"`
		} `json:"method"`
		Params []interface{} `json:"params"`
		ID     string        `json:"id"`
	} `json:"eth_blockNumber,omitempty"`
	EthGetBlockByNumber struct {
		Jsonrpc string `json:"jsonrpc"`
		Method  string `json:"method"`
		Params  []struct {
			Type         string `json:"type"`
			Description  string `json:"description"`
			InputStreams struct {
				Flags     []string `json:"flags"`
				Functions []struct {
					Method string `json:"method"`
					Return struct {
						Type           string `json:"type"`
						ReturnFunction string `json:"returnFunction"`
					} `json:"return"`
				} `json:"functions"`
			} `json:"inputStreams,omitempty"`
		} `json:"params"`
		ID string `json:"id"`
	} `json:"eth_getBlockByNumber,omitempty"`
}

// 													updated notes... 												  //
// - rpcConfig.json will contain the typed rpc call and dependency objects in json format.
// - The user can dynamically subscribe to any subset of rpc api calls via command line or with our custom web
// 	 interface.
// - Use the --cmd flag on launch to default node setup localhost:8454
// - our custom rpc type config allows for a dynamic and resource efficient query system.
func main() {
	typedRpcConfig := unmarshalConfigFile() // the above struct with user defined rpc methods
	// client, endpoint, rpcDependencies, rpcCallStructs := initNode() // all basic setup is done here
	// rpcCallPulse(client, endpoint, rpcDependencies, rpcCallStructs) // call this to generate node stats
	// fmt.Println(rpcCallStructs)
}

func unmarshalConfigFile() *TypedRpc {
	path, _ := filepath.Abs("./src/config/rpcConfig.json")
	file, err := ioutil.ReadFile(path)

	if err != nil {
		fmt.Println("error", err)
	}

	tyRpc := new(TypedRpc)
	err = json.Unmarshal([]byte(file), tyRpc)

	if err != nil {
		fmt.Println("error parsing json file to typed rpc struct", err)
	}

	return tyRpc
}

func initNode() (http.Client, string, map[string][]string, []Call){
	client, endpoint := setupClient()
	rpcCallStructs := setupRpcCalls()
	rpcDependencies := defineRpcDependencies()
	return client, endpoint, rpcDependencies, rpcCallStructs
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
		`{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["eth_blockNumber", "true"],"id":1}`,
		//`{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["eth_blockNumber", true],"id":1}`, //first inp = res (blocknum)
	}
	// parse these strings into structs
	rpcStructs := initializeCallStructs(rpcStrings)
	return rpcStructs
}

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

func rpcCallPulse(client http.Client, url string, rpcDependencies map[string][]string, calls []Call) {
	// loop through calls & execute each rpc method
	// creates method => response mapping to lookup results if needed by other functions as params
	respMapping := make(map[string]Resp)
	for i := 0; i < len(calls); i++ {
		// should add a middleware to check for needed inputs (ex -> block number => get block by number(block num))
		// if params length = 0 no check is needed
		// if there are params, check if the function can be executed independently or needs input from other func input
		if len(calls[i].Params) == 0 {
			fmt.Println("0 params")
			resp, err := executeRpcCall(client, url, calls[i])
			// will check it with our mapping, if it has not been done yet, add it to queue & then execute at end
			// have a basic alg that dynamically executes all methods
			// the end result will be registering all the rpc methods you want, adding their predecessors and execute
			if err != nil {
				fmt.Println(err)
			}

			respMapping[calls[i].Method] = resp
		} else {
			fmt.Println("more than 0 params")
			parseArrayToString(calls[i].Params)
			// checkRpcDependencies(rpcDependencies, respMapping, calls[i])
			// if the dep has not been executed => put in queue
		}
	}
	fmt.Println("respMapping", respMapping)
}

func defineRpcDependencies() map[string][]string {
	// eth_getBlockByNumber needs eth_blockNumber result
	// setup any future method param dependencies here
	var rpcDependencies = map[string][]string{}
	rpcDependencies["eth_getBlockByNumber"] = []string{"eth_blockNumber"}
	return rpcDependencies
}

/**
	checkRpcDependencies is used to determine if a given rpc method being executed in any given go routine needs
	another method's output as input for the current param array. The dependencies mapping can lookup a method and if
	they require other methods to be invoked. The respMapping maps a method to a rpc response struct, which will be used
	to indicate what rpc methods have been invoked.
 */
func checkRpcDependencies(dependencies map[string][]string, respMapping map[string]Resp, call Call) {
	// want to actually change the given call object to alter the params in the struct
}

func executeRpcCall(client http.Client, url string, call Call) (Resp, error) {
	paramString := parseArrayToString(call.Params)
	fmt.Println("param string [executeRpcCall]", paramString)
	jsonStr := `{"jsonrpc":"`+ call.Jsonrpc +`","method":"`+ call.Method +`","params":[`+ paramString +`],"id":`+ call.Id +`}`
	fmt.Println("jsonString:", jsonStr)
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

/**
	parseArrayToString will take a json rpc string and format it such that it will unqoute bool values and
	add quotes to non
	TODO: refactor for functions
 */
func parseArrayToString (params []string) string {
	formattedInput := make([]string, len(params))
	if len(params) == 0 {
		return ""
	}
	for i := 0; i < len(params); i++ {
		if strings.EqualFold("true", params[i]) != true && strings.EqualFold("false", params[i]) != true {
			formattedInput = append(formattedInput, strconv.Quote(params[i]))
		} else {
			formattedInput = append(formattedInput, params[i])
		}
	}
 	justString := strings.Join(formattedInput,",")
	return justString
}


