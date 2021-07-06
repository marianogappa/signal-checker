package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/marianogappa/signal-checker/signalchecker"
	"github.com/marianogappa/signal-checker/types"
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
	var input types.SignalCheckInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	output, _ := signalchecker.CheckSignal(input)
	w.WriteHeader(output.HttpStatus)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(output)
}

func main() {
	inputStr := os.Args[1]
	if inputStr == "serve" {
		serve(os.Args)
	}

	input := types.SignalCheckInput{}
	if err := json.Unmarshal([]byte(inputStr), &input); err != nil {
		log.Fatal(err)
	}

	output, _ := signalchecker.CheckSignal(input)
	byts, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(byts))
}
