package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/time/rate"
	"gopkg.in/natefinch/lumberjack.v2"
)

var once sync.Once

var (
	logger        *lumberjack.Logger
	loggerMu      sync.Mutex
	rateLimiter   *rate.Limiter
	rateLimiterMu sync.Mutex
)

const (
	defaultLogDir      = "/logs"
	defaultMaxSize     = 500 // MB
	defaultMaxBackups  = 20
	defaultMaxAge      = 28    // Days
	defaultWriteRateMB = 100   // MB/s
	defaultWriteBurst  = 10000 // Burst size in bytes
)

func getEnvAsInt(name string, defaultValue int) int {
	valueStr := os.Getenv(name)
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func initLogger(filename string) {
	if logger == nil {
		logDir := os.Getenv("LOG_DIR")
		if logDir == "" {
			logDir = defaultLogDir
		}
		if _, err := os.Stat(logDir); os.IsNotExist(err) {
			os.Mkdir(logDir, 0755)
		}
		maxSize := getEnvAsInt("LOG_MAX_SIZE", defaultMaxSize)
		maxBackups := getEnvAsInt("LOG_MAX_BACKUPS", defaultMaxBackups)
		maxAge := getEnvAsInt("LOG_MAX_AGE", defaultMaxAge)

		logger = &lumberjack.Logger{
			Filename:   filepath.Join(logDir, filename),
			MaxSize:    maxSize, // megabytes
			MaxBackups: maxBackups,
			MaxAge:     maxAge, //days
			Compress:   false,  // disabled by default
		}
	}
}

func appendToFile(filename string, data []byte) {
	if logger == nil {
		once.Do(func() {
			initLogger(filename)
		})
	}

	rateLimiterMu.Lock()
	r := rateLimiter.ReserveN(time.Now(), len(data))
	rateLimiterMu.Unlock()

	if !r.OK() {
		return // Not allowed to write due to rate limiting
	}
	time.Sleep(r.Delay())

	loggerMu.Lock()
	logger.Write(data)
	loggerMu.Unlock()
}

func echoHandler(log string, minute int) {
	var wg sync.WaitGroup
	var processWg sync.WaitGroup
	timer := time.NewTimer(time.Duration(minute) * time.Minute)
	quit := make(chan struct{})

	go func() {
		<-timer.C
		close(quit)
	}()
	var stdoutLogCount int64 = 0
	var fileLogCount int64 = 0

	logChannel1 := make(chan string, 10000)
	processWg.Add(1)
	go func() {
		defer processWg.Done()
		shouldPrint := os.Getenv("SHOULD_PRINT")
		cnt := 0
		for log := range logChannel1 {
			cnt++
			if shouldPrint == "on" {
				fmt.Println("cnt:", cnt, ",", log)
				atomic.AddInt64(&stdoutLogCount, 1)
			}
		}
	}()

	logChannel2 := make(chan string, 10000)
	processWg.Add(1)
	go func() {
		defer processWg.Done()
		shouldAppendToFile := os.Getenv("SHOULD_APPEND_TO_FILE")

		cnt := 0
		for log := range logChannel2 {
			cnt++
			if shouldAppendToFile == "on" {
				appendToFile("output.txt", []byte(fmt.Sprintf("cnt:%d,%s\n", cnt, log)))
				atomic.AddInt64(&fileLogCount, 1)
			}
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-quit:
				return
			default:
				logChannel1 <- fmt.Sprintf("log:%s", log)
				logChannel2 <- fmt.Sprintf("log:%s", log)
			}
		}
	}()
	wg.Wait()

	close(logChannel1)
	close(logChannel2)

	processWg.Wait()
	atomic.AddInt64(&stdoutLogCount, 2)
	fmt.Println("Total stdoutLogCount:", atomic.LoadInt64(&stdoutLogCount))
	fmt.Println("Total fileLogCount:", atomic.LoadInt64(&fileLogCount))
}

func main() {
	minute := getEnvAsInt("MINUTE", 10)
	log := strings.ReplaceAll(os.Getenv("LOG"), "\\n", "\n")

	writeRateMB := getEnvAsInt("LOG_WRITE_RATE_MB", defaultWriteRateMB)
	rateLimiter = rate.NewLimiter(rate.Limit(writeRateMB*1024*1024), defaultWriteBurst)

	echoHandler(log, minute)
}
