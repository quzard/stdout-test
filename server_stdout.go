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
	defaultMaxSize     = 10240 // MB
	defaultMaxBackups  = 10
	defaultMaxAge      = 1     // Days
	defaultWriteRateMB = 500   // MB/s
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
	var processWg sync.WaitGroup
	quit := make(chan struct{})
	if minute != -1 {
		timer := time.NewTimer(time.Duration(minute) * time.Minute)
		go func() {
			<-timer.C
			close(quit)
		}()
	}

	var stdoutLogCount int64 = 0
	var fileLogCount int64 = 0

	logChannel1 := make(chan string, 10000*thread)
	processWg.Add(1)
	go func() {
		defer processWg.Done()
		shouldPrint := os.Getenv("SHOULD_PRINT")
		cnt := 0
		for log := range logChannel1 {
			r := rateLimiter.ReserveN(time.Now(), len(log))
			if !r.OK() {
				continue
			}
			time.Sleep(r.Delay())
			if shouldPrint == "on" {
				cnt++
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
			if shouldAppendToFile == "on" {
				cnt++
				appendToFile("output.txt", []byte(fmt.Sprintf("cnt:%d,%s\n", cnt, log)))
				atomic.AddInt64(&fileLogCount, 1)
			}
		}
	}()
	printDirectly := os.Getenv("PRINT_DIRECTLY")
	for j := 0; j < thread; j++ {

		wg.Add(1)
		go func(threadNo int) {
			defer wg.Done()
			var cnt int64 = 0
			for {
				select {
				case <-quit:
					return
				default:
					if printDirectly == "on" {
						cnt++
						fmt.Printf("cnt:%d thread:%d,log:%s\n", cnt, threadNo, log)
					} else {
						logChannel1 <- fmt.Sprintf("thread:%d,log:%s", threadNo, log)
						logChannel2 <- fmt.Sprintf("thread:%d,log:%s", threadNo, log)
					}
				}
			}
		}(j)
	}
	wg.Wait()

	close(logChannel1)
	close(logChannel2)

	processWg.Wait()
	atomic.AddInt64(&stdoutLogCount, 2)
	fmt.Println("cnt:Total stdoutLogCount:", atomic.LoadInt64(&stdoutLogCount))
	fmt.Println("cnt:Total fileLogCount:", atomic.LoadInt64(&fileLogCount))
}

func main() {
	minute := getEnvAsInt("MINUTE", 10)
	thread := getEnvAsInt("THREAD", 5)
	log := strings.ReplaceAll(os.Getenv("LOG"), "\\n", "\n")

	writeRateMB := getEnvAsInt("LOG_WRITE_RATE_MB", defaultWriteRateMB)
	rateLimiter = rate.NewLimiter(rate.Limit(writeRateMB*1024*1024), defaultWriteBurst)

	echoHandler(thread, log, minute)
}
