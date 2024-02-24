package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
)

var (
	file *os.File
	err  error
	mu   sync.Mutex
)

func appendToFile(filename string, data []byte) {
	mu.Lock()
	defer mu.Unlock()

	if file == nil {
		file, err = os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}
	}

	writer := bufio.NewWriter(file)
	if _, err := writer.Write(data); err != nil {
		log.Fatal(err)
	}
	writer.Flush()
}

func echoHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	s := string(body)
	parts := strings.SplitN(s, ",", 2)

	if len(parts) < 2 {
		fmt.Println("Invalid data format")
		return
	}

	num, err := strconv.Atoi(parts[0])
	if err != nil {
		fmt.Println("Invalid number format:", parts[0])
		return
	}

	log := parts[1]

	go func() {
		for i := 0; i < num; i++ {
			fmt.Println(log)
			appendToFile("output.txt", []byte(log+"\n"))
		}
	}()

	// fmt.Fprintf(w, "%s", data)
}

func main() {
	http.HandleFunc("/", echoHandler)

	fmt.Println("Starting server at port 8000")
	if err := http.ListenAndServe(":8000", nil); err != nil {
		fmt.Println(err)
	}
}

func cleanup() {
	if file != nil {
		file.Close()
	}
}

func init() {
	defer cleanup()
}
