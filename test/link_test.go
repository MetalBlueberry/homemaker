package test_test

import (
	"io"
	"os/exec"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func run(arg ...string) (string, error) {
	cmd := exec.Command("homemaker", arg...)
	output, err := cmd.CombinedOutput()
	io.WriteString(GinkgoWriter, "Start execution of homemaker with args ["+strings.Join(arg, ", ")+"]\n\n")
	GinkgoWriter.Write(output)
	io.WriteString(GinkgoWriter, "\nExecution Done\n")
	return string(output), err
}

var _ = Describe("Link", func() {
	It("Should exect the program", func() {
		output, err := run()
		Expect(err).To(BeNil())
		Expect(output).ToNot(BeNil())
	})

	It("Should execute the run command", func() {
		output, err := run("run")
		Expect(err).To(BeNil())
		Expect(output).ToNot(BeNil())
	})
})
