// Copyright 2019 Adam Shannon
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

type config struct {
	Organizations []organization `yaml:"organizations"`
}

type organization struct {
	Name  string `yaml:"name"`
	Token string `yaml:"token"`

	UseOrg bool `yaml:"org,omitempty"`
}
