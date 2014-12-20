package pshdlWavSim

import "testing"

func TestNewMono_Passthrough(t *testing.T) {
	err := NewMonoFIRFromFile(
		"testfiles/monopassthrough",
		23, []int{1, 2, 3},
		"testfiles/chirp.wav",
		"testfiles/testOut_chirp.wav")
	if err != nil {
		t.Fatalf("NewMonoFIRFromFile() Error %s", err)
	}
}
