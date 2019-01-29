// Copyright 2018 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"
	"strings"

	travis "github.com/kevinburke/travis/lib"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const version = "0.1.0-dev"

var (
	defaultInterval, _ = time.ParseDuration("1m")

	// CLI flags
	flagAddress    = flag.String("address", "0.0.0.0:9099", "HTTP listen address")
	flagConfigFile = flag.String("config", "~/.travis", "Path to file with TravisCI token (in TOML)")
	flagNames = flag.String("names", "", "TravisCI usernames or organizations")
	flagInterval   = flag.Duration("interval", defaultInterval, "Interval to check domains at")
	flagVersion    = flag.Bool("version", false, "Print the rdap_exporter version")

	// Prometheus metrics
	jobDurations = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "travisci_job_duration_seconds",
		Help: "Duration in seconds of each TravisCI job",
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
	if *flagNames == "" {
		fmt.Println("-names is required")
		return
	}

	log.Printf("Starting travisci_exporter:%s", version)

	names := strings.Split(*flagNames, ",")
	for i := range names {
		// Setup TravisCI client
		token, err := travis.GetToken(names[i])
		if err != nil {
			log.Fatal(err)
		}
		client := travis.NewClient(token)

		check := &checker{
			client: client,
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
	client *travis.Client

	t *time.Ticker
	interval time.Duration
}

func (c *checker) checkAll() {
	if c.t == nil {
		c.t = time.NewTicker(c.interval)
		c.checkNow(c.client) // check domains right away after ticker setup
	}
	for range c.t.C {
		c.checkNow(c.client)
	}
}

func (c *checker) checkNow(client *travis.Client) {
	req, err := client.NewRequest("GET", "/repo/moov-io%2Fach/builds?branch.name=master", nil)
	if err != nil {
		log.Printf("ERROR: checkNow: %v", err)
	}
	req = req.WithContext(context.TODO())

	builds := make([]*travis.Build, 0)
	resp := &travis.ListResponse{
		Data: &builds,
	}
	if err := client.Do(req, resp); err != nil {
		log.Printf("ERROR: checkNow: %v", err)
	}

	for i := range builds {
		dur, _ := time.ParseDuration(fmt.Sprintf("%ds", builds[i].Duration))
		for k := range builds[i].Jobs {
			id := fmt.Sprintf("%d", builds[i].Jobs[k].ID)
			jobDurations.WithLabelValues(id, builds[i].Repository.Slug).Observe(dur.Seconds())
		}
	}
}

// func readDomainFile(where string) ([]string, error) {
// 	fullPath, err := filepath.Abs(where)
// 	if err != nil {
// 		return nil, fmt.Errorf("when expanding %s: %v", *flagDomainFile, err)
// 	}

// 	fd, err := os.Open(fullPath)
// 	if err != nil {
// 		return nil, fmt.Errorf("when opening %s: %v", fullPath, err)
// 	}
// 	defer fd.Close()
// 	r := bufio.NewScanner(fd)

// 	var domains []string
// 	for r.Scan() {
// 		domains = append(domains, strings.TrimSpace(r.Text()))
// 	}
// 	if len(domains) == 0 {
// 		return nil, fmt.Errorf("no domains found in %s", fullPath)
// 	}
// 	return domains, nil
// }
