// Copyright 2019 Adam Shannon
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/shuheiktgw/go-travis"
	"gopkg.in/yaml.v2"
)

const version = "0.1.0"

var (
	defaultInterval, _ = time.ParseDuration("1m")

	// CLI flags
	flagAddress    = flag.String("address", "0.0.0.0:9099", "HTTP listen address")
	flagConfigFile = flag.String("config.file", "config.yaml", "Path to file with TravisCI token (in TOML)")
	flagInterval   = flag.Duration("interval", defaultInterval, "Interval to check domains at")
	flagVersion    = flag.Bool("version", false, "Print the rdap_exporter version")

	// Prometheus metrics
	jobDurations = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "travisci_job_duration_seconds",
		Help:    "Duration in seconds of each TravisCI job",
		Buckets: []float64{5.0, 10.0, 20.0, 30.0, 60.0, 300.0, 600.0},
	}, []string{"id", "slug"})
)

func init() {
	prometheus.MustRegister(jobDurations)
}

func main() {
	flag.Parse()

	// Flags that quit after running
	if *flagVersion {
		fmt.Println(version)
		return
	}

	log.Printf("Starting travisci_exporter:%s", version)

	// Read our config file
	var config *Config
	if *flagConfigFile == "" {
		log.Println("-config.file is empty so using example config")
		config = &Config{
			Organizations: []Organization{
				{
					Name:  "adamdecaf",
					Token: "",
				},
			},
		}
	} else {
		bs, err := ioutil.ReadFile(*flagConfigFile)
		if err != nil {
			log.Fatalf("problem reading %s: %v", *flagConfigFile, err)
		}
		if err := yaml.Unmarshal(bs, &config); err != nil {
			log.Fatalf("problem unmarshaling %s: %v", *flagConfigFile, err)
		}
	}

	for i := range config.Organizations {
		org := config.Organizations[i]

		var client *travis.Client
		if org.UseOrg {
			client = travis.NewClient(travis.ApiOrgUrl, org.Token)
		} else {
			client = travis.NewClient(travis.ApiComUrl, org.Token)
		}
		check := &checker{
			name:     org.Name,
			client:   client,
			interval: *flagInterval,
		}
		go check.checkAll()
	}

	// Add Prometheus metrics HTTP handler
	h := promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{})
	http.Handle("/metrics", h)

	// Block on HTTP server
	log.Printf("listenting on %s", *flagAddress)
	if err := http.ListenAndServe(*flagAddress, nil); err != nil {
		log.Fatalf("ERROR binding to %s: %v", *flagAddress, err)
	}
}

type checker struct {
	name   string
	client *travis.Client

	t        *time.Ticker
	interval time.Duration
}

func (c *checker) checkAll() {
	if c.t == nil {
		c.t = time.NewTicker(c.interval)
		c.checkNow() // check domains right away after ticker setup
	}
	for range c.t.C {
		c.checkNow()
	}
}

func (c *checker) checkNow() {
	builds, resp, err := c.client.Builds.List(context.Background(), &travis.BuildsOption{
		Limit: 10, // TODO(adam): pagination of all?
	})
	if resp != nil && resp.Body != nil {
		resp.Body.Close()
	}
	if err != nil {
		log.Printf("ERROR: %s from travis-ci api: %v", c.name, err)
	}
	for i := range builds {
		for k := range builds[i].Jobs {
			jobDurations.WithLabelValues(fmt.Sprintf("%d", builds[i].Jobs[k].Id), builds[i].Repository.Slug).Observe(float64(builds[i].Duration))
		}
	}
}
