package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/time/rate"
)

var (
	rateLimiter        *rate.Limiter
	defaultWriteRateMB = 300   // MB/s
	defaultWriteBurst  = 10000 // Burst size in bytes
)

func getEnvAsInt(name string, defaultValue int) int {
	valueStr := os.Getenv(name)
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
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
	logChannel := make(chan string, 10000*thread)
	processWg.Add(1)
	go func() {
		defer processWg.Done()
		shouldPrint := os.Getenv("SHOULD_PRINT")
		cnt := 0
		for log := range logChannel {
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

	for j := 0; j < thread; j++ {
		wg.Add(1)
		go func(threadNo int) {
			defer wg.Done()
			for {
				select {
				case <-quit:
					return
				default:
					logChannel <- fmt.Sprintf("thread:%d,log:%s", threadNo, log)
				}
			}
		}(j)
	}
	wg.Wait()

	close(logChannel)

	processWg.Wait()
	atomic.AddInt64(&stdoutLogCount, 1)
	fmt.Println("cnt:Total stdoutLogCount:", atomic.LoadInt64(&stdoutLogCount))
}

func main() {
	minute := getEnvAsInt("MINUTE", 10)
	thread := getEnvAsInt("THREAD", 5)
	log := strings.ReplaceAll(os.Getenv("LOG"), "\\n", "\n")

	writeRateMB := getEnvAsInt("LOG_WRITE_RATE_MB", defaultWriteRateMB)
	rateLimiter = rate.NewLimiter(rate.Limit(writeRateMB*1024*1024), defaultWriteBurst)

	echoHandler(thread, log, minute)
}
