#include <stdio.h>          // for fdopen
#include <stdlib.h>         // for NULL and iperf.h
#include <iperf.h>          // for iperf_test definition
#include <iperf_api.h>      // for iperf_*
#include "runner.h"         // for IRResult, IRServerConfig and IR_RESULT_*

IRResult ir_listen_and_serve(IRServerConfig cfg, int json_fd) {
  struct iperf_test *test;

  // create test
  test = iperf_new_test();
  if (test == NULL) {
    return IR_RESULT_EIPERF;
  }

  // configuration
  iperf_defaults(test);
  iperf_set_test_role(test, 's');

  // server
  iperf_set_test_server_port(test, cfg.listen_port);

  // output
  test->outfile = fdopen(json_fd, "w");
  if (test->outfile == NULL) {
    return IR_RESULT_EOPENJSONFD;
  }
  iperf_set_test_json_output(test, 1);

  // serve
  if (iperf_run_server(test) < 0) {
    iperf_free_test(test);
    return IR_RESULT_EIPERF;
  }

  // cleanup
  iperf_free_test(test);

  return IR_RESULT_SUCCESS;
}
