pshdlWavSim
===========

helper for audio processing of PSHDL simulation code  - Pipes a WAV file through an external process using os/exec


## Todo

- [x] don't pass coeffs through argv[]
	can be done with os.Pipe and exec.Cmd ExtraFiles [ex](https://github.com/cryptix/randomcode/blob/13671b73ec55c9fc938547b3b74836c6497f8140/go/passPipe/main.go)
