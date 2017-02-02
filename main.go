// package main provides a simple script to discover the longest running orca
// pipeline executions in Redis.
package main

import (
	"flag"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	redis "gopkg.in/redis.v5"
)

var (
	log = logrus.New()
	now = time.Now()

	executionType = flag.String("type", "orchestration", "orchestration or pipeline")
	statusFilter  = flag.String("status", "RUNNING", "the execution status to filter on")
)

type execution struct {
	Key       string
	StartTime time.Time
}

func (e execution) TimeSince() string {
	return now.Sub(e.StartTime).String()
}

type ByStartTime []execution

func (s ByStartTime) Len() int {
	return len(s)
}

func (s ByStartTime) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s ByStartTime) Less(i, j int) bool {
	return s[i].StartTime.After(s[j].StartTime)
}

func main() {
	flag.Parse()

	log.Infof("Filtering %s on status: %s", *executionType, *statusFilter)

	c, err := createClient()
	if err != nil {
		log.WithField("cause", err.Error()).Fatal("failed creating Redis client")
	}

	keys, err := c.Keys(fmt.Sprintf("%s:*", *executionType)).Result()
	if err != nil {
		log.WithField("cause", err.Error()).Fatal("failed listing all orchestration keys")
	}

	var executions []execution
	for _, key := range keys {
		if strings.HasPrefix(key, fmt.Sprintf("%s:app:", *executionType)) ||
			strings.HasPrefix(key, fmt.Sprintf("%s:executions", *executionType)) {
			continue
		}

		status, err := c.HGet(key, "status").Result()
		if err != nil {
			log.WithField("cause", err.Error()).Warn("execution does not have status")
			continue
		}

		if status != *statusFilter {
			continue
		}

		start, err := c.HGet(key, "startTime").Result()
		if err != nil {
			log.WithField("cause", err.Error()).Warn("failed getting startTime on key")
			continue
		}

		timestamp, err := strconv.ParseInt(start, 10, 64)
		if err != nil {
			log.WithField("cause", err.Error()).Error("could not convert starttime to int")
			continue
		}

		executions = append(executions, execution{
			Key:       key,
			StartTime: time.Unix(timestamp/1000, 0),
		})
	}

	sort.Sort(ByStartTime(executions))
	for _, execution := range executions {
		fmt.Printf("%s  %s\n", execution.Key, execution.TimeSince())
	}
}

func createClient() (*redis.Client, error) {
	c := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	if _, err := c.Ping().Result(); err != nil {
		return nil, errors.Wrap(err, "could not connect to redis")
	}
	return c, nil
}
