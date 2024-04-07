package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
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

func echoHandler(thread int, log string, minute int) {
	var wg sync.WaitGroup
	// Create a timer that will send a signal after 'minute' minutes
	timer := time.NewTimer(time.Duration(minute) * time.Minute)
	// Create a quit channel that will be closed when the timer fires
	quit := make(chan struct{})

	go func() {
		<-timer.C
		close(quit)
	}()

	logChannel1 := make(chan string, 10000*thread) // 创建一个带缓冲的 channel
	go func() {
		for log := range logChannel1 {
			fmt.Println(log)
		}
	}()
	logChannel2 := make(chan string, 10000*thread) // 创建一个带缓冲的 channel
	go func() {
		for log := range logChannel2 {
			appendToFile("output.txt", []byte(log+"\n"))
		}
	}()
	for j := 0; j < thread; j++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-quit:
					return
				default:
					logChannel1 <- log
					logChannel2 <- log
				}
			}
		}()
	}
	wg.Wait()
}

func main() {
	minute := getEnvAsInt("MINUTE", 10)
	thread := getEnvAsInt("THREAD", 5)
	log := strings.ReplaceAll(os.Getenv("LOG"), "\\n", "\n")

	writeRateMB := getEnvAsInt("LOG_WRITE_RATE_MB", defaultWriteRateMB)
	rateLimiter = rate.NewLimiter(rate.Limit(writeRateMB*1024*1024), defaultWriteBurst)

	echoHandler(thread, log, minute)
}
