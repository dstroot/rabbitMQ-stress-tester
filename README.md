rabbit-mq-stress-tester
=======================

Introduction
------------

This program is designed for load testing with RabbitMQ. RabbitMQ is an AMQP-compliant queue broker. It exists to facilitate the passing of messages between or within systems.

To use this program it is helpful to know some basic AMQP concepts.  AMQP 0-9-1 (Advanced Message Queuing Protocol) is a messaging protocol that enables conforming client applications to communicate with conforming messaging middleware brokers. Messaging brokers receive messages from publishers (applications that publish them, also known as producers) and route them to consumers (applications that process them).  Since it is a network protocol, the publishers, consumers and the broker can all reside on different machines.

https://www.rabbitmq.com/tutorials/amqp-concepts.html

Terminology
-----------

If you’re unfamiliar with AMQP, here’s some terminology to help understand what’s possible with a queue broker and what the words mean.

* **Connection:** A connection is a long-lived TCP connection between an AMQP client and a queue broker. Maintaining a connection reduces TCP overhead. A client can re-use a connection, and can share a connection among threads.
* **Channel:** A channel is a short-lived sub-connection between a client and a broker. The client can create and dispose of channels without incurring a lot of overhead.
* **Exchange:** A client writes messages to an exchange. The exchange forwards each message on to zero or more queues based on the message’s routing key.
* **Queue:** A queue is a first-in, first out holder of messages. A client reads messages from a queue. The client can specify a queue name (useful, for example, for a work queue where multiple clients are consuming from the same queue), or allow the queue broker to assign it a queue name (useful if you want to distribute copies of a message to multiple clients).
Routing Key: A string (optionally) attached to each message. Depending on the exchange type, the exchange may or may not use the Routing Key to determine the queues to which it should publish the message.

Exchange types:

* **Direct:** Delivers all messages with the same routing key to the same queue(s).
* **Fanout:** Ignores the routing key, delivers a copy of the message to each queue bound to it.
* **Topic:** Each queue subscribes to a topic, which is a regular expression. The exchange delivers the message to a queue if the queue’s subscribed topic matches the message.
* **Header:** Ignores the routing key and delivers the message based on the AMQP header. Useful for certain kinds of messages.

Compiling
---------

    $ go build tester.go

That will produce the executable `tester`

Running
-------

    $ ./tester -h

Examples
--------
Open two terminal windows. In one, run

    ./tester -s test-rmq-server -c 100000

That will launch the in Consumer mode. It defaults to 50 Goroutines, and will consume 100,000 messages before quitting.

In the other terminal window, run

    ./tester -s test-rmq-server -p 100000 -b 10000 -n 100 -q

This will run the tester in Producer mode. It will (-p)roduce 100,000 messages of 10,000 (-b)ytes each. It will launch a pool of 100 Goroutines (-n), and it will work in (-q)uiet mode, only printing NACKs and final statistics to stdout.

    ./tester -s test-rmq-server -p 100000 -b 10000 -n 100 -q -a

With the -a flag each Goroutine waits for an ACK or NACK from the RabbitMQ server before publishing the next message. I have never seen a missing message in this mode.

Consume messages forever:

    ./tester -s rabbit-mq-test.cs1cloud.internal -c 0

Produce 100,000 messages of 10KB each, using 50 concurrent goroutines, waiting 100 nanoseconds between each message. Only print to stdout if there is a nack or when you finish.

    ./tester -s rabbit-mq-test.cs1cloud.internal -p 100000 -b 10000 -w 100 -n 50 -q
	
	
Also checkout:	
https://github.com/agocs/rabbit-mq-stress-tester
