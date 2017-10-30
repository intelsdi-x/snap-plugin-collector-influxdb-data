/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2017 Intel Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package influxdbdata

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/influxdb/client/v2"
	log "github.com/sirupsen/logrus"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
)

const (
	//Name of the plugin
	Name = "influxdb-data"
	//Version of the plugin
	Version = 1
	//timeStampPrecision time precision
	timeStampPrecision = "ns"
	//nsSeriesPosition indicates position of series in namespace
	nsSeriesPosition = 2
	//nsDataTypePosition indicates position of data_type in namespace
	nsDataTypePosition = 3
)

var (
	//maxConnectionIdle the maximum time a connection can sit around unused.
	maxConnectionIdle = time.Minute * 30
	//watchConnctionWait defines how frequently idle connections are checked
	watchConnctionWait = time.Minute * 15
	//connPool idicates connection pool
	connPool = make(map[string]*clientConnection)
	//mutex indicates mutex for synchronizing connection pool changes
	mutex = &sync.Mutex{}
)

func init() {
	go watchConnections()
}

//Plugin represents instance of the plugin
type Plugin struct {
}

//GetMetricTypes returns metric types for testing.
func (p *Plugin) GetMetricTypes(cfg plugin.Config) ([]plugin.Metric, error) {
	metrics := []plugin.Metric{
		plugin.Metric{
			Namespace: plugin.NewNamespace("intel", "influxdb-data").
				AddDynamicElement("series", "name of series in influxdb").
				AddDynamicElement("data_type", "type of data").
				AddStaticElement("data"),
			Version: 1,
		},
	}
	return metrics, nil
}

//CollectMetrics collects metrics for testing.
func (p *Plugin) CollectMetrics(mts []plugin.Metric) ([]plugin.Metric, error) {
	metrics := []plugin.Metric{}
	for _, m := range mts {
		config, err := getConfig(m.Config)
		if err != nil {
			return nil, err
		}

		cCon, err := openOrSelectConnection(config)
		if err != nil {
			return nil, err
		}

		query := client.NewQuery(config["query"], config["database"], timeStampPrecision)
		response, err := (*cCon.Conn).Query(query)
		if err != nil || response.Error() != nil {
			errFields := map[string]interface{}{
				"requested_metric": "/" + strings.Join(m.Namespace.Strings(), "/"),
				"query":            config["query"],
				"query_error":      err,
				"response_error":   response.Error(),
			}
			return nil, fmt.Errorf("error in response from InfluxDB, %v", errFields)
		}

		if len(m.Namespace) != 5 {
			return nil, fmt.Errorf("incorrect format of namespace, namespace length %v ", len(m.Namespace))
		}

		//get requested series
		reqSeries := m.Namespace.Element(nsSeriesPosition).Value

		for _, result := range response.Results {
			for _, series := range result.Series {

				seriesName := strings.Replace(series.Name, "/", "_", -1)

				if seriesName == reqSeries || reqSeries == "*" {
					//prepare metric namespace
					ns := make([]plugin.NamespaceElement, len(m.Namespace))
					copy(ns, m.Namespace)
					ns[nsSeriesPosition].Value = seriesName
					ns[nsDataTypePosition].Value = config["data_type"]
					//get values
					for _, val := range series.Values {
						if len(series.Columns) != len(val) {
							warnFields := map[string]interface{}{
								"requested_series": reqSeries,
								"columns":          series.Columns,
								"values":           val,
							}
							log.WithFields(warnFields).
								Warn("incorrect format of response from InfluxDB, " +
									"number of columns should equal number of values")
							continue
						}

						mt := plugin.Metric{
							Namespace: ns,
							Tags:      m.Tags,
							Timestamp: time.Now(),
							Version:   Version,
						}

						for i, columns := range series.Columns {

							switch columns {
							case config["data_type"]:
								v, err := convertType(val[i])
								if err != nil {
									warnFields := map[string]interface{}{
										"err":        err,
										"value_type": reflect.TypeOf(val[i]),
										"namespace":  "/" + strings.Join(mt.Namespace.Strings(), "/"),
									}
									log.WithFields(warnFields).Warn("cannot convert type of value")
									continue
								}
								mt.Data = v
							case "time":
								t, err := val[i].(json.Number).Int64()
								if err != nil {
									warnFields := map[string]interface{}{
										"err":       err,
										"namespace": "/" + strings.Join(mt.Namespace.Strings(), "/"),
									}
									log.WithFields(warnFields).Warn("cannot convert time")
									continue
								}
								//set timestamp only if time is set in the result received from influxdb
								if t != 0 {
									timestamp := time.Unix(0, t)
									m.Timestamp = timestamp
								}
							default:
								m.Tags[columns] = fmt.Sprintf("%s", val[i])
							}
						}
						metrics = append(metrics, mt)
					}
				}
			}
		}
	}

	if len(metrics) == 0 {
		warnFields := map[string]interface{}{
			"requested_metrics": mts,
		}
		log.WithFields(warnFields).Warn("nothing has been collected")
	}

	return metrics, nil
}

//GetConfigPolicy returns the configPolicy for your plugin.
func (p *Plugin) GetConfigPolicy() (plugin.ConfigPolicy, error) {
	policy := plugin.NewConfigPolicy()
	policy.AddNewStringRule([]string{"intel", Name}, "host", true)
	policy.AddNewStringRule([]string{"intel", Name}, "database", true)
	policy.AddNewStringRule([]string{"intel", Name}, "user", true)
	policy.AddNewStringRule([]string{"intel", Name}, "password", true)
	policy.AddNewStringRule([]string{"intel", Name}, "query", true)
	policy.AddNewStringRule([]string{"intel", Name}, "data_type", false, plugin.SetDefaultString("value"))
	return *policy, nil
}

func getConfig(cfg plugin.Config) (map[string]string, error) {
	params := []string{"host", "database", "user", "password", "query", "data_type"}
	config := make(map[string]string, 0)
	for _, param := range params {
		prm, err := cfg.GetString(param)
		if err != nil {
			return nil, fmt.Errorf("%s : %s", err, param)
		}
		config[param] = prm
	}

	if strings.Contains(config["query"], "drop") || strings.Contains(config["query"], "delete") {
		return nil, fmt.Errorf("the plugin is intended to receive data, `drop` and `delete` are not allowed in query, query: %v ", config["query"])
	}

	return config, nil
}

func convertType(val interface{}) (v interface{}, err error) {
	switch val.(type) {
	case json.Number:
		v, err = val.(json.Number).Float64()
		if err != nil {
			return nil, err
		}
	default:
		v = fmt.Sprintf("%s", val)
	}
	return v, nil
}

func watchConnections() {
	for {
		time.Sleep(watchConnctionWait)
		for k, c := range connPool {

			if time.Now().Sub(c.LastUsed) > maxConnectionIdle {
				mutex.Lock()
				//close the connection
				c.closeConnection()
				//remove from the pool
				delete(connPool, k)
				mutex.Unlock()
			}
		}
	}
}

func connectionKey(host string, user string, db string) string {
	return fmt.Sprintf("%s:%s:%s", host, user, db)
}

type clientConnection struct {
	Key      string
	Conn     *client.Client
	LastUsed time.Time
}

func openOrSelectConnection(pluginCfg map[string]string) (*clientConnection, error) {
	key := connectionKey(pluginCfg["host"], pluginCfg["user"], pluginCfg["password"])
	if connPool[key] == nil {
		cfg := client.HTTPConfig{
			Addr:     pluginCfg["host"],
			Username: pluginCfg["user"],
			Password: pluginCfg["password"]}

		con, err := client.NewHTTPClient(cfg)
		if err != nil {
			return nil, err
		}

		cCon := &clientConnection{
			Key:      key,
			Conn:     &con,
			LastUsed: time.Now(),
		}

		//add to the pool
		connPool[key] = cCon
		return connPool[key], nil
	}
	//update when it was accessed
	connPool[key].LastUsed = time.Now()
	return connPool[key], nil
}

func (c *clientConnection) closeConnection() error {
	return (*c.Conn).Close()
}
