package test_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

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

type InteractiveExecCommand struct {
	io.WriteCloser
	io.ReadCloser
	Command *exec.Cmd
	Output  *bytes.Buffer
}

func interactive(arg ...string) InteractiveExecCommand {
	fmt.Fprintf(GinkgoWriter, "Start interactive execution of homemaker with args %s\n\n", arg)

	outBuffer := &bytes.Buffer{}

	cmd := exec.Command("homemaker", arg...)

	outWritter := io.MultiWriter(GinkgoWriter, outBuffer)
	cmd.Stdout = outWritter
	cmd.Stderr = outWritter

	r, w := io.Pipe()
	icmd := InteractiveExecCommand{
		WriteCloser: w,
		ReadCloser:  r,
		Command:     cmd,
		Output:      outBuffer,
	}
	cmd.Stdin = icmd

	Expect(cmd.Start()).To(Succeed())
	return icmd
}

func (i InteractiveExecCommand) Close() (err error) {
	err = i.WriteCloser.Close()
	if err != nil {
		return
	}
	err = i.ReadCloser.Close()
	return
}

func (i InteractiveExecCommand) interact(msg string) InteractiveExecCommand {
	fmt.Fprintf(GinkgoWriter, "Typing %s \n", msg)
	fmt.Fprint(i, msg)
	return i
}
func (i InteractiveExecCommand) wait(duration time.Duration) InteractiveExecCommand {
	fmt.Fprintf(GinkgoWriter, "Giving %s to the command to progress\n", duration)
	time.Sleep(duration)
	return i
}

func (i InteractiveExecCommand) complete() (*os.ProcessState, error) {
	err := i.Close()
	if err != nil {
		return nil, err
	}
	err = i.Command.Wait()
	fmt.Fprintf(GinkgoWriter, "\nExecution Done\n")
	return i.Command.ProcessState, err
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

	It("Should return ask user what to do in case of file already exist", func() {
		printTree()
		Expect(notExistingConf).ToNot(BeAnExistingFile())

		By("Create the file")
		err := ioutil.WriteFile(linkToSampleConf, []byte{}, os.ModePerm)
		Expect(err).To(BeNil())

		By("Checking file exist")
		printTree()
		Expect(linkToSampleConf).To(BeAnExistingFile())

		By("Running default task")
		cmd := interactive("run", "-v").wait(time.Millisecond * 100)

		By("Answer no to clobber")
		printTree()
		cmd.interact("n\n").wait(time.Millisecond * 100)
		Expect(linkToSampleConf).To(BeAnExistingFile())

		By("Answer Abort to failed task")
		printTree()
		cmd.interact("a\n").wait(time.Millisecond * 100)
		Expect(linkToSampleConf).To(BeAnExistingFile())

		state, err := cmd.complete()
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
