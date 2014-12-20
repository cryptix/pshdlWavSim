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



int main(int argc, char const *argv[])
{
  ssize_t ret;

  // we need the sample count to know when we are done
  if (argc != 6) {
    fprintf(stderr, "Usage: %s <samplecount>\nGot:%d Wanted:2\n", argv[0], argc);
    exit(1);
  }

  // our union buffers for easy casting between byte[] and long
  union sample input;//, output;
  int samplecount = atoi(argv[1]);


  while(samplecount-- > 0) {

    // read a sample from stdin
    ret = read(0, input.buf, bufSize);
    if (ret != bufSize) {
      fprintf(stderr, "Sim.Read Error\nn[%d] ret[%zd]\n",samplecount,ret);
      exit(1);
    }

    // do something with the sample...
    // output.number = input.number;

    // write it out to stdout
    ret = write(1,input.buf, bufSize);
    if(ret != bufSize) {
      fprintf(stderr, "Sim.Write Error\nn[%d] ret[%zd]\n",samplecount,ret);
      exit(1);
    }
  }


  exit(0);
}
