#include <stdio.h>          // for fdopen
#include <stdlib.h>         // for NULL and iperf.h
#include <unistd.h>         // for usleep
#include <iperf.h>          // for iperf_test definition
#include <iperf_api.h>      // for iperf_*
#include "runner.h"         // for IRResult, IRServerConfig and IR_RESULT_*

IRResult ir_run_test(IRClientConfig cfg, int json_fd) {
  struct iperf_test *test;

  // create test
  test = iperf_new_test();
  if (test == NULL) {
    return IR_RESULT_EIPERF;
  }

  // configuration
  iperf_defaults(test);
  iperf_set_test_role(test, 'c');
  if (cfg.duration_secs > 0)
    iperf_set_test_duration(test, cfg.duration_secs);
  if (cfg.buffer_size > 0)
    iperf_set_test_blksize(test, cfg.buffer_size);
  if (cfg.bytes_amt > 0)
    test->settings->bytes = cfg.bytes_amt;
  if (cfg.packets_amt > 0)
    test->settings->blocks = cfg.packets_amt;

  // server
  iperf_set_test_server_hostname(test, cfg.target_host_ip);
  iperf_set_test_server_port(test, cfg.target_host_port);

  // output
  test->outfile = fdopen(json_fd, "w");
  if (test->outfile == NULL) {
    return IR_RESULT_EOPENJSONFD;
  }
  iperf_set_test_json_output(test, 1);

  // connect
  while (1) {
    if (iperf_run_client(test) < 0) {
      // sometimes the server takes a while to come up - retry
      if (i_errno == IECONNECT) {
        usleep(50); // 50ms
        continue;
      }

      iperf_free_test(test);
      return IR_RESULT_EIPERF;
    }
    break;
  }

  // cleanup
  iperf_free_test(test);

  return IR_RESULT_SUCCESS;
}
