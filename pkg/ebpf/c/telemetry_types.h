#ifndef TELEMETRY_TYPES_H
#define TELEMETRY_TYPES_H

//#include <uapi/asm-generic/errno-base.h>

//#define MAX_ERRNO (ERANGE + 1)
#define MAX_ERRNO 35
typedef struct {
    unsigned int err_count[MAX_ERRNO];
} map_err_telemetry_t;

#define read_indx 0
#define read_user_indx 1
#define read_kernel_indx 2
#define MAX_TELEMETRY_INDX read_kernel_indx
typedef struct {
    unsigned int err_count[MAX_TELEMETRY_INDX * MAX_ERRNO];
} helper_err_telemetry_t;

#endif
