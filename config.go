// Copyright 2019 Adam Shannon
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

type Config struct {
	Organizations []Organization `yaml:"organizations"`
}

type Organization struct {
	Name  string `yaml:"name"`
	Token string `yaml:"token"`
}
