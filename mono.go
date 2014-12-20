package pshdlWavSim

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/cryptix/wav"
)

func NewMonoFIRFromFile(binPath string, scaler int, coeffs []int, inputFname, outputFname string) error {
	// open input file
	// BUG(Henry): close this file after run
	inputFile, err := os.Open(inputFname)
	if err != nil {
		return err
	}

	// get stat for file Size
	inputStat, err := inputFile.Stat()
	if err != nil {
		return err
	}

	// create wav.Reader
	input, err := wav.NewReader(inputFile, inputStat.Size())
	if err != nil {
		return err
	}

	// create output file
	outputFile, err := os.Create(outputFname)
	if err != nil {
		return err
	}

	// get meta information from input
	meta := input.GetFile()

	// create output with the same characteristics
	output, err := meta.NewWriter(outputFile)
	if err != nil {
		return err
	}

	return NewMonoFIR(binPath, scaler, coeffs, input, output)
}

func NewMonoFIR(binPath string, scaler int, coeffs []int, input *wav.Reader, output *wav.Writer) error {
	var err error

	// construct arguments
	args := make([]string, len(coeffs)+2)
	args[0] = fmt.Sprintf("%d", input.GetSampleCount())
	args[1] = strconv.Itoa(scaler)
	for i, c := range coeffs {
		args[i+2] = strconv.Itoa(c)
	}

	cmd := exec.Command(binPath, args...)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	in, err := input.GetDumbReader()
	if err != nil {
		return err
	}

	out, writtenPtr, err := output.GetDumbWriter()
	if err != nil {
		return err
	}
	*writtenPtr = int32(input.GetSampleCount())

	if err := cmd.Start(); err != nil {
		return err
	}

	rc := make(chan copyJob)
	wc := make(chan copyJob)
	go copyPump(rc, in, stdin)
	go copyPump(wc, stdout, out)

	rjob := <-rc
	if rjob.Err != nil {
		return fmt.Errorf("inCopy failed: %q", rjob.Err)
	}

	wjob := <-wc
	if wjob.Err != nil {
		return fmt.Errorf("outCopy failed: %q", wjob.Err)
	}

	if err = cmd.Wait(); err != nil {
		return err
	}

	if s := stderr.String(); s != "" {
		return fmt.Errorf("Stderr not empty: %q", s)
	}

	if wjob.N != rjob.N {
		return fmt.Errorf("i/o missmatch - r[%d] w[%d]", rjob.N, wjob.N)
	}

	return nil
}
