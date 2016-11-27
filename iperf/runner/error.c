#include <stdio.h>          // for sprintf
#include <stdlib.h>         // for malloc
#include <stdint.h>         // for uint64_t of iperf_api.h
#include <errno.h>          // for errno
#include <string.h>         // for strerror
#include <iperf_api.h>      // for i_errno and iperf_strerror
#include "runner.h"         // for IRResult and IR_RESULT_*

char *ir_strerror(IRResult result) {
  char *msg = malloc(sizeof(char) * 256);

  switch (result) {
    case IR_RESULT_SUCCESS:
      return "";

    case IR_RESULT_EIPERF:
      return iperf_strerror(i_errno);

    case IR_RESULT_EOPENJSONFD:
      sprintf(msg, "Failed to use JSON file FD: %s", strerror(errno));
      return msg;

    default:
      return "Unknown error occurred";
  }
}
