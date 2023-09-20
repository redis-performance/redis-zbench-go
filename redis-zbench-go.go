package main

import (
	"flag"
	"fmt"
	hdrhistogram "github.com/HdrHistogram/hdrhistogram-go"
	"github.com/mediocregopher/radix/v3"
	"golang.org/x/time/rate"
	"log"
	"math"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"time"
)

var totalCommands uint64
var totalAddedElements uint64
var totalErrors uint64
var latencies *hdrhistogram.Histogram
var replySizes []uint64

const Inf = rate.Limit(math.MaxFloat64)
const charset = "abcdefghijklmnopqrstuvwxyz"

func stringWithCharset(length int, charset string, r *rand.Rand) string {

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[r.Intn(len(charset))]
	}
	return string(b)
}

func main() {
	host := flag.String("h", "127.0.0.1", "Server hostname.")
	port := flag.Int("p", 12000, "Server port.")
	rps := flag.Int64("rps", 0, "Max rps. If 0 no limit is applied and the DB is stressed up to maximum.")
	password := flag.String("a", "", "Password for Redis Auth.")
	seed := flag.Int64("random-seed", 12345, "random seed to be used.")
	clients := flag.Uint64("c", 50, "number of clients.")
	keyspacelen := flag.Uint64("r", 1000000, "keyspace length.")
	keyspacestart := flag.Uint64("r-start", 0, "keyspace start.")
	numberRequests := flag.Uint64("n", 10000000, "Total number of requests. Only used in case of -mode=query")
	debug := flag.Int("debug", 0, "Client debug level.")
	multi := flag.Bool("multi", false, "Run each command in multi-exec.")
	benchMode := flag.String("mode", "", "Bechmark mode. One of [load,query]. `load` will populate the db with sorted sets. `query` will run the zrangebylexscore command .")
	perKeyElmRangeStart := flag.Uint64("key-elements-min", 10, "Use zipfian random-sized items in the specified range (min-max).")
	perKeyElmRangeEnd := flag.Uint64("key-elements-max", 100, "Use zipfian random-sized items in the specified range (min-max).")
	perKeyElmDataSize := flag.Uint64("d", 10, "Data size of each sorted set element.")
	pipeline := flag.Uint64("pipeline", 1, "Redis pipeline value.")
	version := flag.Bool("v", false, "Output version and exit")
	printReplyHistogram := flag.Bool("print-histogram", false, "Print reply histogram")
	clusterMode := flag.Bool("oss-cluster", false, "Enable OSS cluster mode.")
	query := flag.String("query", "zrangebyscore", "Query type.")

	flag.Parse()

	git_sha := toolGitSHA1()
	git_dirty_str := ""
	if toolGitDirty() {
		git_dirty_str = "-dirty"
	}
	if *version {
		fmt.Fprintf(os.Stdout, "redis-zbench-go (git_sha1:%s%s)\n", git_sha, git_dirty_str)
		os.Exit(0)
	}
	if *benchMode != "load" && *benchMode != "query" {
		log.Fatal("Please specify a valid -mode option. Either `load` or `query`")
	}
	isLoad := false
	if *benchMode == "load" {
		isLoad = true
	}
	var requestRate = Inf
	var requestBurst = 1
	useRateLimiter := false
	if *rps != 0 {
		requestRate = rate.Limit(*rps)
		requestBurst = int(*clients) * int(*pipeline)
		useRateLimiter = true
	}

	var rateLimiter = rate.NewLimiter(requestRate, requestBurst)
	totalCmds := *numberRequests
	if isLoad {
		totalCmds = *keyspacelen
	}
	samplesPerClient := totalCmds / *clients
	client_update_tick := 1
	latencies = hdrhistogram.New(1, 90000000, 3)
	opts := make([]radix.DialOpt, 0)
	replySizes = make([]uint64, *perKeyElmRangeEnd*100)
	if *password != "" {
		opts = append(opts, radix.DialAuthPass(*password))
	}
	connectionStr := fmt.Sprintf("%s:%d", *host, *port)
	stopChan := make(chan struct{})
	// a WaitGroup for the goroutines to tell us they've stopped
	wg := sync.WaitGroup{}
	keyspaceend := uint64(*keyspacestart) + uint64(*keyspacelen)
	fmt.Printf("Using redis-zbench-go (git_sha1:%s%s)\n", git_sha, git_dirty_str)
	fmt.Printf("Total clients: %d. Commands per client: %d Total commands: %d\n", *clients, samplesPerClient, totalCmds)
	fmt.Printf("Using random seed: %d\n", *seed)
	if isLoad {
		fmt.Printf("Each ZSET contains between %d and %d elements.\n", *perKeyElmRangeStart, *perKeyElmRangeEnd)
		fmt.Printf("Keyspace range: %d keys. [%d ; %d]\n", *keyspacelen, uint64(*keyspacestart), keyspaceend)
	}
	var cluster *radix.Cluster
	if *clusterMode {
		cluster = getOSSClusterConn(connectionStr, opts, *clients)
	}
	var connectionPool *radix.Pool = getStandaloneConn(connectionStr, opts, *clients)
	for client_id := 1; uint64(client_id) <= *clients; client_id++ {
		wg.Add(1)
		keyspace_client_start := uint64(*keyspacestart) + (uint64(client_id-1) * samplesPerClient)
		keyspace_client_end := uint64(*keyspacestart) + (uint64(client_id) * samplesPerClient)
		if uint64(client_id) == *clients {
			keyspace_client_end = keyspaceend
		}
		if isLoad {
			if *clusterMode {
				go loadGoRoutime(cluster, keyspace_client_start, keyspace_client_end, samplesPerClient, *pipeline, *perKeyElmDataSize, *perKeyElmRangeStart, *perKeyElmRangeEnd, int(*debug), &wg, useRateLimiter, rateLimiter, *seed + int64(client_id))
			} else {
				go loadGoRoutime(connectionPool, keyspace_client_start, keyspace_client_end, samplesPerClient, *pipeline, *perKeyElmDataSize, *perKeyElmRangeStart, *perKeyElmRangeEnd, int(*debug), &wg, useRateLimiter, rateLimiter, *seed + int64(client_id))
			}
		} else {
			switch *query {
			case "zrange-byscore-rev":
				if *clusterMode {
					go queryGoRoutimeZrevrangeByScore(cluster, *multi, uint64(*keyspacelen), samplesPerClient, *pipeline, *perKeyElmDataSize, *perKeyElmRangeStart, *perKeyElmRangeEnd, int(*debug), &wg, useRateLimiter, rateLimiter, *seed + int64(client_id))
				} else {
					go queryGoRoutimeZrevrangeByScore(connectionPool, *multi, uint64(*keyspacelen), samplesPerClient, *pipeline, *perKeyElmDataSize, *perKeyElmRangeStart, *perKeyElmRangeEnd, int(*debug), &wg, useRateLimiter, rateLimiter, *seed + int64(client_id))
				}
			case "zrange-byscore":
				if *clusterMode {
					go queryGoRoutimeZrangeByScore(cluster, *multi, uint64(*keyspacelen), samplesPerClient, *pipeline, *perKeyElmDataSize, *perKeyElmRangeStart, *perKeyElmRangeEnd, int(*debug), &wg, useRateLimiter, rateLimiter, *seed + int64(client_id))
				} else {
					go queryGoRoutimeZrangeByScore(connectionPool, *multi, uint64(*keyspacelen), samplesPerClient, *pipeline, *perKeyElmDataSize, *perKeyElmRangeStart, *perKeyElmRangeEnd, int(*debug), &wg, useRateLimiter, rateLimiter, *seed + int64(client_id))
				}
			case "zrevrangebylex":
				if *clusterMode {
					go queryGoRoutimeZrangeByLex(cluster, *multi, uint64(*keyspacelen), samplesPerClient, *pipeline, *perKeyElmDataSize, *perKeyElmRangeStart, *perKeyElmRangeEnd, int(*debug), &wg, useRateLimiter, rateLimiter, *seed + int64(client_id))
				} else {
					go queryGoRoutimeZrangeByLex(connectionPool, *multi, uint64(*keyspacelen), samplesPerClient, *pipeline, *perKeyElmDataSize, *perKeyElmRangeStart, *perKeyElmRangeEnd, int(*debug), &wg, useRateLimiter, rateLimiter, *seed + int64(client_id))
				}
			}
		}
	}

	// listen for C-c
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	tick := time.NewTicker(time.Duration(client_update_tick) * time.Second)
	closed, _, duration, totalMessages, _ := updateCLI(tick, c, totalCmds)
	messageRate := float64(totalMessages) / float64(duration.Seconds())
	p50IngestionMs := float64(latencies.ValueAtQuantile(50.0)) / 1000.0
	p95IngestionMs := float64(latencies.ValueAtQuantile(95.0)) / 1000.0
	p99IngestionMs := float64(latencies.ValueAtQuantile(99.0)) / 1000.0

	fmt.Printf("\n")
	fmt.Printf("#################################################\n")
	fmt.Printf("Total Duration %.3f Seconds\n", duration.Seconds())
	fmt.Printf("Total Issued commands %d\n", totalMessages)
	fmt.Printf("Total Errors %d\n", totalErrors)
	fmt.Printf("Throughput summary: %.0f requests per second\n", messageRate)
	if isLoad {
		avgZcard := float64(totalAddedElements) / float64(totalCommands)
		fmt.Printf("Average zcard %.0f elements\n", avgZcard)
	}
	fmt.Printf("Latency summary (msec):\n")
	fmt.Printf("    %9s %9s %9s\n", "p50", "p95", "p99")
	fmt.Printf("    %9.3f %9.3f %9.3f\n", p50IngestionMs, p95IngestionMs, p99IngestionMs)
	if !isLoad && *printReplyHistogram {
		fmt.Printf("#################################################\n")
		fmt.Printf("Printing reply histogram\n")
		var total_replies uint64 = 0
		for reply_size, count := range replySizes {
			if count > 0 {
				percent := float32(count) / float32(totalMessages) * 100.0
				fmt.Printf("Size: %d\tCount: %d. (%% %.2f)\n", reply_size, count, percent)
			}
			total_replies += count
		}
		fmt.Printf("--------------------------------------------------\n")
		fmt.Printf("Total processed replies %d\n", total_replies)
	}

	if closed {
		return
	}

	// tell the goroutine to stop
	close(stopChan)
	// and wait for them both to reply back
	wg.Wait()
}

func queryGoRoutimeZrangeByLex(conn radix.Client, multi bool, keyspace_len uint64, samplesPerClient uint64, pipeline uint64, perKeyElmDataSize uint64, perKeyElmRangeStart uint64, perKeyElmRangeEnd uint64, debug int, w *sync.WaitGroup, useRateLimiter bool, rateLimiter *rate.Limiter, seed int64) {
	defer w.Done()

	r := rand.New(rand.NewSource(seed))

	var i uint64 = 0
	var multiIncr uint64 = 0
	var multiPad uint64 = 0
	if multi {
		multiIncr += 2
		multiPad += 1
	}
	cmds := make([]radix.CmdAction, pipeline+multiIncr)
	cmdReplies := make([][]string, pipeline)
	for i < samplesPerClient {
		key_n := r.Int63n(int64(keyspace_len))

		if useRateLimiter {
			r := rateLimiter.ReserveN(time.Now(), int(pipeline))
			time.Sleep(r.Delay())
		}
		var j uint64 = 0
		if multi {
			cmds[0] = radix.Cmd(nil, "MULTI")
			cmds[pipeline+multiPad] = radix.Cmd(&cmdReplies, "EXEC")
		}
		for j < pipeline {
			keyname := getBenchKeyName(uint64(key_n))
			cmdArgs := []string{keyname, fmt.Sprintf("[%c", charset[r.Intn(len(charset))]), "-"}
			cmds[j+multiPad] = radix.Cmd(nil, "ZREVRANGEBYLEX", cmdArgs...)
			j = j + 1
			key_n = key_n + crc16_num_slots
			if uint64(key_n) > keyspace_len {
				key_n = key_n % crc16_num_slots
			}
		}
		var err error
		startT := time.Now()
		err = conn.Do(radix.Pipeline(cmds...))
		endT := time.Now()
		if err != nil {
			log.Fatalf("Received an error with the following command(s): %v, error: %v", cmds, err)
		}
		duration := endT.Sub(startT)
		err = latencies.RecordValue(duration.Microseconds())
		if err != nil {
			log.Fatalf("Received an error while recording latencies: %v", err)
		}
		for _, reply := range cmdReplies {
			atomic.AddUint64(&replySizes[len(reply)], 1)
		}
		atomic.AddUint64(&totalCommands, uint64(pipeline))
		i = i + pipeline
	}
}

func queryGoRoutimeZrangeByScore(conn radix.Client, multi bool, keyspace_len uint64, samplesPerClient uint64, pipeline uint64, perKeyElmDataSize uint64, perKeyElmRangeStart uint64, perKeyElmRangeEnd uint64, debug int, w *sync.WaitGroup, useRateLimiter bool, rateLimiter *rate.Limiter, seed int64) {
	defer w.Done()

	r := rand.New(rand.NewSource(seed))

	var i uint64 = 0
	var multiIncr uint64 = 0
	var multiPad uint64 = 0
	if multi {
		multiIncr += 2
		multiPad += 1
	}
	cmds := make([]radix.CmdAction, pipeline+multiIncr)
	cmdReplies := make([][]string, pipeline)
	for i < samplesPerClient {
		key_n := r.Int63n(int64(keyspace_len))

		if useRateLimiter {
			r := rateLimiter.ReserveN(time.Now(), int(pipeline))
			time.Sleep(r.Delay())
		}
		var j uint64 = 0
		if multi {
			cmds[0] = radix.Cmd(nil, "MULTI")
			cmds[pipeline+multiPad] = radix.Cmd(&cmdReplies, "EXEC")
		}
		for j < pipeline {
			keyname := getBenchKeyName(uint64(key_n))
			cmdArgs := []string{keyname, "0", "1", "BYSCORE"}
			cmds[j+multiPad] = radix.Cmd(&cmdReplies[j], "ZRANGE", cmdArgs...)
			j = j + 1
			key_n = key_n + crc16_num_slots
			if uint64(key_n) > keyspace_len {
				key_n = key_n % crc16_num_slots
			}
		}
		var err error
		startT := time.Now()
		err = conn.Do(radix.Pipeline(cmds...))
		endT := time.Now()
		if err != nil {
			log.Fatalf("Received an error with the following command(s): %v, error: %v", cmds, err)
		}
		duration := endT.Sub(startT)
		err = latencies.RecordValue(duration.Microseconds())
		if err != nil {
			log.Fatalf("Received an error while recording latencies: %v", err)
		}
		for _, reply := range cmdReplies {
			atomic.AddUint64(&replySizes[len(reply)/2], 1)
		}
		atomic.AddUint64(&totalCommands, uint64(pipeline))
		i = i + pipeline
	}
}

func queryGoRoutimeZrevrangeByScore(conn radix.Client, multi bool, keyspace_len uint64, samplesPerClient uint64, pipeline uint64, perKeyElmDataSize uint64, perKeyElmRangeStart uint64, perKeyElmRangeEnd uint64, debug int, w *sync.WaitGroup, useRateLimiter bool, rateLimiter *rate.Limiter, seed int64) {
	defer w.Done()

	r := rand.New(rand.NewSource(seed))

	var i uint64 = 0
	var multiIncr uint64 = 0
	var multiPad uint64 = 0
	if multi {
		multiIncr += 2
		multiPad += 1
	}
	cmds := make([]radix.CmdAction, pipeline+multiIncr)
	cmdReplies := make([][]string, pipeline)
	for i < samplesPerClient {
		key_n := r.Int63n(int64(keyspace_len))

		if useRateLimiter {
			r := rateLimiter.ReserveN(time.Now(), int(pipeline))
			time.Sleep(r.Delay())
		}
		var j uint64 = 0
		if multi {
			cmds[0] = radix.Cmd(nil, "MULTI")
			cmds[pipeline+multiPad] = radix.Cmd(&cmdReplies, "EXEC")
		}
		for j < pipeline {
			keyname := getBenchKeyName(uint64(key_n))
			cmdArgs := []string{keyname, "1", "0"}
			cmds[j+multiPad] = radix.Cmd(&cmdReplies[j], "ZREVRANGEBYSCORE", cmdArgs...)
			j = j + 1
			key_n = key_n + crc16_num_slots
			if uint64(key_n) > keyspace_len {
				key_n = key_n % crc16_num_slots
			}
		}
		var err error
		startT := time.Now()
		err = conn.Do(radix.Pipeline(cmds...))
		endT := time.Now()
		if err != nil {
			log.Fatalf("Received an error with the following command(s): %v, error: %v", cmds, err)
		}
		duration := endT.Sub(startT)
		err = latencies.RecordValue(duration.Microseconds())
		if err != nil {
			log.Fatalf("Received an error while recording latencies: %v", err)
		}
		for _, reply := range cmdReplies {
			atomic.AddUint64(&replySizes[len(reply)/2], 1)
		}
		atomic.AddUint64(&totalCommands, uint64(pipeline))
		i = i + pipeline
	}
}

func loadGoRoutime(conn radix.Client, keyspace_client_start uint64, keyspace_client_end uint64, samplesPerClient uint64, pipeline uint64, perKeyElmDataSize uint64, perKeyElmRangeStart uint64, perKeyElmRangeEnd uint64, debug int, w *sync.WaitGroup, useRateLimiter bool, rateLimiter *rate.Limiter, seed int64) {
	defer w.Done()

	r := rand.New(rand.NewSource(seed))

	var i uint64 = 0
	var keypos uint64 = keyspace_client_start
	cmds := make([]radix.CmdAction, pipeline)
	for i < samplesPerClient {
		if useRateLimiter {
			r := rateLimiter.ReserveN(time.Now(), int(pipeline))
			time.Sleep(r.Delay())
		}
		var j uint64 = 0
		for ; j < pipeline; j++ {
			keyname := getBenchKeyName(keypos)
			cmdArgs := []string{keyname}
			nElements := r.Int63n(int64(perKeyElmRangeEnd-perKeyElmRangeStart)) + int64(perKeyElmRangeStart)
			var k int64 = 0
			for ; k < nElements; k++ {
				cmdArgs = append(cmdArgs, fmt.Sprintf("%f", r.Float32()), stringWithCharset(int(perKeyElmDataSize), charset, r))
			}
			atomic.AddUint64(&totalAddedElements, uint64(nElements))
			cmds[j] = radix.Cmd(nil, "ZADD", cmdArgs...)
			keypos++
		}
		var err error
		startT := time.Now()
		err = conn.Do(radix.Pipeline(cmds...))
		endT := time.Now()
		if err != nil {
			log.Fatalf("Received an error with the following command(s): %v, error: %v", cmds, err)
		}
		duration := endT.Sub(startT)
		err = latencies.RecordValue(duration.Microseconds())
		if err != nil {
			log.Fatalf("Received an error while recording latencies: %v", err)
		}
		atomic.AddUint64(&totalCommands, uint64(pipeline))
		i = i + pipeline
	}
}

func getBenchKeyName(keypos uint64) string {
	keyname := fmt.Sprintf("zbench:{%s}:%d", crc16_slot_table[keypos%crc16_num_slots], keypos)
	return keyname
}

func updateCLI(tick *time.Ticker, c chan os.Signal, message_limit uint64) (bool, time.Time, time.Duration, uint64, []float64) {

	start := time.Now()
	prevTime := time.Now()
	prevMessageCount := uint64(0)
	messageRateTs := []float64{}
	fmt.Printf("%26s %7s %25s %25s %7s %25s %25s\n", "Test time", " ", "Total Commands", "Total Errors", "", "Command Rate", "p50 lat. (msec)")
	for {
		select {
		case <-tick.C:
			{
				now := time.Now()
				took := now.Sub(prevTime)
				messageRate := float64(totalCommands-prevMessageCount) / float64(took.Seconds())
				completionPercent := float64(totalCommands) / float64(message_limit) * 100.0
				completionPercentStr := fmt.Sprintf("[%3.1f%%]", completionPercent)
				errorPercent := float64(totalErrors) / float64(totalCommands) * 100.0

				p50 := float64(latencies.ValueAtQuantile(50.0)) / 1000.0

				if prevMessageCount == 0 && totalCommands != 0 {
					start = time.Now()
				}
				if totalCommands != 0 {
					messageRateTs = append(messageRateTs, messageRate)
				}
				prevMessageCount = totalCommands
				prevTime = now

				fmt.Printf("%25.0fs %s %25d %25d [%3.1f%%] %25.2f %25.2f\t", time.Since(start).Seconds(), completionPercentStr, totalCommands, totalErrors, errorPercent, messageRate, p50)
				fmt.Printf("\r")
				if message_limit > 0 && totalCommands >= uint64(message_limit) {
					return true, start, time.Since(start), totalCommands, messageRateTs
				}

				break
			}

		case <-c:
			fmt.Println("\nreceived Ctrl-c - shutting down")
			return true, start, time.Since(start), totalCommands, messageRateTs
		}
	}
}
