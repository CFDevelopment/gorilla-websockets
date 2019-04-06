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

type Block struct {
	Number           string        `json:"number"`
	Hash             string        `json:"hash"`
	ParentHash       string        `json:"parentHash"`
	Nonce            string        `json:"nonce"`
	Sha3Uncles       string        `json:"sha3Uncles"`
	LogsBloom        string        `json:"logsBloom"`
	TransactionsRoot string        `json:"transactionsRoot"`
	StateRoot        string        `json:"stateRoot"`
	Miner            string        `json:"miner"`
	Difficulty       string        `json:"difficulty"`
	TotalDifficulty  string        `json:"totalDifficulty"`
	ExtraData        string        `json:"extraData"`
	Size             string        `json:"size"`
	GasLimit         string        `json:"gasLimit"`
	GasUsed          string        `json:"gasUsed"`
	Timestamp        string        `json:"timestamp"`
	Transactions     []interface{} `json:"transactions"`
	Uncles           []string      `json:"uncles"`
}

// "result": {
//     "number": "0x1b4", // 436
//     "hash": "0xe670ec64341771606e55d6b4ca35a1a6b75ee3d5145a99d05921026d1527331",
//     "parentHash": "0x9646252be9520f6e71339a8df9c55e4d7619deeb018d2a3f2d21fc165dde5eb5",
//     "nonce": "0xe04d296d2460cfb8472af2c5fd05b5a214109c25688d3704aed5484f9a7792f2",
//     "sha3Uncles": "0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347",
//     "logsBloom": "0xe670ec64341771606e55d6b4ca35a1a6b75ee3d5145a99d05921026d1527331",
//     "transactionsRoot": "0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421",
//     "stateRoot": "0xd5855eb08b3387c0af375e9cdb6acfc05eb8f519e419b874b6ff2ffda7ed1dff",
//     "miner": "0x4e65fda2159562a496f9f3522f89122a3088497a",
//     "difficulty": "0x027f07", // 163591
//     "totalDifficulty":  "0x027f07", // 163591
//     "extraData": "0x0000000000000000000000000000000000000000000000000000000000000000",
//     "size":  "0x027f07", // 163591
//     "gasLimit": "0x9f759", // 653145
//     "gasUsed": "0x9f759", // 653145
//     "timestamp": "0x54e34e8e" // 1424182926
//     "transactions": [{...},{ ... }]
//     "uncles": ["0x1606e5...", "0xd5145a9..."]
//   }

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
		setupOption = "gui"
	}
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
