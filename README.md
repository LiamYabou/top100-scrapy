# About this project
Top100 is a project that presents the top 100 categories in this world. 

The explanation of the architecture can be found [here](https://github.com/LiamYabou/top100-scrapy/wiki/Architecture).

## Top100 Scrapy
The top100-scrapy is a microservice of the Top100 project. It scrapes various entries form the popular websites.

# Devlopment
## Dependencies
- golang 1.14
- rabbitmq 3.8
- postgresql 12

## Environment Variables
We use [direnv](https://direnv.net/) to streamline the loads of the env variables in the project.
```
export ENV=development
export APP_NAME=top100-scrapy-development
export DB_NAME=top100_development
export DB_USER=postgres
export DB_PASSWORD=
export DB_PORT=5432
export DB_HOST=localhost
export SSL_MODE=disable
export APP_URI=.../top100-scrapy
export CLOUDAMQP_URL=amqp://guest:guest@localhost:5672
export TEST_DB_DSN=postgres://postgres:@localhost:5432/top100_test?sslmode=disable
export MAX_POOL_CONNECTIONS=25
export MIN_POOL_CONNECTIONS=5
export AWS_S3_REGION=
export AWS_S3_BUCKET_NAME=
export AWS_ACCESS_KEY_ID=
export AWS_SECRET_ACCESS_KEY=
export AWS_S3_BUCKET_ENDPOINT=
export HTTP_CLIENT_MAX_IDLE_CONNECTIONS_PER_HOST=25
export GOROUTINE_CONCURRENCY=25
```

## Database Initialization
> If you have some inquires, please ask for help from the `make help` command first.
```
make init
```

## Database Migration
```
make migrate
```

## Data Population
```
make populate
```

## Log Monitor
```
tail -f logs/development.log
```

## Pub / Sub
You can run the following schedule tasks to publish the jobs into the message queue on Rabbitmq.
```
bin/enqueue_categories_insertion
bin/enqueue_products_insertion
```
If you want to view the details of the above, you can open the dashboard of the Rabbitmq with http://localhost:15672 (username: guest, password: guest)

You can run the following command to launch the worker to consume the preceding jobs in the message queue.
```
bin/consume
```

## Testing
```
make test
```

# Contributing
If you have any suggestions or any issues you discovered, you can contact me via hello@mengliu.dev or commit a new `pull request`. I appreciate your help!
