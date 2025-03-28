rabbit
======
[![](https://godoc.org/github.com/streamdal/rabbit?status.svg)](http://godoc.org/github.com/streamdal/rabbit) [![Master build status](https://github.com/streamdal/rabbit/workflows/main/badge.svg)](https://github.com/streamdal/rabbit/actions) [![Go Report Card](https://goreportcard.com/badge/github.com/streamdal/rabbit)](https://goreportcard.com/report/github.com/streamdal/rabbit)

A RabbitMQ wrapper lib around ~[streadway/amqp](https://github.com/streadway/amqp)~ [rabbitmq/amqp091-go](https://github.com/rabbitmq/amqp091-go) 
with some bells and whistles.

NOTE: `streadway/amqp` is no longer maintained and RabbitMQ team have forked `streadway/amqp` and created `rabbitmq/amqp091-go`. You can read about this change [here](https://github.com/streadway/amqp/issues/497). This library uses `rabbitmq/amqp091-go`.

* Support for auto-reconnect
* Support for context (ie. cancel/timeout)
* Support for using multiple binding keys
* Support Producer, Consumer or both modes

# Motivation

We (Streamdal, formerly Batch.sh), make heavy use of RabbitMQ - we use it as 
the primary method for facilitating inter-service communication. Due to this, 
all services make use of RabbitMQ and are both publishers and consumers.

We wrote this lib to ensure that all of our services make use of Rabbit in a
consistent, predictable way AND are able to survive network blips.

**NOTE**: This library works only with non-default exchanges. If you need support
for default exchange - open a PR!

# Usage
```go
package main

import (
    "fmt"
    "log"  

    "github.com/streamdal/rabbit"
)

func main() { 
    r, err := rabbit.New(&rabbit.Options{
        URL:          "amqp://localhost",
        QueueName:    "my-queue",
        ExchangeName: "messages",
        BindingKeys:   []string{"messages"},
    })
    if err != nil {
        log.Fatalf("unable to instantiate rabbit: %s", err)
    }
    
    routingKey := "messages"
    data := []byte("pumpkins")

    // Publish something
    if err := r.Publish(context.Background(), routingKey, data); err != nil {
        log.Fatalf("unable to publish message: ")
    }

    // Consume once
    if err := r.ConsumeOnce(nil, func(amqp.Delivery) error {
        fmt.Printf("Received new message: %+v\n", msg)
    }); err != nil {
        log.Fatalf("unable to consume once: %s", err),
    }

    var numReceived int

    // Consume forever (blocks)
    ctx, cancel := context.WithCancel(context.Background())

    r.Consume(ctx, nil, func(msg amqp.Delivery) error {
        fmt.Printf("Received new message: %+v\n", msg)
        
        numReceived++
        
        if numReceived > 1 {
            r.Stop()
        }
    })

    // Or stop via ctx 
    r.Consume(..)
    cancel()
}
```

### Retry Policies

You can specify a retry policy for the consumer.
A pre-made ACK retry policy is available in the library at `rp := rabbit.DefaultAckPolicy()`. This policy will retry
acknowledgement unlimited times

You can also create a new policy using the `rabbit.NewRetryPolicy(maxAttempts, time.Millisecond * 200, time.Second, ...)` function.

The retry policy can then be passed to consume functions as an argument:

```go
consumeFunc := func(msg amqp.Delivery) error {
    fmt.Printf("Received new message: %+v\n", msg)
    
    numReceived++
    
    if numReceived > 1 {
            r.Stop()
        }
    }

rp := rabbit.DefaultAckPolicy()

r.Consume(ctx, nil, consumeFunc, rp)
```