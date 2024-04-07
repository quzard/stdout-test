package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	logger *lumberjack.Logger
	mu     sync.Mutex
)

func appendToFile(filename string, data []byte) {
	mu.Lock()
	defer mu.Unlock()

	if logger == nil {
		logDir := os.Getenv("LOG_DIR")
		if logDir == "" {
			logDir = "/logs"
		}
		if _, err := os.Stat(logDir); os.IsNotExist(err) {
			os.Mkdir(logDir, 0755)
		}

		maxSizeStr := os.Getenv("LOG_MAX_SIZE")
		maxSize, err := strconv.Atoi(maxSizeStr)
		if err != nil {
			maxSize = 10 // default value
		}

		maxBackupsStr := os.Getenv("LOG_MAX_BACKUPS")
		maxBackups, err := strconv.Atoi(maxBackupsStr)
		if err != nil {
			maxBackups = 5 // default value
		}

		maxAgeStr := os.Getenv("LOG_MAX_AGE")
		maxAge, err := strconv.Atoi(maxAgeStr)
		if err != nil {
			maxAge = 28 // default value
		}

		compressStr := os.Getenv("LOG_COMPRESS")
		compress := compressStr == "true" || compressStr == "1"

		logger = &lumberjack.Logger{
			Filename:   filepath.Join(logDir, filename),
			MaxSize:    maxSize, // megabytes
			MaxBackups: maxBackups,
			MaxAge:     maxAge,   //days
			Compress:   compress, // disabled by default
		}
	}

	logger.Write(data)
}

func echoHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	s := string(body)
	parts := strings.SplitN(s, ",", 3)

	if len(parts) < 3 {
		fmt.Println("Invalid data format")
		return
	}

	num, err := strconv.Atoi(parts[0])
	if err != nil {
		fmt.Println("Invalid number format:", parts[0])
		return
	}

	thread, err := strconv.Atoi(parts[1])
	if err != nil {
		fmt.Println("Invalid number format:", parts[0])
		return
	}

	log := parts[2]

	go func() {
		var wg sync.WaitGroup
		for j := 0; j < thread-1; j++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < num/thread; i++ {
					fmt.Println(log)
				}
			}()
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < num/thread; i++ {
					appendToFile("output.txt", []byte(log+"\n"))
				}
			}()
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < num-num/thread*(thread-1); i++ {
				fmt.Println(log)
			}
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < num-num/thread*(thread-1); i++ {
				appendToFile("output.txt", []byte(log+"\n"))
			}
		}()
		wg.Wait()
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
	if logger != nil {
		logger.Close()
	}
}

func init() {
	defer cleanup()
}
