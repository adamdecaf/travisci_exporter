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

const version = "0.2.1-dev"

var (
	defaultInterval, _ = time.ParseDuration("1m")

	timestampFormat = "2006-01-02T15:04:05Z"

	// CLI flags
	flagAddress    = flag.String("address", "0.0.0.0:9099", "HTTP listen address")
	flagConfigFile = flag.String("config.file", "", "Path to file with TravisCI token (in TOML)")
	flagInterval   = flag.Duration("interval", defaultInterval, "Interval to check domains at")
	flagVersion    = flag.Bool("version", false, "Print the rdap_exporter version")

	// Prometheus metrics
	jobDurations = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "travisci_job_duration_seconds",
		Help: "Duration in seconds of each TravisCI job",
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
	var config *config
	if *flagConfigFile == "" {
		log.Fatalf("-config.file is empty")
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
		Limit: 100, // TODO(adam): paginate
	})
	if resp != nil && resp.Body != nil {
		resp.Body.Close()
	}
	if err != nil {
		log.Printf("ERROR: %s from travis-ci api: %v", c.name, err)
	}
	for i := range builds {
		for k := range builds[i].Jobs {
			job, resp, err := c.client.Jobs.Find(context.Background(), builds[i].Jobs[k].Id)
			if resp != nil && resp.Body != nil {
				resp.Body.Close()
			}
			if err != nil {
				continue
			}
			if job.FinishedAt == "" {
				continue // can't measure job duration if it's not finished
			}

			start, err := time.Parse(timestampFormat, job.StartedAt)
			if err != nil {
				continue
			}
			end, err := time.Parse(timestampFormat, job.FinishedAt)
			if err != nil {
				continue
			}

			jobDurations.WithLabelValues(fmt.Sprintf("%d", builds[i].Jobs[k].Id), builds[i].Repository.Slug).Set(float64(end.Sub(start).Seconds()))
		}
	}
}
