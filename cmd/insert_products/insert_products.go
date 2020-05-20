package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"top100-scrapy/pkg/app"
	"top100-scrapy/pkg/crawler"
	"top100-scrapy/pkg/logger"
	"top100-scrapy/pkg/model"
	"top100-scrapy/pkg/model/category"

	"github.com/lib/pq"
	"github.com/panjf2000/ants"
	"github.com/streadway/amqp"
)

type optionInterface interface{}

type options struct {
	msgs <-chan amqp.Delivery
}

// TODO: Figure out the usage.
type jobs struct {
	args *[]options
}

func main() {
	logger.Debug("Debug starts - insert products")
	defer app.Finalize()
	performJob()
	logger.Debug("Debug stops -  insert products")
}

func performJob() {
	ch, err := app.AMQPconn.Channel()
	if err != nil {
		logger.Error("Failed to open a channel.", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"products_insertion", // name
		true,                 // durable
		false,                // delete when unused
		false,                // exclusive
		false,                // no-wait
		nil,                  // arguments
	)
	if err != nil {
		logger.Error("Failed to declare a queue.", err)
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		logger.Error("Failed to register a consumer.", err)
	}

	fmt.Println(" [*] Waiting for messages. To exit press CTRL+C")
	fmt.Printf(" [*] The PID of the consumer is: %d\n", os.Getpid())

	var concurrency = 25
	prefetchCount := concurrency * 4
	err = ch.Qos(prefetchCount, 0, false)
	if err != nil {
		logger.Error("Failed to set QoS.", err)
	}

	var wg sync.WaitGroup
	o := &options{
		msgs: msgs,
	}
	p, _ := ants.NewPoolWithFunc(concurrency, func(optionInterface interface{}) {
		options, ok := optionInterface.(*options)
		if !ok {
			// TODO: Replace with logger.
			fmt.Printf("error: %s\n", "there is not the instance of the type `options` in the interface `optionInterface`.")
		}
		worker(options)
		wg.Done()
	})
	defer p.Release()
	for {
		wg.Add(1)
		_ = p.Invoke(o)
	}
}

func worker(options *options) {
	for d := range options.msgs {
		fmt.Printf("Received a message: %s\n", d.Body)
		args := strings.Split(string(d.Body), ",")
		// args[0]: The id of the row of category
		// args[1]: The number of the page expected to scrape
		categoryId, _ := strconv.Atoi(args[0])
		page, _ := strconv.Atoi(args[1])
		category, err := category.NewRow().FetchRow(categoryId, app.DBconn)
		if err != nil {
			logger.Error("Failed to query on DB or failed to assign a value by the Scan.", err)
		}
		if page == 2 {
			category.Url = category.Url + fmt.Sprintf("?_encoding=UTF8&pg=%d", page)
		}
		set, err := app.InitCrawler(category).WithPage(page).ScrapeProducts()
		if err, ok := err.(*crawler.EmptyError); ok {
			logger.Info(fmt.Sprintf("The names scraped from the url `%s` are empty, the category id stored into the DB is %d", err.Category.Url, err.Category.Id))
			if err := d.Ack(false); err != nil { // Acknowledge a message maunally.
				logger.Error("Failed to acknowledge a message.", err)
			}
			fmt.Println("Done")
			return
		}
		modelOptions := &model.Options{
			DB:       app.DBconn,
			Category: category,
			Set:      set,
			Page:     page,
		}
		msg, err := model.New().WithOptions(modelOptions).BulkilyInsertRelations()
		if pqErr, ok := err.(*pq.Error); ok {
			factors := logger.Factors{
				"pq_err_code":   pqErr.Code,
				"pq_err_msg":    pqErr.Message,
				"pq_err_detail": pqErr.Detail,
				"pq_err_hint":   pqErr.Hint,
				"pq_err_query":  pqErr.InternalQuery,
				"category_id":   category.Id,
				"category_url":  category.Url,
			}
			switch pqErr.Code {
			case "23505": // Violate unique constraint
				logger.Info("Could not insert a row.", factors)
			default:
				logger.Error(msg, err, factors)
			}
		} else if err != nil {
			logger.Error(msg, err)
		}

		if err := d.Ack(false); err != nil { // Acknowledge a message maunally.
			logger.Error("Failed to acknowledge a message.", err)
		}
		fmt.Println("Done")
	}
}
