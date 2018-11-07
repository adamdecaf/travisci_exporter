// Copyright 2018 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const version = "0.1.0-dev"

var (
	defaultInterval, _ = time.ParseDuration("1m")

	// CLI flags
	flagAddress    = flag.String("address", "0.0.0.0:9099", "HTTP listen address")
	flagConfigFile = flag.String("config", "", "Path to file with TravisCI token (in TOML)")
	flagInterval   = flag.Duration("interval", defaultInterval, "Interval to check domains at")
	flagVersion    = flag.Bool("version", false, "Print the rdap_exporter version")

	// TODO(adam): define metrics
	// domainExpiration = prometheus.NewGaugeVec(
	// 	prometheus.GaugeOpts{
	// 		Name: "domain_expiration",
	// 		Help: "Days until the RDAP expiration event states this domain will expire",
	// 	},
	// 	[]string{"domain"},
	// )
)

func init() {
	// prometheus.MustRegister(domainExpiration)
}

func main() {
	flag.Parse()

	if *flagVersion {
		fmt.Println(version)
		return
	}
	log.Printf("Starting moov-io/travisci_exporter:%s", version)

	// Sanity checks
	if *flagConfigFile == "" {
		log.Fatal("no -config file specified")
	}
	check := &checker{
		// TODO(adam): add travis-ci client
		interval: *flagInterval,
	}
	go check.checkAll()

	// Add Prometheus metrics HTTP handler
	h := promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{})
	http.Handle("/metrics", h)

	log.Printf("listenting on %s", *flagAddress)
	if err := http.ListenAndServe(*flagAddress, nil); err != nil {
		log.Fatalf("ERROR binding to %s: %v", *flagAddress, err)
	}
}

type checker struct {
	t *time.Ticker
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
	log.Println("checking...")
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
