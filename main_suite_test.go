package main_test

import (
	"io/ioutil"

	log "github.com/sirupsen/logrus"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestHussar(t *testing.T) {
	log.SetOutput(ioutil.Discard)

	RegisterFailHandler(Fail)
	RunSpecs(t, "Main Suite")
}
