package main

import (
	"bufio"
	"bytes"
	"fmt"
	"market_replay/aes"
	"net"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func ReadConfig(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	configs := &ProcessConfig{}
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		entries := strings.Split(line, ",")
		if len(entries) < 1 {
			fmt.Println("Invalid line:", line)
			continue
		}

		traderCount, err0 := strconv.Atoi(entries[0])
		throttleLimit, err1 := strconv.Atoi(entries[1])
		runtimeLimit, err2 := strconv.Atoi(entries[2])
		targetEnv, err3 := strconv.Atoi(entries[3])
		fileType, err4 := strconv.Atoi(entries[4])
		orderMod, err5 := strconv.Atoi(entries[5])
		orderCancel, err6 := strconv.Atoi(entries[6])
		heartbeat, err7 := strconv.Atoi(entries[7])

		if err0 != nil || err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil || err6 != nil || err7 != nil {
			fmt.Println("Error parsing limits: ", err1, err2)
			continue
		}

		configs.TraderCount = traderCount
		configs.ThrottleLimit = throttleLimit
		configs.RunTimeLimit = runtimeLimit
		configs.TargetEnv = targetEnv
		configs.FileType = fileType
		configs.OrderModPercent = orderMod
		configs.OrderCancelPercent = orderCancel
		configs.HeartBeat = heartbeat

		TARGET_ENV = configs.TargetEnv
		TOTAL_ORDER_COUNT = configs.TraderCount * configs.ThrottleLimit * configs.RunTimeLimit
		THROTTLE_VALUE = configs.ThrottleLimit
		FILE_TYPE = configs.FileType
		MODIFY_PERCENT = configs.OrderModPercent
		CANCEL_PERCENT = configs.OrderCancelPercent
		TRADERCOUNT = configs.TraderCount
		HEARTBEAT_VALUE = configs.HeartBeat

		fmt.Println("Total No. of Orders :", TOTAL_ORDER_COUNT)
		fmt.Println("Throttle Value:", THROTTLE_VALUE)
		fmt.Println("Run time (seconds) :", configs.RunTimeLimit)
		fmt.Println("Modify Order Percentage :", MODIFY_PERCENT)
		fmt.Println("Cancel Order Percentage :", CANCEL_PERCENT)
		fmt.Println("HeartBeat value :", HEARTBEAT_VALUE)

	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func ReadRecords(filename string) ([]Record, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var records []Record
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		entries := strings.Split(line, ",")

		if len(entries) < 9 {
			fmt.Println("Invalid line:", line)
			continue
		}

		instrumentID := entries[0]
		lowerLimit, err1 := strconv.Atoi(entries[5])
		upperLimit, err2 := strconv.Atoi(entries[6])
		minLot, err3 := strconv.Atoi(entries[7])
		product_id, err4 := strconv.Atoi(entries[8])

		if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
			fmt.Println("Error parsing limits:", err1, err2, err3, err4)
			continue
		}

		records = append(records, Record{
			InstrumentID: instrumentID,
			LowerLimit:   lowerLimit,
			UpperLimit:   upperLimit,
			MinLot:       minLot,
			Product_ID:   product_id,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return records, nil
}

func ByteToIP(bytes []byte) (string, error) {

	if len(bytes) != 4 {
		return "", fmt.Errorf("invalid byte sequence length: expected 4, got %d", len(bytes))
	}

	bigEndianBytes := []byte{bytes[3], bytes[2], bytes[1], bytes[0]}

	ip := net.IP(bigEndianBytes).String()

	return ip, nil
}

func GetTemplateID(data []byte) (uint16, error) {

	buff := bytes.NewBuffer(data)
	buff.Next(4)

	templateID, err := ReadUnsignedInt(buff, 2, true)
	if err != nil {
		return 0, fmt.Errorf("error reading template id: %w", err)
	}

	return uint16(templateID), nil
}

func ReadRejectedMsgNonDecrypted(data []byte) (string, error) {

	body_data := data[8:]

	buff := bytes.NewBuffer(body_data)

	buff.Next(64)

	str := buff.String()

	return str, nil
}

func DecryptData(data []byte, key, iv string, CTX *aes.CipherContext) ([]byte, error) {

	body_data := data[8:]

	decrypted_data, err := CTX.DecryptUpdate(body_data)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return decrypted_data, nil
}

func ReadRejectedMsg(data []byte, CTX *aes.CipherContext) (string, error) {

	body_data := data[8:]

	decrypted_data, err := CTX.DecryptUpdate(body_data)
	if err != nil {
		return "", fmt.Errorf("decryption failed: %w", err)
	}

	buff := bytes.NewBuffer(decrypted_data)

	buff.Next(64)

	str := buff.String()

	return str, nil
}

func GetTimestamps(data []byte) (*TimeStamps, error) {

	res := &TimeStamps{}

	buff := bytes.NewBuffer(data)

	gw_req_in, err := ReadUTCTimestamp(buff, true)
	if err != nil {
		return nil, fmt.Errorf("GW REQ IN: %w", err)
	}

	res.GW_Req_In = gw_req_in

	gw_me_req, err := ReadUTCTimestamp(buff, true)
	if err != nil {
		return nil, fmt.Errorf("GW to ME REQ: %w", err)
	}

	res.GW_ME_Req = gw_me_req

	me_req_in, err := ReadUTCTimestamp(buff, true)
	if err != nil {
		return nil, fmt.Errorf("ME REQ IN: %w", err)
	}

	res.ME_Req_In = me_req_in

	me_res_out, err := ReadUTCTimestamp(buff, true)
	if err != nil {
		return nil, fmt.Errorf("ME RES OUT: %w", err)
	}

	res.ME_Res_Out = me_res_out

	gw_me_res, err := ReadUTCTimestamp(buff, true)
	if err != nil {
		return nil, fmt.Errorf("GW to ME Res: %w", err)
	}

	res.GW_ME_Res = gw_me_res

	gw_res_out, err := ReadUTCTimestamp(buff, true)
	if err != nil {
		return nil, fmt.Errorf("GW RES OUT: %w", err)
	}

	res.GW_Res_Out = gw_res_out

	fmt.Printf("GW REQ IN: %v | GW to ME REQ: %v | ME REQ IN: %v | ME RES OUT: %v | GW to ME Res: %v | GW RES OUT: %v \n", convertTimestamp(res.GW_Req_In), convertTimestamp(res.GW_ME_Req), convertTimestamp(res.ME_Req_In), convertTimestamp(res.ME_Res_Out), convertTimestamp(res.GW_ME_Res), convertTimestamp(res.GW_Res_Out))

	return res, nil
}

func convertTimestamp(timeValue uint64) string {
	epoch := time.Unix(0, int64(timeValue))
	return epoch.Format(EPOCH_FORMAT)

}

func SetSocketBuffer(conn *net.TCPConn, sendBufSize, recvBufSize int) error {
	rawConn, err := conn.SyscallConn()
	if err != nil {
		return fmt.Errorf("Failed to get raw connection: %v", err)
	}

	// Set send buffer size
	err = rawConn.Control(func(fd uintptr) {

		if err := syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_RCVBUF, recvBufSize); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to set SO_RCVBUF: %v\n", err)
		}
	})
	if err != nil {
		return fmt.Errorf("Failed to set receive buffer size: %v", err)
	}

	// Set receive buffer size
	err = rawConn.Control(func(fd uintptr) {
		if err := syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_SNDBUF, sendBufSize); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to set SO_SNDBUF: %v\n", err)
		}
	})
	if err != nil {
		return fmt.Errorf("Failed to set send buffer size: %v", err)
	}

	// Now retrieve and print the actual buffer sizes
	err = rawConn.Control(func(fd uintptr) {

		_, err = syscall.GetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_SNDBUF)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get SO_SNDBUF: %v\n", err)
		}

		_, err = syscall.GetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_RCVBUF)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get SO_RCVBUF: %v\n", err)
		}
		// else {
		// 	fmt.Printf("Actual Receive Buffer Size: %d bytes\n", actualRecvBufSize)
		// }
	})
	if err != nil {
		return fmt.Errorf("Failed to retrieve buffer sizes: %v", err)
	}

	return nil
}
