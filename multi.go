package pshdlWavSim

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/cryptix/wav"
)

type Multi struct {
	cmd *exec.Cmd

	out          *wav.WavWriter
	samplesTotal uint32

	files []*os.File
}

func NewMultiFromFile(binPath string, inAfname, inBfname, outputFname string) (*Multi, error) {
	// open first input file
	fileA, err := os.Open(inAfname)
	if err != nil {
		return nil, err
	}

	statA, err := fileA.Stat()
	if err != nil {
		return nil, err
	}

	wavA, err := wav.NewWavReader(fileA, statA.Size())
	if err != nil {
		return nil, err
	}

	_, err = fileA.Seek(int64(wavA.FirstSampleOffset()), os.SEEK_SET)
	if err != nil {
		return nil, err
	}

	// open 2nd input file
	fileB, err := os.Open(inBfname)
	if err != nil {
		return nil, err
	}

	statB, err := fileB.Stat()
	if err != nil {
		return nil, err
	}

	wavB, err := wav.NewWavReader(fileB, statB.Size())
	if err != nil {
		return nil, err
	}

	_, err = fileB.Seek(int64(wavB.FirstSampleOffset()), os.SEEK_SET)
	if err != nil {
		return nil, err
	}

	// create output file
	outputFile, err := os.Create(outputFname)
	if err != nil {
		return nil, err
	}

	// get meta information from input
	meta := wavA.GetWavFile()

	// create output with the same characteristics
	output, err := meta.NewWriter(outputFile)
	if err != nil {
		return nil, err
	}

	// create a MonoFIR based in these files
	files := []*os.File{fileA, fileB, outputFile}
	sim, err := NewMulti(binPath, int(wavA.GetSampleCount()), files, output)
	if err != nil {
		return nil, err
	}

	// // store file handles for closing after sim

	return sim, nil
}

func NewMulti(binPath string, samplesTotal int, files []*os.File, output *wav.WavWriter) (*Multi, error) {
	m := &Multi{
		out: output,
	}

	m.samplesTotal = uint32(samplesTotal)

	m.cmd = exec.Command(binPath, strconv.Itoa(samplesTotal))
	m.cmd.ExtraFiles = files

	return m, nil
}

func (m Multi) Run() (err error) {
	_, wPtr, err := m.out.GetDumbWriter()
	if err != nil {
		return err
	}

	out, err := m.cmd.CombinedOutput()
	if err != nil {
		return err
	}

	if len(out) != 0 {
		return fmt.Errorf("Sim output: %q", string(out))
	}

	*wPtr = int32(m.samplesTotal)
	err = m.out.Close()
	if err != nil {
		return err
	}

	for _, f := range m.files {
		if err := f.Close(); err != nil {
			return err
		}
	}

	return nil
}
