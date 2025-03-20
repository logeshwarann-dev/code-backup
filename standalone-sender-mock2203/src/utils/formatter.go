package utils

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"unsafe"
)

//====================================================================WRITE==================================================================================

// Signed Int: Used for signed integers (1, 2, 4, and 8-byte) with little-endian byte order.
// Unsigned Int: Used for unsigned integers (1, 2, 4, and 8-byte) in little-endian.
// Float: Encodes float values as 8-byte signed integers with implicit decimals.
// FixedString: Encodes a fixed-length string, padded with spaces or null bytes if required.
// FixedString(0-terminable): Encodes a fixed-length string, optionally terminated by a null byte.
// VariableString: Used for variable-length strings, includes length and padding.
// Counter: Represents counters using unsigned integers (1, 2 bytes).
// LocalMktDate: Used for market date representation in YYYYMMDD format.
// PriceType: Prices in integer format with 8 decimals.
// Amount: Represents amounts in integer format with 2 decimals.
// Qty: Represents quantities as 4-byte signed integers.
// SeqNum: Message sequence numbers encoded as 8-byte unsigned integers.
// UTCTimestamp: Timestamps encoded as nanoseconds past the UNIX epoch.
// char: A single-character string.
// Data: Raw byte arrays (values between 0 and 255).
// chartext: Encodes text fields with optional null termination and padding.

// Helper function to write data into bytes
func writeToBuffer(buf *bytes.Buffer, data interface{}, littleEndian bool) error {
	if littleEndian {
		return binary.Write(buf, binary.LittleEndian, data)
	}
	return binary.Write(buf, binary.BigEndian, data)
}

// 1. Signed Int
// - Used for signed integers in 1, 2, 4, and 8-byte representations.
func ParseSignedInt(buf *bytes.Buffer, value int64, size int) error {
	switch size {
	case 1:
		return writeToBuffer(buf, int8(value), true)
	case 2:
		return writeToBuffer(buf, int16(value), true)
	case 4:
		return writeToBuffer(buf, int32(value), true)
	case 8:
		return writeToBuffer(buf, value, true)
	default:
		return fmt.Errorf("unsupported signed integer size")
	}
}

// 2. Unsigned Int
// - Used for unsigned integers in 1, 2, 4, and 8-byte representations.
func ParseUnsignedInt(buf *bytes.Buffer, value uint64, size int) error {
	switch size {
	case 1:
		return writeToBuffer(buf, uint8(value), true)
	case 2:
		return writeToBuffer(buf, uint16(value), true)
	case 4:
		return writeToBuffer(buf, uint32(value), true)
	case 8:
		return writeToBuffer(buf, value, true)
	default:
		return fmt.Errorf("unsupported unsigned integer size")
	}
}

// 3. Float
// - Used for floats encoded as 8-byte signed integers with 8 decimals.
func ParseFloat(buf *bytes.Buffer, value float64, decimals int) error {
	scaledValue := int64(value * float64(unsafe.Sizeof(decimals)))
	return writeToBuffer(buf, scaledValue, true)
}

// 4. FixedString
// - Used for fixed-length strings with space padding if required.
func ParseFixedString(buf *bytes.Buffer, value string, size int) error {
	if len(value) > size {
		return fmt.Errorf("string length exceeds fixed size")
	}
	buf.WriteString(value)
	for i := len(value); i < size; i++ {
		buf.WriteByte(0x00) // Padding with null bytes
	}
	return nil
}

// 5. FixedString(0-terminable)
// - Used for fixed-length strings that are optionally 0-terminated.
func ParseFixedStringTerminable(buf *bytes.Buffer, value string, size int) error {
	if len(value) > size {
		return fmt.Errorf("string length exceeds fixed size")
	}
	buf.WriteString(value)
	buf.WriteByte(0x00) // Null terminator
	for i := len(value) + 1; i < size; i++ {
		buf.WriteByte(0x00) // Padding with null bytes
	}
	return nil
}

// 6. VariableString
// - Used for variable-length strings with length information.
func ParseVariableString(buf *bytes.Buffer, value string, maxSize int) error {
	strLen := len(value)
	if strLen > maxSize {
		return fmt.Errorf("string length exceeds maximum size")
	}
	// Write length first
	if err := writeToBuffer(buf, uint8(strLen), true); err != nil {
		return err
	}
	buf.WriteString(value)
	// Padding to a multiple of 8 bytes
	for i := strLen; i < maxSize; i++ {
		buf.WriteByte(0x00)
	}
	return nil
}

// 7. Counter
// - Used for counters, represented as unsigned integers.
func ParseCounter(buf *bytes.Buffer, value uint64, size int) error {
	return ParseUnsignedInt(buf, value, size)
}

// 8. LocalMktDate
// - Used for LocalMktDate in YYYYMMDD format as 4-byte unsigned binary.
func ParseLocalMktDate(buf *bytes.Buffer, date int) error {
	return writeToBuffer(buf, uint32(date), true)
}

// 9. PriceType
// - Price as 8-byte signed integer with 8 decimal places.
func ParsePriceType(buf *bytes.Buffer, value int64) error {
	return writeToBuffer(buf, value, true)
}

// 10. Amount
// - Amount as 8-byte signed integer with 2 decimal places.
func ParseAmount(buf *bytes.Buffer, value int64) error {
	scaledValue := value * 100 // 2 decimals
	return writeToBuffer(buf, scaledValue, true)
}

// 11. Qty
// - Quantity as signed 4-byte integer.
func ParseQty(buf *bytes.Buffer, value int32) error {
	return writeToBuffer(buf, value, true)
}

// 12. SeqNum
// - Sequence number in 8-byte unsigned binary format.
func ParseSeqNum(buf *bytes.Buffer, value uint64) error {
	return writeToBuffer(buf, value, true)
}

// 13. UTCTimestamp
// - UTC timestamp in nanoseconds since UNIX epoch (8-byte unsigned).
func ParseUTCTimestamp(buf *bytes.Buffer, timestamp uint64) error {
	return writeToBuffer(buf, timestamp, true)
}

// 14. char
// - A string of length 1.
func ParseChar(buf *bytes.Buffer, value string) error {
	if len(value) != 1 {
		return fmt.Errorf("char must be 1 byte")
	}
	buf.WriteByte(value[0])
	return nil
}

// 15. Data
// - Byte array where each byte has a value from 0 to 255.
func ParseData(buf *bytes.Buffer, data []byte) error {
	buf.Write(data)
	return nil
}

// 16. chartext
// - Used for text fields with optional null termination.
func ParseCharText(buf *bytes.Buffer, value string, size int) error {
	return ParseFixedStringTerminable(buf, value, size)
}

//====================================================================READ===================================================================================

// Helper function to read data from bytes
func readFromBuffer(buf *bytes.Buffer, data interface{}, littleEndian bool) error {
	if littleEndian {
		return binary.Read(buf, binary.LittleEndian, data)
	}
	return binary.Read(buf, binary.BigEndian, data)
}

// 1. Read Signed Int
func ReadSignedInt(buf *bytes.Buffer, size int, littleEndian bool) (int64, error) {
	var value int64
	switch size {
	case 1:
		var v int8
		if err := readFromBuffer(buf, &v, littleEndian); err != nil {
			return 0, err
		}
		value = int64(v)
	case 2:
		var v int16
		if err := readFromBuffer(buf, &v, littleEndian); err != nil {
			return 0, err
		}
		value = int64(v)
	case 4:
		var v int32
		if err := readFromBuffer(buf, &v, littleEndian); err != nil {
			return 0, err
		}
		value = int64(v)
	case 8:
		if err := readFromBuffer(buf, &value, littleEndian); err != nil {
			return 0, err
		}
	default:
		return 0, fmt.Errorf("unsupported signed integer size")
	}
	return value, nil
}

// 2. Read Unsigned Int
func ReadUnsignedInt(buf *bytes.Buffer, size int, littleEndian bool) (uint64, error) {
	var value uint64
	switch size {
	case 1:
		var v uint8
		if err := readFromBuffer(buf, &v, littleEndian); err != nil {
			return 0, err
		}
		value = uint64(v)
	case 2:
		var v uint16
		if err := readFromBuffer(buf, &v, littleEndian); err != nil {
			return 0, err
		}
		value = uint64(v)
	case 4:
		var v uint32
		if err := readFromBuffer(buf, &v, littleEndian); err != nil {
			return 0, err
		}
		value = uint64(v)
	case 8:
		if err := readFromBuffer(buf, &value, littleEndian); err != nil {
			return 0, err
		}
	default:
		return 0, fmt.Errorf("unsupported unsigned integer size")
	}
	return value, nil
}

// 3. Read Float
func ReadFloat(buf *bytes.Buffer, decimals int, littleEndian bool) (float64, error) {
	var scaledValue int64
	if err := readFromBuffer(buf, &scaledValue, littleEndian); err != nil {
		return 0, err
	}
	return float64(scaledValue) / float64(unsafe.Sizeof(decimals)), nil
}

// 4. Read FixedString
func ReadFixedString(buf *bytes.Buffer, size int) (string, error) {
	data := make([]byte, size)
	if _, err := buf.Read(data); err != nil {
		return "", err
	}
	return string(bytes.TrimRight(data, "\x00")), nil // Trim null padding
}

// 5. Read FixedString(0-terminable)
func ReadFixedStringTerminable(buf *bytes.Buffer, size int) (string, error) {
	data := make([]byte, size)
	if _, err := buf.Read(data); err != nil {
		return "", err
	}
	return string(data[:bytes.IndexByte(data, 0)]), nil // Read until null terminator
}

// 6. Read VariableString
func ReadVariableString(buf *bytes.Buffer, maxSize int) (string, error) {
	var strLen uint8
	if err := readFromBuffer(buf, &strLen, true); err != nil {
		return "", err
	}
	if int(strLen) > maxSize {
		return "", fmt.Errorf("string length exceeds maximum size")
	}
	data := make([]byte, strLen)
	if _, err := buf.Read(data); err != nil {
		return "", err
	}
	// Read padding to a multiple of 8 bytes
	_, err := buf.Read(make([]byte, maxSize-int(strLen)))
	// utils.Printf(static.LOG_FLAG, data,"REASON")
	return string(data), err
}

// 7. Read Counter
func ReadCounter(buf *bytes.Buffer, size int, littleEndian bool) (uint64, error) {
	return ReadUnsignedInt(buf, size, littleEndian)
}

// 8. Read LocalMktDate
func ReadLocalMktDate(buf *bytes.Buffer, littleEndian bool) (int, error) {
	var date uint32
	if err := readFromBuffer(buf, &date, littleEndian); err != nil {
		return 0, err
	}
	return int(date), nil
}

// 9. Read PriceType
func ReadPriceType(buf *bytes.Buffer, littleEndian bool) (int64, error) {
	var value int64
	if err := readFromBuffer(buf, &value, littleEndian); err != nil {
		return 0, err
	}
	return value, nil
}

// 10. Read Amount
func ReadAmount(buf *bytes.Buffer, littleEndian bool) (int64, error) {
	var scaledValue int64
	if err := readFromBuffer(buf, &scaledValue, littleEndian); err != nil {
		return 0, err
	}
	return scaledValue / 100, nil // 2 decimals
}

// 11. Read Qty
func ReadQty(buf *bytes.Buffer, littleEndian bool) (int32, error) {
	var value int32
	if err := readFromBuffer(buf, &value, littleEndian); err != nil {
		return 0, err
	}
	return value, nil
}

// 12. Read SeqNum
func ReadSeqNum(buf *bytes.Buffer, littleEndian bool) (uint64, error) {
	var seq uint64
	if err := readFromBuffer(buf, &seq, littleEndian); err != nil {
		return 0, err
	}
	return seq, nil
}

// 13. Read UTCTimestamp
func ReadUTCTimestamp(buf *bytes.Buffer, littleEndian bool) (uint64, error) {
	var timestamp uint64
	if err := readFromBuffer(buf, &timestamp, littleEndian); err != nil {
		return 0, err
	}
	return timestamp, nil
}

// 14. Read Char
func ReadChar(buf *bytes.Buffer) (string, error) {
	var value [1]byte
	if err := readFromBuffer(buf, &value, true); err != nil {
		return "", err
	}
	return string(value[:]), nil
}

// 15. Read Data
func ReadData(buf *bytes.Buffer, size int) ([]byte, error) {
	data := make([]byte, size)
	if _, err := buf.Read(data); err != nil {
		return nil, err
	}
	return data, nil
}

// 16. Read CharText
func ReadCharText(buf *bytes.Buffer, size int) (string, error) {
	return ReadFixedStringTerminable(buf, size)
}
