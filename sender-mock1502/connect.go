package main

import (
	"bufio"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"io"
	"market_replay/aes"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

var (
	dataChan         chan DataPacket
	connThrottleChan chan int
	// members          []Member
	// recordFile       string
	recordsChan chan []DataPacket
)

const (
	retryConn = 5
)

/*
func StartOld() {

	err := ReadConfig(CONFIG_FILE_PATH)
	if err != nil {
		fmt.Println("[ERR] Error reading config file.")
	}

	switch TARGET_ENV {
	case PRODUCTION_ENV:
		members = SessionIdSets[SESSIONSETVALUE]
		recordFile = PRODUCTION_MOCK_RECORDS
		if FILE_TYPE == 0 {
			recordFile = PRODUCTION_RECORDS_FILE
		}
		fmt.Println("[INFO] Environment : PRODUCTION")
	case LAB_ENV:
		members = SessionIdSets[SESSIONSETVALUE]
		recordFile = PRODUCTION_MOCK_RECORDS
		if FILE_TYPE == 0 {
			recordFile = LAB_RECORDS_FILE
		}
		fmt.Println("[INFO] Environment : LAB")
	case SIMULATION_ENV:
		members = SIMULATION_ENV_ID
		recordFile = SIMULATION_RECORDS1_FILE
		if FILE_TYPE == 0 {
			recordFile = SIMULATION_RECORDS_FILE
		}
		fmt.Println("[INFO] Environment : SIMULATION")
	}

	var wg sync.WaitGroup
	dataChan = make(chan DataPacket, DATA_CHANNEL_LENGTH)
	recordsChan = make(chan []DataPacket, DATA_CHANNEL_LENGTH)
	connThrottleChan = make(chan int, PATTERN_CHANNEL_LENGTH)

	err = ReadRecords(recordFile)
	if err != nil {
		fmt.Println("[ERR] Error reading records:", err)
		return
	}

	if TRADE {
		go PumpData()
	} else {
		go PumpRecords()
	}

	go PatternThrottle()
	go ChangeThrottle()

	for i := range members {
		for j := range members[i].Traders {
			wg.Add(1)
			trader := &members[i].Traders[j]
			go func(t *Trader) {
				defer wg.Done()
				retryAttempt := 0
				for retry := 0; retry < 10; retry++ {
					err := t.establishConnection(retryAttempt)
					if err == nil {
						fmt.Println("[SUCCESS] Goroutine completes processing. Exiting now.")
						break
					} else if strings.Contains(err.Error(), "[ERR] Error in rejected message parser") {
						break
					} else if strings.Contains(err.Error(), "An existing connection was forcibly closed by the remote host") {
						fmt.Println("[WARN] Retrying Connection: ", err.Error())
						continue
					} else {
						fmt.Printf("[ERR] In Main Goroutine, Error establishing connection for Trader %d_%d: %v", t.MID, t.TID, err)
						continue
					}
				}
			}(trader)
		}
	}
	wg.Wait()
}
*/

func Start() {

	var wg sync.WaitGroup
	dataChan = make(chan DataPacket, DATA_CHANNEL_LENGTH)
	recordsChan = make(chan []DataPacket, DATA_CHANNEL_LENGTH)
	connThrottleChan = make(chan int, PATTERN_CHANNEL_LENGTH)

	if TRADE {
		go PumpData()
	} else {
		go PumpRecords()
	}

	go PatternThrottle()
	go ChangeThrottle()

	ORDERPUMPINGSTATUS = true

	for _, trader := range TRADERS {
		wg.Add(1)
		// trader := &members[i].Traders[j]
		go func(t *Trader) {
			defer wg.Done()
			retryAttempt := 0
			for retry := 0; retry < 10; retry++ {
				err := t.establishConnection(retryAttempt)
				if err == nil {
					fmt.Println("[SUCCESS] Goroutine completes processing. Exiting now.")
					break
				} else if strings.Contains(err.Error(), "[ERR] Error in rejected message parser") {
					break
				} else if strings.Contains(err.Error(), "An existing connection was forcibly closed by the remote host") {
					fmt.Println("[WARN] Retrying Connection: ", err.Error())
					continue
				} else {
					fmt.Printf("[ERR] In Main Goroutine, Error establishing connection for Trader %d_%d: %v\n", t.MID, t.TID, err)
					continue
				}
			}
		}(&trader)
	}
	wg.Wait()

}

func PumpData() {
	for i := 0; i < TOTAL_ORDER_COUNT; i++ {
		record := RECORDS[i%len(RECORDS)]
		dataChan <- DataPacket{InstrumentId: record.InstrumentID, Price: record.LowerLimit, MaxPrice: record.UpperLimit, ProductId: record.Product_ID, Qty: record.MinLot, BidIntrvl: record.BidInterval, MaxOrderQty: record.MaxTrdQty}
	}
}

func PumpRecords() {
	fmt.Println("In PUMP RECORDS...")
	for i := 0; i < TOTAL_ORDER_COUNT; i++ {
		var dataPackets []DataPacket

		for _, rec := range RECORDS {
			dataPackets = append(dataPackets, DataPacket{
				InstrumentId: rec.InstrumentID,
				Price:        rec.LowerLimit,
				ProductId:    rec.Product_ID,
				Qty:          rec.MinLot,
				MaxPrice:     rec.UpperLimit,
				BidIntrvl:    rec.BidInterval,
				MaxOrderQty:  rec.MaxTrdQty,
			})
		}

		recordsChan <- dataPackets
	}
}

func (t *Trader) establishConnection(attempt int) error {

	attempt++
	t.orderMap = make(map[uint64]OrderResponse)
	t.uiOrderMap = make(map[uint64]UiOrderResponse)
	t.session_id = (t.MID * 100000) + t.TID
	t.msg_seq = 1

	sslConn, err := t.createConnectionTLS()
	if err != nil {
		return err
	}

	// Step 1: Connection Request/Response
	if err := t.handleConnectionRequest(sslConn); err != nil {
		return err
	}

	sslConn.Close()
	fmt.Println("[INFO] SSL Connection Closed for session id :", t.session_id)

	conn, err := t.createConnection()
	if err != nil {
		return err
	}
	t.conn = conn

	// Initialize the CipherContext
	cipherCtx, err := aes.NewCipherContext([]byte(t.CGW_Res.SecKey), []byte(t.CGW_Res.IV))
	if err != nil {
		fmt.Printf("[ERR] Error initializing CipherContext: %v", err)
	}
	t.CTX = cipherCtx

	// Step 2: Session Registration
	if err := t.handleSessionRegistration(); err != nil {
		return err
	}
	t.msg_seq++

	// Step 3: Session Login
	if err := t.handleSessionLogon(); err != nil {
		return err
	}
	t.msg_seq++

	// Step 4: User Login
	if err := t.handleUserLogin(); err != nil {
		return err
	}
	t.msg_seq++

	if TRADE {
		if err := t.handleOrderEntryTrade(); err != nil {
			return err
		}
	} else {

		// res_ch := make(chan struct{})
		// go t.NewOrderResponseListnerAsync(res_ch, DUMMY_NO)

		// if err := t.NewHandleOrderEntryRecordsAsync(); err != nil {
		// 	return err
		// }

		if err := t.NewHandleOrderEntryRecords(); err != nil {

			return err
		}

		// res_ch := make(chan struct{})
		// go t.NewOrderResponseListnerAsync(res_ch, DUMMY_NO)

		// if err := t.AsyncHandleOrderEntryRecordsOE(); err != nil {
		// 	return err
		// }

		// close(res_ch)

	}
	t.msg_seq++

	time.Sleep(1000 * time.Millisecond)
	fmt.Printf("[INFO] Ack received by session id  : %v  is %v. \n", t.session_id, t.AckCount)

	return nil
}

func (t *Trader) createConnectionTLS() (*tls.Conn, error) {

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		MinVersion:         tls.VersionTLS13,
	}

	dialer := &net.Dialer{
		Timeout:   60 * time.Second,
		KeepAlive: 60 * time.Second,
	}
	address := net.JoinHostPort(t.HostIP, t.Port)
	var sslConn *tls.Conn
	var err error
	sendBufferSize := 64 * 1024 * 1024
	recvBufferSize := 64 * 1024 * 1024
	for try := 1; try <= retryConn; try++ {
		sslConn, err = tls.DialWithDialer(dialer, "tcp", address, tlsConfig)
		if err == nil {

			tcpConn, ok := sslConn.NetConn().(*net.TCPConn)
			if ok {
				err = setSocketBuffer(tcpConn, sendBufferSize, recvBufferSize)
				if err != nil {
					return sslConn, fmt.Errorf("failed to set socket buffer sizes: %v", err)
				}
			} else {
				fmt.Println("[ERR] Failed to get TCP Conn.")
			}
			return sslConn, nil
		}
		fmt.Printf("[ERR] Trader %v Failed TLS connection (attempt %d/%d): %v\n", t.TID, try, retryConn, err)
		time.Sleep(500 * time.Millisecond)

	}
	return nil, fmt.Errorf("[Trader %v Failed TLS Connection after %d retries]", t.TID, retryConn)
}

func (t *Trader) createConnection() (net.Conn, error) {

	sendBufferSize := 64 * 1024 * 1024
	recvBufferSize := 64 * 1024 * 1024

	fmt.Println("[INFO] Primary GW IP and Port: ", t.CGW_Res.GW1, t.CGW_Res.Port1)
	address := net.JoinHostPort(t.CGW_Res.GW1, strconv.FormatUint(uint64(t.CGW_Res.Port1), 10))

	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %v", err)
	}

	tcpConn, ok := conn.(*net.TCPConn)
	if ok {
		err = setSocketBuffer(tcpConn, sendBufferSize, recvBufferSize)
		if err != nil {
			return conn, fmt.Errorf("failed to set socket buffer sizes: %v", err)
		}
		// err = tcpConn.SetNoDelay(true)
		// if err != nil {
		// 	return conn, fmt.Errorf("failed to set TCP No Delay: %v", err)
		// }
		// fmt.Println("")
	} else {
		fmt.Println("[ERR] Failed to get TCP Conn.")
	}
	return conn, nil
}

func (t *Trader) handleConnectionRequest(sslConn *tls.Conn) error {

	gatewayRequest, err := CreateGatewayConnectionRequest(t.MID, t.TID, t.Password)
	if err != nil {
		return fmt.Errorf("failed to create gateway connection request with session id: %v %v", t.session_id, err)
	}

	_, err = sslConn.Write(gatewayRequest)
	if err != nil {
		return fmt.Errorf("failed to send tls connection request with session id: %v %v", t.session_id, err)
	}

	buffer := make([]byte, 100*1024)
	n, err := sslConn.Read(buffer)
	if err != nil {
		return fmt.Errorf("error reading tls connection response for session id: %v %v", t.session_id, err)
	}

	response := buffer[:n]

	val, err := GetTemplateID(response)
	if err != nil {
		fmt.Println("[ERR] Error in template id parser in connection gateway response :", err)
		return nil
	}

	if val == REJECTED {
		rejectedMsg, err := ReadRejectedMsgNonDecrypted(response)
		if err != nil {
			fmt.Println("[ERR] Error in rejected message parser connection gateway response:", err)
			return nil
		}
		fmt.Printf("[ERR] Connection Gateway Response for session id %v  has been rejected.\n", t.session_id)
		return fmt.Errorf("reason: %v", rejectedMsg)
	}

	gatewayResponse, err := ConnectionGatewayResponse(response)
	if err != nil {
		return fmt.Errorf("error parsing connection gateway response: %v", err)
	}

	t.CGW_Res = *gatewayResponse

	fmt.Println("[INFO] Gateway Connected Successfully for session id: ", t.session_id)
	return nil
}

func setSocketBuffer(conn *net.TCPConn, sendBufSize, recvBufSize int) error {
	rawConn, err := conn.SyscallConn()
	if err != nil {
		return fmt.Errorf("failed to get raw connection: %v", err)
	}

	// Set send buffer size
	err = rawConn.Control(func(fd uintptr) {
		// handle := syscall.Handle(fd)
		// if err := syscall.SetsockoptInt(handle, syscall.SOL_SOCKET, syscall.SO_RCVBUF, recvBufSize); err != nil {
		// 	fmt.Fprintf(os.Stderr, "Failed to set SO_RCVBUF: %v\n", err)
		// }

		//Linux
		if err := syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_RCVBUF, recvBufSize); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to set SO_RCVBUF: %v\n", err)
		}
	})
	if err != nil {
		return fmt.Errorf("failed to set receive buffer size: %v", err)
	}

	// Set receive buffer size
	err = rawConn.Control(func(fd uintptr) {
		// handle := syscall.Handle(fd)
		// if err := syscall.SetsockoptInt(handle, syscall.SOL_SOCKET, syscall.SO_SNDBUF, sendBufSize); err != nil {
		// 	fmt.Fprintf(os.Stderr, "Failed to set SO_SNDBUF: %v\n", err)
		// }

		// Linux
		if err := syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_SNDBUF, sendBufSize); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to set SO_SNDBUF: %v\n", err)
		}
	})
	if err != nil {
		return fmt.Errorf("failed to set send buffer size: %v", err)
	}

	// Now retrieve and print the actual buffer sizes
	err = rawConn.Control(func(fd uintptr) {
		// handle := syscall.Handle(fd)
		// _, err = syscall.GetsockoptInt(handle, syscall.SOL_SOCKET, syscall.SO_SNDBUF)
		// if err != nil {
		// 	fmt.Fprintf(os.Stderr, "Failed to get SO_SNDBUF: %v\n", err)
		// }

		// _, err = syscall.GetsockoptInt(handle, syscall.SOL_SOCKET, syscall.SO_RCVBUF)
		// if err != nil {
		// 	fmt.Fprintf(os.Stderr, "Failed to get SO_RCVBUF: %v\n", err)
		// }

		// Linux
		_, err = syscall.GetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_SNDBUF)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get SO_SNDBUF: %v\n", err)
		}

		_, err = syscall.GetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_RCVBUF)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get SO_RCVBUF: %v\n", err)
		}

	})
	if err != nil {
		return fmt.Errorf("failed to retrieve buffer sizes: %v", err)
	}

	return nil
}

func (t *Trader) handleSessionRegistration() error {

	sessionRegRequest, err := CreateSessionRegistrationRequest(t.MID, t.TID, t.msg_seq)
	if err != nil {
		return fmt.Errorf("failed to create session registration request with session id: %v %v", t.session_id, err)
	}

	_, err = t.conn.Write(sessionRegRequest)
	if err != nil {
		return fmt.Errorf("failed to send session registration request with session id: %v %v", t.session_id, err)
	}

	buffer := make([]byte, 100*1024)
	n, err := t.conn.Read(buffer)
	if err != nil {
		return fmt.Errorf("error reading session registration response for session id: %v %v", t.session_id, err)
	}

	response := buffer[:n]

	val, err := GetTemplateID(response)
	if err != nil {
		fmt.Println("[ERR] Error in template id parser in session registration :", err)
		return nil
	}

	if val == REJECTED {
		rejectedMsg, err := ReadRejectedMsgNonDecrypted(response)
		if err != nil {
			fmt.Println("[ERR] Error in rejected message parser:", err)
			return nil
		}
		fmt.Printf("[ERR] Session Registration for session id %v  has been rejected.\n", t.session_id)
		return fmt.Errorf("reason: %v", rejectedMsg)
	}

	fmt.Println("[INFO] Session Registration Completed Successfully for session id:", t.session_id)
	return nil
}

func (t *Trader) handleSessionLogon() error {

	sessionLogonRequest, err := CreateSessionLogonRequest(t.MID, t.TID, t.msg_seq, t.Password, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
	if err != nil {
		return fmt.Errorf("failed to create session logon request with session id: %v %v", t.session_id, err)
	}

	_, err = t.conn.Write(sessionLogonRequest)
	if err != nil {
		return fmt.Errorf("failed to send session logon request with session id: %v %v", t.session_id, err)
	}

	buffer := make([]byte, 100*1024)
	n, err := t.conn.Read(buffer)
	if err != nil {
		return fmt.Errorf("error reading session logon response for session id: %v %v", t.session_id, err)
	}

	response := buffer[:n]

	val, err := GetTemplateID(response)
	if err != nil {
		fmt.Println("[ERR] Error in template id parser :", err)
		return nil
	}

	if val == REJECTED {
		rejectedMsg, err := ReadRejectedMsg(response, t.CTX)
		if err != nil {
			fmt.Println("[ERR] Error in rejected message parser in session logon:", err)
			return nil
		}
		fmt.Printf("[ERR] Session Logon for session id %v  has been rejected.\n", t.session_id)
		return fmt.Errorf("reason: %v", rejectedMsg)
	}

	session_res, err := DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
	if err != nil {
		fmt.Println("[ERR] Error in decrypting session logon response :", err)
		return nil
	}

	sessionLogonResponse, err := SessionLogonResponse(session_res)
	if err != nil {
		return fmt.Errorf("error parsing session logon response: %v", err)
	}
	t.throttle = sessionLogonResponse

	fmt.Println("[INFO] Session Logon Completed Successfully for Session id:", t.session_id)
	return nil
}

func (t *Trader) handleUserLogin() error {

	userLogonRequest, err := CreateUserLogonRequest(t.MID, t.TID, t.msg_seq, t.Password, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
	if err != nil {
		return fmt.Errorf("failed to create user logon request with session id: %v %v", t.session_id, err)
	}

	_, err = t.conn.Write(userLogonRequest)
	if err != nil {
		return fmt.Errorf("failed to send user logon request with session id: %v %v", t.session_id, err)
	}

	buffer := make([]byte, 100*1024)
	n, err := t.conn.Read(buffer)
	if err != nil {
		return fmt.Errorf("error reading user logon response with session id: %v %v", t.session_id, err)
	}

	response := buffer[:n]
	val, err := GetTemplateID(response)
	if err != nil {
		fmt.Println("[ERR] Error in template id parser :", err)
		return nil
	}

	if val == REJECTED {
		rejectedMsg, err := ReadRejectedMsg(response, t.CTX)
		if err != nil {
			fmt.Println("[ERR] Error in rejected message parser:", err)
			return nil
		}
		fmt.Printf("[ERR] User Logon for session id %v  has been rejected.\n", t.session_id)
		return fmt.Errorf("reason: %v", rejectedMsg)
	}

	_, err = DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
	if err != nil {
		fmt.Println("[ERR] Error in response user logon parser :", err)
		return nil
	}

	fmt.Println("[INFO] User Logon Completed Successfully for session id:", t.session_id)
	return nil
}

func (t *Trader) sendDataToServer(order []byte) error {

	_, err := t.conn.Write(order)
	if err != nil {
		fmt.Printf("[ERR] Error sending data with session id %d: %v", t.session_id, err)
	}

	return nil
}

func (t *Trader) readReceivedResponse(val uint16, inst_id int, response []byte) (uint16, error) {

	switch val {
	case REJECTED:
		rejectedMsg, err := ReadRejectedMsg(response, t.CTX)
		if err != nil {
			fmt.Println("[ERR] Error in rejected message parser:", err)
			return 0, err
		}
		t.AckCount++
		fmt.Printf("[ERR] Order placed by session id %v has been rejected for instrument id - %v.\n", t.session_id, inst_id)
		fmt.Println("[ERR] Reason:", rejectedMsg)
		return REJECTED, nil

	case STANDARD_ORDER, LEAN_ORDER:
		orderType := "Standard Order"
		if val == LEAN_ORDER {
			orderType = "Lean Order"
		}
		fmt.Printf("[PASS] Order placed successfully by session id %v for instrument id - %v (%s).\n", t.session_id, inst_id, orderType)
		t.AckCount++
		return LEAN_ORDER, nil

	case HEARTBEAT:
		fmt.Println("[INFO] HeartBeat Received for session id:", t.session_id)
		return HEARTBEAT, nil

	case ORDER_CONFIRMATION:
		_, err := DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
		if err != nil {
			fmt.Println("[ERR] Error in decrypting order confirmation :", err)
			return 0, err
		}
		return ORDER_CONFIRMATION, nil

	case MODIFY_LEAN_ORDER:
		fmt.Printf("[PASS] Order Modified successfully by session id %v for instrument id - %v (Lean Order).\n", t.session_id, inst_id)
		t.AckCount++
		return MODIFY_LEAN_ORDER, nil

	case CANCEL_LEAN_ORDER:
		fmt.Printf("[PASS] Order Cancelled successfully by session id %v for instrument id - %v (Lean Order).\n", t.session_id, inst_id)
		t.AckCount++
		return CANCEL_LEAN_ORDER, nil
	case IMMEDIATE_EXECUTION:
		fmt.Printf("[PASS] Received IMMEDIATE EXECUTION for session id %v.\n", t.session_id)
		t.AckCount++
		return IMMEDIATE_EXECUTION, nil
	case ORDER_BOOK_EXECUTION:
		fmt.Printf("[PASS] Received ORDER BOOK EXECUTION for session id %v.\n", t.session_id)
		t.AckCount++
		return ORDER_BOOK_EXECUTION, nil
	case TRADING_SESSION_EVENT:
		fmt.Printf("[PASS] Received TRADING_SESSION_EVENT for session id %v.\n", t.session_id)
		t.AckCount++
		return TRADING_SESSION_EVENT, nil
	case DELETE_ALL_ORDERS:
		fmt.Printf("[PASS] Received DELETE_ALL_ORDERS for session id %v.\n", t.session_id)
		t.AckCount++
		return DELETE_ALL_ORDERS, nil
	case DELETE_ALL_ORDERS_NO_HITS:
		fmt.Printf("[PASS] Received DELETE_ALL_ORDERS_NO_HITS for session id %v.\n", t.session_id)
		t.AckCount++
		return DELETE_ALL_ORDERS_NO_HITS, nil

	default:
		return 0, nil
	}
}

func (t *Trader) heartBeatListener() error {
	for {
		_, _, order_template, err := t.NewOrderResponseListner(DUMMY_ID)
		if err != nil {
			fmt.Println("[ERR] Error in reading received response :", err)
		}

		if order_template == HEARTBEAT {
			heartbeat, err := HeartBeatRequest()
			if err != nil {
				return err
			}

			t.sendDataToServer(heartbeat)

			if THROTTLE_VALUE > 0 {
				fmt.Println("[INFO] Throttle value is updated. Exiting heartbeat case ")
				return nil
			}
		}
	}
}

func (t *Trader) sendOrderModify(inst_id, price, qty, order_type int, order_id, act_time uint64) (uint64, uint64, uint16, error) {

	modified_data, err := ModifyLeanOrder(t.MID, t.TID, t.msg_seq, inst_id, price, qty, order_type, order_id, act_time, t.CTX, t.ClOrdID, order_id)
	if err != nil {
		fmt.Println("[ERR] Error in order modify parser :", err)
		// return nil
		return 0, 0, 0, nil
	}

	modOrderTime := time.Now()
	err = t.sendDataToServer(modified_data)
	if err != nil {
		fmt.Printf("[ERR] Error sending data for session id %d: %v", t.session_id, err)
		return 0, 0, 0, nil
	}

	order_id, act_time, order_template, err := t.NewOrderResponseListner(inst_id)
	fmt.Println("[INFO] Mod Order Request RTT: ", time.Since(modOrderTime))
	if err != nil {
		fmt.Println("[ERR] Error in reading received response :", err)
	}

	return order_id, act_time, order_template, nil
}

func (t *Trader) sendOrderCancel(inst_id, product_id int, order_id uint64) error {

	cancelOrder, err := CancelSingleLegOrderRequest(t.MID, t.TID, t.msg_seq, inst_id, product_id, order_id, t.CTX, t.ClOrdID, order_id)
	if err != nil {
		fmt.Println("[ERR] Error in cancel order", err)
	}

	cancelOrderTime := time.Now()
	err = t.sendDataToServer(cancelOrder)
	if err != nil {
		fmt.Printf("[ERR] Error sending cancel order for session id %d: %v", t.session_id, err)
		return err
	}

	_, _, _, err = t.NewOrderResponseListner(inst_id)
	fmt.Println("[INFO] Cancel Order Request RTT: ", time.Since(cancelOrderTime))
	if err != nil {
		fmt.Println("[ERR] Error in reading received response :", err)
	}

	return nil
}

func (t *Trader) sendOrderDelete(inst_id, product_id int) error {

	deleteOrder, err := DeleteAllOrderComplexRequest(inst_id, product_id, t.MID, t.TID, t.msg_seq, t.CTX)
	if err != nil {
		fmt.Println("[ERR] Error in Delete order", err)
	}

	deleteOrderTime := time.Now()
	err = t.sendDataToServer(deleteOrder)
	if err != nil {
		fmt.Printf("[ERR] Error sending Delete order for session id %d: %v", t.session_id, err)
		return err
	}

	_, _, _, err = t.NewOrderResponseListner(inst_id)
	fmt.Println("[INFO] Delete Order Request RTT: ", time.Since(deleteOrderTime))
	if err != nil {
		fmt.Println("[ERR] Error in reading received response :", err)
	}

	return nil
}

func (t *Trader) DeleteOrdersByType() error {

	if DELETE_ORDERS_ALL_SESSIONS != t.delete_order {
		switch DELETE_ORDERS_DETAILS.Type {

		case DELETE_SPECIFIC_MEMBER:
			if _, exists := DELETE_ORDERS_DETAILS.Ids[t.MID]; exists {
				if err := t.sendOrderDelete(DELETE_ORDERS_DETAILS.InstrumentID, DELETE_ORDERS_DETAILS.ProductID); err != nil {
					return fmt.Errorf("delete error for specific member session id %d: %v", t.session_id, err)
				}

				// delete(DELETE_ORDERS_DETAILS.Ids, t.MID)
				t.delete_order = DELETE_ORDERS_ALL_SESSIONS
				t.msg_seq++
				t.orderMap = make(map[uint64]OrderResponse)
			}

		case DELETE_SPECIFIC_SESSION:
			if _, exists := DELETE_ORDERS_DETAILS.Ids[t.session_id]; exists {
				if err := t.sendOrderDelete(DELETE_ORDERS_DETAILS.InstrumentID, DELETE_ORDERS_DETAILS.ProductID); err != nil {
					return fmt.Errorf("delete error for specific session id %d: %v", t.session_id, err)
				}
				delete(DELETE_ORDERS_DETAILS.Ids, t.session_id)
				t.delete_order = DELETE_ORDERS_ALL_SESSIONS
				t.msg_seq++
				t.orderMap = make(map[uint64]OrderResponse)
			}

		case DELETE_ALL:
			if err := t.sendOrderDelete(DELETE_ORDERS_DETAILS.InstrumentID, DELETE_ORDERS_DETAILS.ProductID); err != nil {
				return fmt.Errorf("delete error for session id %d: %v", t.session_id, err)
			}

			t.delete_order = DELETE_ORDERS_ALL_SESSIONS
			t.msg_seq++
			t.orderMap = make(map[uint64]OrderResponse)

		default:
			fmt.Println("DELETE ORDERS BY TYPE NO CASE MATCHED:", DELETE_ORDERS_DETAILS.Type)
		}
	}

	return nil
}

func (t *Trader) NewOrderResponseListner(inst_id int) (uint64, uint64, uint16, error) {
	reader := bufio.NewReader(t.conn)
	buffer := make([]byte, 100*1024)
	for {

		n, err := reader.Read(buffer)
		if err != nil {
			if err == io.EOF {
				fmt.Printf("[INFO] Connection closed for session id %d: %v\n", t.session_id, err)
				return 0, 0, 0, err
			}
			fmt.Printf("[ERR] Error reading responses for session id %d: %v", t.session_id, err)
			return 0, 0, 0, nil
		}

		responseBuf := buffer[:n]

		responseLength := len(responseBuf)
		if responseLength == 0 {
			fmt.Println("Response len is 0.")
			return 0, 0, 0, nil
		}

		var responseArr [][]byte
		count := 0

		for len(responseBuf) > 0 {

			count++

			if len(responseBuf) < 4 {
				fmt.Println("[ERR] Response buffer too short to read size")
				return 0, 0, 0, fmt.Errorf("response buffer too short")
			}

			size := int(binary.LittleEndian.Uint32(responseBuf[:4]))
			if len(responseBuf) < size {
				fmt.Println("[ERR] Response buffer too short for the indicated size")
				return 0, 0, 0, fmt.Errorf("response buffer too short for indicated size")
			}

			responseArr = append(responseArr, responseBuf[:size])
			responseBuf = responseBuf[size:]
		}

		for i, response := range responseArr {

			if len(response) == 0 {
				break
			}

			val, err := GetTemplateID(response)
			if err != nil {
				fmt.Println("[ERR] Error in template id parser :", err)
			}

			read_response, err := t.readReceivedResponse(val, inst_id, response)
			if err != nil {
				fmt.Println("[ERR] Error in reading received response :", err)
			}

			if read_response == ORDER_CONFIRMATION {
				fmt.Println("[PASS] Order Confirmation Received for session id:", t.session_id)
				continue
			}

			switch read_response {
			case LEAN_ORDER:
				order_res, err := DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
				if err != nil {
					fmt.Println("[ERR] Error in decryption for order response parser :", err)
				}

				server_res, err1 := GetTimestamps(order_res)
				if err1 != nil {
					fmt.Println("[ERR] Error while getting timestamps")
				} else {
					fmt.Printf("[INFO] Order Entry Server Processing Time(T6-T1):  %v microseconds\n", (server_res.GW_Res_Out-server_res.GW_Req_In)/1000)
				}

				order_id, act_time, err := SingleLegLeanOrderResponse(order_res, t.CTX)
				if err != nil {
					fmt.Println("[ERR] Error in lean order response parser :", err)
				}

				if i == len(responseArr)-1 {
					return order_id, act_time, LEAN_ORDER, nil
				}
			case CANCEL_LEAN_ORDER:
				order_res, err := DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
				if err != nil {
					fmt.Println("[ERR] Error in decryption for order response parser :", err)
				}

				server_res, err1 := GetTimestamps(order_res)
				if err1 != nil {
					fmt.Println("[ERR] Error while getting timestamps")
				} else {
					fmt.Printf("[INFO] Order Cancel Server Processing Time(T6-T1):  %v microseconds\n", (server_res.GW_Res_Out-server_res.GW_Req_In)/1000)
				}

				if i == len(responseArr)-1 {
					return 0, 0, CANCEL_LEAN_ORDER, nil
				}
			case MODIFY_LEAN_ORDER:
				order_res, err := DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
				if err != nil {
					fmt.Println("[ERR] Error in decryption for order response parser :", err)
				}

				server_res, err1 := GetTimestamps(order_res)
				if err1 != nil {
					fmt.Println("[ERR] Error while getting timestamps")
				} else {
					fmt.Printf("[INFO] Order Mod Server Processing Time(T6-T1):  %v microseconds\n", (server_res.GW_Res_Out-server_res.GW_Req_In)/1000)
				}

				order_id, act_time, err := SingleLegLeanOrderModifiedResponse(order_res, t.CTX)
				if err != nil {
					fmt.Println("[ERR] Error in lean order response parser :", err)
				}

				if i == len(responseArr)-1 {
					return order_id, act_time, MODIFY_LEAN_ORDER, nil
				}
			case HEARTBEAT:
				_, err := DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
				if err != nil {
					fmt.Println("[ERR] Error in decryption for order response parser :", err)
				}

				if i == len(responseArr)-1 {
					return 0, 0, HEARTBEAT, nil
				}
			case REJECTED:
				if i == len(responseArr)-1 {
					return 0, 0, REJECTED, nil
				}
			case IMMEDIATE_EXECUTION:
				_, err := DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
				if err != nil {
					fmt.Println("[ERR] Error in decryption for order response parser :", err)
				}

				if i == len(responseArr)-1 {
					return 0, 0, IMMEDIATE_EXECUTION, nil
				}
			case ORDER_BOOK_EXECUTION:
				_, err := DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
				if err != nil {
					fmt.Println("[ERR] Error in decryption for order response parser :", err)
				}

				if i == len(responseArr)-1 {
					return 0, 0, ORDER_BOOK_EXECUTION, nil
				}
			case TRADING_SESSION_EVENT:
				_, err := DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
				if err != nil {
					fmt.Println("[ERR] Error in decryption for order response parser :", err)
				}

				if i == len(responseArr)-1 {
					return 0, 0, TRADING_SESSION_EVENT, nil
				}

			case DELETE_ALL_ORDERS:
				_, err := DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
				if err != nil {
					fmt.Println("[ERR] Error in decryption for order response parser :", err)
				}

				if i == len(responseArr)-1 {
					return 0, 0, DELETE_ALL_ORDERS, nil
				}

			case DELETE_ALL_ORDERS_NO_HITS:
				_, err := DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
				if err != nil {
					fmt.Println("[ERR] Error in decryption for order response parser :", err)
				}

				if i == len(responseArr)-1 {
					return 0, 0, DELETE_ALL_ORDERS_NO_HITS, nil
				}
			default:
				fmt.Println("Other Template ID: ", read_response)
				if i == len(responseArr)-1 {
					return 0, 0, 0, nil
				}
			}
		}
	}
}

func (t *Trader) handleOrderEntryTrade() error {
	startTime := time.Now()
	totalTime := time.Now()
	packetCount := 0
	totalPacketCount := 0
	count := 0

	for {
		if TRADE_THROTTLE > 0 {
		traderChanLoop:
			for data := range dataChan {

				elapsedTime := time.Since(startTime)

				if packetCount >= TRADE_THROTTLE || elapsedTime >= time.Second {
					if packetCount >= TRADE_THROTTLE && elapsedTime < time.Second {
						fmt.Printf("Session id %d , sleeping for %v\n", t.session_id, time.Second-elapsedTime)
						time.Sleep(time.Second - elapsedTime)
					}
					fmt.Printf("Session id %d sent %d packets.\n", t.session_id, packetCount)
					startTime = time.Now()
					packetCount = 0
				}

				if THROTTLE_VALUE == 0 {
					fmt.Println("Exiting Trade Channel as trade throttle is set to zero and msg seq is", t.msg_seq)
					break traderChanLoop
				}

				_, buy_max_price, sell_min_price, _ := GetPriceRange(data.Price, data.MaxPrice, data.BidIntrvl)
				startPrice := buy_max_price + data.BidIntrvl
				price := startPrice + (count * data.BidIntrvl)

				if price >= sell_min_price {
					price = startPrice
					count = 0
				}

				if UNIQUE_CLIENT_IDENTIFIER {
					t.ClOrdID++
				}

				order_type := BUY
				client_code := CLIENT3
				if t.session_id%2 != 0 {
					order_type = SELL
					client_code = CLIENT4
				}

				orderRequest, err := SingleLegLeanOrder(t.MID, t.TID, t.msg_seq, data.InstrumentId, price, data.Qty, order_type, t.CTX, t.ClOrdID, client_code)
				if err != nil {
					return err
				}

				if err := t.sendDataToServer(orderRequest); err != nil {
					fmt.Printf("Error sending data with session id %d: %v", t.session_id, err)
				}

				_, _, _, err = t.NewOrderResponseListner(data.InstrumentId)
				if err != nil {
					return fmt.Errorf("response error for session id %d: %v", t.session_id, err)
				}

				t.msg_seq++
				packetCount++
				totalPacketCount++
				count++

				if len(dataChan) == 0 && !SWITCH_CH {
					fmt.Printf("Session id %d took total time: %v and sent packets: %v\n",
						t.session_id, time.Since(totalTime), totalPacketCount)

					return nil
				}
			}
		} else {
			fmt.Println(" Trade Throttle is 0 and Heartbeat started!")
			if err := t.heartBeatListener(); err != nil {
				return fmt.Errorf("HeartBeat err: %v ", err)
			}
		}
	}
}

func (t *Trader) NewHandleOrderEntryRecords() error {

	startTime := time.Now()
	totalTime := time.Now()
	packetCount := 0
	totalPacketCount := 0
	count := 0

	err := setOrderPumpingParameters()
	if err != nil {
		fmt.Println("Error in setting Order Pumping Parameters.")
	}

	for {
		if THROTTLE_VALUE > 0 {
		dataChanLoop:
			for data := range recordsChan {

				throttleRate := THROTTLE_VALUE

				if err := t.DeleteOrdersByType(); err != nil {
					fmt.Printf("Delete all orders error for session id %d: %v", t.session_id, err)
				}

				if THROTTLE_VALUE == 0 {
					fmt.Println("1. Exiting Data Channel as throttle is set to zero and msg seq is", t.msg_seq)
					t.orderMap = make(map[uint64]OrderResponse)
					t.uiOrderMap = make(map[uint64]UiOrderResponse)
					break dataChanLoop
				}

				count++
				fmt.Printf("Packet No. (Iteration) in data channel %d for session id: %v \n", count, t.session_id)

				client_code := CLIENT3
				order_type := BUY
				if t.session_id%2 == 0 {
					order_type = SELL
					client_code = CLIENT4
				}

				price := 0
				qty := 0

				maxInterval := time.Second
				if PATTERN_GENERATOR && PATTERN_TYPE == PEAK_GENERATOR {
					maxInterval = time.Duration(INTERVAL_TIME) * time.Millisecond
				}

				elapsedTime := time.Since(startTime)
				fmt.Println("Elapsed Time :", elapsedTime)

				loopCount := 0
				dataLen := len(data)

				for loopCount < ORDER_ENTRY_COUNT {

					record := data[loopCount%dataLen]
					elapsedTime = time.Since(startTime)
					fmt.Println("OE ELAPSED TIME :", elapsedTime)
					if packetCount >= THROTTLE_VALUE || elapsedTime >= maxInterval {

						if THROTTLE_VALUE == 0 {
							fmt.Println("2. Exiting Data Channel as throttle is set to zero and msg seq is", t.msg_seq)
							t.orderMap = make(map[uint64]OrderResponse)
							t.uiOrderMap = make(map[uint64]UiOrderResponse)
							break dataChanLoop
						} else if packetCount >= THROTTLE_VALUE && elapsedTime < maxInterval {
							// fmt.Printf("Session id %d , sleeping for %v\n", t.session_id, 1000*time.Millisecond-elapsedTime)
							// time.Sleep(time.Second - elapsedTime)

							fmt.Println("Packet count :", packetCount, "| THROTTLE value : ", THROTTLE_VALUE, "Delay :", DELAY_TIME)

							if PATTERN_GENERATOR && PATTERN_TYPE == PEAK_GENERATOR {
								fmt.Printf("1. OE Session id %d , sleeping for %v (peak generator)\n", t.session_id, time.Duration(DELAY_TIME)*time.Millisecond)
								time.Sleep(time.Duration(DELAY_TIME) * time.Millisecond)
							} else {
								fmt.Printf("2. OE Session id %d , sleeping for %v\n", t.session_id, 1000*time.Millisecond-elapsedTime)
								time.Sleep(time.Second - elapsedTime)
							}

						}

						fmt.Printf("3. OE Session id %d sent %d packets.\n", t.session_id, packetCount)

						startTime = time.Now()
						packetCount = 0
						SEND_MOD = true
						t.orderMap = make(map[uint64]OrderResponse)
						t.uiOrderMap = make(map[uint64]UiOrderResponse)
					}

					if UNIQUE_CLIENT_IDENTIFIER {
						t.ClOrdID++
					}

					if throttleRate != THROTTLE_VALUE && PATTERN_TYPE != PEAK_GENERATOR {
						fmt.Println("1. Throttle value changed, execution will begin from start!")
						t.orderMap = make(map[uint64]OrderResponse)
						t.uiOrderMap = make(map[uint64]UiOrderResponse)
						break
					}

					price_factor := record.BidIntrvl
					buy_min_price, buy_max_price, sell_min_price, sell_max_price := GetPriceRange(record.Price, record.MaxPrice, price_factor)

					mid_qty := record.Qty * (1 + record.MaxOrderQty/(record.Qty*2))
					if mid_qty > record.MaxOrderQty {
						mid_qty = record.MaxOrderQty
					}

					order_details := OrderDetails{price_factor: price_factor, buy_min_price: buy_min_price, buy_max_price: buy_max_price, sell_min_price: sell_min_price, sell_max_price: sell_max_price, min_qty: record.Qty, mid_qty: mid_qty, max_qty: record.MaxOrderQty}

					if order_type == BUY {
						price = GenerateRandomPrice(buy_min_price, buy_max_price, price_factor)
						qty = GenerateRandomQty(record.Qty, mid_qty)
					} else if order_type == SELL {
						price = GenerateRandomPrice(sell_min_price, sell_max_price, price_factor)
						qty = GenerateRandomQty(mid_qty, record.MaxOrderQty)
					}

					orderRequest, err := SingleLegLeanOrder(t.MID, t.TID, t.msg_seq, record.InstrumentId, price, qty, order_type, t.CTX, t.ClOrdID, client_code)
					if err != nil {
						return err
					}

					fmt.Println("Client Order ID: ", t.ClOrdID)

					if err := t.sendDataToServer(orderRequest); err != nil {
						fmt.Printf("Error sending record with session id %d: %v", t.session_id, err)
					}

					orderID, actTime, template, err := t.NewOrderResponseListner(record.InstrumentId)
					if err != nil {
						return fmt.Errorf("response error for session id %d: %v", t.session_id, err)
					}

					if template == LEAN_ORDER {
						if UNIQUE_CLIENT_IDENTIFIER {
							t.uiOrderMap[t.ClOrdID] = UiOrderResponse{origClOrdID: t.ClOrdID, actTime: actTime, instrumentID: record.InstrumentId, productID: record.ProductId, orderDetails: order_details}
							fmt.Println("Unique identifier Map :", t.uiOrderMap)
						} else {
							t.orderMap[orderID] = OrderResponse{orderID: orderID, actTime: actTime, instrumentID: record.InstrumentId, productID: record.ProductId, orderDetails: order_details}
						}
					}

					t.msg_seq++
					packetCount++
					totalPacketCount++
					loopCount++

					fmt.Println("OE PACKET COUNT :", packetCount)
				}

				map_length := len(t.orderMap)
				if UNIQUE_CLIENT_IDENTIFIER {
					map_length = len(t.uiOrderMap)
				}

				if SEND_MOD && map_length > 0 {
				outerLoopMod:
					for n := 0; n < MOD_PER_ORDER; n++ {
						if UNIQUE_CLIENT_IDENTIFIER {
							for _, order_map := range t.uiOrderMap {
								t.ClOrdID++

								if throttleRate != THROTTLE_VALUE && PATTERN_TYPE != PEAK_GENERATOR {
									fmt.Println("2. Throttle value changed, execution will begin from start!")
									t.uiOrderMap = make(map[uint64]UiOrderResponse)
									break outerLoopMod
								}

								if order_map.orderDetails.buy_min_price == 0 || order_map.orderDetails.buy_max_price == 0 || order_map.orderDetails.price_factor == 0 ||
									order_map.orderDetails.sell_min_price == 0 || order_map.orderDetails.sell_max_price == 0 || order_map.orderDetails.min_qty == 0 || order_map.orderDetails.max_qty == 0 {
									delete(t.uiOrderMap, order_map.origClOrdID)
									continue
								}

								if order_type == BUY {
									price = GenerateRandomPrice(order_map.orderDetails.buy_min_price, order_map.orderDetails.buy_max_price, order_map.orderDetails.price_factor)
									qty = GenerateRandomQty(order_map.orderDetails.min_qty, order_map.orderDetails.mid_qty)
								} else if order_type == SELL {
									price = GenerateRandomPrice(order_map.orderDetails.sell_min_price, order_map.orderDetails.sell_max_price, order_map.orderDetails.price_factor)
									qty = GenerateRandomQty(order_map.orderDetails.mid_qty, order_map.orderDetails.max_qty)
								}

								fmt.Println("Client Order ID: ", t.ClOrdID, " | Org Client Order ID: ", order_map.origClOrdID)

								_, actTime, template, err := t.sendOrderModify(order_map.instrumentID, price, qty, order_type, order_map.origClOrdID, order_map.actTime)
								if err != nil {
									fmt.Printf("Modify error for session id %d: %v", t.session_id, err)
								}

								if template == MODIFY_LEAN_ORDER {
									oMap := t.uiOrderMap[order_map.origClOrdID]
									delete(t.uiOrderMap, order_map.origClOrdID)
									oMap.origClOrdID = t.ClOrdID
									oMap.actTime = actTime
									t.uiOrderMap[oMap.origClOrdID] = oMap
									fmt.Println("Unique identifier Map Mod :", t.uiOrderMap)
								} else {
									delete(t.uiOrderMap, order_map.origClOrdID)
								}

								packetCount++
								t.msg_seq++
								totalPacketCount++

								// fmt.Println("MODIFY ORDER iteration success Packet count :", packetCount)
								fmt.Println("OM PACKET COUNT :", packetCount)

								elapsedTime = time.Since(startTime)
								fmt.Println("OM ELAPSED TIME :", elapsedTime)

								if packetCount >= THROTTLE_VALUE || elapsedTime >= maxInterval {
									if THROTTLE_VALUE == 0 {
										fmt.Println("3. Exiting Data Channel as throttle is set to zero and msg seq is", t.msg_seq)
										t.orderMap = make(map[uint64]OrderResponse)
										t.uiOrderMap = make(map[uint64]UiOrderResponse)
										break dataChanLoop
									} else if elapsedTime >= maxInterval {
										oldPacketCount := packetCount

										fmt.Println("MOD ELAPSED TIME := Packet count :", packetCount, "| THROTTLE value : ", THROTTLE_VALUE, "Delay :", DELAY_TIME)

										if err := t.handleCancelOrders(&packetCount, &totalPacketCount); err != nil {
											fmt.Printf("Cancel error for session id %d: %v", t.session_id, err)
										}

										packetCount = oldPacketCount

										fmt.Println("OM => OC PACKET COUNT :", packetCount)
									} else if packetCount >= THROTTLE_VALUE && elapsedTime < maxInterval {
										// fmt.Printf("Session id %d , sleeping for %v\n", t.session_id, 1000*time.Millisecond-elapsedTime)
										// time.Sleep(time.Second - elapsedTime)

										fmt.Println("Packet count :", packetCount, "| THROTTLE value : ", THROTTLE_VALUE, "Delay :", DELAY_TIME)

										if PATTERN_GENERATOR && PATTERN_TYPE == PEAK_GENERATOR {
											fmt.Printf("OM Session id %d , sleeping for %v (peak generator)\n", t.session_id, time.Duration(DELAY_TIME)*time.Millisecond)
											time.Sleep(time.Duration(DELAY_TIME) * time.Millisecond)
										} else {
											fmt.Printf("OM Session id %d , sleeping for %v\n", t.session_id, 1000*time.Millisecond-elapsedTime)
											time.Sleep(time.Second - elapsedTime)
										}

									}

									fmt.Printf("OM Session id %d sent %d packets.\n", t.session_id, packetCount)

									startTime = time.Now()
									packetCount = 0
									t.uiOrderMap = make(map[uint64]UiOrderResponse)

									fmt.Println("Exiting Mod for loop!", len(t.uiOrderMap))
									break outerLoopMod
								}
							}

						} else {
							for _, order_map := range t.orderMap {

								if throttleRate != THROTTLE_VALUE && PATTERN_TYPE != PEAK_GENERATOR {
									fmt.Println("2. Throttle value changed, execution will begin from start!")
									t.orderMap = make(map[uint64]OrderResponse)
									break outerLoopMod
								}

								if order_map.orderDetails.buy_min_price == 0 || order_map.orderDetails.buy_max_price == 0 || order_map.orderDetails.price_factor == 0 ||
									order_map.orderDetails.sell_min_price == 0 || order_map.orderDetails.sell_max_price == 0 || order_map.orderDetails.min_qty == 0 || order_map.orderDetails.max_qty == 0 {
									delete(t.orderMap, order_map.orderID)
									continue
								}

								if order_type == BUY {
									price = GenerateRandomPrice(order_map.orderDetails.buy_min_price, order_map.orderDetails.buy_max_price, order_map.orderDetails.price_factor)
									qty = GenerateRandomQty(order_map.orderDetails.min_qty, order_map.orderDetails.mid_qty)
								} else if order_type == SELL {
									price = GenerateRandomPrice(order_map.orderDetails.sell_min_price, order_map.orderDetails.sell_max_price, order_map.orderDetails.price_factor)
									qty = GenerateRandomQty(order_map.orderDetails.mid_qty, order_map.orderDetails.max_qty)
								}

								orderID, actTime, template, err := t.sendOrderModify(order_map.instrumentID, price, qty, order_type, order_map.orderID, order_map.actTime)
								if err != nil {
									fmt.Printf("Modify error for session id %d: %v", t.session_id, err)
								}

								if template == MODIFY_LEAN_ORDER {
									oMap := t.orderMap[orderID]
									oMap.orderID = orderID
									oMap.actTime = actTime
									t.orderMap[orderID] = oMap
								} else {
									delete(t.orderMap, order_map.orderID)

								}

								packetCount++
								t.msg_seq++
								totalPacketCount++

								fmt.Println("OM PACKET COUNT :", packetCount)

								elapsedTime = time.Since(startTime)
								fmt.Println("OE ELAPSED TIME :", elapsedTime)

								if packetCount >= THROTTLE_VALUE || elapsedTime >= maxInterval {
									if THROTTLE_VALUE == 0 {
										fmt.Println("3. Exiting Data Channel as throttle is set to zero and msg seq is", t.msg_seq)
										t.orderMap = make(map[uint64]OrderResponse)
										t.uiOrderMap = make(map[uint64]UiOrderResponse)
										break dataChanLoop
									} else if elapsedTime >= maxInterval {
										oldPacketCount := packetCount

										if err := t.handleCancelOrders(&packetCount, &totalPacketCount); err != nil {
											fmt.Printf("Cancel error for session id %d: %v", t.session_id, err)
										}

										packetCount = oldPacketCount

										fmt.Println("OM => OC PACKET COUNT :", packetCount)
									} else if packetCount >= THROTTLE_VALUE && elapsedTime < maxInterval {
										// fmt.Printf("Session id %d , sleeping for %v\n", t.session_id, 1000*time.Millisecond-elapsedTime)
										// time.Sleep(time.Second - elapsedTime)

										fmt.Println("Packet count :", packetCount, "| THROTTLE value : ", THROTTLE_VALUE, "Delay :", DELAY_TIME)

										if PATTERN_GENERATOR && PATTERN_TYPE == PEAK_GENERATOR {
											fmt.Printf("OM Session id %d , sleeping for %v (peak generator)\n", t.session_id, time.Duration(DELAY_TIME)*time.Millisecond)
											time.Sleep(time.Duration(DELAY_TIME) * time.Millisecond)
										} else {
											fmt.Printf("OM Session id %d , sleeping for %v\n", t.session_id, 1000*time.Millisecond-elapsedTime)
											time.Sleep(time.Second - elapsedTime)
										}
									}

									fmt.Printf("OM Session id %d sent %d packets.\n", t.session_id, packetCount)

									startTime = time.Now()
									packetCount = 0
									t.orderMap = make(map[uint64]OrderResponse)

									fmt.Println("Exiting Mod for loop!", len(t.orderMap))
									break outerLoopMod
								}
							}
						}
					}
				}

				if err := t.handleCancelOrders(&packetCount, &totalPacketCount); err != nil {
					fmt.Printf("Cancel error for session id %d: %v", t.session_id, err)
				}

				t.orderMap = make(map[uint64]OrderResponse)
				t.uiOrderMap = make(map[uint64]UiOrderResponse)

				if len(recordsChan) == 0 && !SWITCH_CH {

					PATTERN_GENERATOR = false

					fmt.Printf("Session id %d took total time: %v and sent packets: %v\n",
						t.session_id, time.Since(totalTime), totalPacketCount)
					return nil
				}
			}
		} else {
			PATTERN_GENERATOR = false

			fmt.Println("Throttle is 0 and Heartbeat started!")
			if err := t.heartBeatListener(); err != nil {
				return fmt.Errorf("HeartBeat err: %v ", err)
			}
		}
	}
}

// func (t *Trader) handleCancelOrders(packetCount, totalPacketCount *int) error {

// 	map_length := len(t.orderMap)
// 	if UNIQUE_CLIENT_IDENTIFIER {
// 		map_length = len(t.uiOrderMap)
// 	}

// 	if CANCEL_PERCENT != 0 && map_length > 0 {
// 		cancel_count := 0

// 		if UNIQUE_CLIENT_IDENTIFIER {

// 			quit := make(chan struct{})
// 			go t.NewOrderResponseListnerAsync(quit, DUMMY_NO)

// 			for _, order_map := range t.uiOrderMap {
// 				if cancel_count == CANCEL_ORDER_COUNT {
// 					t.uiOrderMap = make(map[uint64]UiOrderResponse)
// 					fmt.Println("Exiting cancel loop and cancel order count:", cancel_count)
// 					break
// 				}
// 				t.ClOrdID++

// 				t.sendOrderCancelAsync(order_map.instrumentID, order_map.productID, order_map.origClOrdID)

// 				delete(t.uiOrderMap, order_map.origClOrdID)
// 				t.msg_seq++
// 				*packetCount++
// 				*totalPacketCount++
// 				cancel_count++
// 			}

// 		} else {
// 			for _, order_map := range t.orderMap {
// 				if cancel_count == CANCEL_ORDER_COUNT {
// 					t.orderMap = make(map[uint64]OrderResponse)
// 					fmt.Println("Exiting cancel loop and cancel order count:", cancel_count)
// 					break
// 				}

// 				if err := t.sendOrderCancel(order_map.instrumentID, order_map.productID, order_map.orderID); err != nil {
// 					fmt.Printf("[ERR] Cancel error for session id %d: %v", t.session_id, err)
// 					// return err
// 				}
// 				delete(t.orderMap, order_map.orderID)
// 				t.msg_seq++
// 				*packetCount++
// 				*totalPacketCount++
// 				cancel_count++
// 			}
// 		}

// 		fmt.Println("Total CANCEL ORDERS :", cancel_count)
// 	}
// 	return nil
// }

func (t *Trader) handleCancelOrders(packetCount, totalPacketCount *int) error {
	map_length := len(t.orderMap)
	if UNIQUE_CLIENT_IDENTIFIER {
		map_length = len(t.uiOrderMap)
	}
	if CANCEL_PERCENT != 0 && map_length > 0 {
		cancel_count := 0
		if UNIQUE_CLIENT_IDENTIFIER {
			for _, order_map := range t.uiOrderMap {
				if cancel_count == CANCEL_ORDER_COUNT {
					t.uiOrderMap = make(map[uint64]UiOrderResponse)
					fmt.Println("Exiting cancel loop and cancel order count:", cancel_count)
					break
				}
				t.ClOrdID++
				if err := t.sendOrderCancel(order_map.instrumentID, order_map.productID, order_map.origClOrdID); err != nil {
					fmt.Printf("[ERR] Cancel error for session id %d: %v", t.session_id, err)
					// return err
				}
				delete(t.uiOrderMap, order_map.origClOrdID)
				t.msg_seq++
				*packetCount++
				*totalPacketCount++
				cancel_count++
			}
		} else {
			for _, order_map := range t.orderMap {
				if cancel_count == CANCEL_ORDER_COUNT {
					t.orderMap = make(map[uint64]OrderResponse)
					fmt.Println("Exiting cancel loop and cancel order count:", cancel_count)
					break
				}
				if err := t.sendOrderCancel(order_map.instrumentID, order_map.productID, order_map.orderID); err != nil {
					fmt.Printf("[ERR] Cancel error for session id %d: %v", t.session_id, err)
					// return err
				}
				delete(t.orderMap, order_map.orderID)
				t.msg_seq++
				*packetCount++
				*totalPacketCount++
				cancel_count++
			}
		}
		fmt.Println("Total CANCEL ORDERS :", cancel_count)
	}
	return nil
}

//================================================================== ASYNC IMLEMENTATION ======================================================================

func (t *Trader) NewHandleOrderEntryRecordsAsync() error {

	startTime := time.Now()
	totalTime := time.Now()
	packetCount := 0
	totalPacketCount := 0
	count := 0

	err := setOrderPumpingParameters()
	if err != nil {
		fmt.Println("Error in setting Order Pumping Parameters.")
	}

	for {
		if THROTTLE_VALUE > 0 {
		dataChanLoop:
			for data := range recordsChan {

				throttleRate := THROTTLE_VALUE

				if err := t.DeleteOrdersByType(); err != nil {
					fmt.Printf("Delete all orders error for session id %d: %v", t.session_id, err)
				}

				if THROTTLE_VALUE == 0 {
					fmt.Println("1. Exiting Data Channel as throttle is set to zero and msg seq is", t.msg_seq)
					t.uiOrderMap = make(map[uint64]UiOrderResponse)
					break dataChanLoop
				}

				count++
				fmt.Printf("Packet No. (Iteration) in data channel %d for session id: %v \n", count, t.session_id)

				client_code := CLIENT3
				order_type := BUY
				if t.session_id%2 == 0 {
					order_type = SELL
					client_code = CLIENT4
				}

				price := 0
				qty := 0

				maxInterval := time.Second
				if PATTERN_GENERATOR && PATTERN_TYPE == PEAK_GENERATOR {
					maxInterval = time.Duration(INTERVAL_TIME) * time.Millisecond
				}

				elapsedTime := time.Since(startTime)
				fmt.Println("Elapsed Time :", elapsedTime)

				loopCount := 0
				dataLen := len(data)

				for loopCount < ORDER_ENTRY_COUNT {

					record := data[loopCount%dataLen]
					elapsedTime = time.Since(startTime)
					if packetCount >= THROTTLE_VALUE || elapsedTime >= maxInterval {

						if THROTTLE_VALUE == 0 {
							fmt.Println("2. Exiting Data Channel as throttle is set to zero and msg seq is", t.msg_seq)
							t.uiOrderMap = make(map[uint64]UiOrderResponse)
							break dataChanLoop
						} else if packetCount >= THROTTLE_VALUE && elapsedTime < maxInterval {
							if PATTERN_GENERATOR && PATTERN_TYPE == PEAK_GENERATOR {
								fmt.Printf("Session id %d , sleeping for %v\n", t.session_id, time.Duration(DELAY_TIME)*time.Millisecond)
								time.Sleep(time.Duration(DELAY_TIME) * time.Millisecond)
							} else {
								fmt.Printf("Session id %d , sleeping for %v\n", t.session_id, 1000*time.Millisecond-elapsedTime)
								time.Sleep(time.Second - elapsedTime)
							}
						}

						fmt.Printf("Session id %d sent %d packets.\n", t.session_id, packetCount)

						startTime = time.Now()
						packetCount = 0
						SEND_MOD = true
						t.uiOrderMap = make(map[uint64]UiOrderResponse)
					}

					t.ClOrdID++

					if throttleRate != THROTTLE_VALUE && PATTERN_TYPE != PEAK_GENERATOR {
						fmt.Println("1. Throttle value changed, execution will begin from start!")
						t.uiOrderMap = make(map[uint64]UiOrderResponse)
						break
					}

					price_factor := record.BidIntrvl
					buy_min_price, buy_max_price, sell_min_price, sell_max_price := GetPriceRange(record.Price, record.MaxPrice, price_factor)

					mid_qty := record.Qty * (1 + record.MaxOrderQty/(record.Qty*2))
					if mid_qty > record.MaxOrderQty {
						mid_qty = record.MaxOrderQty
					}

					order_details := OrderDetails{price_factor: price_factor, buy_min_price: buy_min_price, buy_max_price: buy_max_price, sell_min_price: sell_min_price, sell_max_price: sell_max_price, min_qty: record.Qty, mid_qty: mid_qty, max_qty: record.MaxOrderQty}

					if order_type == BUY {
						price = GenerateRandomPrice(buy_min_price, buy_max_price, price_factor)
						qty = GenerateRandomQty(record.Qty, mid_qty)
					} else if order_type == SELL {
						price = GenerateRandomPrice(sell_min_price, sell_max_price, price_factor)
						qty = GenerateRandomQty(mid_qty, record.MaxOrderQty)
					}

					orderRequest, err := SingleLegLeanOrder(t.MID, t.TID, t.msg_seq, record.InstrumentId, price, qty, order_type, t.CTX, t.ClOrdID, client_code)
					if err != nil {
						return err
					}

					if err := t.sendDataToServer(orderRequest); err != nil {
						fmt.Printf("Error sending record with session id %d: %v", t.session_id, err)
					}

					orderID, actTime, template, err := t.NewOrderResponseListner(record.InstrumentId)
					if err != nil {
						return fmt.Errorf("response error for session id %d: %v", t.session_id, err)
					}

					if template == LEAN_ORDER {
						if UNIQUE_CLIENT_IDENTIFIER {
							t.uiOrderMap[t.ClOrdID] = UiOrderResponse{origClOrdID: t.ClOrdID, actTime: actTime, instrumentID: record.InstrumentId, productID: record.ProductId, orderDetails: order_details}
							fmt.Println("Unique identifier Map :", t.uiOrderMap)
						} else {
							t.orderMap[orderID] = OrderResponse{orderID: orderID, actTime: actTime, instrumentID: record.InstrumentId, productID: record.ProductId, orderDetails: order_details}
						}
					}

					t.msg_seq++
					packetCount++
					totalPacketCount++
					loopCount++
				}

				map_length := len(t.orderMap)
				if UNIQUE_CLIENT_IDENTIFIER {
					map_length = len(t.uiOrderMap)
				}

				if SEND_MOD && map_length > 0 {
					// quit := make(chan struct{})
					// go t.NewOrderResponseListnerAsync(quit, DUMMY_NO)
				outerLoopMod:
					for n := 0; n < MOD_PER_ORDER; n++ {
						if UNIQUE_CLIENT_IDENTIFIER {
							fmt.Println("Before Mod Unique identifier Map :", t.uiOrderMap)
							for _, order_map := range t.uiOrderMap {
								t.ClOrdID++

								if throttleRate != THROTTLE_VALUE && PATTERN_TYPE != PEAK_GENERATOR {
									fmt.Println("2. Throttle value changed, execution will begin from start!")
									t.uiOrderMap = make(map[uint64]UiOrderResponse)
									// close(quit)
									break outerLoopMod
								}

								if order_map.orderDetails.buy_min_price == 0 || order_map.orderDetails.buy_max_price == 0 || order_map.orderDetails.price_factor == 0 ||
									order_map.orderDetails.sell_min_price == 0 || order_map.orderDetails.sell_max_price == 0 || order_map.orderDetails.min_qty == 0 || order_map.orderDetails.max_qty == 0 {
									delete(t.uiOrderMap, order_map.origClOrdID)
									continue
								}

								if order_type == BUY {
									price = GenerateRandomPrice(order_map.orderDetails.buy_min_price, order_map.orderDetails.buy_max_price, order_map.orderDetails.price_factor)
									qty = GenerateRandomQty(order_map.orderDetails.min_qty, order_map.orderDetails.mid_qty)
								} else if order_type == SELL {
									price = GenerateRandomPrice(order_map.orderDetails.sell_min_price, order_map.orderDetails.sell_max_price, order_map.orderDetails.price_factor)
									qty = GenerateRandomQty(order_map.orderDetails.mid_qty, order_map.orderDetails.max_qty)
								}

								t.sendOrderModifyAsync(order_map.instrumentID, price, qty, order_type, order_map.origClOrdID, order_map.actTime)
								if err != nil {
									fmt.Printf("Modify error for session id %d: %v", t.session_id, err)
								}

								oMap := t.uiOrderMap[order_map.origClOrdID]
								delete(t.uiOrderMap, order_map.origClOrdID)
								oMap.origClOrdID = t.ClOrdID
								oMap.actTime = UINT64_MAX
								t.uiOrderMap[oMap.origClOrdID] = oMap

								packetCount++
								t.msg_seq++
								totalPacketCount++

								fmt.Println("MODIFY ORDER iteration success Packet count :", packetCount)

								elapsedTime = time.Since(startTime)
								fmt.Println("Elapsed Time in Modify :", elapsedTime)

								if packetCount >= THROTTLE_VALUE || elapsedTime >= maxInterval {
									if THROTTLE_VALUE == 0 {
										fmt.Println("3. Exiting Data Channel as throttle is set to zero and msg seq is", t.msg_seq)
										t.uiOrderMap = make(map[uint64]UiOrderResponse)
										// close(quit)
										break dataChanLoop
									} else if elapsedTime >= maxInterval {
										oldPacketCount := packetCount

										if err := t.handleCancelOrdersAsync(&packetCount, &totalPacketCount); err != nil {
											fmt.Printf("Cancel error for session id %d: %v", t.session_id, err)
										}

										packetCount = oldPacketCount
									} else if packetCount >= THROTTLE_VALUE && elapsedTime < maxInterval {
										if PATTERN_GENERATOR && PATTERN_TYPE == PEAK_GENERATOR {
											fmt.Printf("Session id %d , sleeping for %v\n", t.session_id, time.Duration(DELAY_TIME)*time.Millisecond)
											time.Sleep(time.Duration(DELAY_TIME) * time.Millisecond)
										} else {
											fmt.Printf("Session id %d , sleeping for %v\n", t.session_id, 1000*time.Millisecond-elapsedTime)
											time.Sleep(time.Second - elapsedTime)
										}
									}

									fmt.Printf("Session id %d sent %d packets.\n", t.session_id, packetCount)

									startTime = time.Now()
									packetCount = 0
									t.uiOrderMap = make(map[uint64]UiOrderResponse)
									// close(quit)
									fmt.Println("Exiting Mod for loop!", len(t.uiOrderMap))
									break outerLoopMod
								}
							}

						}
					}
				}

				if err := t.handleCancelOrdersAsync(&packetCount, &totalPacketCount); err != nil {
					fmt.Printf("Cancel error for session id %d: %v", t.session_id, err)
				}

				t.uiOrderMap = make(map[uint64]UiOrderResponse)

				if len(recordsChan) == 0 && !SWITCH_CH {

					PATTERN_GENERATOR = false

					fmt.Printf("Session id %d took total time: %v and sent packets: %v\n",
						t.session_id, time.Since(totalTime), totalPacketCount)
					return nil
				}
			}
		} else {
			PATTERN_GENERATOR = false

			fmt.Println("Throttle is 0 and Heartbeat started!")
			if err := t.heartBeatListener(); err != nil {
				return fmt.Errorf("HeartBeat err: %v ", err)
			}
		}
	}
}

func (t *Trader) sendOrderModifyAsync(inst_id, price, qty, order_type int, order_id, act_time uint64) {

	fmt.Println("Send Modify :", act_time)
	modified_data, err := ModifyLeanOrder(t.MID, t.TID, t.msg_seq, inst_id, price, qty, order_type, order_id, act_time, t.CTX, t.ClOrdID, order_id)
	if err != nil {
		fmt.Println("[ERR] Error in order modify parser :", err)
		return
	}

	err = t.sendDataToServer(modified_data)
	if err != nil {
		fmt.Printf("[ERR] Error sending data for session id %d: %v", t.session_id, err)
		return
	}
}

func (t *Trader) sendOrderCancelAsync(inst_id, product_id int, order_id uint64) {

	cancelOrder, err := CancelSingleLegOrderRequest(t.MID, t.TID, t.msg_seq, inst_id, product_id, order_id, t.CTX, t.ClOrdID, order_id)
	if err != nil {
		fmt.Println("[ERR] Error in cancel order", err)
	}

	err = t.sendDataToServer(cancelOrder)
	if err != nil {
		fmt.Printf("[ERR] Error sending cancel order for session id %d: %v", t.session_id, err)
	}
}

func (t *Trader) NewOrderResponseListnerAsync(quit chan struct{}, inst_id int) {
	buffer := make([]byte, 100*1024)
	reader := bufio.NewReader(t.conn)
	for {

		// select {
		// case <-quit:
		// 	fmt.Printf("[INFO] Stopping Order Response Listener for session id %d\n", t.session_id)
		// 	return
		// default:
		n, err := reader.Read(buffer)
		if err != nil {
			if err == io.EOF {
				fmt.Printf("[INFO] Connection closed for session id %d: %v\n", t.session_id, err)

				return
			}
			fmt.Printf("[ERR] Error reading responses for session id %d: %v", t.session_id, err)
			return
		}

		responseBuf := buffer[:n]

		responseLength := len(responseBuf)
		if responseLength == 0 {
			fmt.Println("Response len is 0.")
			return
		}

		var responseArr [][]byte
		count := 0

		for len(responseBuf) > 0 {

			count++

			if len(responseBuf) < 4 {
				fmt.Println("[ERR] Response buffer too short to read size")
				return
			}

			size := int(binary.LittleEndian.Uint32(responseBuf[:4]))
			if len(responseBuf) < size {
				fmt.Println("[ERR] Response buffer too short for the indicated size")
				return
			}

			responseArr = append(responseArr, responseBuf[:size])
			responseBuf = responseBuf[size:]
		}

		fmt.Println("Async Response Listner count:", count)

		for i, response := range responseArr {

			if len(response) == 0 {
				break

			}

			val, err := GetTemplateID(response)
			if err != nil {
				fmt.Println("[ERR] Error in template id parser :", err)
			}

			read_response, err := t.readReceivedResponse(val, inst_id, response)
			if err != nil {
				fmt.Println("[ERR] Error in reading received response :", err)
			}

			if read_response == ORDER_CONFIRMATION {
				fmt.Println("[PASS] Order Confirmation Received for session id:", t.session_id)
				continue
			}

			switch read_response {
			case LEAN_ORDER:
				order_res, err := DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
				if err != nil {
					fmt.Println("[ERR] Error in decryption for order response parser :", err)
				}

				server_res, err1 := GetTimestamps(order_res)
				if err1 != nil {
					fmt.Println("[ERR] Error while getting timestamps")
				} else {
					fmt.Printf("[INFO] Order Entry Server Processing Time(T6-T1):  %v microseconds\n", (server_res.GW_Res_Out-server_res.GW_Req_In)/1000)
				}

				_, _, err = SingleLegLeanOrderResponse(order_res, t.CTX)
				if err != nil {
					fmt.Println("[ERR] Error in lean order response parser :", err)
				}

			case CANCEL_LEAN_ORDER:
				order_res, err := DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
				if err != nil {
					fmt.Println("[ERR] Error in decryption for order response parser :", err)
				}

				server_res, err1 := GetTimestamps(order_res)
				if err1 != nil {
					fmt.Println("[ERR] Error while getting timestamps")
				} else {
					fmt.Printf("[INFO] Order Cancel Server Processing Time(T6-T1):  %v microseconds\n", (server_res.GW_Res_Out-server_res.GW_Req_In)/1000)

				}

			case MODIFY_LEAN_ORDER:
				order_res, err := DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
				if err != nil {
					fmt.Println("[ERR] Error in decryption for order response parser :", err)
				}

				server_res, err1 := GetTimestamps(order_res)

				if err1 != nil {
					fmt.Println("[ERR] Error while getting timestamps")
				} else {
					fmt.Printf("[INFO] Order Mod Server Processing Time(T6-T1):  %v microseconds\n", (server_res.GW_Res_Out-server_res.GW_Req_In)/1000)
				}

				_, _, err = SingleLegLeanOrderModifiedResponse(order_res, t.CTX)

				if err != nil {
					fmt.Println("[ERR] Error in lean order response parser :", err)
				}

			case HEARTBEAT:
				_, err := DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
				if err != nil {
					fmt.Println("[ERR] Error in decryption for order response parser :", err)
				}

			case REJECTED:
				fmt.Println("Rejected case!!")
			case IMMEDIATE_EXECUTION:
				_, err := DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
				if err != nil {
					fmt.Println("[ERR] Error in decryption for order response parser :", err)
				}

			case ORDER_BOOK_EXECUTION:
				_, err := DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
				if err != nil {
					fmt.Println("[ERR] Error in decryption for order response parser :", err)
				}

			case TRADING_SESSION_EVENT:
				_, err := DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
				if err != nil {
					fmt.Println("[ERR] Error in decryption for order response parser :", err)
				}

			case DELETE_ALL_ORDERS:
				_, err := DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
				if err != nil {
					fmt.Println("[ERR] Error in decryption for order response parser :", err)
				}

			case DELETE_ALL_ORDERS_NO_HITS:
				_, err := DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
				if err != nil {
					fmt.Println("[ERR] Error in decryption for order response parser :", err)
				}

			default:
				fmt.Println("Other Template ID: ", read_response)
			}

			if i == len(responseArr)-1 {
				break
			}
			// }
		}
	}
}

func (t *Trader) handleCancelOrdersAsync(packetCount, totalPacketCount *int) error {
	if CANCEL_PERCENT != 0 && len(t.uiOrderMap) > 0 {
		cancel_count := 0
		// quit := make(chan struct{})
		// go t.NewOrderResponseListnerAsync(quit, DUMMY_NO)
		for _, order_map := range t.uiOrderMap {
			if cancel_count == CANCEL_ORDER_COUNT {
				t.uiOrderMap = make(map[uint64]UiOrderResponse)
				// close(quit)
				fmt.Println("Exiting cancel loop and cancel order count:", cancel_count)
				break
			}
			t.ClOrdID++
			t.sendOrderCancelAsync(order_map.instrumentID, order_map.productID, order_map.origClOrdID)
			delete(t.uiOrderMap, order_map.origClOrdID)
			t.msg_seq++
			*packetCount++
			*totalPacketCount++
			cancel_count++
		}
		// close(quit)
		fmt.Println("Total CANCEL ORDERS :", cancel_count)
	}
	return nil
}

func (t *Trader) AsyncHandleOrderEntryRecordsOE() error {
	startTime := time.Now()
	totalTime := time.Now()
	packetCount := 0
	totalPacketCount := 0
	count := 0
	err := setOrderPumpingParameters()
	if err != nil {
		fmt.Println("Error in setting Order Pumping Parameters.")
	}

	for {
		if THROTTLE_VALUE > 0 {
		dataChanLoop:
			for data := range recordsChan {
				throttleRate := THROTTLE_VALUE
				if err := t.DeleteOrdersByType(); err != nil {
					fmt.Printf("Delete all orders error for session id %d: %v", t.session_id, err)
				}
				if THROTTLE_VALUE == 0 {
					fmt.Println("1. Exiting Data Channel as throttle is set to zero and msg seq is", t.msg_seq)
					t.orderMap = make(map[uint64]OrderResponse)
					t.uiOrderMap = make(map[uint64]UiOrderResponse)
					break dataChanLoop
				}
				count++
				fmt.Printf("Packet No. (Iteration) in data channel %d for session id: %v \n", count, t.session_id)
				client_code := CLIENT3
				order_type := BUY
				if t.session_id%2 == 0 {
					order_type = SELL
					client_code = CLIENT4
				}
				price := 0
				qty := 0
				maxInterval := time.Second
				if PATTERN_GENERATOR && PATTERN_TYPE == PEAK_GENERATOR {
					maxInterval = time.Duration(INTERVAL_TIME) * time.Millisecond
				}
				elapsedTime := time.Since(startTime)
				fmt.Println("Elapsed Time :", elapsedTime)
				loopCount := 0
				dataLen := len(data)
				for loopCount < ORDER_ENTRY_COUNT {
					record := data[loopCount%dataLen]
					elapsedTime = time.Since(startTime)
					if packetCount >= THROTTLE_VALUE || elapsedTime >= maxInterval {
						if THROTTLE_VALUE == 0 {
							fmt.Println("2. Exiting Data Channel as throttle is set to zero and msg seq is", t.msg_seq)
							t.orderMap = make(map[uint64]OrderResponse)
							t.uiOrderMap = make(map[uint64]UiOrderResponse)
							break dataChanLoop
						} else if packetCount >= THROTTLE_VALUE && elapsedTime < maxInterval {
							// fmt.Printf("Session id %d , sleeping for %v\n", t.session_id, 1000*time.Millisecond-elapsedTime)
							// time.Sleep(time.Second - elapsedTime)
							if PATTERN_GENERATOR && PATTERN_TYPE == PEAK_GENERATOR {
								fmt.Printf("Session id %d , sleeping for %v\n", t.session_id, time.Duration(DELAY_TIME)*time.Millisecond)
								time.Sleep(time.Duration(DELAY_TIME) * time.Millisecond)
							} else {
								fmt.Printf("Session id %d , sleeping for %v\n", t.session_id, 1000*time.Millisecond-elapsedTime)
								time.Sleep(time.Second - elapsedTime)
							}
						}
						fmt.Printf("Session id %d sent %d packets.\n", t.session_id, packetCount)
						startTime = time.Now()
						packetCount = 0
						SEND_MOD = true
						t.orderMap = make(map[uint64]OrderResponse)
						t.uiOrderMap = make(map[uint64]UiOrderResponse)
					}
					if UNIQUE_CLIENT_IDENTIFIER {
						t.ClOrdID++
					}
					if throttleRate != THROTTLE_VALUE && PATTERN_TYPE != PEAK_GENERATOR {
						fmt.Println("1. Throttle value changed, execution will begin from start!")
						t.orderMap = make(map[uint64]OrderResponse)
						t.uiOrderMap = make(map[uint64]UiOrderResponse)
						break
					}
					price_factor := record.BidIntrvl
					buy_min_price, buy_max_price, sell_min_price, sell_max_price := GetPriceRange(record.Price, record.MaxPrice, price_factor)
					mid_qty := record.Qty * (1 + record.MaxOrderQty/(record.Qty*2))
					if mid_qty > record.MaxOrderQty {
						mid_qty = record.MaxOrderQty
					}
					if order_type == BUY {
						price = GenerateRandomPrice(buy_min_price, buy_max_price, price_factor)
						qty = GenerateRandomQty(record.Qty, mid_qty)
					} else if order_type == SELL {
						price = GenerateRandomPrice(sell_min_price, sell_max_price, price_factor)
						qty = GenerateRandomQty(mid_qty, record.MaxOrderQty)
					}
					orderRequest, err := SingleLegLeanOrder(t.MID, t.TID, t.msg_seq, record.InstrumentId, price, qty, order_type, t.CTX, t.ClOrdID, client_code)
					if err != nil {
						return err
					}
					if err := t.sendDataToServer(orderRequest); err != nil {
						fmt.Printf("Error sending record with session id %d: %v", t.session_id, err)
					}
					t.msg_seq++
					packetCount++
					totalPacketCount++
					loopCount++
				}
				t.orderMap = make(map[uint64]OrderResponse)
				t.uiOrderMap = make(map[uint64]UiOrderResponse)
				if len(recordsChan) == 0 && !SWITCH_CH {
					PATTERN_GENERATOR = false
					fmt.Printf("Session id %d took total time: %v and sent packets: %v\n",
						t.session_id, time.Since(totalTime), totalPacketCount)
					return nil
				}
			}
		} else {
			PATTERN_GENERATOR = false
			fmt.Println("Throttle is 0 and Heartbeat started!")
			if err := t.heartBeatListener(); err != nil {
				return fmt.Errorf("HeartBeat err: %v ", err)
			}
		}
	}
}
