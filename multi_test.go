package pshdlWavSim

import "testing"

func TestNewMulti(t *testing.T) {
	m, err := NewMulti(
		"testfiles/naiveMixDown",
		SetInputs([]string{"testfiles/chirp.wav", "testfiles/noise.wav"}),
		SetOutputs([]string{"testfiles/testOut_multi1.wav"}),
	)
	if err != nil {
		t.Fatalf("NewMulti() Error %s", err)
	}
	if err := m.Run(); err != nil {
		t.Fatalf("Run() Error %s", err)
	}
}

func TestNewMultiForFilter(t *testing.T) {
	m, err := NewMulti(
		"testfiles/filterReader",
		SetInputs([]string{"testfiles/chirp.wav", "testfiles/noise.wav"}),
		SetOutputs([]string{"testfiles/testOut_multi1.wav"}),
		SetFilter(Filter{Scaler: 10, Vals: []int{23, 42, 666}}),
	)
	if err != nil {
		t.Fatalf("NewMulti() Error %s", err)
	}
	if err := m.Run(); err != nil {
		t.Fatalf("Run() Error %s", err)
	}
}
