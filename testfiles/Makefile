CC=clang

all: monopassthrough naiveMixDown

monopassthrough: monopassthrough.c
	$(CC) -o $@ monopassthrough.c

naiveMixDown: naiveMixDown.c
	$(CC) -o $@ naiveMixDown.c


clean:
	rm naiveMixDown
	rm monopassthrough
	rm testOut_*.wav
