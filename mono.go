package pshdlWavSim

import (
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
	stderr io.ReadCloser

	samplesTotal uint32

	bytesWritten, bytesRead int64
	writtenSamplesPtr       *int32

	iFile, oFile *os.File
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

	// store file handles for closing after sim
	sim.iFile = inputFile
	sim.oFile = outputFile

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

	w.stdin, err = w.cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	w.stdout, err = w.cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	w.stderr, err = w.cmd.StderrPipe()
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

	if err := w.cmd.Start(); err != nil {
		return err
	}

	rc := make(chan copyJob)
	wc := make(chan copyJob)

	go stderrLog(w.stderr)
	go copyPump(rc, in, w.stdin)
	go copyPump(wc, w.stdout, nopCloseWriter{out})

	wjob := <-wc
	if wjob.Err != nil {
		return wjob.Err
	}

	rjob := <-rc
	if rjob.Err != nil {
		return rjob.Err
	}

	if err = w.cmd.Wait(); err != nil {
		return err
	}

	if wjob.N != rjob.N {
		return fmt.Errorf("i/o missmatch - r[%d] w[%d]", rjob.N, wjob.N)
	}

	*writtenPtr = int32(w.samplesTotal)
	err = w.output.Close()
	if err != nil {
		return err
	}

	if w.iFile != nil {
		if err := w.iFile.Close(); err != nil {
			return err
		}
	}

	if w.oFile != nil {
		if err := w.oFile.Close(); err != nil {
			return err
		}
	}

	return nil
}
