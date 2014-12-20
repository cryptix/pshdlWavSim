package pshdlWavSim

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/cryptix/wav"
)

// NewMultiFromFile does the preperation and then calls NewMulti
func NewMultiFromFile(binPath string, inAfname, inBfname, outputFname string) error {
	// open first input file
	fileA, err := os.Open(inAfname)
	if err != nil {
		return err
	}

	statA, err := fileA.Stat()
	if err != nil {
		return err
	}

	wavA, err := wav.NewReader(fileA, statA.Size())
	if err != nil {
		return err
	}

	_, err = fileA.Seek(int64(wavA.FirstSampleOffset()), os.SEEK_SET)
	if err != nil {
		return err
	}

	// open 2nd input file
	fileB, err := os.Open(inBfname)
	if err != nil {
		return err
	}

	statB, err := fileB.Stat()
	if err != nil {
		return err
	}

	wavB, err := wav.NewReader(fileB, statB.Size())
	if err != nil {
		return err
	}

	_, err = fileB.Seek(int64(wavB.FirstSampleOffset()), os.SEEK_SET)
	if err != nil {
		return err
	}

	// create output file
	outputFile, err := os.Create(outputFname)
	if err != nil {
		return err
	}

	// get meta information from input
	meta := wavA.GetFile()

	// create output with the same characteristics
	output, err := meta.NewWriter(outputFile)
	if err != nil {
		return err
	}

	// create a MonoFIR based in these files
	files := []*os.File{fileA, fileB, outputFile}

	return NewMulti(binPath, int(wavA.GetSampleCount()), files, output)
}

func NewMulti(binPath string, samplesTotal int, files []*os.File, output *wav.Writer) error {

	cmd := exec.Command(binPath, strconv.Itoa(samplesTotal))
	cmd.ExtraFiles = files

	_, wPtr, err := output.GetDumbWriter()
	if err != nil {
		return fmt.Errorf("output.GetDumbWriter() Err:%q", err)
	}

	cmtOut, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("cmd.CombinedOutput() Err:%q", err)
	}

	if len(cmtOut) != 0 {
		fmt.Fprintf(os.Stderr, "sim.stderr output: %q\n", string(cmtOut))
	}

	*wPtr = int32(samplesTotal)
	err = output.Close()
	if err != nil {
		return err
	}

	for _, f := range files[:1] { // dont close outputFile
		if err := f.Close(); err != nil {
			return err
		}
	}

	return nil
}
