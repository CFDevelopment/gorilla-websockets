package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
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
type Resp struct {
	jsonrpc string
	id      int
	result  string
}

// used to json format an RPC call
type Call struct {
	Jsonrpc string   `json:"jsonrpc"`
	Method  string   `json:"method"`
	Params  []string `json:"params"`
	Id      string   `json:"id"`
}

func main() {
	var url = "http://localhost:8545"
	client := http.Client{}

	// refactor out to structs and generic rpc method
	peerCount, err := getPeerCount(client, url)

	if err != nil {
		fmt.Println("[error] peer count")
	} else {
		fmt.Println(peerCount)
	}

	lastBlock, err := getLatestBlock(client, url)

	if err != nil {
		fmt.Println("[error] last block")
	} else {
		fmt.Println(lastBlock)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	http.HandleFunc("/v1/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}
		go func(conn *websocket.Conn) {
			for {
				mType, msg, err := conn.ReadMessage()
				if err != nil {
					log.Println(err)
					return
				}

				conn.WriteMessage(mType, msg)
			}
		}(conn)
	})

	http.HandleFunc("/v2/ws", func(w http.ResponseWriter, r *http.Request) {
		var conn, err = upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}
		go func(conn *websocket.Conn) {
			for {
				_, msg, _ := conn.ReadMessage()
				println(string(msg))
			}
		}(conn)
	})

	// every json every 5 minutes to channel
	// http.HandleFunc("/v3/ws", func(w http.ResponseWriter, r *http.Request) {
	// 	var conn, err = upgrader.Upgrade(w, r, nil)
	// 	if err != nil {
	// 		log.Println(err)
	// 		return
	// 	}
	// 	go func(conn *websocket.Conn) {
	// 		ch := time.Tick(5 * time.Second)

	// 		// for range ch {
	// 		// 	conn.WriteJSON(myStruct{
	// 		// 		Username:  "cristobal",
	// 		// 		FirstName: "chris",
	// 		// 		LastName:  "Ffffff",
	// 		// 	})
	// 		// }
	// 	}(conn)
	// })

	http.ListenAndServe(":3000", nil)
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
