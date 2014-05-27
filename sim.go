package pshdlWavSim

import (
	"bufio"
	"fmt"
	"io"
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
}

func NewWavSim(input *wav.WavReader, output *wav.WavWriter) (w *WavSim, err error) {
	w = &WavSim{
		input:  input,
		output: output,

		doneWrite: make(chan error),
		doneRead:  make(chan error),
	}

	w.samplesTotal = input.GetSampleCount()

	w.cmd = exec.Command("./sim", strconv.Itoa(int(w.samplesTotal)))

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

	go w.stderrLog()

	return w, nil
}

// starts executing the Simulation
func (w WavSim) Run() (err error) {
	err = w.cmd.Start()
	if err != nil {
		return err
	}

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
