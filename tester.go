package main

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/codegangsta/cli"
	"github.com/streadway/amqp"
)

var totalTime int64
var totalCount int64

// MqMessage ...
type MqMessage struct {
	TimeNow        time.Time
	SequenceNumber int
	Payload        string
}

func main() {
	app := cli.NewApp()
	app.Name = "Tester"
	app.Version = "0.0.1"
	app.Usage = "Make the rabbit cry"
	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "server, s", Value: "localhost:5672", Usage: "hostname for RabbitMQ server"},
		cli.IntFlag{Name: "produce, p", Value: 0, Usage: "number of messages to produce, -1 to produce forever"},
		cli.IntFlag{Name: "wait, w", Value: 0, Usage: "number of nanoseconds to wait between publish events"},
		cli.IntFlag{Name: "consume, c", Value: -1, Usage: "number of messages to consume. 0 consumes forever"},
		cli.IntFlag{Name: "bytes, b", Value: 0, Usage: "number of extra bytes to add to the RabbitMQ message payload. ~50K max"},
		cli.IntFlag{Name: "concurrency, n", Value: 50, Usage: "number of reader/writer Goroutines"},
		cli.BoolFlag{Name: "quiet, q", Usage: "print only errors to stdout"},
		cli.BoolFlag{Name: "wait-for-ack, a", Usage: "wait for an ack or nack after enqueueing a message"},
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
	println("Running!")

	uri := "amqp://guest:guest@" + c.String("server")

	if c.Int("consume") > -1 {
		makeConsumers(uri, c.Int("concurrency"), c.Int("consume"))
	}

	if c.Int("produce") != 0 {
		config := producerConfig{uri, c.Int("bytes"), c.Bool("quiet"), c.Bool("wait-for-ack")}
		makeProducers(c.Int("produce"), c.Int("wait"), c.Int("concurrency"), config)
	}
}

/**
 * [makeQueue description]
 * @param  c *amqp.Channel [pointer to amqp.Channel]
 * @return   ampq.Queue    [returns a queue]
 */
func makeQueue(c *amqp.Channel) amqp.Queue {

	// Queues in the AMQP model are very similar to queues in other message
	// and task-queueing systems: they store messages that are consumed by
	// applications.
	//
	// Before a queue can be used it has to be declared. QueueDeclare declares
	// a queue to hold messages and deliver to consumers. Declaring creates a
	// queue if it doesn't already exist, or ensures that an existing queue
	// matches the same parameters.
	//
	// Durable and Non-Auto-Deleted queues will survive server restarts and
	// remain when there are no remaining consumers or bindings. Persistent
	// publishings will be restored in this queue on server restart.
	// These queues are only able to be bound to durable exchanges.
	//
	// Non-Durable and Auto-Deleted queues will not be redeclared on server
	// restart and will be deleted by the server after a short time when the
	// last consumer is canceled or the last consumer's channel is closed.
	// Queues with this lifetime can also be deleted normally with QueueDelete.
	// These durable queues can only be bound to non-durable exchanges.
	//
	// Non-Durable and Non-Auto-Deleted queues will remain declared as long
	// as the server is running regardless of how many consumers. This lifetime
	// is useful for temporary topologies that may have long delays between
	// consumer activity. These queues can only be bound to non-durable exchanges.
	//
	// Durable and Auto-Deleted queues will be restored on server restart,
	// but without active consumers, will not survive and be removed. This
	// Lifetime is unlikely to be useful.
	//
	// Exclusive queues are only accessible by the connection that declares
	// them and will be deleted when the connection closes. Channels on other
	// connections will receive an error when attempting declare, bind,
	// consume, purge or delete a queue with the same name.
	//
	// When noWait is true, the queue will assume to be declared on the server.
	// A channel exception will arrive if the conditions are met for existing
	// queues or attempting to modify an existing queue from a different
	// connection.
	//
	// When the error return value is not nil, you can assume the queue
	// could not be declared with these parameters and the channel will be closed.

	/**
	 * [QueueDeclare description]
	 * @param name         string [name of the queue]
	 * @param durable      bool   [durable flag]
	 * @param autoDelete   bool   [auto delete flag]
	 * @param exclusive    bool   [exclusive flag]
	 * @param noWait       bool   [no wait flag]
	 * @param args         Table  [additional arguments]
	 */
	q, err2 := c.QueueDeclare("stress-test-exchange", true, false, false, false, nil)
	if err2 != nil {
		panic(err2)
	}
	return q
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
func makeProducers(n int, wait int, concurrency int, config producerConfig) {

	println("Make Producers")
	println("producers: ", n)
	println("wait: ", wait)
	println("concurrency: ", concurrency)

	taskChan := make(chan int)

	for i := 0; i < concurrency; i++ {
		go producer(config, taskChan)
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
		go consume(uri, doneChan)
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

func consume(uri string, doneChan chan bool) {
	log.Println("Consuming...")

	connection, err := amqp.Dial(uri)
	if err != nil {
		println(err.Error())
		panic(err.Error())
	}
	defer connection.Close()

	channel, err1 := connection.Channel()
	if err1 != nil {
		println(err1.Error())
		panic(err1.Error())
	}
	defer channel.Close()

	q := makeQueue(channel)

	msgs, err3 := channel.Consume(q.Name, "", true, false, false, false, nil)
	if err3 != nil {
		panic(err3)
	}

	for d := range msgs {
		doneChan <- true
		var thisMessage MqMessage
		err4 := json.Unmarshal(d.Body, &thisMessage)
		if err4 != nil {
			log.Printf("Error unmarshalling! %s", err.Error())
		}
		log.Printf("Message age: %s", time.Since(thisMessage.TimeNow))

	}

	log.Println("done recieving")
}
