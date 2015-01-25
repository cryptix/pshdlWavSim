package pshdlWavSim

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/cryptix/wav"
)

type Filter struct {
	Vals   []int `bson:"vals"`
	Scaler int   `bson:"scaleBits"`
}

type Multi struct {
	bin         string
	sampleCount uint32
	files       []*os.File
	wavWriters  []*wav.Writer
	tmpWav      *wav.Reader
}

type optionFunc func(*Multi) error

func SetInputs(inputFnames []string) optionFunc {
	return func(m *Multi) error {
		if len(inputFnames) < 1 {
			return errors.New("need at least one input")
		}
		if m.files != nil {
			return errors.New("inputs allready set..!")
		}
		m.files = make([]*os.File, len(inputFnames))
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
			m.tmpWav, err = wav.NewReader(f, s.Size())
			if err != nil {
				return err
			}
			if i == 0 {
				m.sampleCount = m.tmpWav.GetSampleCount()
			}
			if m.sampleCount != m.tmpWav.GetSampleCount() {
				return fmt.Errorf("input[%s] - inconsistent sample count: %d vs %d", iname, m.sampleCount, m.tmpWav.GetSampleCount())
			}
			_, err = f.Seek(int64(m.tmpWav.FirstSampleOffset()), os.SEEK_SET)
			if err != nil {
				return err
			}
			m.files[i] = f
		}

		return nil
	}
}

func SetOutputs(outputFnames []string) optionFunc {
	return func(m *Multi) error {
		if len(outputFnames) < 1 {
			return errors.New("need at least one output filename")
		}
		if m.wavWriters != nil {
			return errors.New("outputs allready set..!")
		}
		// prepare output files
		m.wavWriters = make([]*wav.Writer, len(outputFnames))
		for i, oname := range outputFnames {
			// create output file
			f, err := os.Create(oname)
			if err != nil {
				return err
			}
			// get meta information from input
			// BUG(Henry): check that all wav files have the same characteristics
			meta := m.tmpWav.GetFile()
			// create output with the same characteristics
			m.wavWriters[i], err = meta.NewWriter(f)
			if err != nil {
				return err
			}
			m.files = append(m.files, f)
		}
		return nil
	}
}

func SetFilter(f Filter) optionFunc {
	return func(m *Multi) error {
		filterFile, writeMe, err := os.Pipe()
		if err != nil {
			return err
		}
		fmt.Fprintln(writeMe, f.Scaler)
		for _, v := range f.Vals {
			fmt.Fprintln(writeMe, v)
		}
		m.files = append(m.files, filterFile)
		return writeMe.Close()
	}
}

func NewMulti(binPath string, options ...optionFunc) (*Multi, error) {
	m := &Multi{
		bin: binPath,
	}
	if len(m.bin) == 0 {
		return nil, errors.New("no binary specified")
	}
	for _, o := range options {
		if err := o(m); err != nil {
			return nil, err
		}
	}
	if len(m.files) == 0 {
		return nil, errors.New("no inputs")
	}
	return m, nil
}

// Run start up the binary, run it and check for errors.
// close down the writers afterwards.
func (m *Multi) Run() error {
	cmd := exec.Command(m.bin, strconv.FormatUint(uint64(m.sampleCount), 10))
	cmd.ExtraFiles = m.files
	cmdOut, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("cmd.CombinedOutput() Err:%q\nOutput:%q", err, cmdOut)
	}
	if len(cmdOut) != 0 {
		fmt.Fprintf(os.Stderr, "sim.stderr output: %q\n", string(cmdOut))
	}
	for _, w := range m.wavWriters {
		_, wPtr, err := w.GetDumbWriter()
		if err != nil {
			return fmt.Errorf("output.GetDumbWriter() Err:%q", err)
		}
		*wPtr = int32(m.sampleCount)
		if err = w.Close(); err != nil {
			return err
		}
	}
	for i, f := range m.files[:len(m.files)-len(m.wavWriters)] { // dont close outputWriters
		if err := f.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to close input %d - Error:%s\n", i, err)
		}
	}
	return nil
}
