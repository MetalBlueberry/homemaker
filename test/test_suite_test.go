package test_test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestTest(t *testing.T) {
	if _, exist := os.LookupEnv("HOMEMAKER_DOCKER_TEST_ENV"); !exist {
		t.Log("This test must be run inside Docker test environment defined by the Docker file in this repository")
		t.SkipNow()
	}
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test Suite")
}
