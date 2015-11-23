// Copyright 2015 Dan Stroot. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// RabbitMQ is an AMQP-compliant queue broker. It exists to facilitate
// the passing of messages between or within systems. This program is
// designed for load testing with RabbitMQ.
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

// function with parameters (types go after params), and named return values
// func functionName(param1 string, param2 int) (n int, s string) {}

/**
 * Since main is taken up by command line handling, this is really the "main"
 * procedure of the application - it fires up the producers and consumers.
 *
 * Parameter/name/type/description
 * Return/name/type/description
 *
 * @param  c    *cli.Context  [pointer to cli.Context]
 * @return nil
 */
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
