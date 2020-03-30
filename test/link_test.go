package test_test

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func run(arg ...string) *os.ProcessState {
	io.WriteString(GinkgoWriter, "Start execution of homemaker with args ["+strings.Join(arg, ", ")+"]\n\n")
	cmd := exec.Command("homemaker", arg...)
	cmd.Stdout = GinkgoWriter
	cmd.Stderr = GinkgoWriter
	Expect(cmd.Start()).To(Succeed())
	Expect(cmd.Wait()).To(Succeed())
	io.WriteString(GinkgoWriter, "\nExecution Done\n")
	return cmd.ProcessState
}

func logCommand(command string, arg ...string) {
	cmd := exec.Command(command, arg...)
	output, err := cmd.CombinedOutput()
	Expect(err).To(BeNil())
	GinkgoWriter.Write(output)
}

func printTree() {
	fmt.Fprintf(GinkgoWriter, "Running tree on HOME=%s\n\n", os.Getenv("HOME"))
	logCommand("tree", "-a", os.Getenv("HOME"))
}

var _ = Describe("Link", func() {
	It("Should exec the program", func() {
		state := run()
		Expect(state.Success()).To(BeTrue())
	})

	It("Should execute the run command", func() {

		printTree()
		state := run("run", "-v")
		Expect(state.Success()).To(BeTrue())

		printTree()

		fileLinked, err := os.Readlink("/home/gopher/.config/app/sample.conf")
		Expect(err).To(BeNil())

		dst, err := os.Lstat(fileLinked)
		Expect(err).To(BeNil())

		srcPath, err := filepath.Abs("./.config/app/sample.conf")
		Expect(err).To(BeNil())
		log.Printf("srcPath %s", srcPath)

		src, err := os.Lstat(srcPath)
		Expect(err).To(BeNil())

		Expect(os.SameFile(src, dst)).To(BeTrue())
	})
})
