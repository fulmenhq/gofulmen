package fulencode

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"

	"github.com/zeebo/xxh3"
)

func computeChecksumIfRequested(data []byte, algorithm string) (checksum string, checksumAlgorithm string, err *FulencodeError) {
	if algorithm == "" {
		return "", "", nil
	}

	switch algorithm {
	case "xxh3-128":
		sum := xxh3.Hash128(data)
		b := sum.Bytes()
		return fmt.Sprintf("%s:%s", algorithm, hex.EncodeToString(b[:])), algorithm, nil
	case "sha256":
		sum := sha256.Sum256(data)
		return fmt.Sprintf("%s:%s", algorithm, hex.EncodeToString(sum[:])), algorithm, nil
	case "sha512":
		sum := sha512.Sum512(data)
		return fmt.Sprintf("%s:%s", algorithm, hex.EncodeToString(sum[:])), algorithm, nil
	case "sha1":
		sum := sha1.Sum(data)
		return fmt.Sprintf("%s:%s", algorithm, hex.EncodeToString(sum[:])), algorithm, nil
	case "md5":
		sum := md5.Sum(data)
		return fmt.Sprintf("%s:%s", algorithm, hex.EncodeToString(sum[:])), algorithm, nil
	default:
		return "", "", newError(OperationEncode, "UNSUPPORTED_CHECKSUM", fmt.Sprintf("unsupported checksum algorithm: %s", algorithm))
	}
}
