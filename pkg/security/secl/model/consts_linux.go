// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

// Package model holds model related files
package model

import (
	"syscall"

	"golang.org/x/sys/unix"
)

var (
	// errorConstants are the supported error constants
	// generate_constants:Error constants,Error constants are the supported error constants.
	errorConstants = map[string]int{
		"E2BIG":           -int(syscall.E2BIG),
		"EACCES":          -int(syscall.EACCES),
		"EADDRINUSE":      -int(syscall.EADDRINUSE),
		"EADDRNOTAVAIL":   -int(syscall.EADDRNOTAVAIL),
		"EADV":            -int(syscall.EADV),
		"EAFNOSUPPORT":    -int(syscall.EAFNOSUPPORT),
		"EAGAIN":          -int(syscall.EAGAIN),
		"EALREADY":        -int(syscall.EALREADY),
		"EBADE":           -int(syscall.EBADE),
		"EBADF":           -int(syscall.EBADF),
		"EBADFD":          -int(syscall.EBADFD),
		"EBADMSG":         -int(syscall.EBADMSG),
		"EBADR":           -int(syscall.EBADR),
		"EBADRQC":         -int(syscall.EBADRQC),
		"EBADSLT":         -int(syscall.EBADSLT),
		"EBFONT":          -int(syscall.EBFONT),
		"EBUSY":           -int(syscall.EBUSY),
		"ECANCELED":       -int(syscall.ECANCELED),
		"ECHILD":          -int(syscall.ECHILD),
		"ECHRNG":          -int(syscall.ECHRNG),
		"ECOMM":           -int(syscall.ECOMM),
		"ECONNABORTED":    -int(syscall.ECONNABORTED),
		"ECONNREFUSED":    -int(syscall.ECONNREFUSED),
		"ECONNRESET":      -int(syscall.ECONNRESET),
		"EDEADLK":         -int(syscall.EDEADLK),
		"EDEADLOCK":       -int(syscall.EDEADLOCK),
		"EDESTADDRREQ":    -int(syscall.EDESTADDRREQ),
		"EDOM":            -int(syscall.EDOM),
		"EDOTDOT":         -int(syscall.EDOTDOT),
		"EDQUOT":          -int(syscall.EDQUOT),
		"EEXIST":          -int(syscall.EEXIST),
		"EFAULT":          -int(syscall.EFAULT),
		"EFBIG":           -int(syscall.EFBIG),
		"EHOSTDOWN":       -int(syscall.EHOSTDOWN),
		"EHOSTUNREACH":    -int(syscall.EHOSTUNREACH),
		"EIDRM":           -int(syscall.EIDRM),
		"EILSEQ":          -int(syscall.EIDRM),
		"EINPROGRESS":     -int(syscall.EINPROGRESS),
		"EINTR":           -int(syscall.EINTR),
		"EINVAL":          -int(syscall.EINVAL),
		"EIO":             -int(syscall.EIO),
		"EISCONN":         -int(syscall.EISCONN),
		"EISDIR":          -int(syscall.EISDIR),
		"EISNAM":          -int(syscall.EISNAM),
		"EKEYEXPIRED":     -int(syscall.EKEYEXPIRED),
		"EKEYREJECTED":    -int(syscall.EKEYREJECTED),
		"EKEYREVOKED":     -int(syscall.EKEYREVOKED),
		"EL2HLT":          -int(syscall.EL2HLT),
		"EL2NSYNC":        -int(syscall.EL2NSYNC),
		"EL3HLT":          -int(syscall.EL3HLT),
		"EL3RST":          -int(syscall.EL3RST),
		"ELIBACC":         -int(syscall.ELIBACC),
		"ELIBBAD":         -int(syscall.ELIBBAD),
		"ELIBEXEC":        -int(syscall.ELIBEXEC),
		"ELIBMAX":         -int(syscall.ELIBMAX),
		"ELIBSCN":         -int(syscall.ELIBSCN),
		"ELNRNG":          -int(syscall.ELNRNG),
		"ELOOP":           -int(syscall.ELOOP),
		"EMEDIUMTYPE":     -int(syscall.EMEDIUMTYPE),
		"EMFILE":          -int(syscall.EMFILE),
		"EMLINK":          -int(syscall.EMLINK),
		"EMSGSIZE":        -int(syscall.EMSGSIZE),
		"EMULTIHOP":       -int(syscall.EMULTIHOP),
		"ENAMETOOLONG":    -int(syscall.ENAMETOOLONG),
		"ENAVAIL":         -int(syscall.ENAVAIL),
		"ENETDOWN":        -int(syscall.ENETDOWN),
		"ENETRESET":       -int(syscall.ENETRESET),
		"ENETUNREACH":     -int(syscall.ENETUNREACH),
		"ENFILE":          -int(syscall.ENFILE),
		"ENOANO":          -int(syscall.ENOANO),
		"ENOBUFS":         -int(syscall.ENOBUFS),
		"ENOCSI":          -int(syscall.ENOCSI),
		"ENODATA":         -int(syscall.ENODATA),
		"ENODEV":          -int(syscall.ENODEV),
		"ENOENT":          -int(syscall.ENOENT),
		"ENOEXEC":         -int(syscall.ENOEXEC),
		"ENOKEY":          -int(syscall.ENOKEY),
		"ENOLCK":          -int(syscall.ENOLCK),
		"ENOLINK":         -int(syscall.ENOLINK),
		"ENOMEDIUM":       -int(syscall.ENOMEDIUM),
		"ENOMEM":          -int(syscall.ENOMEM),
		"ENOMSG":          -int(syscall.ENOMSG),
		"ENONET":          -int(syscall.ENONET),
		"ENOPKG":          -int(syscall.ENOPKG),
		"ENOPROTOOPT":     -int(syscall.ENOPROTOOPT),
		"ENOSPC":          -int(syscall.ENOSPC),
		"ENOSR":           -int(syscall.ENOSR),
		"ENOSTR":          -int(syscall.ENOSTR),
		"ENOSYS":          -int(syscall.ENOSYS),
		"ENOTBLK":         -int(syscall.ENOTBLK),
		"ENOTCONN":        -int(syscall.ENOTCONN),
		"ENOTDIR":         -int(syscall.ENOTDIR),
		"ENOTEMPTY":       -int(syscall.ENOTEMPTY),
		"ENOTNAM":         -int(syscall.ENOTNAM),
		"ENOTRECOVERABLE": -int(syscall.ENOTRECOVERABLE),
		"ENOTSOCK":        -int(syscall.ENOTSOCK),
		"ENOTSUP":         -int(syscall.ENOTSUP),
		"ENOTTY":          -int(syscall.ENOTTY),
		"ENOTUNIQ":        -int(syscall.ENOTUNIQ),
		"ENXIO":           -int(syscall.ENXIO),
		"EOPNOTSUPP":      -int(syscall.EOPNOTSUPP),
		"EOVERFLOW":       -int(syscall.EOVERFLOW),
		"EOWNERDEAD":      -int(syscall.EOWNERDEAD),
		"EPERM":           -int(syscall.EPERM),
		"EPFNOSUPPORT":    -int(syscall.EPFNOSUPPORT),
		"EPIPE":           -int(syscall.EPIPE),
		"EPROTO":          -int(syscall.EPROTO),
		"EPROTONOSUPPORT": -int(syscall.EPROTONOSUPPORT),
		"EPROTOTYPE":      -int(syscall.EPROTOTYPE),
		"ERANGE":          -int(syscall.ERANGE),
		"EREMCHG":         -int(syscall.EREMCHG),
		"EREMOTE":         -int(syscall.EREMOTE),
		"EREMOTEIO":       -int(syscall.EREMOTEIO),
		"ERESTART":        -int(syscall.ERESTART),
		"ERFKILL":         -int(syscall.ERFKILL),
		"EROFS":           -int(syscall.EROFS),
		"ESHUTDOWN":       -int(syscall.ESHUTDOWN),
		"ESOCKTNOSUPPORT": -int(syscall.ESOCKTNOSUPPORT),
		"ESPIPE":          -int(syscall.ESPIPE),
		"ESRCH":           -int(syscall.ESRCH),
		"ESRMNT":          -int(syscall.ESRMNT),
		"ESTALE":          -int(syscall.ESTALE),
		"ESTRPIPE":        -int(syscall.ESTRPIPE),
		"ETIME":           -int(syscall.ETIME),
		"ETIMEDOUT":       -int(syscall.ETIMEDOUT),
		"ETOOMANYREFS":    -int(syscall.ETOOMANYREFS),
		"ETXTBSY":         -int(syscall.ETXTBSY),
		"EUCLEAN":         -int(syscall.EUCLEAN),
		"EUNATCH":         -int(syscall.EUNATCH),
		"EUSERS":          -int(syscall.EUSERS),
		"EWOULDBLOCK":     -int(syscall.EWOULDBLOCK),
		"EXDEV":           -int(syscall.EXDEV),
		"EXFULL":          -int(syscall.EXFULL),
	}

	// openFlagsConstants are the supported flags for the open syscall
	// generate_constants:Open flags,Open flags are the supported flags for the open syscall.
	openFlagsConstants = map[string]int{
		"O_RDONLY":    syscall.O_RDONLY,
		"O_WRONLY":    syscall.O_WRONLY,
		"O_RDWR":      syscall.O_RDWR,
		"O_APPEND":    syscall.O_APPEND,
		"O_CREAT":     syscall.O_CREAT,
		"O_EXCL":      syscall.O_EXCL,
		"O_SYNC":      syscall.O_SYNC,
		"O_TRUNC":     syscall.O_TRUNC,
		"O_ACCMODE":   syscall.O_ACCMODE,
		"O_ASYNC":     syscall.O_ASYNC,
		"O_CLOEXEC":   syscall.O_CLOEXEC,
		"O_DIRECT":    syscall.O_DIRECT,
		"O_DIRECTORY": syscall.O_DIRECTORY,
		"O_DSYNC":     syscall.O_DSYNC,
		"O_FSYNC":     syscall.O_FSYNC,
		// "O_LARGEFILE": syscall.O_LARGEFILE, golang defines this as 0
		"O_NDELAY":   syscall.O_NDELAY,
		"O_NOATIME":  syscall.O_NOATIME,
		"O_NOCTTY":   syscall.O_NOCTTY,
		"O_NOFOLLOW": syscall.O_NOFOLLOW,
		"O_NONBLOCK": syscall.O_NONBLOCK,
		"O_RSYNC":    syscall.O_RSYNC,
	}

	// fileModeConstants contains the constants describing file permissions as well as the set-user-ID, set-group-ID, and sticky bits.
	// generate_constants:File mode constants,File mode constants are the supported file permissions as well as constants for the set-user-ID, set-group-ID, and sticky bits.
	fileModeConstants = map[string]int{
		// "S_IREAD":  syscall.S_IREAD, deprecated
		"S_ISUID": syscall.S_ISUID,
		"S_ISGID": syscall.S_ISGID,
		"S_ISVTX": syscall.S_ISVTX,
		"S_IRWXU": syscall.S_IRWXU,
		"S_IRUSR": syscall.S_IRUSR,
		"S_IWUSR": syscall.S_IWUSR,
		"S_IXUSR": syscall.S_IXUSR,
		"S_IRWXG": syscall.S_IRWXG,
		"S_IRGRP": syscall.S_IRGRP,
		"S_IWGRP": syscall.S_IWGRP,
		"S_IXGRP": syscall.S_IXGRP,
		"S_IRWXO": syscall.S_IRWXO,
		"S_IROTH": syscall.S_IROTH,
		"S_IWOTH": syscall.S_IWOTH,
		"S_IXOTH": syscall.S_IXOTH,
		// "S_IWRITE": syscall.S_IWRITE, deprecated
	}

	// inodeModeConstants are the supported file types and file modes
	// generate_constants:Inode mode constants,Inode mode constants are the supported file type constants as well as the file mode constants.
	inodeModeConstants = map[string]int{
		// "S_IEXEC":  syscall.S_IEXEC, deprecated
		"S_IFMT":   syscall.S_IFMT,
		"S_IFSOCK": syscall.S_IFSOCK,
		"S_IFLNK":  syscall.S_IFLNK,
		"S_IFREG":  syscall.S_IFREG,
		"S_IFBLK":  syscall.S_IFBLK,
		"S_IFDIR":  syscall.S_IFDIR,
		"S_IFCHR":  syscall.S_IFCHR,
		"S_IFIFO":  syscall.S_IFIFO,
		"S_ISUID":  syscall.S_ISUID,
		"S_ISGID":  syscall.S_ISGID,
		"S_ISVTX":  syscall.S_ISVTX,
		"S_IRWXU":  syscall.S_IRWXU,
		"S_IRUSR":  syscall.S_IRUSR,
		"S_IWUSR":  syscall.S_IWUSR,
		"S_IXUSR":  syscall.S_IXUSR,
		"S_IRWXG":  syscall.S_IRWXG,
		"S_IRGRP":  syscall.S_IRGRP,
		"S_IWGRP":  syscall.S_IWGRP,
		"S_IXGRP":  syscall.S_IXGRP,
		"S_IRWXO":  syscall.S_IRWXO,
		"S_IROTH":  syscall.S_IROTH,
		"S_IWOTH":  syscall.S_IWOTH,
		"S_IXOTH":  syscall.S_IXOTH,
	}

	// KernelCapabilityConstants list of kernel capabilities
	// generate_constants:Kernel Capability constants,Kernel Capability constants are the supported Linux Kernel Capability.
	KernelCapabilityConstants = map[string]uint64{
		"CAP_AUDIT_CONTROL":      1 << unix.CAP_AUDIT_CONTROL,
		"CAP_AUDIT_READ":         1 << unix.CAP_AUDIT_READ,
		"CAP_AUDIT_WRITE":        1 << unix.CAP_AUDIT_WRITE,
		"CAP_BLOCK_SUSPEND":      1 << unix.CAP_BLOCK_SUSPEND,
		"CAP_BPF":                1 << unix.CAP_BPF,
		"CAP_CHECKPOINT_RESTORE": 1 << unix.CAP_CHECKPOINT_RESTORE,
		"CAP_CHOWN":              1 << unix.CAP_CHOWN,
		"CAP_DAC_OVERRIDE":       1 << unix.CAP_DAC_OVERRIDE,
		"CAP_DAC_READ_SEARCH":    1 << unix.CAP_DAC_READ_SEARCH,
		"CAP_FOWNER":             1 << unix.CAP_FOWNER,
		"CAP_FSETID":             1 << unix.CAP_FSETID,
		"CAP_IPC_LOCK":           1 << unix.CAP_IPC_LOCK,
		"CAP_IPC_OWNER":          1 << unix.CAP_IPC_OWNER,
		"CAP_KILL":               1 << unix.CAP_KILL,
		"CAP_LEASE":              1 << unix.CAP_LEASE,
		"CAP_LINUX_IMMUTABLE":    1 << unix.CAP_LINUX_IMMUTABLE,
		"CAP_MAC_ADMIN":          1 << unix.CAP_MAC_ADMIN,
		"CAP_MAC_OVERRIDE":       1 << unix.CAP_MAC_OVERRIDE,
		"CAP_MKNOD":              1 << unix.CAP_MKNOD,
		"CAP_NET_ADMIN":          1 << unix.CAP_NET_ADMIN,
		"CAP_NET_BIND_SERVICE":   1 << unix.CAP_NET_BIND_SERVICE,
		"CAP_NET_BROADCAST":      1 << unix.CAP_NET_BROADCAST,
		"CAP_NET_RAW":            1 << unix.CAP_NET_RAW,
		"CAP_PERFMON":            1 << unix.CAP_PERFMON,
		"CAP_SETFCAP":            1 << unix.CAP_SETFCAP,
		"CAP_SETGID":             1 << unix.CAP_SETGID,
		"CAP_SETPCAP":            1 << unix.CAP_SETPCAP,
		"CAP_SETUID":             1 << unix.CAP_SETUID,
		"CAP_SYSLOG":             1 << unix.CAP_SYSLOG,
		"CAP_SYS_ADMIN":          1 << unix.CAP_SYS_ADMIN,
		"CAP_SYS_BOOT":           1 << unix.CAP_SYS_BOOT,
		"CAP_SYS_CHROOT":         1 << unix.CAP_SYS_CHROOT,
		"CAP_SYS_MODULE":         1 << unix.CAP_SYS_MODULE,
		"CAP_SYS_NICE":           1 << unix.CAP_SYS_NICE,
		"CAP_SYS_PACCT":          1 << unix.CAP_SYS_PACCT,
		"CAP_SYS_PTRACE":         1 << unix.CAP_SYS_PTRACE,
		"CAP_SYS_RAWIO":          1 << unix.CAP_SYS_RAWIO,
		"CAP_SYS_RESOURCE":       1 << unix.CAP_SYS_RESOURCE,
		"CAP_SYS_TIME":           1 << unix.CAP_SYS_TIME,
		"CAP_SYS_TTY_CONFIG":     1 << unix.CAP_SYS_TTY_CONFIG,
		"CAP_WAKE_ALARM":         1 << unix.CAP_WAKE_ALARM,
	}

	// ptraceConstants are the supported ptrace commands for the ptrace syscall
	// generate_constants:Ptrace constants,Ptrace constants are the supported ptrace commands for the ptrace syscall.
	ptraceConstants = map[string]uint32{
		"PTRACE_TRACEME":    unix.PTRACE_TRACEME,
		"PTRACE_PEEKTEXT":   unix.PTRACE_PEEKTEXT,
		"PTRACE_PEEKDATA":   unix.PTRACE_PEEKDATA,
		"PTRACE_PEEKUSR":    unix.PTRACE_PEEKUSR,
		"PTRACE_POKETEXT":   unix.PTRACE_POKETEXT,
		"PTRACE_POKEDATA":   unix.PTRACE_POKEDATA,
		"PTRACE_POKEUSR":    unix.PTRACE_POKEUSR,
		"PTRACE_CONT":       unix.PTRACE_CONT,
		"PTRACE_KILL":       unix.PTRACE_KILL,
		"PTRACE_SINGLESTEP": unix.PTRACE_SINGLESTEP,
		"PTRACE_ATTACH":     unix.PTRACE_ATTACH,
		"PTRACE_DETACH":     unix.PTRACE_DETACH,
		"PTRACE_SYSCALL":    unix.PTRACE_SYSCALL,

		"PTRACE_SETOPTIONS":           unix.PTRACE_SETOPTIONS,
		"PTRACE_GETEVENTMSG":          unix.PTRACE_GETEVENTMSG,
		"PTRACE_GETSIGINFO":           unix.PTRACE_GETSIGINFO,
		"PTRACE_SETSIGINFO":           unix.PTRACE_SETSIGINFO,
		"PTRACE_GETREGSET":            unix.PTRACE_GETREGSET,
		"PTRACE_SETREGSET":            unix.PTRACE_SETREGSET,
		"PTRACE_SEIZE":                unix.PTRACE_SEIZE,
		"PTRACE_INTERRUPT":            unix.PTRACE_INTERRUPT,
		"PTRACE_LISTEN":               unix.PTRACE_LISTEN,
		"PTRACE_PEEKSIGINFO":          unix.PTRACE_PEEKSIGINFO,
		"PTRACE_GETSIGMASK":           unix.PTRACE_GETSIGMASK,
		"PTRACE_SETSIGMASK":           unix.PTRACE_SETSIGMASK,
		"PTRACE_SECCOMP_GET_FILTER":   unix.PTRACE_SECCOMP_GET_FILTER,
		"PTRACE_SECCOMP_GET_METADATA": unix.PTRACE_SECCOMP_GET_METADATA,
		"PTRACE_GET_SYSCALL_INFO":     unix.PTRACE_GET_SYSCALL_INFO,
	}

	// protConstants are the supported protections for the mmap syscall
	// generate_constants:Protection constants,Protection constants are the supported protections for the mmap syscall.
	protConstants = map[string]int{
		"PROT_NONE":      unix.PROT_NONE,
		"PROT_READ":      unix.PROT_READ,
		"PROT_WRITE":     unix.PROT_WRITE,
		"PROT_EXEC":      unix.PROT_EXEC,
		"PROT_GROWSDOWN": unix.PROT_GROWSDOWN,
		"PROT_GROWSUP":   unix.PROT_GROWSUP,
	}

	// mmapFlagConstants are the supported flags for the mmap syscall
	// generate_constants:MMap flags,MMap flags are the supported flags for the mmap syscall.
	mmapFlagConstants = map[string]uint64{
		"MAP_SHARED":          unix.MAP_SHARED,          /* Share changes */
		"MAP_PRIVATE":         unix.MAP_PRIVATE,         /* Changes are private */
		"MAP_SHARED_VALIDATE": unix.MAP_SHARED_VALIDATE, /* share + validate extension flags */
		"MAP_ANON":            unix.MAP_ANON,
		"MAP_ANONYMOUS":       unix.MAP_ANONYMOUS,       /* don't use a file */
		"MAP_DENYWRITE":       unix.MAP_DENYWRITE,       /* ETXTBSY */
		"MAP_EXECUTABLE":      unix.MAP_EXECUTABLE,      /* mark it as an executable */
		"MAP_FIXED":           unix.MAP_FIXED,           /* Interpret addr exactly */
		"MAP_FIXED_NOREPLACE": unix.MAP_FIXED_NOREPLACE, /* MAP_FIXED which doesn't unmap underlying mapping */
		"MAP_GROWSDOWN":       unix.MAP_GROWSDOWN,       /* stack-like segment */
		"MAP_HUGETLB":         unix.MAP_HUGETLB,         /* create a huge page mapping */
		"MAP_LOCKED":          unix.MAP_LOCKED,          /* pages are locked */
		"MAP_NONBLOCK":        unix.MAP_NONBLOCK,        /* do not block on IO */
		"MAP_NORESERVE":       unix.MAP_NORESERVE,       /* don't check for reservations */
		"MAP_POPULATE":        unix.MAP_POPULATE,        /* populate (prefault) pagetables */
		"MAP_STACK":           unix.MAP_STACK,           /* give out an address that is best suited for process/thread stacks */
		"MAP_SYNC":            unix.MAP_SYNC,            /* perform synchronous page faults for the mapping */
		"MAP_UNINITIALIZED":   0x4000000,                /* For anonymous mmap, memory could be uninitialized */
		"MAP_HUGE_16KB":       14 << unix.MAP_HUGE_SHIFT,
		"MAP_HUGE_64KB":       16 << unix.MAP_HUGE_SHIFT,
		"MAP_HUGE_512KB":      19 << unix.MAP_HUGE_SHIFT,
		"MAP_HUGE_1MB":        20 << unix.MAP_HUGE_SHIFT,
		"MAP_HUGE_2MB":        21 << unix.MAP_HUGE_SHIFT,
		"MAP_HUGE_8MB":        23 << unix.MAP_HUGE_SHIFT,
		"MAP_HUGE_16MB":       24 << unix.MAP_HUGE_SHIFT,
		"MAP_HUGE_32MB":       25 << unix.MAP_HUGE_SHIFT,
		"MAP_HUGE_256MB":      28 << unix.MAP_HUGE_SHIFT,
		"MAP_HUGE_512MB":      29 << unix.MAP_HUGE_SHIFT,
		"MAP_HUGE_1GB":        30 << unix.MAP_HUGE_SHIFT,
		"MAP_HUGE_2GB":        31 << unix.MAP_HUGE_SHIFT,
		"MAP_HUGE_16GB":       34 << unix.MAP_HUGE_SHIFT,
	}

	// signalConstants are the supported signals for the kill syscall
	// generate_constants:Signal constants,Signal constants are the supported signals for the kill syscall.
	signalConstants = map[string]int{
		"SIGHUP":    int(unix.SIGHUP),
		"SIGINT":    int(unix.SIGINT),
		"SIGQUIT":   int(unix.SIGQUIT),
		"SIGILL":    int(unix.SIGILL),
		"SIGTRAP":   int(unix.SIGTRAP),
		"SIGABRT":   int(unix.SIGABRT),
		"SIGIOT":    int(unix.SIGIOT),
		"SIGBUS":    int(unix.SIGBUS),
		"SIGFPE":    int(unix.SIGFPE),
		"SIGKILL":   int(unix.SIGKILL),
		"SIGUSR1":   int(unix.SIGUSR1),
		"SIGSEGV":   int(unix.SIGSEGV),
		"SIGUSR2":   int(unix.SIGUSR2),
		"SIGPIPE":   int(unix.SIGPIPE),
		"SIGALRM":   int(unix.SIGALRM),
		"SIGTERM":   int(unix.SIGTERM),
		"SIGSTKFLT": int(unix.SIGSTKFLT),
		"SIGCHLD":   int(unix.SIGCHLD),
		"SIGCONT":   int(unix.SIGCONT),
		"SIGSTOP":   int(unix.SIGSTOP),
		"SIGTSTP":   int(unix.SIGTSTP),
		"SIGTTIN":   int(unix.SIGTTIN),
		"SIGTTOU":   int(unix.SIGTTOU),
		"SIGURG":    int(unix.SIGURG),
		"SIGXCPU":   int(unix.SIGXCPU),
		"SIGXFSZ":   int(unix.SIGXFSZ),
		"SIGVTALRM": int(unix.SIGVTALRM),
		"SIGPROF":   int(unix.SIGPROF),
		"SIGWINCH":  int(unix.SIGWINCH),
		"SIGIO":     int(unix.SIGIO),
		"SIGPOLL":   int(unix.SIGPOLL),
		"SIGPWR":    int(unix.SIGPWR),
		"SIGSYS":    int(unix.SIGSYS),
	}

	// unlinkFlagsConstants are the supported unlink flags for the unlink syscall
	// generate_constants:Unlink flags,Unlink flags are the supported flags for the unlink syscall.
	unlinkFlagsConstants = map[string]int{
		"AT_REMOVEDIR": unix.AT_REMOVEDIR,
	}

	// addressFamilyConstants are the supported network address families
	// generate_constants:Network Address Family constants,Network Address Family constants are the supported network address families.
	addressFamilyConstants = map[string]uint16{
		"AF_UNSPEC":     unix.AF_UNSPEC,
		"AF_LOCAL":      unix.AF_LOCAL,
		"AF_UNIX":       unix.AF_UNIX,
		"AF_FILE":       unix.AF_FILE,
		"AF_INET":       unix.AF_INET,
		"AF_AX25":       unix.AF_AX25,
		"AF_IPX":        unix.AF_IPX,
		"AF_APPLETALK":  unix.AF_APPLETALK,
		"AF_NETROM":     unix.AF_NETROM,
		"AF_BRIDGE":     unix.AF_BRIDGE,
		"AF_ATMPVC":     unix.AF_ATMPVC,
		"AF_X25":        unix.AF_X25,
		"AF_INET6":      unix.AF_INET6,
		"AF_ROSE":       unix.AF_ROSE,
		"AF_DECnet":     unix.AF_DECnet,
		"AF_NETBEUI":    unix.AF_NETBEUI,
		"AF_SECURITY":   unix.AF_SECURITY,
		"AF_KEY":        unix.AF_KEY,
		"AF_NETLINK":    unix.AF_NETLINK,
		"AF_ROUTE":      unix.AF_ROUTE,
		"AF_PACKET":     unix.AF_PACKET,
		"AF_ASH":        unix.AF_ASH,
		"AF_ECONET":     unix.AF_ECONET,
		"AF_ATMSVC":     unix.AF_ATMSVC,
		"AF_RDS":        unix.AF_RDS,
		"AF_SNA":        unix.AF_SNA,
		"AF_IRDA":       unix.AF_IRDA,
		"AF_PPPOX":      unix.AF_PPPOX,
		"AF_WANPIPE":    unix.AF_WANPIPE,
		"AF_LLC":        unix.AF_LLC,
		"AF_IB":         unix.AF_IB,
		"AF_MPLS":       unix.AF_MPLS,
		"AF_CAN":        unix.AF_CAN,
		"AF_TIPC":       unix.AF_TIPC,
		"AF_BLUETOOTH":  unix.AF_BLUETOOTH,
		"AF_IUCV":       unix.AF_IUCV,
		"AF_RXRPC":      unix.AF_RXRPC,
		"AF_ISDN":       unix.AF_ISDN,
		"AF_PHONET":     unix.AF_PHONET,
		"AF_IEEE802154": unix.AF_IEEE802154,
		"AF_CAIF":       unix.AF_CAIF,
		"AF_ALG":        unix.AF_ALG,
		"AF_NFC":        unix.AF_NFC,
		"AF_VSOCK":      unix.AF_VSOCK,
		"AF_KCM":        unix.AF_KCM,
		"AF_QIPCRTR":    unix.AF_QIPCRTR,
		"AF_SMC":        unix.AF_SMC,
		"AF_XDP":        unix.AF_XDP,
		"AF_MAX":        unix.AF_MAX,
	}
)
