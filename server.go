package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func appendToFile(filename string, data []byte) {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	if _, err := file.Write(data); err != nil {
		log.Fatal(err)
	}
}

func echoHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	data := append(body, '\n')

	fmt.Println(string(data))

	appendToFile("output.txt", data)

	fmt.Fprintf(w, "%s", data)
}

func main() {
	http.HandleFunc("/", echoHandler)

	fmt.Println("Starting server at port 8000")
	if err := http.ListenAndServe(":8000", nil); err != nil {
		fmt.Println(err)
	}
}
