package test_test

import (
	"errors"
	"fmt"
	"io"
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
	var (
		linkToSampleConf = "/home/gopher/.config/app/sample.conf"
		sampleConf       = "./.config/app/sample.conf"
		notExistingConf  = "./.config/app/not_exist.conf"
	)

	It("Should exec the program", func() {
		state := run()
		Expect(state.Success()).To(BeTrue())
	})

	It("Should execute the run command with the default task", func() {

		printTree()
		Expect(linkToSampleConf).ToNot(BeAnExistingFile())

		By("Running default task")
		state := run("run", "-v")
		Expect(state.Success()).To(BeTrue())

		By("Checking output")
		printTree()
		Expect(linkToSampleConf).To(BeAnExistingFile())
		Expect(linkToSampleConf).To(BeALinkOf(sampleConf))

		By("Running unlink")
		state = run("run", "-v", "--unlink")
		Expect(state.Success()).To(BeTrue())

		By("Checking that environment is clean")
		printTree()
		Expect(linkToSampleConf).ToNot(BeAnExistingFile())
	})
	It("Should return an error if the referenced file not exists", func() {
		printTree()
		Expect(notExistingConf).ToNot(BeAnExistingFile())

		By("Running task link_unexisting_file")
		state := run("run", "-v", "link_unexisting_file")
		Expect(state.Success()).ToNot(BeTrue())
	})
})

type fileLinkMatcher struct {
	source string
	target string
}

func BeALinkOf(path string) *fileLinkMatcher {
	return &fileLinkMatcher{
		source: path,
	}
}

func (fm *fileLinkMatcher) Match(actual interface{}) (success bool, err error) {
	linkPath, ok := actual.(string)
	if !ok {
		return false, errors.New("Value is not string")
	}

	fileLink, err := os.Readlink(linkPath)
	if err != nil {
		return false, err
	}
	fm.target = fileLink

	dst, err := os.Lstat(fileLink)
	if err != nil {
		return false, err
	}
	src, err := os.Lstat(fm.source)
	if err != nil {
		return false, err
	}

	return os.SameFile(src, dst), nil

}

func (fm *fileLinkMatcher) FailureMessage(actual interface{}) (message string) {
	sourcePath, err := filepath.Abs(fm.source)
	if err != nil {
		panic(err)
	}
	targetPath, err := filepath.Abs(fm.target)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("File %s is not linked by %s", targetPath, sourcePath)
}

func (fm *fileLinkMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	sourcePath, err := filepath.Abs(fm.source)
	if err != nil {
		panic(err)
	}
	targetPath, err := filepath.Abs(fm.target)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("File %s is linked by %s", targetPath, sourcePath)
}
