# redis-zbench-go
Redis Sorted Sets Benchmark


## Overview

This repo contains code to trigger `load` ( `ZADD` ) or `query` (`ZRANGEBYLEX key min max`) benchmarks, 
with/without transaction, with/without pipelining, and with diversified ways of controlling the per-key size and total keyspace length.

## Install 
### Standalone binaries ( no Golang needed )

If you don't have go on your machine and just want to use the produced binaries you can download the following prebuilt bins:

| OS | Arch | Link |
| :---         |     :---:      |          ---: |
| Windows   | amd64     | [redis--go_windows_amd64.exe](https://s3.amazonaws.com/benchmarks.redislabs/tools/redis-zbench-go/unstable/redis-zbench-go_windows_amd64.exe)    |
| Linux   | amd64     | [redis-zbench-go_linux_amd64](https://s3.amazonaws.com/benchmarks.redislabs/tools/redis-zbench-go/unstable/redis-zbench-go_linux_amd64)    |
| Linux   | arm64     | [redis-zbench-go_linux_arm64](https://s3.amazonaws.com/benchmarks.redislabs/tools/redis-zbench-go/unstable/redis-zbench-go_linux_arm64)    |
| Darwin   | amd64     | [redis-zbench-go_darwin_amd64](https://s3.amazonaws.com/benchmarks.redislabs/tools/redis-zbench-go/unstable/redis-zbench-go_darwin_amd64)    |
| Darwin   | arm64     | [redis-zbench-go_darwin_arm64](https://s3.amazonaws.com/benchmarks.redislabs/tools/redis-zbench-go/unstable/redis-zbench-go_darwin_arm64)    |



Here's an example on how to use the above links:
```bash
# Fetch this repo
wget https://s3.amazonaws.com/benchmarks.redislabs/tools/redis-zbench-go/unstable/redis-zbench-go_linux_amd64

# change permissions
chmod 755 redis-zbench-go_linux_amd64

# give it a try 
./redis-zbench-go_linux_amd64 --help
```

### Installation in a Golang env

The easiest way to get and install the benchmark utility with a Go Env is to use
`go get` and then `go install`:
```bash
# Fetch this repo
go get github.com/filipecosta90/redis-zbench-go
cd $GOPATH/src/github.com/filipecosta90/redis-zbench-go
make
```

## Usage of redis-zbench-go

To get a full grasp of the tool capabilities please check the `--help` output

```
$ ./redis-zbench-go --help
Usage of ./redis-zbench-go:
  -a string
        Password for Redis Auth.
  -c uint
        number of clients. (default 50)
  -d uint
        Data size of each sorted set element. (default 10)
  -debug int
        Client debug level.
  -h string
        Server hostname. (default "127.0.0.1")
  -key-elements-max uint
        Use zipfian random-sized items in the specified range (min-max). (default 10)
  -key-elements-min uint
        Use zipfian random-sized items in the specified range (min-max). (default 1)
  -mode load
        Bechmark mode. One of [load,query]. load will populate the db with sorted sets. `query` will run the zrangebylexscore command .
  -multi
        Run each command in multi-exec.
  -n uint
        Total number of requests. Only used in case of -mode=query (default 10000000)
  -p int
        Server port. (default 12000)
  -pipeline uint
        Redis pipeline value. (default 1)
  -r uint
        keyspace length. (default 1000000)
  -random-seed int
        random seed to be used. (default 12345)
  -rps int
        Max rps. If 0 no limit is applied and the DB is stressed up to maximum.
```

## Sample output - 1M Keys keyspace, 100K issued commands, pipeline of 100 with transaction enabled, while querying at a limit of @10K RPS

### Prepopulation

```
$ ./redis-zbench-go -mode load -r 1000000 -p 6379 
Total clients: 50. Commands per client: 20000 Total commands: 1000000
Using random seed: 12345
                 Test time                    Total Commands              Total Errors                      Command Rate           p50 lat. (msec)
                       18s [100.0%]                   1000000                         0 [0.0%]                   1370.82                      0.81      
#################################################
Total Duration 18.000 Seconds
Total Issued commands 1000000
Total Errors 0
Throughput summary: 55555 requests per second
Latency summary (msec):
          p50       p95       p99
        0.808     1.200     1.582

```


### Querying with piepline of 100, and transction, @10K rps

```
$ ./redis-zbench-go  -mode query -p 6379 -pipeline 100 -c 10 -multi -rps 10000 -n 100000 -r 1000000
Total clients: 10. Commands per client: 10000 Total commands: 100000
Using random seed: 12345
                 Test time                    Total Commands              Total Errors                      Command Rate           p50 lat. (msec)
                        9s [100.0%]                    100000                         0 [0.0%]                   9095.29                      2.53      
#################################################
Total Duration 9.001 Seconds
Total Issued commands 100000
Total Errors 0
Throughput summary: 11110 requests per second
Latency summary (msec):
          p50       p95       p99
        2.529     3.509     6.083
#################################################
Printing reply histogram
Size: 0 Count: 51962. (% 51.96)
Size: 1 Count: 9774. (% 9.77)
Size: 2 Count: 5124. (% 5.12)
Size: 3 Count: 5276. (% 5.28)
Size: 4 Count: 5088. (% 5.09)
Size: 5 Count: 5152. (% 5.15)
Size: 6 Count: 4943. (% 4.94)
Size: 7 Count: 4694. (% 4.69)
Size: 8 Count: 4547. (% 4.55)
Size: 9 Count: 3440. (% 3.44)
--------------------------------------------------
Total processed replies 100000
```