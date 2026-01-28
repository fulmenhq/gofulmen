package fulencode

func DetectBOM(data []byte) BOMResult {
	if len(data) >= 4 {
		if data[0] == 0xFF && data[1] == 0xFE && data[2] == 0x00 && data[3] == 0x00 {
			bom := "utf-32le"
			return BOMResult{BOMType: &bom, ByteLength: 4, EncodingImplied: &bom}
		}
		if data[0] == 0x00 && data[1] == 0x00 && data[2] == 0xFE && data[3] == 0xFF {
			bom := "utf-32be"
			return BOMResult{BOMType: &bom, ByteLength: 4, EncodingImplied: &bom}
		}
	}
	if len(data) >= 3 {
		if data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
			bom := "utf-8"
			return BOMResult{BOMType: &bom, ByteLength: 3, EncodingImplied: &bom}
		}
	}
	if len(data) >= 2 {
		if data[0] == 0xFF && data[1] == 0xFE {
			bom := "utf-16le"
			return BOMResult{BOMType: &bom, ByteLength: 2, EncodingImplied: &bom}
		}
		if data[0] == 0xFE && data[1] == 0xFF {
			bom := "utf-16be"
			return BOMResult{BOMType: &bom, ByteLength: 2, EncodingImplied: &bom}
		}
	}
	return BOMResult{BOMType: nil, ByteLength: 0, EncodingImplied: nil}
}

func RemoveBOM(data []byte, expectedEncoding *string) ([]byte, error) {
	bom := DetectBOM(data)
	if bom.ByteLength == 0 {
		return data, nil
	}
	if expectedEncoding != nil && bom.EncodingImplied != nil && *expectedEncoding != *bom.EncodingImplied {
		err := newError(OperationBOM, "BOM_MISMATCH", "bom does not match expected encoding")
		err = addDetails(err, map[string]any{"expected": *expectedEncoding, "actual": *bom.EncodingImplied})
		return nil, err
	}
	return data[bom.ByteLength:], nil
}

func AddBOM(data []byte, encoding string) ([]byte, error) {
	var bom []byte
	switch encoding {
	case "utf-8":
		bom = []byte{0xEF, 0xBB, 0xBF}
	case "utf-16le":
		bom = []byte{0xFF, 0xFE}
	case "utf-16be":
		bom = []byte{0xFE, 0xFF}
	case "utf-32le":
		bom = []byte{0xFF, 0xFE, 0x00, 0x00}
	case "utf-32be":
		bom = []byte{0x00, 0x00, 0xFE, 0xFF}
	default:
		return nil, newError(OperationBOM, "UNSUPPORTED_FORMAT", "unsupported bom encoding")
	}
	out := make([]byte, 0, len(bom)+len(data))
	out = append(out, bom...)
	out = append(out, data...)
	return out, nil
}
