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
	"github.com/robzienert/orca-executions/filter"
	redis "gopkg.in/redis.v5"
)

var (
	log = logrus.New()

	redisAddr     = flag.String("redisAddr", "localhost:6379", "The address for orca's redis instance")
	executionType = flag.String("type", "orchestration", "orchestration or pipeline")
	statusFilter  = flag.String("status", "RUNNING", "the execution status to filter on")
	extraFilters  = flag.String("filters", "", "Extra filters in comma-delimited Key=Value format")
	fields        = flag.String("fields", "", "Extra fields to return in the data table for each record, comma-delimited")
	quiet         = flag.Bool("quiet", false, "Set if you do not want logging enabled")
	debug         = flag.Bool("debug", false, "Set if you want debug level logging")
)

type execution struct {
	Key       string
	StartTime time.Time
	Fields    []string
}

func (e execution) TimeSince() string {
	return time.Now().Sub(e.StartTime).String()
}

func (e execution) String() string {
	s := []string{e.Key, e.TimeSince()}
	for _, f := range e.Fields {
		s = append(s, f)
	}
	return strings.Join(s, "  ")
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

	if *quiet {
		// Not quite, but meh
		log.Level = logrus.PanicLevel
	}
	if *debug {
		log.Level = logrus.DebugLevel
	}

	filters, err := filter.Parse(*extraFilters, *statusFilter)
	if err != nil {
		log.WithField("cause", err.Error()).Fatal("could not parse given filters")
	}
	for _, filter := range filters {
		log.WithField(filter.Key, filter.Value).Info("Adding result filter")
	}

	c, err := createClient()
	if err != nil {
		log.WithField("cause", err.Error()).Fatal("failed creating Redis client")
	}

	log.Infof("Finding all keys for type: %s...", *executionType)
	keys, err := c.Keys(fmt.Sprintf("%s:*", *executionType)).Result()
	if err != nil {
		log.WithError(err).Fatal("failed listing all keys")
	}

	keys = filterSlice(keys, func(s string) bool {
		return !strings.HasSuffix(s, ":stageIndex")
	})

	log.Info("Filtering...")
	var executions []execution
	for _, key := range keys {
		if strings.HasPrefix(key, fmt.Sprintf("%s:app:", *executionType)) ||
			strings.HasPrefix(key, fmt.Sprintf("%s:executions", *executionType)) {
			continue
		}

		var shouldPass bool
		for _, f := range filters {
			isMatching, err := filter.Get(f)(c, key, f)
			log.WithFields(logrus.Fields{"match": isMatching, "key": key}).Debugf("%#v", f)
			if err != nil {
				log.WithError(err).Warn("could not complete filter")
				shouldPass = true
				break
			}
			if !isMatching {
				shouldPass = true
				break
			}
		}
		if shouldPass {
			continue
		}

		start, err := c.HGet(key, "startTime").Result()
		if err != nil {
			log.WithError(err).Warn("failed getting startTime on key")
			continue
		}

		timestamp, err := strconv.ParseInt(start, 10, 64)
		if err != nil {
			log.WithError(err).Error("could not convert starttime to int")
			continue
		}

		execution := execution{
			Key:       key,
			StartTime: time.Unix(timestamp/1000, 0),
		}

		if *fields != "" {
			for _, f := range strings.Split(*fields, ",") {
				fv, err := c.HGet(key, f).Result()
				if err != nil {
					fv = "ERR"
					log.WithError(err).Warnf("could not get field %s", f)
				}
				execution.Fields = append(execution.Fields, fv)
			}
		}

		executions = append(executions, execution)
	}

	sort.Sort(ByStartTime(executions))

	legend := []string{"key", "runTime"}
	if *fields != "" {
		for _, f := range strings.Split(*fields, ",") {
			legend = append(legend, f)
		}
	}
	fmt.Println(strings.Join(legend, "  "))
	for _, execution := range executions {
		fmt.Println(execution.String())
	}
}

func createClient() (*redis.Client, error) {
	c := redis.NewClient(&redis.Options{
		Addr:     *redisAddr,
		Password: "",
		DB:       0,
	})

	if _, err := c.Ping().Result(); err != nil {
		return nil, errors.Wrap(err, "could not connect to redis")
	}
	return c, nil
}

func filterSlice(ss []string, test func(string) bool) (ret []string) {
	for _, s := range ss {
		if test(s) {
			ret = append(ret, s)
		}
	}
	return
}
