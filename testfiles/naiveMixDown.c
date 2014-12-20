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
  if (argc != 2) {
    fprintf(stderr, "Usage: %s <samplecount>\nGot:%d Wanted:2\n", argv[0], argc);
    exit(1);
  }


  // our union buffers for easy casting between byte[] and long
  union sample inA,inB, output;
  int samplecount = atoi(argv[1]);


  while(samplecount-- > 0) {

    // read a sample from extra FDs 3 and 4
    ret = read(3, inA.buf, bufSize);
    if (ret != bufSize) {
      fprintf(stderr, "Sim.Read Error\nn[%d] ret[%zd]\n",samplecount,ret);
      exit(1);
    }

    ret = read(4, inB.buf, bufSize);
    if (ret != bufSize) {
      fprintf(stderr, "Sim.Read Error\nn[%d] ret[%zd]\n",samplecount,ret);
      exit(1);
    }

    // do something with the sample...
    output.number = inA.number+inB.number;

    // write it out to stdout
    ret = write(5,output.buf, bufSize);
    if(ret != bufSize) {
      fprintf(stderr, "Sim.Write Error\nn[%d] ret[%zd]\n",samplecount,ret);
      exit(1);
    }
  }


  exit(0);
}
