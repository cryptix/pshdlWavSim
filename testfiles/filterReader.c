#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <stdint.h>
#include <stdbool.h>


#define bufSize 3 // 24bit samples
union sample
{
	uint8_t buf[bufSize];
	long number;
};

#define wantArgc 2

int main(int argc, char const *argv[])
{
	// we need the sample count to know when we are done
	if (argc != wantArgc) {
		fprintf(stderr, "Usage: %s <samplecount>\nGot:%d Wanted:%d\n", argv[0], argc,wantArgc);
		exit(1);
	}

	// our union buffers for easy casting between byte[] and long
	union sample input;//, output;
	int samplecount = atoi(argv[1]);
	FILE *filter;
	if ((filter = fdopen(6, "r")) == NULL)
	{
		fprintf(stderr, "fdopen failed\n");
		exit(1);
	}

	size_t readNum =0;
	size_t len = 0;
	ssize_t bytes;
	char *line = NULL;

	while ((bytes = getline(&line, &len, filter)) != -1) {
		int f = atoi(line);
		switch (readNum) {

		case 0:
			if (f != 10)
			{
				fprintf(stderr, "wrong scaler");
				exit(1);
			}
			break;

		case 1:
			if (f != 23)
			{
				fprintf(stderr, "wrong val0");
				exit(1);
			}
			break;

		case 2:
			if (f != 42)
			{
				fprintf(stderr, "wrong val1");
				exit(1);
			}
			break;

		case 3:
			if (f != 666)
			{
				fprintf(stderr, "wrong val2");
				exit(1);
			}
			break;
		}
		readNum++;
	}

	if (ferror(filter)) {
		fprintf(stderr, "ferror");
		exit(1);
	}

	free(line);
	fclose(filter);
	exit(0);
}
