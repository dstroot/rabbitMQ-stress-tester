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
	"os"
	"sync"
	"time"

	"github.com/backstop/rabbit-mq-stress-tester/consumer"
	"github.com/backstop/rabbit-mq-stress-tester/logging"
	"github.com/backstop/rabbit-mq-stress-tester/producer"
	"github.com/codegangsta/cli"
)

var totalTime int64
var totalCount int64

// main parses all the command flags and the calls the function "runApp" to
// actually run the application
func main() {
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
			Name:   "consumers, c",
			Value:  0,
			Usage:  "number of consumers to create. each is a seperate goroutine",
			EnvVar: "RABBITMQ_CONSUMER",
		},
		cli.IntFlag{
			Name:   "producers, p",
			Value:  0,
			Usage:  "number of producers to create. each is a seperate goroutine",
			EnvVar: "RABBITMQ_PRODUCER",
		},
		cli.IntFlag{
			Name:   "messages, m",
			Value:  10,
			Usage:  "number of messages to process. -1 is infinite",
			EnvVar: "RABBITMQ_MESSAGES",
		},
		cli.IntFlag{
			Name:   "bytes, b",
			Value:  1000,
			Usage:  "number of bytes to deliver in the RabbitMQ message payload (~50K max)",
			EnvVar: "RABBITMQ_BYTES",
		},
		cli.BoolFlag{
			Name:   "reliable, r",
			Usage:  "wait for an ack or nack after enqueueing a message",
			EnvVar: "RABBITMQ_RELIABLE",
		},
		cli.BoolFlag{
			Name:   "verbose, V",
			Usage:  "show all logging output",
			EnvVar: "RABBITMQ_VERBOSE",
		},
	}
	app.Action = func(c *cli.Context) {
		runApp(c)
	}
	app.Run(os.Args)
}

// runApp is really the "main" function of the application - it fires up
// the producers and consumers. (Since main is taken up by command line handling)
func runApp(c *cli.Context) {

	// setup logging
	// logging.SetLogFile("./tester.txt")
	if c.Bool("verbose") {
		logging.SetLogThreshold(logging.LevelTrace)
		logging.SetStdoutThreshold(logging.LevelInfo)
	} else {
		logging.SetStdoutThreshold(logging.LevelWarn)
	}

	uri := "amqp://guest:guest@" + c.String("server")

	if c.Int("consumers") > 0 && c.Int("producers") > 0 {
		logging.ERROR.Println("Cannot specify both producer and consumer options together. Start this up as either a producer *or* consumer.")
		cli.ShowAppHelp(c)
		os.Exit(1)

	} else if c.Int("messages") < 1 {
		logging.ERROR.Println("Messages must be one or more.")
		cli.ShowAppHelp(c)
		os.Exit(1)

	} else if c.Int("consumers") > 0 {
		logging.WARN.Println("Consumers to create: ", c.Int("consumers"))
		logging.WARN.Println("Messages to consume: ", c.Int("messages"))
		makeConsumers(uri, c.Int("consumers"), c.Int("messages"))

	} else if c.Int("producers") > 0 {
		logging.WARN.Println("Producers to create: ", c.Int("producers"))
		logging.WARN.Println("Messages to send: ", c.Int("messages"))
		logging.WARN.Println("Bytes per message: ", c.Int("bytes"))
		// logging.WARN.Println("Wait between messages: ", c.Int("wait"))

		config := producer.MyConfig{URI: uri, Bytes: c.Int("bytes"), Quiet: c.Bool("quiet"), Reliable: c.Bool("reliable")}
		makeProducers(c.Int("messages"), c.Int("producers"), config)

	} else {
		logging.ERROR.Println("Something was specified incorrectly.")
		cli.ShowAppHelp(c)
		os.Exit(0)
	}

}

// makeProducers creates a variable number of "producer" goroutines and
// sends a variable number of messages to the taskChan channel to make the
// producers create that number of messages.
func makeProducers(messages int, producers int, config producer.MyConfig) {

	start := time.Now()
	taskChan := make(chan int, messages) // buffered channel, # of messages
	var wg sync.WaitGroup                // track number of Goroutines

	// fill tasks channel
	go func() {
		for i := 0; i < messages; i++ {
			logging.INFO.Printf("Making message %d", i+1)
			taskChan <- i
		}
		// closing the channel allows our producer goroutines to know
		// when they are done.
		close(taskChan)
	}()

	// create producers to create messages from the task channel
	for i := 0; i < producers; i++ {
		wg.Add(1)
		// make a producer
		go func(c producer.MyConfig, tch chan int, i int) {
			defer wg.Done() // Decrement the counter when the goroutine completes.
			logging.WARN.Printf("Making producer %d", i+1)
			producer.Produce(c, tch, i)
		}(config, taskChan, i)
	}

	// wait for producers to finish
	wg.Wait()
	logging.WARN.Printf("Producing finished: %s", time.Since(start))
}

// makeConsumers creates a variable number of "consumer" goroutines and
// receives a variable number of messages.
func makeConsumers(uri string, consumers int, messages int) {

	doneChan := make(chan bool)

	// create consumers
	for i := 0; i < consumers; i++ {
		logging.WARN.Printf("Making consumer %d", i+1)
		go consumer.Consume(uri, doneChan)
	}

	start := time.Now()

	// get messages
	if messages > 0 {
		for i := 0; i < messages; i++ {
			<-doneChan
			if i == 1 {
				start = time.Now()
			}
			logging.INFO.Printf("Number of messages consumed %d", i+1)
		}
	} else {
		for {
			<-doneChan
		}
	}

	logging.WARN.Printf("Done consuming! %s", time.Since(start))
}
