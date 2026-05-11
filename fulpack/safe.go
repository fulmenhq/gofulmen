package fulpack

import (
	"fmt"
	"os"
)

const maxInt64Uint = uint64(1<<63 - 1)

func checkedUint64ToInt64(value uint64, field string) (int64, error) {
	if value > maxInt64Uint {
		return 0, fmt.Errorf("%s exceeds maximum supported size: %d", field, value)
	}
	return int64(value), nil
}

func checkedAddUint64(total int64, value uint64, field string) (int64, error) {
	addend, err := checkedUint64ToInt64(value, field)
	if err != nil {
		return 0, err
	}
	if total > int64(maxInt64Uint)-addend {
		return 0, fmt.Errorf("%s total exceeds maximum supported size", field)
	}
	return total + addend, nil
}

func tarPermissionMode(mode int64) os.FileMode {
	if mode <= 0 {
		return 0
	}
	if mode > 0o777 {
		mode &= 0o777
	}
	return os.FileMode(mode)
}

func archiveEntryMode(mode os.FileMode) uint32 {
	return uint32(mode.Perm())
}

func extractionMode(mode os.FileMode, fallback os.FileMode, preserve bool) os.FileMode {
	if !preserve || mode == 0 {
		return fallback
	}
	if perm := mode.Perm(); perm != 0 {
		return perm
	}
	return fallback
}
