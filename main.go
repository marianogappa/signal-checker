// The main package is only used as a proxy to the signalchecker package to be used as a cli application or as a server.
//
// If you're looking for documentation on how to import and use signalchecker in your program, please review the
// signalchecker package.
//
// If you're looking for how to use the cli & server versions, please review the README.
//
// If you're looking for the spec for input & output, please review the types package.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/marianogappa/signal-checker/common"
	"github.com/marianogappa/signal-checker/signalchecker"
)

func serve(args []string) {
	port := 8080
	if len(args) >= 3 {
		port, _ = strconv.Atoi(args[2])
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/check", serveCheck)

	if err := http.ListenAndServe(fmt.Sprintf(":%v", port), mux); err != nil {
		log.Fatal(err)
	}
}

func serveCheck(w http.ResponseWriter, r *http.Request) {
	var input common.SignalCheckInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	output, _ := signalchecker.NewSignalChecker(input).Check()
	w.WriteHeader(output.HttpStatus)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(output)
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	inputStr := os.Args[1]
	if inputStr == "serve" {
		serve(os.Args)
	}

	input := common.SignalCheckInput{}
	if err := json.Unmarshal([]byte(inputStr), &input); err != nil {
		log.Fatal(err)
	}

	output, _ := signalchecker.NewSignalChecker(input).Check()
	byts, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(byts))
}
