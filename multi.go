package pshdlWavSim

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/cryptix/wav"
)

// NewMultiFromFile does the preperation and then calls NewMulti
func NewMultiFromFile(binPath string, inputFnames, outputFnames []string) error {
	var iwav *wav.Reader

	if len(inputFnames) < 1 || len(outputFnames) < 1 {
		return errors.New("need at least one input and output filename")
	}

	files := make([]*os.File, len(inputFnames))
	// prepare input files
	for i, iname := range inputFnames {

		f, err := os.Open(iname)
		if err != nil {
			return err
		}

		s, err := f.Stat()
		if err != nil {
			return err
		}

		iwav, err = wav.NewReader(f, s.Size())
		if err != nil {
			return err
		}

		_, err = f.Seek(int64(iwav.FirstSampleOffset()), os.SEEK_SET)
		if err != nil {
			return err
		}

		files[i] = f
	}

	// prepare output files
	outputWriters := make([]*wav.Writer, len(outputFnames))
	for i, oname := range outputFnames {
		// create output file
		f, err := os.Create(oname)
		if err != nil {
			return err
		}

		// get meta information from input
		// BUG(Henry): check that all wav files have the same characteristics
		meta := iwav.GetFile()

		// create output with the same characteristics
		outputWriters[i], err = meta.NewWriter(f)
		if err != nil {
			return err
		}

		files = append(files, f)
	}

	// BUG(Henry): check that all wav files have the same length
	return NewMulti(binPath, int(iwav.GetSampleCount()), files, outputWriters)
}

func NewMulti(binPath string, samplesTotal int, files []*os.File, wr []*wav.Writer) error {

	cmd := exec.Command(binPath, strconv.Itoa(samplesTotal))
	cmd.ExtraFiles = files

	cmdOut, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("cmd.CombinedOutput() Err:%q\nOutput:%q", err, cmdOut)
	}

	if len(cmdOut) != 0 {
		fmt.Fprintf(os.Stderr, "sim.stderr output: %q\n", string(cmdOut))
	}

	for _, w := range wr {
		_, wPtr, err := w.GetDumbWriter()
		if err != nil {
			return fmt.Errorf("output.GetDumbWriter() Err:%q", err)
		}

		*wPtr = int32(samplesTotal)

		if err = w.Close(); err != nil {
			return err
		}
	}

	for _, f := range files[:len(files)-len(wr)] { // dont close outputWriters
		if err := f.Close(); err != nil {
			return err
		}
	}

	return nil
}
