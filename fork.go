package grp

/*
#include <unistd.h>
*/
import (
	"C"
)

func forkInCContext() int64 {
	return int64(C.fork())
}
