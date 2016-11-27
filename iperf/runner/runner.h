/**
 * Error codes
 */

typedef enum {
  IR_RESULT_SUCCESS = 0,
  IR_RESULT_EIPERF = 1,        // An iperf API call has failed
  IR_RESULT_EOPENJSONFD = 2,   // Failed to open the JSON file
} IRResult;

char *ir_strerror(IRResult result);

/**
 * Types
 */

typedef struct {
  // The interval time in seconds between periodic bandwidth, jitter, and loss
  // measurements.
  int measurement_interval;
} IRConfig;

typedef struct {
  IRConfig ir_config;
  // Port to listen to
  int listen_port;
} IRServerConfig;

typedef enum {
  IR_PROTOCOL_TCP = 0,
  IR_PROTOCOL_UDP = 1,
  IR_PROTOCOL_SCTP = 2,
} IRProtocol;

typedef struct {
  IRConfig ir_config;
	// Target Iperf to connect to
  char *target_host_ip;
  int target_host_port;
	// Transport protocol to use for the measurements.
  IRProtocol protocol;
	// Duration, in seconds, of the stream. Default: 10 seconds.
  int duration_secs;
	// Amount of bytes to send.
  int bytes_amt;
	// Size of transmission buffer. Default: 128 KB for TCP and 8 KB for UDP.
  int buffer_size;
	// Amount of packets to send.
  int packets_amt;
} IRClientConfig;

/**
 * Functions
 */

IRResult ir_run_test(IRClientConfig cfg, int json_fd);

IRResult ir_listen_and_serve(IRServerConfig cfg, int json_fd);
