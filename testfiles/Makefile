CC=clang

all: filterReader monopassthrough naiveMixDown

monopassthrough: monopassthrough.c
	$(CC) -o $@ $<

naiveMixDown: naiveMixDown.c
	$(CC) -o $@ $<

filterReader: filterReader.c
	$(CC) -o $@ $<

clean:
	rm naiveMixDown
	rm monopassthrough
	rm testOut_*.wav
