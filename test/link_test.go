package test_test

import (
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func run(arg ...string) (string, error) {
	cmd := exec.Command("homemaker", arg...)
	output, err := cmd.CombinedOutput()

	return string(output), err
}

var _ = Describe("Link", func() {
	It("Should exect the program", func() {
		output, err := run()
		Expect(err).To(BeNil())
		Expect(output).ToNot(BeNil())
	})
})
