package logger

// It's a new wrapper of logger, extends several general methods.

import (
	log "github.com/sirupsen/logrus"
)

// Construct the factors as the first item of args passed into the methods.
type Factors map[string]interface{}

// args: Receive the factors as the first item of args, it's optional.
// e.g.
// resp, err := http.Get(seedURL)
// if err != nil {
//   logger.Error("Failed to crawl the data from the seed URL.", err)
// }
// defer resp.Body.Close()
//
// if resp.StatusCode != 200 {
//   factors := logger.Factors{"status_code": resp.StatusCode, "status": resp.Status}
//   logger.Error("The status of the code error occurs!", err, factors)
// }
func Error(msg string, err error, args ...Factors) {
	NewFactorsEntry(args...).Panic(msg)
}

func NewFactorsEntry(args ...Factors) *log.Entry {
	l := log.Fields{}
	if len(args) > 0 {
		f := args[0]
		for k, v := range f {
			l[k] = v
		}
	}
	return log.WithFields(l)
}
