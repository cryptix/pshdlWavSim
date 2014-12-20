package pshdlWavSim

import "testing"

func TestNewMultiFromFile_MixDown(t *testing.T) {
	err := NewMultiFromFile(
		"testfiles/naiveMixDown",
		"testfiles/chirp.wav",
		"testfiles/noise.wav",
		"testfiles/testOut_multi1.wav")
	if err != nil {
		t.Fatalf("NewMultiFromFile() Error %s", err)
	}

}
