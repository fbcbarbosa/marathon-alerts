package main

import (
	"github.com/gambol99/go-marathon"
)

type Checker interface {
	Name() string
	Check(marathon.Application) AppCheck
}
