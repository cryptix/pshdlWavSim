package pshdlWavSim

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"

	"github.com/cryptix/wav"
)

type WavSim struct {
	cmd *exec.Cmd

	input  *wav.WavReader
	output *wav.WavWriter

	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser

	samplesTotal uint32

	bytesWritten, bytesRead int64
	writtenSamplesPtr       *int32

	doneWrite chan error
	doneRead  chan error

	iFile, oFile *os.File
}

func NewWavSimFromFile(binPath string, scaler int, coeffs []int, inputFname, outputFname string) (w *WavSim, err error) {
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

	// create a WavSim based in these files
	sim, err := NewWavSim(binPath, scaler, coeffs, input, output)
	if err != nil {
		return nil, err
	}

	// store file handles for closing after sim
	sim.iFile = inputFile
	sim.oFile = outputFile

	return sim, nil
}

func NewWavSim(binPath string, scaler int, coeffs []int, input *wav.WavReader, output *wav.WavWriter) (w *WavSim, err error) {
	w = &WavSim{
		input:  input,
		output: output,

		doneWrite: make(chan error),
		doneRead:  make(chan error),
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
func (w WavSim) Run() (err error) {
	err = w.cmd.Start()
	if err != nil {
		return err
	}

	go w.stderrLog()

	go w.readPump()
	go w.writePump()

	err = <-w.doneWrite
	if err != nil {
		return err
	}

	err = <-w.doneRead
	if err != nil {
		return err
	}

	err = w.cmd.Wait()
	if err != nil {
		return err
	}

	if w.bytesRead != w.bytesWritten {
		return fmt.Errorf("i/o missmatch - r[%d] w[%d]", w.bytesRead, w.bytesWritten)
	}

	*w.writtenSamplesPtr = int32(w.samplesTotal)
	err = w.output.CloseFile()
	if err != nil {
		return err
	}

	if w.iFile != nil {
		w.iFile.Close()
	}

	if w.oFile != nil {
		w.oFile.Close()
	}

	return nil
}

// writePump pumps samples to wavWriter from sim's stdout
func (w *WavSim) writePump() {
	var (
		err       error
		wavWriter io.Writer
	)
	wavWriter, w.writtenSamplesPtr, err = w.output.GetDumbWriter()
	if err != nil {
		w.doneRead <- err
	}

	w.bytesRead, err = io.Copy(wavWriter, w.stdout)
	if err != nil {
		w.doneRead <- err
	}

	w.doneRead <- nil
}

// readPump pumps samples to sim's stdin
func (w *WavSim) readPump() {
	in, err := w.input.GetDumbReader()
	if err != nil {
		w.doneWrite <- err
		return
	}

	w.bytesWritten, err = io.Copy(w.stdin, in)
	if err != nil {
		w.doneWrite <- err
		return
	}

	err = w.stdin.Close()
	if err != nil {
		w.doneWrite <- err
		return
	}

	w.doneWrite <- nil
}

func (w WavSim) stderrLog() {
	buffedErr := bufio.NewReader(w.stderr)

errLoop:
	for {
		line, err := buffedErr.ReadString('\n')
		switch {
		case err == nil:
			fmt.Print(line)
		case err == io.EOF:
			break errLoop
		default:
			panic(err)
		}
	}

}
