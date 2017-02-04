package filter

import (
	"strings"

	"github.com/pkg/errors"
	redis "gopkg.in/redis.v5"
)

// Filter is a Key/Value pair
type Filter struct {
	Key   string
	Value string
}

// Parse will convert a comma-delimited list of Key=Value pairs into a slice of
// Filter objects.
func Parse(userInput string, status string) ([]Filter, error) {
	filters := []Filter{{"status", status}}
	if userInput != "" {
		for _, p := range strings.Split(userInput, ",") {
			kv := strings.Split(p, "=")
			if len(kv) != 2 {
				return nil, errors.Errorf("invalid filter format '%s', require 'Key=Value'", p)
			}
			filters = append(filters, Filter{Key: kv[0], Value: kv[1]})
		}
	}

	return filters, nil
}

// Func defines the interface of a filtering function.
type Func func(c *redis.Client, key string, f Filter) (bool, error)

// Get returns the applicable Func given a Filter struct.
func Get(f Filter) Func {
	switch f.Key {
	case "ContainsStage":
		return containsStageTypeFilter
	default:
		return hashKeyFilter
	}
}

func hashKeyFilter(c *redis.Client, key string, f Filter) (bool, error) {
	v, err := c.HGet(key, f.Key).Result()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, errors.Wrapf(err, "failed getting hash field '%s' from hash '%s'", f.Key, key)
	}
	return v == f.Value, nil
}

func containsStageTypeFilter(c *redis.Client, key string, f Filter) (bool, error) {
	allKeys, err := c.HKeys(key).Result()
	if err != nil {
		return false, errors.Wrapf(err, "failed getting all hash keys from hash '%s'", key)
	}
	var stageTypeKeys []string
	for _, k := range allKeys {
		if strings.HasPrefix(k, "stage.") && strings.HasSuffix(k, ".type") {
			stageTypeKeys = append(stageTypeKeys, k)
		}
	}

	for _, k := range stageTypeKeys {
		stageType, err := c.HGet(key, k).Result()
		if err != nil {
			return false, errors.Wrapf(err, "failed getting hash field '%s' from hash '%s'", k, key)
		}
		if stageType == f.Value {
			return true, nil
		}
	}
	return false, nil
}
