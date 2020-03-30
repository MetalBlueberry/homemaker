package test_test

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func run(arg ...string) (*os.ProcessState, error) {
	io.WriteString(GinkgoWriter, "Start execution of homemaker with args ["+strings.Join(arg, ", ")+"]\n\n")
	cmd := exec.Command("homemaker", arg...)
	cmd.Stdout = GinkgoWriter
	cmd.Stderr = GinkgoWriter
	Expect(cmd.Start()).To(Succeed())
	err := cmd.Wait()
	io.WriteString(GinkgoWriter, "\nExecution Done\n")
	return cmd.ProcessState, err
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
		state, err := run()
		Expect(err).To(BeNil())
		Expect(state.Success()).To(BeTrue())
	})

	It("Should execute the run command with the default task", func() {

		printTree()
		Expect(linkToSampleConf).ToNot(BeAnExistingFile())

		By("Running default task")
		state, err := run("run", "-v")
		Expect(err).To(BeNil())
		Expect(state.Success()).To(BeTrue())

		By("Checking output")
		printTree()
		Expect(linkToSampleConf).To(BeAnExistingFile())
		Expect(linkToSampleConf).To(BeALinkOf(sampleConf))

		By("Running unlink")
		state, err = run("run", "-v", "--unlink")
		Expect(err).To(BeNil())
		Expect(state.Success()).To(BeTrue())

		By("Checking that environment is clean")
		printTree()
		Expect(linkToSampleConf).ToNot(BeAnExistingFile())
	})

	It("Should return an error if the referenced file not exists", func() {
		printTree()
		Expect(notExistingConf).ToNot(BeAnExistingFile())

		By("Running task link_unexisting_file")
		state, err := run("run", "-v", "link_unexisting_file")
		Expect(err).ToNot(BeNil())
		Expect(state.Success()).ToNot(BeTrue())
	})

	It("Should return an error if the target file exists", func() {
		printTree()
		Expect(notExistingConf).ToNot(BeAnExistingFile())

		By("Create the file")
		err := ioutil.WriteFile(linkToSampleConf, []byte{}, os.ModePerm)
		Expect(err).To(BeNil())

		By("Checking file exist")
		printTree()
		Expect(linkToSampleConf).To(BeAnExistingFile())

		By("Running default task")
		state, err := run("run", "-v")
		Expect(err).ToNot(BeNil())
		Expect(state.Success()).To(BeFalse())
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
