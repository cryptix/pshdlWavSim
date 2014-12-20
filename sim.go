package pshdlWavSim

import "io"

// readPump pumps samples to sim's stdin
type copyJob struct {
	Err error
	N   int64
}

func copyPump(done chan<- copyJob, in io.Reader, out io.WriteCloser) {
	var j copyJob

	j.N, j.Err = io.Copy(out, in)
	if j.Err != nil {
		done <- j
		return
	}

	j.Err = out.Close()
	done <- j
	close(done)
}
