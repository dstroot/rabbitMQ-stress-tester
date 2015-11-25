// Copyright 2015 by Dan Stroot. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

/*
Package coyote is a RabbitMQ stress test utility.  RabbitMQ is an
AMQP-compliant queue broker. It facilitates the passing of messages
between or within systems. This program was created for load/stress
testing RabbitMQ.

Usage:

To use this program it is helpful to know some basic AMQP concepts.
AMQP 0-9-1 (Advanced Message Queuing Protocol) is a messaging protocol
that enables conforming client applications to communicate with
conforming messaging middleware brokers. Messaging brokers receive
messages from publishers (applications that publish them, also known
as producers) and route them to consumers (applications that process
them).  Since it is a network protocol, the publishers, consumers and
the broker can all reside on different machines.

Compiling:
	$ go build tester.go

Running:
	$ ./tester -h

Examples
	Open two terminal windows. In one, run

		./tester -s test-rmq-server -c 100000

	That will launch in Consumer mode. It defaults to 50 Goroutines,
	and will (-c)consume 100,000 messages before quitting.

	In the other terminal window, run

		./tester -s test-rmq-server -p 100000 -b 10000 -n 100 -q

	This will run the tester in Producer mode. It will (-p)roduce 100,000
	messages of 10,000 (-b)ytes each. It will launch a pool of 100
	Goroutines (-n), and it will work in (-q)uiet mode, only printing
	NACKs and final statistics to stdout.

		./tester -s test-rmq-server -p 100000 -b 10000 -n 100 -q -a

	With the -a flag each Goroutine waits for an ACK or NACK from
	the RabbitMQ server before publishing the next message. I have
	never seen a missing message in this mode.

	Consume messages forever:

		./tester -s rabbit-mq-test.cs1cloud.internal -c 0

	Produce 100,000 messages of 10KB each, using 50 concurrent
	goroutines, waiting 100 nanoseconds between each message.
	Only print to stdout if there is a nack or when you finish.

		./tester -s rabbit-mq-test.cs1cloud.internal -p 100000 -b 10000 -w 100 -n 50 -q
*/
package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/backstop/rabbit-mq-stress-tester/consumer"
	"github.com/backstop/rabbit-mq-stress-tester/producer"
	"github.com/codegangsta/cli"
)

var totalTime int64
var totalCount int64

func main() {
	// runtime.GOMAXPROCS(runtime.NumCPU())
	app := cli.NewApp()
	app.Name = "Coyote"
	app.HelpName = "coyote"
	app.Usage = "Coyote makes the rabbit run! (RabbitMQ Stress Tester)"
	app.Version = "0.0.1"
	app.Authors = []cli.Author{cli.Author{Name: "Dan Stroot", Email: "dan.stroot@gmail.com"}}
	app.Copyright = "None"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "server, s",
			Value:  "amqp://guest:guest@localhost:5672",
			Usage:  "RabbitMQ connection string",
			EnvVar: "RABBITMQ_SERVER",
		},
		cli.IntFlag{
			Name:   "produce, p",
			Value:  0,
			Usage:  "number of messages to produce, -1 to produce forever",
			EnvVar: "RABBITMQ_PRODUCE",
		},
		cli.IntFlag{
			Name:   "wait, w",
			Value:  0,
			Usage:  "number of nanoseconds to wait between publish events",
			EnvVar: "RABBITMQ_WAIT",
		},
		cli.IntFlag{
			Name:   "consume, c",
			Value:  -1,
			Usage:  "number of messages to consume. 0 consumes forever",
			EnvVar: "RABBITMQ_CONSUME",
		},
		cli.IntFlag{
			Name:   "bytes, b",
			Value:  1000,
			Usage:  "number of extra bytes to add to the RabbitMQ message payload. ~50K max",
			EnvVar: "RABBITMQ_BYTES",
		},
		cli.IntFlag{
			Name:   "concurrency, n",
			Value:  50,
			Usage:  "number of reader/writer Goroutines",
			EnvVar: "RABBITMQ_CONCURRECY",
		},
		cli.BoolFlag{
			Name:   "quiet, q",
			Usage:  "print only errors to stdout",
			EnvVar: "RABBITMQ_QUIET",
		},
		cli.BoolFlag{
			Name:   "wait-for-ack, a",
			Usage:  "wait for an ack or nack after enqueueing a message",
			EnvVar: "RABBITMQ_ACK",
		},
	}
	app.Action = func(c *cli.Context) {
		runApp(c)
	}
	app.Run(os.Args)
}

// runApp function with parameters (types go after params), and named return values
// func functionName(param1 string, param2 int) (n int, s string) {}
//
// Since main is taken up by command line handling, this is really the "main"
// procedure of the application - it fires up the producers and consumers.
//
// Parameter/name/type/description
// Return/name/type/description
//
// @param  c    *cli.Context  [pointer to cli.Context]
// @return nil
func runApp(c *cli.Context) {

	uri := "amqp://guest:guest@" + c.String("server")

	if c.Int("consume") > -1 && c.Int("produce") != 0 {
		fmt.Println("Error: Cannot specify both producer and consumer options together")
		cli.ShowAppHelp(c)
		os.Exit(1)
	} else if c.Int("consume") > -1 {
		fmt.Println("Running in consumer mode!")
		makeConsumers(uri, c.Int("concurrency"), c.Int("consume"))
	} else if c.Int("produce") != 0 {
		fmt.Println("Running in producer mode!")
		config := producer.MyConfig{URI: uri, Bytes: c.Int("bytes"), Quiet: c.Bool("quiet"), WaitForAck: c.Bool("wait-for-ack")}
		makeProducers(c.Int("produce"), c.Int("wait"), c.Int("concurrency"), config)
	} else {
		cli.ShowAppHelp(c)
		os.Exit(0)
	}

}

/**
 * Create a task channel and a variable number of goroutines (passed in via the concurrency variable)
 *
 * @param   n           int            [number of messages to produce, -1 to produce forever]
 * @param   wait        int            [number of nanoseconds to wait between publish events]
 * @param   concurrency int            [number of reader/writer Goroutines]
 * @param   config      producerConfig [struct containing configuration information]
 * @return              [description]
 */
func makeProducers(n int, wait int, concurrency int, config producer.MyConfig) {
	println("Make Producers")
	println("producers: ", n)
	println("wait: ", wait)
	println("concurrency: ", concurrency)

	taskChan := make(chan int)

	for i := 0; i < concurrency; i++ {
		go producer.Produce(config, taskChan)
	}

	start := time.Now()

	for i := 0; i < n; i++ {
		taskChan <- i
		time.Sleep(time.Duration(int64(wait)))
	}

	time.Sleep(time.Duration(10000))

	close(taskChan)

	log.Printf("Finished: %s", time.Since(start))
}

func makeConsumers(uri string, concurrency int, toConsume int) {

	doneChan := make(chan bool)

	for i := 0; i < concurrency; i++ {
		go consumer.Consume(uri, doneChan)
	}

	start := time.Now()

	if toConsume > 0 {
		for i := 0; i < toConsume; i++ {
			<-doneChan
			if i == 1 {
				start = time.Now()
			}
			numComsumed := i + 1 // 1-100 easier than 0-99
			log.Println("Consumed: ", numComsumed)
		}
	} else {
		for {
			<-doneChan
		}
	}

	log.Printf("Done consuming! %s", time.Since(start))
}
