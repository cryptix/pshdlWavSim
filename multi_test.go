package pshdlWavSim

import "testing"

func TestNewMultiFromFile_MixDown(t *testing.T) {
	err := NewMultiFromFile(
		"testfiles/naiveMixDown", // bin
		[]string{"testfiles/chirp.wav", "testfiles/noise.wav"},
		[]string{"testfiles/testOut_multi1.wav"},
	)
	if err != nil {
		t.Fatalf("NewMultiFromFile() Error %s", err)
	}

}
