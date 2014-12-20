package pshdlWavSim

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"

	"github.com/cryptix/wav"
)

type MonoFIR struct {
	cmd *exec.Cmd

	input  *wav.WavReader
	output *wav.WavWriter

	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr bytes.Buffer

	samplesTotal uint32

	bytesWritten, bytesRead int64
	writtenSamplesPtr       *int32
}

func NewMonoFIRFromFile(binPath string, scaler int, coeffs []int, inputFname, outputFname string) (w *MonoFIR, err error) {
	// open input file
	inputFile, err := os.Open(inputFname)
	if err != nil {
		return nil, err
	}

	// get stat for file Size
	inputStat, err := inputFile.Stat()
	if err != nil {
		return nil, err
	}

	// create wav.Reader
	input, err := wav.NewWavReader(inputFile, inputStat.Size())
	if err != nil {
		return nil, err
	}

	// create output file
	outputFile, err := os.Create(outputFname)
	if err != nil {
		return nil, err
	}

	// get meta information from input
	meta := input.GetWavFile()

	// create output with the same characteristics
	output, err := meta.NewWriter(outputFile)
	if err != nil {
		return nil, err
	}

	// create a MonoFIR based in these files
	sim, err := NewMonoFIR(binPath, scaler, coeffs, input, output)
	if err != nil {
		return nil, err
	}

	return sim, nil
}

func NewMonoFIR(binPath string, scaler int, coeffs []int, input *wav.WavReader, output *wav.WavWriter) (w *MonoFIR, err error) {
	w = &MonoFIR{
		input:  input,
		output: output,
	}

	w.samplesTotal = input.GetSampleCount()

	// construct arguments
	args := make([]string, len(coeffs)+2)
	args[0] = strconv.Itoa(int(w.samplesTotal))
	args[1] = strconv.Itoa(scaler)
	for i, c := range coeffs {
		args[i+2] = strconv.Itoa(c)
	}

	w.cmd = exec.Command(binPath, args...)

	w.cmd.Stderr = &w.stderr
	w.stdin, err = w.cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	w.stdout, err = w.cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	return w, nil
}

// starts executing the Simulation
func (w MonoFIR) Run() (err error) {

	in, err := w.input.GetDumbReader()
	if err != nil {
		return err
	}

	out, writtenPtr, err := w.output.GetDumbWriter()
	if err != nil {
		return err
	}
	*writtenPtr = int32(w.samplesTotal)

	if err := w.cmd.Start(); err != nil {
		return err
	}

	rc := make(chan copyJob)
	wc := make(chan copyJob)
	go copyPump(rc, in, w.stdin)
	go copyPump(wc, w.stdout, out)

	wjob := <-wc
	if wjob.Err != nil {
		return fmt.Errorf("outCopy failed: %q", wjob.Err)
	}

	rjob := <-rc
	if rjob.Err != nil {
		return fmt.Errorf("inCopy failed: %q", rjob.Err)
	}

	if err = w.cmd.Wait(); err != nil {
		return err
	}

	if s := w.stderr.String(); s != "" {
		return fmt.Errorf("Stderr not empty: %q", s)
	}

	if wjob.N != rjob.N {
		return fmt.Errorf("i/o missmatch - r[%d] w[%d]", rjob.N, wjob.N)
	}

	return nil
}
