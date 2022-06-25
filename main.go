package main

import (
	"context"
	"fmt"
	"log"
	"math"

	"github.com/go-redis/redis/v8"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-plugin-sdk/sensu"
)

// Config represents a handler config.
type Config struct {
	sensu.PluginConfig
	Host     string
	Port     int
	Password string
}

const (
	host     = "host"
	port     = "port"
	password = "password"
)

var (
	handlerConfig = Config{
		PluginConfig: sensu.PluginConfig{
			Name:     "sensu-metrics-stats-handler",
			Short:    "Calculates and stores metric statistics in Redis for use with metrics-top tool",
			Keyspace: "sensu.io/plugins/sensu-metrics-stats-handler/config",
		},
	}

	opts = []sensu.ConfigOption{
		&sensu.PluginConfigOption[string]{
			Path:      host,
			Env:       "REDIS_HOST",
			Argument:  host,
			Shorthand: "",
			Default:   "127.0.0.1",
			Usage:     "the host address of the Redis server",
			Value:     &handlerConfig.Host,
		},
		&sensu.PluginConfigOption[int]{
			Path:      port,
			Env:       "REDIS_PORT",
			Argument:  port,
			Shorthand: "",
			Default:   6379,
			Usage:     "the port of the Redis server",
			Value:     &handlerConfig.Port,
		},
		&sensu.PluginConfigOption[string]{
			Path:      password,
			Env:       "REDIS_PASSWORD",
			Argument:  password,
			Shorthand: "",
			Default:   "",
			Usage:     "the password for the Redis server",
			Value:     &handlerConfig.Password,
		},
	}
)

func main() {
	handler := sensu.NewGoHandler(&handlerConfig.PluginConfig, opts, checkArgs, executeHandler)
	handler.Execute()
}

func checkArgs(event *corev2.Event) error {
	if !event.HasMetrics() {
		return fmt.Errorf("event does not contain metrics")
	}
	return nil
}

func executeHandler(event *corev2.Event) error {
	if len(event.Metrics.Points) == 0 {
		log.Println("event does not contain metric points")
		return nil
	}

	log.Printf("connection info: %s:%d", handlerConfig.Host, handlerConfig.Port)

	ctx := context.TODO()
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", handlerConfig.Host, handlerConfig.Port),
		Password: handlerConfig.Password,
		DB:       0, // use default DB
	})

	failureCount := 0
	for _, point := range event.Metrics.Points {
		// prefix metric name if provided as config option
		err := storeMetricPoint(client, point.Name, point.Value, point.Timestamp, point.Tags)
		if err != nil {
			log.Printf("error sending metric: %s", err)
			failureCount++
		}
	}
	log.Printf("sent %d metric points with %d failures", len(event.Metrics.Points), failureCount)

	// publish event to the follow-metrics channel using redis pub/sub
	eventJSON, _ := event.MarshalJSON()
	err := client.Publish(ctx, "follow-metrics", eventJSON).Err()
	if err != nil {
		log.Printf("error sending event to pub/sub: %s", err)
	}

	return err
}

func storeMetricPoint(client *redis.Client, name string, value float64, timestamp int64, tags []*corev2.MetricTag) error {

	// store metric name and timestamp
	storeMetricName(client, name, timestamp)

	// add/update the stats entry for this metric

	// max, min, current value
	err := storeStats(client, name, value)
	if err != nil && err != redis.Nil {
		log.Printf("error recording stats: %s", err)
	}

	return err
}

func storeMetricName(client *redis.Client, metricName string, timestamp int64) error {
	ctx := context.TODO()

	// add/update the metric name to a sorted set, ranked by current time
	err := client.ZAdd(ctx, "metrics", &redis.Z{
		Score:  float64(secTimestamp(timestamp)),
		Member: metricName,
	}).Err()

	if err != nil {
		log.Printf("error recording metric: %s", err)
	}

	return err
}

func storeStats(client *redis.Client, metricName string, observation float64) error {

	ctx := context.TODO()

	var statsScript = redis.NewScript(`
local key, value = KEYS[1], tonumber(ARGV[1]);

local min;
local max;
local mean;
local count;
local sumOfSquares;
local sum;
local stddev;
local variance;

local values = redis.call('HMGET', key, 'min', 'max', 'mean', 'count', 'sumOfSquares', 'sum');

if(tonumber(values[1]) == nil) then
  min = 9223372036854775807;
  max = -9223372036854775808;
  mean = 0.0
  stddev = 0.0
  variance = 0.0
  sumOfSquares = 0.0
  sum = 0.0
  count = 0
else
  min = math.min(value, tonumber(values[1]));
  max = math.max(value, tonumber(values[2]));
  mean = tonumber(values[3]);
  count = tonumber(values[4]) + 1;
  sumOfSquares = tonumber(values[5]) + value * value;
  sum = tonumber(values[6]) + value;
  stddev = 0.0;
  variance = 0.0;
end;

if(count > 1) then
  mean = mean + (value  - mean) / count;
  
  stddev = math.sqrt((count * sumOfSquares - sum * sum) / (count * (count -1)));
  variance = stddev * stddev;
else
  mean = value;
end;

redis.call('HMSET', key, 'min', min, 'max', max, 'current', value, 'mean', mean, 'variance', variance, 'stddev', stddev, 'count', count, 'sum', sum, 'sumOfSquares', sumOfSquares);

if(ARGV[2]=='get_stats') then
  return {'current', value, 'min', min, 'max', max, 'mean', mean, 'stddev', stddev, 'variance', variance, 'sum', sum, 'count', count};
end;

return true

`)
	keys := []string{fmt.Sprintf("%s.stats", metricName)}

	_, err := statsScript.Run(ctx, client, keys, observation).Int()

	return err
}

// msTimestamp auto-detection of metric point timestamp precision using a heuristic with a 250-ish year cutoff
func secTimestamp(ts int64) int64 {
	timestamp := ts
	switch ts := math.Log10(float64(timestamp)); {
	case ts < 10:
		// assume timestamp is seconds
	case ts < 13:
		// assume timestamp is milliseconds
		timestamp = (timestamp / 1e3)
	case ts < 16:
		// assume timestamp is microseconds
		timestamp = (timestamp / 1e6)
	default:
		// assume timestamp is nanoseconds
		timestamp = (timestamp / 1e9)
	}

	return timestamp
}
