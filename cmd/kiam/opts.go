// Copyright 2017 uSwitch
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package main

import (
	"context"
	"net"
	"time"

	statsd "github.com/pubnub/go-metrics-statsd"
	metrics "github.com/rcrowley/go-metrics"
	log "github.com/sirupsen/logrus"
	"github.com/uswitch/kiam/pkg/prometheus"
)

type logOptions struct {
	jsonLog  bool
	logLevel string
}

func (o *logOptions) bind(parser parser) {
	parser.Flag("json-log", "Output log in JSON").BoolVar(&o.jsonLog)
	parser.Flag("level", "Log level: debug, info, warn, error.").Default("info").EnumVar(&o.logLevel, "debug", "info", "warn", "error")
}

func (o *logOptions) configureLogger() {
	if o.jsonLog {
		log.SetFormatter(&log.JSONFormatter{})
	}

	switch o.logLevel {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	}
}

type telemetryOptions struct {
	statsd           string
	statsdInterval   time.Duration
	prometheusListen string
	prometheusSync   time.Duration
}

func (o *telemetryOptions) bind(parser parser) {
	parser.Flag("statsd", "UDP address to publish StatsD metrics. e.g. 127.0.0.1:8125").Default("").StringVar(&o.statsd)
	parser.Flag("statsd-interval", "Interval to publish to StatsD").Default("10s").DurationVar(&o.statsdInterval)

	parser.Flag("prometheus-listen-addr", "Prometheus HTTP listen address. e.g. localhost:9620").StringVar(&o.prometheusListen)
	parser.Flag("prometheus-sync-interval", "How frequently to update Prometheus metrics").Default("5s").DurationVar(&o.prometheusSync)
}

func (o telemetryOptions) start(ctx context.Context, identifier string) {
	if o.statsd != "" {
		addr, err := net.ResolveUDPAddr("udp", o.statsd)
		if err != nil {
			log.Fatal("error parsing statsd address:", err.Error())
		}
		go statsd.StatsD(metrics.DefaultRegistry, o.statsdInterval, "kiam."+identifier, addr)
	}

	if o.prometheusListen != "" {
		metrics := prometheus.NewServer(identifier, o.prometheusListen, o.prometheusSync)
		metrics.Listen(ctx)
	}
}

type tlsOptions struct {
	certificatePath string
	keyPath         string
	caPath          string
}

func (o *tlsOptions) bind(parser parser) {
	parser.Flag("cert", "Certificate path").Required().ExistingFileVar(&o.certificatePath)
	parser.Flag("key", "Key path").Required().ExistingFileVar(&o.keyPath)
	parser.Flag("ca", "CA certificate path").Required().ExistingFileVar(&o.caPath)
}

type clientOptions struct {
	serverAddress        string
	serverAddressRefresh time.Duration
	timeoutKiamGateway   time.Duration
}

func (o *clientOptions) bind(parser parser) {
	parser.Flag("server-address", "gRPC address to Kiam server service").Default("localhost:9610").StringVar(&o.serverAddress)
	parser.Flag("server-address-refresh", "Interval to refresh server service endpoints").Default("10s").DurationVar(&o.serverAddressRefresh)
	parser.Flag("gateway-timeout-creation", "Timeout to create the kiam gateway ").Default("50ms").DurationVar(&o.timeoutKiamGateway)
}
