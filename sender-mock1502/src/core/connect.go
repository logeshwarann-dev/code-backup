package core

import (
	"bufio"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"io"
	"market_replay/src/encryption/aes"
	"market_replay/src/static"
	"market_replay/src/utils"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

var (

// members          []Member
// recordFile       string

)

const (
	retryConn = 5
)

type Traders struct {
	*static.Trader
}

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
	static.DataChan = make(chan static.DataPacket, static.DATA_CHANNEL_LENGTH)
	static.RecordsChan = make(chan []static.DataPacket, static.DATA_CHANNEL_LENGTH)
	connThrottleChan = make(chan int, static.PATTERN_CHANNEL_LENGTH)

	err = ReadRecords(recordFile)
	if err != nil {
		fmt.Println("[ERR] Error reading records:", err)
		return
	}

	if static.TRADE {
		go PumpData()
	} else {
		go PumpRecords()
	}

	go PatternThrottle()
	go ChangeThrottle()
	// go CheckRejectedSessions()

	for i := range members {
		for j := range members[i].Traders {
			wg.Add(1)
			trader := &members[i].Traders[j]
			go func(t Traders) {
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
	static.DataChan = make(chan static.DataPacket, static.DATA_CHANNEL_LENGTH)
	static.RecordsChan = make(chan []static.DataPacket, static.DATA_CHANNEL_LENGTH)
	static.ConnThrottleChan = make(chan int, static.PATTERN_CHANNEL_LENGTH)

	// ORDER_PUMPING_TYPE = SYNC_ORDER_PUMPING_TYPE //Sync order pumping

	if static.TRADE {
		go PumpData()
	} else {
		go PumpRecords()
	}

	go utils.PatternThrottle()
	go utils.ChangeThrottle()

	static.ORDERPUMPINGSTATUS = true

	for _, tradr := range static.TRADERS {
		wg.Add(1)
		trader := Traders{Trader: &tradr}
		go func(t Traders) {
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
		}(trader)
	}
	wg.Wait()

}

func PumpData() {
	for i := 0; i < static.TOTAL_ORDER_COUNT; i++ {
		record := static.RECORDS[i%len(static.RECORDS)]
		// fmt.Println("Record: ", record)
		static.DataChan <- static.DataPacket{InstrumentId: record.InstrumentID, Price: record.LowerLimit, MaxPrice: record.UpperLimit, ProductId: record.Product_ID, Qty: record.MinLot, BidIntrvl: record.BidInterval, MaxOrderQty: record.MaxTrdQty}
	}
}

func PumpRecords() {
	fmt.Println("In PUMP static.RECORDS...")
	for i := 0; i < static.TOTAL_ORDER_COUNT; i++ {
		var dataPackets []static.DataPacket

		for _, rec := range static.RECORDS {
			dataPackets = append(dataPackets, static.DataPacket{
				InstrumentId: rec.InstrumentID,
				Price:        rec.LowerLimit,
				ProductId:    rec.Product_ID,
				Qty:          rec.MinLot,
				MaxPrice:     rec.UpperLimit,
				BidIntrvl:    rec.BidInterval,
				MaxOrderQty:  rec.MaxTrdQty,
			})
		}

		// fmt.Println("Data packets: ", dataPackets)

		static.RecordsChan <- dataPackets
	}
}

func (t Traders) establishConnection(attempt int) error {

	attempt++
	t.OrderMap = make(map[uint64]static.OrderResponse)
	t.UiOrderMap = make(map[uint64]static.UiOrderResponse)
	t.Session_id = (t.MID * 100000) + t.TID
	t.Msg_seq = 1

	// fmt.Println("Order Pumping Type: ", static.ORDER_PUMPING_TYPE, "| System Vendor: ", static.DISABLED_ACTIVITY_SYSTEM_VENDOR, "| System Version: ", static.LAB_DISABLED_ACTIVITY_SYSTEM_VERSION, "| ", static.PROD_DISABLED_ACTIVITY_SYSTEM_VERSION)
	sslConn, err := t.createConnectionTLS()
	if err != nil {
		return err
	}

	// Step 1: Connection Request/Response
	if err := t.handleConnectionRequest(sslConn); err != nil {
		return err
	}

	sslConn.Close()
	fmt.Println("[INFO] SSL Connection Closed for session id :", t.Session_id)

	conn, err := t.createConnection()
	if err != nil {
		return err
	}
	t.Conn = conn

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
	t.Msg_seq++

	// Step 3: Session Login
	if err := t.handleSessionLogon(); err != nil {
		return err
	}
	t.Msg_seq++

	// Step 4: User Login
	if err := t.handleUserLogin(); err != nil {
		return err
	}
	t.Msg_seq++

	if static.TRADE {
		if err := t.handleOrderEntryTrade(); err != nil {
			return err
		}
	} else {
		switch static.ORDER_PUMPING_TYPE {
		case static.SYNC_ORDER_PUMPING_TYPE:

			if err := t.NewHandleOrderEntryRecords(); err != nil {
				return err
			}
		case static.ASYNC_ORDER_PUMPING_OEOMOC_TYPE:

			res_ch := make(chan struct{})
			go t.NewOrderResponseListnerAsync(res_ch, static.DUMMY_NO)

			if err := t.NewHandleOrderEntryRecordsAsync(); err != nil {
				return err
			}
			close(res_ch)
		case static.ASYNC_ORDER_PUMPING_IOC_TYPE:

			res_ch := make(chan struct{})
			go t.NewOrderResponseListnerAsync(res_ch, static.DUMMY_NO)

			if err := t.AsyncHandleOrderEntryRecordsOE(); err != nil {
				return err
			}
			close(res_ch)

		}

	}
	t.Msg_seq++

	time.Sleep(1000 * time.Millisecond)
	fmt.Printf("[INFO] Ack received by session id  : %v  is %v. \n", t.Session_id, t.AckCount)

	return nil
}

func (t Traders) createConnectionTLS() (*tls.Conn, error) {

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

func (t Traders) createConnection() (net.Conn, error) {

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
		fmt.Println("")
	} else {
		fmt.Println("[ERR] Failed to get TCP Conn.")
	}
	return conn, nil
}

func (t Traders) handleConnectionRequest(sslConn *tls.Conn) error {

	gatewayRequest, err := utils.CreateGatewayConnectionRequest(t.MID, t.TID, t.Password)
	if err != nil {
		return fmt.Errorf("failed to create gateway connection request with session id: %v %v", t.Session_id, err)
	}

	_, err = sslConn.Write(gatewayRequest)
	if err != nil {
		return fmt.Errorf("failed to send tls connection request with session id: %v %v", t.Session_id, err)
	}

	buffer := make([]byte, 100*1024)
	n, err := sslConn.Read(buffer)
	if err != nil {
		return fmt.Errorf("error reading tls connection response for session id: %v %v", t.Session_id, err)
	}

	response := buffer[:n]

	val, err := utils.GetTemplateID(response)
	if err != nil {
		fmt.Println("[ERR] error in template id parser in connection gateway response :", err)
		return nil
	}

	if val == static.REJECTED {
		rejectedMsg, err := utils.ReadRejectedMsgNonDecrypted(response)
		if err != nil {
			fmt.Println("[ERR] Error in rejected message parser connection gateway response:", err)
			return nil
		}
		fmt.Printf("[ERR] Connection Gateway Response for session id %v  has been rejected.\n", t.Session_id)
		return fmt.Errorf("reason: %v", rejectedMsg)
	}

	gatewayResponse, err := utils.ConnectionGatewayResponse(response)
	if err != nil {
		return fmt.Errorf("error parsing connection gateway response: %v", err)
	}

	t.CGW_Res = *gatewayResponse

	fmt.Println("[INFO] Gateway Connected Successfully for session id: ", t.Session_id)
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
		// 	fmt.Fprintf(os.Stderr, "failed to set SO_RCVBUF: %v\n", err)
		// }

		//Linux
		if err := syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_RCVBUF, recvBufSize); err != nil {
			fmt.Fprintf(os.Stderr, "failed to set SO_RCVBUF: %v\n", err)
		}
	})
	if err != nil {
		return fmt.Errorf("failed to set receive buffer size: %v", err)
	}

	// Set receive buffer size
	err = rawConn.Control(func(fd uintptr) {
		// handle := syscall.Handle(fd)
		// if err := syscall.SetsockoptInt(handle, syscall.SOL_SOCKET, syscall.SO_SNDBUF, sendBufSize); err != nil {
		// 	fmt.Fprintf(os.Stderr, "failed to set SO_SNDBUF: %v\n", err)
		// }

		// Linux
		if err := syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_SNDBUF, sendBufSize); err != nil {
			fmt.Fprintf(os.Stderr, "failed to set SO_SNDBUF: %v\n", err)
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
		// 	fmt.Fprintf(os.Stderr, "failed to get SO_SNDBUF: %v\n", err)
		// }

		// _, err = syscall.GetsockoptInt(handle, syscall.SOL_SOCKET, syscall.SO_RCVBUF)
		// if err != nil {
		// 	fmt.Fprintf(os.Stderr, "failed to get SO_RCVBUF: %v\n", err)
		// }

		// Linux
		_, err = syscall.GetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_SNDBUF)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to get SO_SNDBUF: %v\n", err)
		}

		_, err = syscall.GetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_RCVBUF)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to get SO_RCVBUF: %v\n", err)
		}

	})
	if err != nil {
		return fmt.Errorf("failed to retrieve buffer sizes: %v", err)
	}

	return nil
}

func (t Traders) handleSessionRegistration() error {

	sessionRegRequest, err := utils.CreateSessionRegistrationRequest(t.MID, t.TID, t.Msg_seq)
	if err != nil {
		return fmt.Errorf("failed to create session registration request with session id: %v %v", t.Session_id, err)
	}

	_, err = t.Conn.Write(sessionRegRequest)
	if err != nil {
		return fmt.Errorf("failed to send session registration request with session id: %v %v", t.Session_id, err)
	}

	buffer := make([]byte, 100*1024)
	n, err := t.Conn.Read(buffer)
	if err != nil {
		return fmt.Errorf("error reading session registration response for session id: %v %v", t.Session_id, err)
	}

	response := buffer[:n]

	val, err := utils.GetTemplateID(response)
	if err != nil {
		fmt.Println("[ERR] Error in template id parser in session registration :", err)
		return nil
	}

	if val == static.REJECTED {
		rejectedMsg, err := utils.ReadRejectedMsgNonDecrypted(response)
		if err != nil {
			fmt.Println("[ERR] Error in rejected message parser:", err)
			return nil
		}
		fmt.Printf("[ERR] Session Registration for session id %v  has been rejected.\n", t.Session_id)
		return fmt.Errorf("reason: %v", rejectedMsg)
	}

	fmt.Println("[INFO] Session Registration Completed Successfully for session id:", t.Session_id)
	return nil
}

func (t Traders) handleSessionLogon() error {

	sessionLogonRequest, err := utils.CreateSessionLogonRequest(t.MID, t.TID, t.Msg_seq, t.Password, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
	if err != nil {
		return fmt.Errorf("failed to create session logon request with session id: %v %v", t.Session_id, err)
	}

	_, err = t.Conn.Write(sessionLogonRequest)
	if err != nil {
		return fmt.Errorf("failed to send session logon request with session id: %v %v", t.Session_id, err)
	}

	buffer := make([]byte, 100*1024)
	n, err := t.Conn.Read(buffer)
	if err != nil {
		return fmt.Errorf("error reading session logon response for session id: %v %v", t.Session_id, err)
	}

	response := buffer[:n]

	val, err := utils.GetTemplateID(response)
	if err != nil {
		fmt.Println("[ERR] Error in template id parser :", err)
		return nil
	}

	if val == static.REJECTED {
		rejectedMsg, err := utils.ReadRejectedMsg(response, t.CTX)
		if err != nil {
			fmt.Println("[ERR] Error in rejected message parser in session logon:", err)
			return nil
		}
		fmt.Printf("[ERR] Session Logon for session id %v  has been rejected.\n", t.Session_id)
		return fmt.Errorf("reason: %v", rejectedMsg)
	}

	session_res, err := utils.DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
	if err != nil {
		fmt.Println("[ERR] Error in decrypting session logon response :", err)
		return nil
	}

	sessionLogonResponse, err := utils.SessionLogonResponse(session_res)
	if err != nil {
		return fmt.Errorf("error parsing session logon response: %v", err)
	}
	t.Throttle = sessionLogonResponse

	fmt.Println("[INFO] Session Logon Completed Successfully for Session id:", t.Session_id)
	return nil
}

func (t Traders) handleUserLogin() error {

	userLogonRequest, err := utils.CreateUserLogonRequest(t.MID, t.TID, t.Msg_seq, t.Password, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
	if err != nil {
		return fmt.Errorf("failed to create user logon request with session id: %v %v", t.Session_id, err)
	}

	_, err = t.Conn.Write(userLogonRequest)
	if err != nil {
		return fmt.Errorf("failed to send user logon request with session id: %v %v", t.Session_id, err)
	}

	buffer := make([]byte, 100*1024)
	n, err := t.Conn.Read(buffer)
	if err != nil {
		return fmt.Errorf("error reading user logon response with session id: %v %v", t.Session_id, err)
	}

	response := buffer[:n]
	val, err := utils.GetTemplateID(response)
	if err != nil {
		fmt.Println("[ERR] Error in template id parser :", err)
		return nil
	}

	if val == static.REJECTED {
		rejectedMsg, err := utils.ReadRejectedMsg(response, t.CTX)
		if err != nil {
			fmt.Println("[ERR] Error in rejected message parser:", err)
			return nil
		}
		fmt.Printf("[ERR] User Logon rejected | session id %v | Reason: %v\n", t.Session_id, rejectedMsg)
		return fmt.Errorf("reason: %v", rejectedMsg)
	}

	_, err = utils.DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
	if err != nil {
		fmt.Println("[ERR] Error in response user logon parser :", err)
		return nil
	}

	fmt.Println("[SUCCESS] User Logon Completed | session id:", t.Session_id)
	return nil
}

func (t Traders) sendDataToServer(order []byte) error {

	_, err := t.Conn.Write(order)
	if err != nil {
		fmt.Printf("[ERR] Error sending data with session id %d: %v", t.Session_id, err)
	}

	return nil
}

func (t Traders) readReceivedResponse(val uint16, inst_id int, response []byte) (uint16, error) {

	switch val {
	case static.REJECTED:
		rejectedMsg, err := utils.ReadRejectedMsg(response, t.CTX)
		if err != nil {
			fmt.Println("[ERR] Error in rejected message parser:", err)
			return 0, err
		}
		t.AckCount++
		fmt.Printf("[ERR] Order static.REJECTED | session id %v | instrument id - %v | Reason: %v \n", t.Session_id, inst_id, rejectedMsg)
		return static.REJECTED, nil

	case static.STANDARD_ORDER, static.LEAN_ORDER:
		orderType := "Standard Order"
		if val == static.LEAN_ORDER {
			orderType = "Lean Order"
		}
		fmt.Printf("[SUCCESS] Order Entry | session id %v | instrument id - %v (%s)\n", t.Session_id, inst_id, orderType)
		t.AckCount++
		return static.LEAN_ORDER, nil

	case static.HEARTBEAT:
		fmt.Println("[INFO] HeartBeat | session id:", t.Session_id)
		return static.HEARTBEAT, nil

	case static.ORDER_CONFIRMATION:
		_, err := utils.DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
		if err != nil {
			fmt.Println("[ERR] Error in decrypting order confirmation :", err)
			return 0, err
		}
		return static.ORDER_CONFIRMATION, nil

	case static.MODIFY_LEAN_ORDER:
		fmt.Printf("[SUCCESS] Order Modify | session id %v | instrument id - %v (Lean Order).\n", t.Session_id, inst_id)
		t.AckCount++
		return static.MODIFY_LEAN_ORDER, nil

	case static.CANCEL_LEAN_ORDER:
		fmt.Printf("[SUCCESS] Order Cancel | session id %v | instrument id - %v (Lean Order)\n", t.Session_id, inst_id)
		t.AckCount++
		return static.CANCEL_LEAN_ORDER, nil
	case static.IMMEDIATE_EXECUTION:
		fmt.Printf("[SUCCESS] IMMEDIATE EXECUTION | session id %v\n", t.Session_id)
		t.AckCount++
		return static.IMMEDIATE_EXECUTION, nil
	case static.ORDER_BOOK_EXECUTION:
		fmt.Printf("[SUCCESS] ORDER BOOK EXECUTION | session id %v\n", t.Session_id)
		t.AckCount++
		return static.ORDER_BOOK_EXECUTION, nil
	case static.TRADING_SESSION_EVENT:
		fmt.Printf("[SUCCESS] static.TRADING_SESSION_EVENT | session id %v\n", t.Session_id)
		t.AckCount++
		return static.TRADING_SESSION_EVENT, nil
	case static.DELETE_ALL_ORDERS:
		fmt.Printf("[SUCCESS] static.DELETE_ALL_ORDERS | session id %v\n", t.Session_id)
		t.AckCount++
		return static.DELETE_ALL_ORDERS, nil
	case static.DELETE_ALL_ORDERS_NO_HITS:
		fmt.Printf("[SUCCESS] static.DELETE_ALL_ORDERS_NO_HITS | session id %v\n", t.Session_id)
		t.AckCount++
		return static.DELETE_ALL_ORDERS_NO_HITS, nil

	default:
		return 0, nil
	}
}

func (t Traders) heartBeatListener() error {
	for {
		_, _, order_template, err := t.NewOrderResponseListner(static.DUMMY_ID)
		if err != nil {
			fmt.Println("[ERR] Error in reading received response :", err)
		}

		if order_template == static.HEARTBEAT {
			heartbeat, err := utils.HeartBeatRequest()
			if err != nil {
				return err
			}

			t.sendDataToServer(heartbeat)

			if static.THROTTLE_VALUE > 0 {
				fmt.Println("[INFO] Throttle value is updated. Exiting heartbeat case ")
				return nil
			}
		}
	}
}

func (t Traders) sendOrderModify(inst_id, price, qty, order_type int, order_id, act_time uint64) (uint64, uint64, uint16, error) {

	modified_data, err := utils.ModifyLeanOrder(t.MID, t.TID, t.Msg_seq, inst_id, price, qty, order_type, order_id, act_time, t.CTX, t.ClOrdID, order_id)
	if err != nil {
		fmt.Println("[ERR] Error in order modify parser :", err)
		// return nil
		return 0, 0, 0, nil
	}

	modOrderTime := time.Now()
	err = t.sendDataToServer(modified_data)
	if err != nil {
		fmt.Printf("[ERR] Error sending data for session id %d: %v", t.Session_id, err)
		return 0, 0, 0, nil
	}

	order_id, act_time, order_template, err := t.NewOrderResponseListner(inst_id)
	fmt.Println("[INFO] Mod Order Request RTT: ", time.Since(modOrderTime))
	if err != nil {
		fmt.Println("[ERR] Error in reading received response :", err)
	}

	return order_id, act_time, order_template, nil
}

func (t Traders) sendOrderCancel(inst_id, product_id int, order_id uint64) error {

	cancelOrder, err := utils.CancelSingleLegOrderRequest(t.MID, t.TID, t.Msg_seq, inst_id, product_id, order_id, t.CTX, t.ClOrdID, order_id)
	if err != nil {
		fmt.Println("[ERR] Error in cancel order", err)
	}

	cancelOrderTime := time.Now()
	err = t.sendDataToServer(cancelOrder)
	if err != nil {
		fmt.Printf("[ERR] Error sending cancel order for session id %d: %v", t.Session_id, err)
		return err
	}

	_, _, _, err = t.NewOrderResponseListner(inst_id)
	fmt.Println("[INFO] Cancel Order Request RTT: ", time.Since(cancelOrderTime))
	if err != nil {
		fmt.Println("[ERR] Error in reading received response :", err)
	}

	return nil
}

func (t Traders) sendOrderDelete(inst_id, product_id int) error {

	deleteOrder, err := utils.DeleteAllOrderComplexRequest(inst_id, product_id, t.MID, t.TID, t.Msg_seq, t.CTX)
	if err != nil {
		fmt.Println("[ERR] Error in Delete order", err)
	}

	deleteOrderTime := time.Now()
	err = t.sendDataToServer(deleteOrder)
	if err != nil {
		fmt.Printf("[ERR] Error sending Delete order for session id %d: %v", t.Session_id, err)
		return err
	}

	_, _, _, err = t.NewOrderResponseListner(inst_id)
	fmt.Println("[INFO] Delete Order Request RTT: ", time.Since(deleteOrderTime))
	if err != nil {
		fmt.Println("[ERR] Error in reading received response :", err)
	}

	return nil
}

func (t Traders) DeleteOrdersByType() error {

	if static.DELETE_ORDERS_ALL_SESSIONS != t.Delete_order {
		switch static.DELETE_ORDERS_DETAILS.Type {

		case static.DELETE_SPECIFIC_MEMBER:
			if _, exists := static.DELETE_ORDERS_DETAILS.Ids[t.MID]; exists {
				if err := t.sendOrderDelete(static.DELETE_ORDERS_DETAILS.InstrumentID, static.DELETE_ORDERS_DETAILS.ProductID); err != nil {
					return fmt.Errorf("delete Error for specific member session id %d: %v", t.Session_id, err)
				}

				// delete(static.DELETE_ORDERS_DETAILS.Ids, t.MID)
				t.Delete_order = static.DELETE_ORDERS_ALL_SESSIONS
				t.Msg_seq++
				t.OrderMap = make(map[uint64]static.OrderResponse)
			}

		case static.DELETE_SPECIFIC_SESSION:
			if _, exists := static.DELETE_ORDERS_DETAILS.Ids[t.Session_id]; exists {
				if err := t.sendOrderDelete(static.DELETE_ORDERS_DETAILS.InstrumentID, static.DELETE_ORDERS_DETAILS.ProductID); err != nil {
					return fmt.Errorf("delete Error for specific session id %d: %v", t.Session_id, err)
				}
				delete(static.DELETE_ORDERS_DETAILS.Ids, t.Session_id)
				t.Delete_order = static.DELETE_ORDERS_ALL_SESSIONS
				t.Msg_seq++
				t.OrderMap = make(map[uint64]static.OrderResponse)
			}

		case static.DELETE_ALL:
			if err := t.sendOrderDelete(static.DELETE_ORDERS_DETAILS.InstrumentID, static.DELETE_ORDERS_DETAILS.ProductID); err != nil {
				return fmt.Errorf("delete Error for session id %d: %v", t.Session_id, err)
			}

			t.Delete_order = static.DELETE_ORDERS_ALL_SESSIONS
			t.Msg_seq++
			t.OrderMap = make(map[uint64]static.OrderResponse)

		default:
			fmt.Println("DELETE ORDERS BY TYPE NO CASE MATCHED:", static.DELETE_ORDERS_DETAILS.Type)
		}
	}

	return nil
}

func (t Traders) NewOrderResponseListner(inst_id int) (uint64, uint64, uint16, error) {

	for {
		reader := bufio.NewReader(t.Conn)
		buffer := make([]byte, 100*1024)
		n, err := reader.Read(buffer)
		if err != nil {
			if err == io.EOF {
				fmt.Printf("[INFO] Connection closed for session id %d: %v\n", t.Session_id, err)
				return 0, 0, 0, err
			}
			fmt.Printf("[ERR] Error reading responses for session id %d: %v", t.Session_id, err)
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

			val, err := utils.GetTemplateID(response)
			if err != nil {
				fmt.Println("[ERR] Error in template id parser :", err)
			}

			read_response, err := t.readReceivedResponse(val, inst_id, response)
			if err != nil {
				fmt.Println("[ERR] Error in reading received response :", err)
			}

			if read_response == static.ORDER_CONFIRMATION {
				fmt.Println("[PASS] Order Confirmation Received for session id:", t.Session_id)
				continue
			}

			switch read_response {
			case static.LEAN_ORDER:
				order_res, err := utils.DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
				if err != nil {
					fmt.Println("[ERR] Error in decryption for order response parser :", err)
				}

				server_res, err1 := utils.GetTimestamps(order_res)
				if err1 != nil {
					fmt.Println("[ERR] Error while getting timestamps")
				} else {
					fmt.Printf("[INFO] Order Entry Server Processing Time(T6-T1):  %v microseconds\n", (server_res.GW_Res_Out-server_res.GW_Req_In)/1000)
				}

				order_id, act_time, err := utils.SingleLegLeanOrderResponse(order_res, t.CTX)
				if err != nil {
					fmt.Println("[ERR] Error in lean order response parser :", err)
				}

				if i == len(responseArr)-1 {
					return order_id, act_time, static.LEAN_ORDER, nil
				}
			case static.CANCEL_LEAN_ORDER:
				order_res, err := utils.DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
				if err != nil {
					fmt.Println("[ERR] Error in decryption for order response parser :", err)
				}

				server_res, err1 := utils.GetTimestamps(order_res)
				if err1 != nil {
					fmt.Println("[ERR] Error while getting timestamps")
				} else {
					fmt.Printf("[INFO] Order Cancel Server Processing Time(T6-T1):  %v microseconds\n", (server_res.GW_Res_Out-server_res.GW_Req_In)/1000)
				}

				if i == len(responseArr)-1 {
					return 0, 0, static.CANCEL_LEAN_ORDER, nil
				}
			case static.MODIFY_LEAN_ORDER:
				order_res, err := utils.DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
				if err != nil {
					fmt.Println("[ERR] Error in decryption for order response parser :", err)
				}

				server_res, err1 := utils.GetTimestamps(order_res)
				if err1 != nil {
					fmt.Println("[ERR] Error while getting timestamps")
				} else {
					fmt.Printf("[INFO] Order Mod Server Processing Time(T6-T1):  %v microseconds\n", (server_res.GW_Res_Out-server_res.GW_Req_In)/1000)
				}

				order_id, act_time, err := utils.SingleLegLeanOrderModifiedResponse(order_res, t.CTX)
				if err != nil {
					fmt.Println("[ERR] Error in lean order response parser :", err)
				}

				if i == len(responseArr)-1 {
					return order_id, act_time, static.MODIFY_LEAN_ORDER, nil
				}
			case static.HEARTBEAT:
				_, err := utils.DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
				if err != nil {
					fmt.Println("[ERR] Error in decryption for order response parser :", err)
				}

				if i == len(responseArr)-1 {
					return 0, 0, static.HEARTBEAT, nil
				}
			case static.REJECTED:
				if i == len(responseArr)-1 {
					return 0, 0, static.REJECTED, nil
				}
			case static.IMMEDIATE_EXECUTION:
				_, err := utils.DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
				if err != nil {
					fmt.Println("[ERR] Error in decryption for order response parser :", err)
				}

				if i == len(responseArr)-1 {
					return 0, 0, static.IMMEDIATE_EXECUTION, nil
				}
			case static.ORDER_BOOK_EXECUTION:
				_, err := utils.DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
				if err != nil {
					fmt.Println("[ERR] Error in decryption for order response parser :", err)
				}

				if i == len(responseArr)-1 {
					return 0, 0, static.ORDER_BOOK_EXECUTION, nil
				}
			case static.TRADING_SESSION_EVENT:
				_, err := utils.DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
				if err != nil {
					fmt.Println("[ERR] Error in decryption for order response parser :", err)
				}

				if i == len(responseArr)-1 {
					return 0, 0, static.TRADING_SESSION_EVENT, nil
				}

			case static.DELETE_ALL_ORDERS:
				_, err := utils.DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
				if err != nil {
					fmt.Println("[ERR] Error in decryption for order response parser :", err)
				}

				if i == len(responseArr)-1 {
					return 0, 0, static.DELETE_ALL_ORDERS, nil
				}

			case static.DELETE_ALL_ORDERS_NO_HITS:
				_, err := utils.DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
				if err != nil {
					fmt.Println("[ERR] Error in decryption for order response parser :", err)
				}

				if i == len(responseArr)-1 {
					return 0, 0, static.DELETE_ALL_ORDERS_NO_HITS, nil
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

func (t Traders) handleOrderEntryTrade() error {
	startTime := time.Now()
	totalTime := time.Now()
	packetCount := 0
	totalPacketCount := 0
	count := 0

	for {
		if static.TRADE_THROTTLE > 0 {
		traderChanLoop:
			for data := range static.DataChan {

				elapsedTime := time.Since(startTime)

				if packetCount >= static.TRADE_THROTTLE || elapsedTime >= time.Second {
					if packetCount >= static.TRADE_THROTTLE && elapsedTime < time.Second {
						fmt.Printf("Session id %d , sleeping for %v\n", t.Session_id, time.Second-elapsedTime)
						time.Sleep(time.Second - elapsedTime)
					}
					fmt.Printf("Session id %d sent %d packets.\n", t.Session_id, packetCount)
					startTime = time.Now()
					packetCount = 0
				}

				if static.TRADE_THROTTLE == 0 {
					fmt.Println("Exiting Trade Channel as trade throttle is set to zero and msg seq is", t.Msg_seq)
					break traderChanLoop
				}

				_, buy_max_price, sell_min_price, _ := utils.GetPriceRange(data.Price, data.MaxPrice, data.BidIntrvl)
				startPrice := buy_max_price + data.BidIntrvl
				price := startPrice + (count * data.BidIntrvl)

				if price >= sell_min_price {
					price = startPrice
					count = 0
				}

				if static.UNIQUE_CLIENT_IDENTIFIER {
					t.ClOrdID++
				}

				order_type := static.BUY
				client_code := static.CLIENT3
				if t.Session_id%2 != 0 {
					order_type = static.SELL
					client_code = static.CLIENT4
				}

				orderRequest, err := utils.SingleLegLeanOrder(t.MID, t.TID, t.Msg_seq, data.InstrumentId, price, data.Qty, order_type, t.CTX, t.ClOrdID, client_code)
				if err != nil {
					return err
				}

				if err := t.sendDataToServer(orderRequest); err != nil {
					fmt.Printf("Error sending data with session id %d: %v", t.Session_id, err)
				}

				_, _, _, err = t.NewOrderResponseListner(data.InstrumentId)
				if err != nil {
					return fmt.Errorf("response Error for session id %d: %v", t.Session_id, err)
				}

				t.Msg_seq++
				packetCount++
				totalPacketCount++
				count++

				if len(static.DataChan) == 0 && !static.SWITCH_CH {
					fmt.Printf("Session id %d took total time: %v and sent packets: %v\n", t.Session_id, time.Since(totalTime), totalPacketCount)
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

func (t Traders) NewHandleOrderEntryRecords() error {

	startTime := time.Now()
	totalTime := time.Now()
	packetCount := 0
	totalPacketCount := 0
	count := 0

	err := utils.SetOrderPumpingParameters()
	if err != nil {
		fmt.Println("Error in setting Order Pumping Parameters.")
	}

	for {
		if static.THROTTLE_VALUE > 0 {
		dataChanLoop:
			for data := range static.RecordsChan {

				throttleRate := static.THROTTLE_VALUE

				if err := t.DeleteOrdersByType(); err != nil {
					fmt.Printf("Delete all orders Error for session id %d: %v", t.Session_id, err)
				}

				if static.THROTTLE_VALUE == 0 {
					fmt.Println("1. Exiting Data Channel as throttle is set to zero and msg seq is", t.Msg_seq)
					t.OrderMap = make(map[uint64]static.OrderResponse)
					t.UiOrderMap = make(map[uint64]static.UiOrderResponse)
					break dataChanLoop
				}

				count++

				client_code := static.CLIENT3
				order_type := static.BUY
				if t.Session_id%2 == 0 {
					order_type = static.SELL
					client_code = static.CLIENT4
				}

				price := 0
				qty := 0

				maxInterval := time.Second
				if static.PATTERN_GENERATOR && static.PATTERN_TYPE == static.PEAK_GENERATOR {
					maxInterval = time.Duration(static.INTERVAL_TIME) * time.Millisecond
				}

				elapsedTime := time.Since(startTime)

				loopCount := 0
				dataLen := len(data)

				for loopCount < static.ORDER_ENTRY_COUNT {

					record := data[loopCount%dataLen]
					elapsedTime = time.Since(startTime)

					if packetCount >= static.THROTTLE_VALUE || elapsedTime >= maxInterval {

						if static.THROTTLE_VALUE == 0 {
							fmt.Println("2. Exiting Data Channel as throttle is set to zero and msg seq is", t.Msg_seq)
							t.OrderMap = make(map[uint64]static.OrderResponse)
							t.UiOrderMap = make(map[uint64]static.UiOrderResponse)
							break dataChanLoop
						} else if packetCount >= static.THROTTLE_VALUE && elapsedTime < maxInterval {

							if static.PATTERN_GENERATOR && static.PATTERN_TYPE == static.PEAK_GENERATOR {
								fmt.Printf("1. OE Session id %d , sleeping for %v (peak generator)\n", t.Session_id, time.Duration(static.DELAY_TIME)*time.Millisecond)
								time.Sleep(time.Duration(static.DELAY_TIME) * time.Millisecond)
							} else {
								fmt.Printf("2. OE Session id %d , sleeping for %v\n", t.Session_id, 1000*time.Millisecond-elapsedTime)
								time.Sleep(time.Second - elapsedTime)
							}

						}

						fmt.Printf("3. OE Session id %d sent %d packets.\n", t.Session_id, packetCount)

						startTime = time.Now()
						packetCount = 0
						static.SEND_MOD = true
						t.OrderMap = make(map[uint64]static.OrderResponse)
						t.UiOrderMap = make(map[uint64]static.UiOrderResponse)
					}

					if static.UNIQUE_CLIENT_IDENTIFIER {
						t.ClOrdID++
					}

					if throttleRate != static.THROTTLE_VALUE && static.PATTERN_TYPE != static.PEAK_GENERATOR {
						fmt.Println("1. Throttle value changed, execution will begin from start!")
						t.OrderMap = make(map[uint64]static.OrderResponse)
						t.UiOrderMap = make(map[uint64]static.UiOrderResponse)
						break
					}

					price_factor := record.BidIntrvl
					buy_min_price, buy_max_price, sell_min_price, sell_max_price := utils.GetPriceRange(record.Price, record.MaxPrice, price_factor)

					mid_qty := record.Qty * (1 + record.MaxOrderQty/(record.Qty*2))
					if mid_qty > record.MaxOrderQty {
						mid_qty = record.MaxOrderQty
					}

					order_details := static.OrderDetails{Price_factor: price_factor, Buy_min_price: buy_min_price, Buy_max_price: buy_max_price, Sell_min_price: sell_min_price, Sell_max_price: sell_max_price, Min_qty: record.Qty, Mid_qty: mid_qty, Max_qty: record.MaxOrderQty}

					if order_type == static.BUY {
						price = utils.GenerateRandomPrice(buy_min_price, buy_max_price, price_factor)
						qty = utils.GenerateRandomQty(record.Qty, mid_qty)
					} else if order_type == static.SELL {
						price = utils.GenerateRandomPrice(sell_min_price, sell_max_price, price_factor)
						qty = utils.GenerateRandomQty(mid_qty, record.MaxOrderQty)
					}

					orderRequest, err := utils.SingleLegLeanOrder(t.MID, t.TID, t.Msg_seq, record.InstrumentId, price, qty, order_type, t.CTX, t.ClOrdID, client_code)
					if err != nil {
						return err
					}

					orderEntryRTT := time.Now()

					if err := t.sendDataToServer(orderRequest); err != nil {
						fmt.Printf("Error sending record with session id %d: %v", t.Session_id, err)
					}

					OrderID, ActTime, template, err := t.NewOrderResponseListner(record.InstrumentId)
					if err != nil {
						return fmt.Errorf("response Error for session id %d: %v", t.Session_id, err)
					}
					fmt.Println("[INFO] Order Entry Request RTT: ", time.Since(orderEntryRTT))

					if template == static.LEAN_ORDER {
						if static.UNIQUE_CLIENT_IDENTIFIER {
							t.UiOrderMap[t.ClOrdID] = static.UiOrderResponse{OrigClOrdID: t.ClOrdID, ActTime: ActTime, InstrumentID: record.InstrumentId, ProductID: record.ProductId, OrderDetails: order_details}

						} else {
							t.OrderMap[OrderID] = static.OrderResponse{OrderID: OrderID, ActTime: ActTime, InstrumentID: record.InstrumentId, ProductID: record.ProductId, OrderDetails: order_details}
						}
					}
					//  else if template == static.REJECTED {
					// 	ORDERS_REJECTION[t.Session_id] = 1
					// }

					t.Msg_seq++
					packetCount++
					totalPacketCount++
					loopCount++

				}

				map_length := len(t.OrderMap)
				if static.UNIQUE_CLIENT_IDENTIFIER {
					map_length = len(t.UiOrderMap)
				}

				if static.SEND_MOD && map_length > 0 {
				outerLoopMod:
					for n := 0; n < static.MOD_PER_ORDER; n++ {
						if static.UNIQUE_CLIENT_IDENTIFIER {
							for _, order_map := range t.UiOrderMap {
								t.ClOrdID++

								if throttleRate != static.THROTTLE_VALUE && static.PATTERN_TYPE != static.PEAK_GENERATOR {
									fmt.Println("2. Throttle value changed, execution will begin from start!")
									t.UiOrderMap = make(map[uint64]static.UiOrderResponse)
									break outerLoopMod
								}

								if order_map.OrderDetails.Buy_min_price == 0 || order_map.OrderDetails.Buy_max_price == 0 || order_map.OrderDetails.Price_factor == 0 ||
									order_map.OrderDetails.Sell_min_price == 0 || order_map.OrderDetails.Sell_max_price == 0 || order_map.OrderDetails.Min_qty == 0 || order_map.OrderDetails.Max_qty == 0 {
									delete(t.UiOrderMap, order_map.OrigClOrdID)
									continue
								}

								if order_type == static.BUY {
									price = utils.GenerateRandomPrice(order_map.OrderDetails.Buy_min_price, order_map.OrderDetails.Buy_max_price, order_map.OrderDetails.Price_factor)
									qty = utils.GenerateRandomQty(order_map.OrderDetails.Min_qty, order_map.OrderDetails.Mid_qty)
								} else if order_type == static.SELL {
									price = utils.GenerateRandomPrice(order_map.OrderDetails.Sell_min_price, order_map.OrderDetails.Sell_max_price, order_map.OrderDetails.Price_factor)
									qty = utils.GenerateRandomQty(order_map.OrderDetails.Mid_qty, order_map.OrderDetails.Max_qty)
								}

								_, ActTime, template, err := t.sendOrderModify(order_map.InstrumentID, price, qty, order_type, order_map.OrigClOrdID, order_map.ActTime)
								if err != nil {
									fmt.Printf("Modify Error for session id %d: %v", t.Session_id, err)
								}

								if template == static.MODIFY_LEAN_ORDER {
									oMap := t.UiOrderMap[order_map.OrigClOrdID]
									delete(t.UiOrderMap, order_map.OrigClOrdID)
									oMap.OrigClOrdID = t.ClOrdID
									oMap.ActTime = ActTime
									t.UiOrderMap[oMap.OrigClOrdID] = oMap

									// delete(ORDERS_REJECTION, t.Session_id)

								} else {
									delete(t.UiOrderMap, order_map.OrigClOrdID)

									// if template == static.REJECTED {
									// 	ORDERS_REJECTION[t.Session_id] = 1
									// }
								}

								packetCount++
								t.Msg_seq++
								totalPacketCount++

								elapsedTime = time.Since(startTime)

								if packetCount >= static.THROTTLE_VALUE || elapsedTime >= maxInterval {
									if static.THROTTLE_VALUE == 0 {
										fmt.Println("3. Exiting Data Channel as throttle is set to zero and msg seq is", t.Msg_seq)
										t.OrderMap = make(map[uint64]static.OrderResponse)
										t.UiOrderMap = make(map[uint64]static.UiOrderResponse)
										break dataChanLoop
									} else if elapsedTime >= maxInterval {
										oldPacketCount := packetCount

										if err := t.handleCancelOrders(&packetCount, &totalPacketCount); err != nil {
											fmt.Printf("Cancel Error for session id %d: %v", t.Session_id, err)
										}

										packetCount = oldPacketCount

									} else if packetCount >= static.THROTTLE_VALUE && elapsedTime < maxInterval {

										if static.PATTERN_GENERATOR && static.PATTERN_TYPE == static.PEAK_GENERATOR {
											fmt.Printf("OM Session id %d , sleeping for %v (peak generator)\n", t.Session_id, time.Duration(static.DELAY_TIME)*time.Millisecond)
											time.Sleep(time.Duration(static.DELAY_TIME) * time.Millisecond)
										} else {
											fmt.Printf("OM Session id %d , sleeping for %v\n", t.Session_id, 1000*time.Millisecond-elapsedTime)
											time.Sleep(time.Second - elapsedTime)
										}

									}

									fmt.Printf("OM Session id %d sent %d packets.\n", t.Session_id, packetCount)
									t.UiOrderMap = make(map[uint64]static.UiOrderResponse)
									startTime = time.Now()
									packetCount = 0

									fmt.Println("Exiting Mod for loop!", len(t.UiOrderMap))
									break outerLoopMod
								}
							}

						} else {
							for _, order_map := range t.OrderMap {

								if throttleRate != static.THROTTLE_VALUE && static.PATTERN_TYPE != static.PEAK_GENERATOR {
									fmt.Println("2. Throttle value changed, execution will begin from start!")
									t.OrderMap = make(map[uint64]static.OrderResponse)
									break outerLoopMod
								}

								if order_map.OrderDetails.Buy_min_price == 0 || order_map.OrderDetails.Buy_max_price == 0 || order_map.OrderDetails.Price_factor == 0 ||
									order_map.OrderDetails.Sell_min_price == 0 || order_map.OrderDetails.Sell_max_price == 0 || order_map.OrderDetails.Min_qty == 0 || order_map.OrderDetails.Max_qty == 0 {
									delete(t.OrderMap, order_map.OrderID)
									continue
								}

								if order_type == static.BUY {
									price = utils.GenerateRandomPrice(order_map.OrderDetails.Buy_min_price, order_map.OrderDetails.Buy_max_price, order_map.OrderDetails.Price_factor)
									qty = utils.GenerateRandomQty(order_map.OrderDetails.Min_qty, order_map.OrderDetails.Mid_qty)
								} else if order_type == static.SELL {
									price = utils.GenerateRandomPrice(order_map.OrderDetails.Sell_min_price, order_map.OrderDetails.Sell_max_price, order_map.OrderDetails.Price_factor)
									qty = utils.GenerateRandomQty(order_map.OrderDetails.Mid_qty, order_map.OrderDetails.Max_qty)
								}

								OrderID, ActTime, template, err := t.sendOrderModify(order_map.InstrumentID, price, qty, order_type, order_map.OrderID, order_map.ActTime)
								if err != nil {
									fmt.Printf("Modify Error for session id %d: %v", t.Session_id, err)
								}

								if template == static.MODIFY_LEAN_ORDER {
									oMap := t.OrderMap[OrderID]
									oMap.OrderID = OrderID
									oMap.ActTime = ActTime
									t.OrderMap[OrderID] = oMap

									// delete(ORDERS_REJECTION, t.Session_id)
								} else {
									delete(t.OrderMap, order_map.OrderID)

									// if template == static.REJECTED {
									// 	ORDERS_REJECTION[t.Session_id] = 1
									// }
								}

								packetCount++
								t.Msg_seq++
								totalPacketCount++

								elapsedTime = time.Since(startTime)

								if packetCount >= static.THROTTLE_VALUE || elapsedTime >= maxInterval {
									if static.THROTTLE_VALUE == 0 {
										fmt.Println("3. Exiting Data Channel as throttle is set to zero and msg seq is", t.Msg_seq)
										t.OrderMap = make(map[uint64]static.OrderResponse)
										t.UiOrderMap = make(map[uint64]static.UiOrderResponse)
										break dataChanLoop
									} else if elapsedTime >= maxInterval {
										oldPacketCount := packetCount

										if err := t.handleCancelOrders(&packetCount, &totalPacketCount); err != nil {
											fmt.Printf("Cancel Error for session id %d: %v", t.Session_id, err)
										}

										packetCount = oldPacketCount

									} else if packetCount >= static.THROTTLE_VALUE && elapsedTime < maxInterval {

										if static.PATTERN_GENERATOR && static.PATTERN_TYPE == static.PEAK_GENERATOR {
											fmt.Printf("OM Session id %d , sleeping for %v (peak generator)\n", t.Session_id, time.Duration(static.DELAY_TIME)*time.Millisecond)
											time.Sleep(time.Duration(static.DELAY_TIME) * time.Millisecond)
										} else {
											fmt.Printf("OM Session id %d , sleeping for %v\n", t.Session_id, 1000*time.Millisecond-elapsedTime)
											time.Sleep(time.Second - elapsedTime)
										}
									}

									fmt.Printf("OM Session id %d sent %d packets.\n", t.Session_id, packetCount)

									startTime = time.Now()
									packetCount = 0
									t.OrderMap = make(map[uint64]static.OrderResponse)

									fmt.Println("Exiting Mod for loop!", len(t.OrderMap))
									break outerLoopMod
								}
							}
						}
					}
				}

				if err := t.handleCancelOrders(&packetCount, &totalPacketCount); err != nil {
					fmt.Printf("Cancel Error for session id %d: %v", t.Session_id, err)
				}

				t.OrderMap = make(map[uint64]static.OrderResponse)
				t.UiOrderMap = make(map[uint64]static.UiOrderResponse)

				if len(static.RecordsChan) == 0 && !static.SWITCH_CH {

					static.PATTERN_GENERATOR = false

					fmt.Printf("Session id %d took total time: %v and sent packets: %v\n", t.Session_id, time.Since(totalTime), totalPacketCount)
					return nil
				}
			}
		} else {
			static.PATTERN_GENERATOR = false

			fmt.Println("Throttle is 0 and Heartbeat started!")
			if err := t.heartBeatListener(); err != nil {
				return fmt.Errorf("HeartBeat err: %v ", err)
			}
		}
	}
}

func (t Traders) handleCancelOrders(packetCount, totalPacketCount *int) error {
	map_length := len(t.OrderMap)
	if static.UNIQUE_CLIENT_IDENTIFIER {
		map_length = len(t.UiOrderMap)
	}
	if static.CANCEL_PERCENT != 0 && map_length > 0 {
		cancel_count := 0
		if static.UNIQUE_CLIENT_IDENTIFIER {
			for _, order_map := range t.UiOrderMap {
				if cancel_count == static.CANCEL_ORDER_COUNT {
					t.UiOrderMap = make(map[uint64]static.UiOrderResponse)
					fmt.Println("Exiting cancel loop and cancel order count:", cancel_count)
					break
				}
				t.ClOrdID++
				if err := t.sendOrderCancel(order_map.InstrumentID, order_map.ProductID, order_map.OrigClOrdID); err != nil {
					fmt.Printf("[ERR] Cancel Error for session id %d: %v", t.Session_id, err)
					// return err
				}
				delete(t.UiOrderMap, order_map.OrigClOrdID)
				t.Msg_seq++
				*packetCount++
				*totalPacketCount++
				cancel_count++
			}
		} else {
			for _, order_map := range t.OrderMap {
				if cancel_count == static.CANCEL_ORDER_COUNT {
					t.OrderMap = make(map[uint64]static.OrderResponse)
					fmt.Println("Exiting cancel loop and cancel order count:", cancel_count)
					break
				}
				if err := t.sendOrderCancel(order_map.InstrumentID, order_map.ProductID, order_map.OrderID); err != nil {
					fmt.Printf("[ERR] Cancel Error for session id %d: %v", t.Session_id, err)
					// return err
				}
				delete(t.OrderMap, order_map.OrderID)
				t.Msg_seq++
				*packetCount++
				*totalPacketCount++
				cancel_count++
			}
		}

	}
	return nil
}

//================================================================== ASYNC IMLEMENTATION ======================================================================

func (t Traders) NewHandleOrderEntryRecordsAsync() error {

	startTime := time.Now()
	totalTime := time.Now()
	packetCount := 0
	totalPacketCount := 0
	count := 0

	err := utils.SetOrderPumpingParameters()
	if err != nil {
		fmt.Println("Error in setting Order Pumping Parameters.")
	}

	for {
		if static.THROTTLE_VALUE > 0 {
		dataChanLoop:
			for data := range static.RecordsChan {

				throttleRate := static.THROTTLE_VALUE

				if err := t.DeleteOrdersByType(); err != nil {
					fmt.Printf("Delete all orders Error for session id %d: %v", t.Session_id, err)
				}

				if static.THROTTLE_VALUE == 0 {
					fmt.Println("1. Exiting Data Channel as throttle is set to zero and msg seq is", t.Msg_seq)
					t.UiOrderMap = make(map[uint64]static.UiOrderResponse)
					break dataChanLoop
				}

				count++

				client_code := static.CLIENT3
				order_type := static.BUY
				if t.Session_id%2 == 0 {
					order_type = static.SELL
					client_code = static.CLIENT4
				}

				price := 0
				qty := 0

				maxInterval := time.Second
				if static.PATTERN_GENERATOR && static.PATTERN_TYPE == static.PEAK_GENERATOR {
					maxInterval = time.Duration(static.INTERVAL_TIME) * time.Millisecond
				}

				elapsedTime := time.Since(startTime)

				loopCount := 0
				dataLen := len(data)

				for loopCount < static.ORDER_ENTRY_COUNT {

					record := data[loopCount%dataLen]
					elapsedTime = time.Since(startTime)
					if packetCount >= static.THROTTLE_VALUE || elapsedTime >= maxInterval {

						if static.THROTTLE_VALUE == 0 {
							fmt.Println("2. Exiting Data Channel as throttle is set to zero and msg seq is", t.Msg_seq)
							t.UiOrderMap = make(map[uint64]static.UiOrderResponse)
							break dataChanLoop
						} else if packetCount >= static.THROTTLE_VALUE && elapsedTime < maxInterval {
							if static.PATTERN_GENERATOR && static.PATTERN_TYPE == static.PEAK_GENERATOR {
								fmt.Printf("Session id %d , sleeping for %v\n", t.Session_id, time.Duration(static.DELAY_TIME)*time.Millisecond)
								time.Sleep(time.Duration(static.DELAY_TIME) * time.Millisecond)
							} else {
								fmt.Printf("Session id %d , sleeping for %v\n", t.Session_id, 1000*time.Millisecond-elapsedTime)
								time.Sleep(time.Second - elapsedTime)
							}
						}

						fmt.Printf("Session id %d sent %d packets.\n", t.Session_id, packetCount)

						startTime = time.Now()
						packetCount = 0
						static.SEND_MOD = true
						t.UiOrderMap = make(map[uint64]static.UiOrderResponse)
					}

					t.ClOrdID++

					if throttleRate != static.THROTTLE_VALUE && static.PATTERN_TYPE != static.PEAK_GENERATOR {
						fmt.Println("1. Throttle value changed, execution will begin from start!")
						t.UiOrderMap = make(map[uint64]static.UiOrderResponse)
						break
					}

					price_factor := record.BidIntrvl
					buy_min_price, buy_max_price, sell_min_price, sell_max_price := utils.GetPriceRange(record.Price, record.MaxPrice, price_factor)

					mid_qty := record.Qty * (1 + record.MaxOrderQty/(record.Qty*2))
					if mid_qty > record.MaxOrderQty {
						mid_qty = record.MaxOrderQty
					}

					order_details := static.OrderDetails{Price_factor: price_factor, Buy_min_price: buy_min_price, Buy_max_price: buy_max_price, Sell_min_price: sell_min_price, Sell_max_price: sell_max_price, Min_qty: record.Qty, Mid_qty: mid_qty, Max_qty: record.MaxOrderQty}

					if order_type == static.BUY {
						price = utils.GenerateRandomPrice(buy_min_price, buy_max_price, price_factor)
						qty = utils.GenerateRandomQty(record.Qty, mid_qty)
					} else if order_type == static.SELL {
						price = utils.GenerateRandomPrice(sell_min_price, sell_max_price, price_factor)
						qty = utils.GenerateRandomQty(mid_qty, record.MaxOrderQty)
					}

					orderRequest, err := utils.SingleLegLeanOrder(t.MID, t.TID, t.Msg_seq, record.InstrumentId, price, qty, order_type, t.CTX, t.ClOrdID, client_code)
					if err != nil {
						return err
					}

					if err := t.sendDataToServer(orderRequest); err != nil {
						fmt.Printf("Error sending record with session id %d: %v", t.Session_id, err)
					}

					if static.UNIQUE_CLIENT_IDENTIFIER {

						t.UiOrderMap[t.ClOrdID] = static.UiOrderResponse{OrigClOrdID: t.ClOrdID, ActTime: uint64(static.DUMMY_NO), InstrumentID: record.InstrumentId, ProductID: record.ProductId, OrderDetails: order_details}

					}

					t.Msg_seq++
					packetCount++
					totalPacketCount++
					loopCount++
				}

				map_length := len(t.OrderMap)
				if static.UNIQUE_CLIENT_IDENTIFIER {
					map_length = len(t.UiOrderMap)
				}

				if static.SEND_MOD && map_length > 0 {

				outerLoopMod:
					for n := 0; n < static.MOD_PER_ORDER; n++ {
						newMap := make(map[uint64]static.UiOrderResponse)
						if static.UNIQUE_CLIENT_IDENTIFIER {
							count_order := 0
							for _, order_map := range t.UiOrderMap {
								t.ClOrdID++
								count_order++

								if throttleRate != static.THROTTLE_VALUE && static.PATTERN_TYPE != static.PEAK_GENERATOR {
									fmt.Println("2. Throttle value changed, execution will begin from start!")
									t.UiOrderMap = make(map[uint64]static.UiOrderResponse)

									break outerLoopMod
								}

								if order_map.OrderDetails.Buy_min_price == 0 || order_map.OrderDetails.Buy_max_price == 0 || order_map.OrderDetails.Price_factor == 0 ||
									order_map.OrderDetails.Sell_min_price == 0 || order_map.OrderDetails.Sell_max_price == 0 || order_map.OrderDetails.Min_qty == 0 || order_map.OrderDetails.Max_qty == 0 {
									delete(t.UiOrderMap, order_map.OrigClOrdID)
									continue
								}

								if order_type == static.BUY {
									price = utils.GenerateRandomPrice(order_map.OrderDetails.Buy_min_price, order_map.OrderDetails.Buy_max_price, order_map.OrderDetails.Price_factor)
									qty = utils.GenerateRandomQty(order_map.OrderDetails.Min_qty, order_map.OrderDetails.Mid_qty)
								} else if order_type == static.SELL {
									price = utils.GenerateRandomPrice(order_map.OrderDetails.Sell_min_price, order_map.OrderDetails.Sell_max_price, order_map.OrderDetails.Price_factor)
									qty = utils.GenerateRandomQty(order_map.OrderDetails.Mid_qty, order_map.OrderDetails.Max_qty)
								}

								t.sendOrderModifyAsync(order_map.InstrumentID, price, qty, order_type, order_map.OrigClOrdID, order_map.ActTime)
								if err != nil {
									fmt.Printf("Modify Error for session id %d: %v", t.Session_id, err)
								}

								oMap := t.UiOrderMap[order_map.OrigClOrdID]

								oMap.OrigClOrdID = t.ClOrdID
								oMap.ActTime = uint64(static.DUMMY_NO)
								newMap[oMap.OrigClOrdID] = oMap
								if count_order == len(t.UiOrderMap) {
									t.UiOrderMap = newMap
								}

								packetCount++
								t.Msg_seq++
								totalPacketCount++

								elapsedTime = time.Since(startTime)

								if packetCount >= static.THROTTLE_VALUE || elapsedTime >= maxInterval {
									fmt.Println("Satisfied case:  Packet count greater =>", packetCount >= static.THROTTLE_VALUE, "Elapsed time greater =>", elapsedTime >= maxInterval)
									if static.THROTTLE_VALUE == 0 {
										fmt.Println("3. Exiting Data Channel as throttle is set to zero and msg seq is", t.Msg_seq)
										t.UiOrderMap = make(map[uint64]static.UiOrderResponse)

										break dataChanLoop
									} else if elapsedTime >= maxInterval {
										oldPacketCount := packetCount
										fmt.Println("1. LENGTH OF UIMAP:", len(t.UiOrderMap))
										if err := t.handleCancelOrdersAsync(&packetCount, &totalPacketCount); err != nil {
											fmt.Printf("Cancel Error for session id %d: %v", t.Session_id, err)
										}

										packetCount = oldPacketCount
									} else if packetCount >= static.THROTTLE_VALUE && elapsedTime < maxInterval {
										if static.PATTERN_GENERATOR && static.PATTERN_TYPE == static.PEAK_GENERATOR {
											fmt.Printf("Session id %d , sleeping for %v\n", t.Session_id, time.Duration(static.DELAY_TIME)*time.Millisecond)
											time.Sleep(time.Duration(static.DELAY_TIME) * time.Millisecond)
										} else {
											fmt.Printf("Session id %d , sleeping for %v\n", t.Session_id, 1000*time.Millisecond-elapsedTime)
											time.Sleep(time.Second - elapsedTime)
										}
									}

									fmt.Printf("Session id %d sent %d packets.\n", t.Session_id, packetCount)

									startTime = time.Now()
									packetCount = 0
									t.UiOrderMap = make(map[uint64]static.UiOrderResponse)

									fmt.Println("Exiting Mod for loop!", len(t.UiOrderMap))
									break outerLoopMod
								}
							}

						}
					}
				}

				if err := t.handleCancelOrdersAsync(&packetCount, &totalPacketCount); err != nil {
					fmt.Printf("Cancel Error for session id %d: %v", t.Session_id, err)
				}

				t.UiOrderMap = make(map[uint64]static.UiOrderResponse)

				if len(static.RecordsChan) == 0 && !static.SWITCH_CH {

					static.PATTERN_GENERATOR = false

					fmt.Printf("Session id %d took total time: %v and sent packets: %v\n", t.Session_id, time.Since(totalTime), totalPacketCount)
					return nil
				}
			}
		} else {
			static.PATTERN_GENERATOR = false

			fmt.Println("Throttle is 0 and Heartbeat started!")
			if err := t.heartBeatListener(); err != nil {
				return fmt.Errorf("HeartBeat err: %v ", err)
			}
		}
	}
}

func (t Traders) sendOrderModifyAsync(inst_id, price, qty, order_type int, order_id, act_time uint64) {

	modified_data, err := utils.ModifyLeanOrder(t.MID, t.TID, t.Msg_seq, inst_id, price, qty, order_type, order_id, act_time, t.CTX, t.ClOrdID, order_id)
	if err != nil {
		fmt.Println("[ERR] Error in order modify parser :", err)
		return
	}

	err = t.sendDataToServer(modified_data)
	if err != nil {
		fmt.Printf("[ERR] Error sending data for session id %d: %v", t.Session_id, err)
		return
	}
}

func (t Traders) sendOrderCancelAsync(inst_id, product_id int, order_id uint64) {

	cancelOrder, err := utils.CancelSingleLegOrderRequest(t.MID, t.TID, t.Msg_seq, inst_id, product_id, order_id, t.CTX, t.ClOrdID, order_id)
	if err != nil {
		fmt.Println("[ERR] Error in cancel order", err)
	}

	err = t.sendDataToServer(cancelOrder)
	if err != nil {
		fmt.Printf("[ERR] Error sending cancel order for session id %d: %v", t.Session_id, err)
	}
}

func (t Traders) NewOrderResponseListnerAsync(quit chan struct{}, inst_id int) {

	for {
		buffer := make([]byte, 100*1024)
		reader := bufio.NewReader(t.Conn)

		n, err := reader.Read(buffer)
		if err != nil {
			if err == io.EOF {
				fmt.Printf("[INFO] Connection closed for session id %d: %v\n", t.Session_id, err)
				return
			}
			fmt.Printf("[ERR] Error reading responses for session id %d: %v", t.Session_id, err)
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
				// return
			}

			size := int(binary.LittleEndian.Uint32(responseBuf[:4]))
			if len(responseBuf) < size {
				fmt.Println(" Length Response Buf", len(responseBuf), size)
				fmt.Println("[ERR] Response buffer too short for the indicated size")
				// return
				break
			}

			responseArr = append(responseArr, responseBuf[:size])
			responseBuf = responseBuf[size:]
		}

		for i, response := range responseArr {

			if len(response) == 0 {
				break

			}

			val, err := utils.GetTemplateID(response)
			if err != nil {
				fmt.Println("[ERR] Error in template id parser :", err)
			}

			read_response, err := t.readReceivedResponse(val, inst_id, response)
			if err != nil {
				fmt.Println("[ERR] Error in reading received response :", err)
			}

			if read_response == static.ORDER_CONFIRMATION {
				fmt.Println("[PASS] Order Confirmation Received for session id:", t.Session_id)
				continue
			}

			switch read_response {
			case static.LEAN_ORDER:
				order_res, err := utils.DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
				if err != nil {
					fmt.Println("[ERR] Error in decryption for order response parser :", err)
				}

				server_res, err1 := utils.GetTimestamps(order_res)
				if err1 != nil {
					fmt.Println("[ERR] Error while getting timestamps")
				} else {
					fmt.Printf("[INFO] Order Entry Server Processing Time(T6-T1):  %v microseconds\n", (server_res.GW_Res_Out-server_res.GW_Req_In)/1000)
				}

				_, _, err = utils.SingleLegLeanOrderResponse(order_res, t.CTX)
				if err != nil {
					fmt.Println("[ERR] Error in lean order response parser :", err)
				}

			case static.CANCEL_LEAN_ORDER:
				order_res, err := utils.DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
				if err != nil {
					fmt.Println("[ERR] Error in decryption for order response parser :", err)
				}

				server_res, err1 := utils.GetTimestamps(order_res)
				if err1 != nil {
					fmt.Println("[ERR] Error while getting timestamps")
				} else {
					fmt.Printf("[INFO] Order Cancel Server Processing Time(T6-T1):  %v microseconds\n", (server_res.GW_Res_Out-server_res.GW_Req_In)/1000)

				}

			case static.MODIFY_LEAN_ORDER:
				order_res, err := utils.DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
				if err != nil {
					fmt.Println("[ERR] Error in decryption for order response parser :", err)
				}

				server_res, err1 := utils.GetTimestamps(order_res)

				if err1 != nil {
					fmt.Println("[ERR] Error while getting timestamps")
				} else {
					fmt.Printf("[INFO] Order Mod Server Processing Time(T6-T1):  %v microseconds\n", (server_res.GW_Res_Out-server_res.GW_Req_In)/1000)
				}

				_, _, err = utils.SingleLegLeanOrderModifiedResponse(order_res, t.CTX)

				if err != nil {
					fmt.Println("[ERR] Error in lean order response parser :", err)
				}

			case static.HEARTBEAT:
				_, err := utils.DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
				if err != nil {
					fmt.Println("[ERR] Error in decryption for order response parser :", err)
				}

			case static.REJECTED:
				fmt.Println("Rejected case!!")
			case static.IMMEDIATE_EXECUTION:
				_, err := utils.DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
				if err != nil {
					fmt.Println("[ERR] Error in decryption for order response parser :", err)
				}

			case static.ORDER_BOOK_EXECUTION:
				_, err := utils.DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
				if err != nil {
					fmt.Println("[ERR] Error in decryption for order response parser :", err)
				}

			case static.TRADING_SESSION_EVENT:
				_, err := utils.DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
				if err != nil {
					fmt.Println("[ERR] Error in decryption for order response parser :", err)
				}

			case static.DELETE_ALL_ORDERS:
				_, err := utils.DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
				if err != nil {
					fmt.Println("[ERR] Error in decryption for order response parser :", err)
				}

			case static.DELETE_ALL_ORDERS_NO_HITS:
				_, err := utils.DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
				if err != nil {
					fmt.Println("[ERR] Error in decryption for order response parser :", err)
				}

			default:
				fmt.Println("Other Template ID: ", read_response)
			}

			if i == len(responseArr)-1 {
				// if read_response == static.REJECTED {
				// 	ORDERS_REJECTION[t.Session_id] = 1
				// } else {
				// 	delete(ORDERS_REJECTION, t.Session_id)
				// }

				break
			}

		}
	}
}

func (t Traders) handleCancelOrdersAsync(packetCount, totalPacketCount *int) error {
	if static.CANCEL_PERCENT != 0 && len(t.UiOrderMap) > 0 {
		cancel_count := 0

		for _, order_map := range t.UiOrderMap {
			if cancel_count == static.CANCEL_ORDER_COUNT {
				t.UiOrderMap = make(map[uint64]static.UiOrderResponse)

				fmt.Println("Exiting cancel loop and cancel order count:", cancel_count)
				break
			}
			t.ClOrdID++
			t.sendOrderCancelAsync(order_map.InstrumentID, order_map.ProductID, order_map.OrigClOrdID)
			delete(t.UiOrderMap, order_map.OrigClOrdID)
			t.Msg_seq++
			*packetCount++
			*totalPacketCount++
			cancel_count++
		}

	}
	return nil
}

func (t Traders) AsyncHandleOrderEntryRecordsOE() error {

	startTime := time.Now()
	totalTime := time.Now()
	packetCount := 0
	totalPacketCount := 0
	count := 0

	err := utils.SetOrderPumpingParameters()
	if err != nil {
		fmt.Println("Error in setting Order Pumping Parameters.")
	}

	for {
		if static.THROTTLE_VALUE > 0 {
		dataChanLoop:
			for data := range static.RecordsChan {

				throttleRate := static.THROTTLE_VALUE

				if err := t.DeleteOrdersByType(); err != nil {
					fmt.Printf("Delete all orders Error for session id %d: %v", t.Session_id, err)
				}

				if static.THROTTLE_VALUE == 0 {
					fmt.Println("1. Exiting Data Channel as throttle is set to zero and msg seq is", t.Msg_seq)
					t.OrderMap = make(map[uint64]static.OrderResponse)
					t.UiOrderMap = make(map[uint64]static.UiOrderResponse)
					break dataChanLoop
				}

				count++

				client_code := static.CLIENT3
				order_type := static.BUY
				if t.Session_id%2 == 0 {
					order_type = static.SELL
					client_code = static.CLIENT4
				}

				price := 0
				qty := 0

				maxInterval := time.Second
				if static.PATTERN_GENERATOR && static.PATTERN_TYPE == static.PEAK_GENERATOR {
					maxInterval = time.Duration(static.INTERVAL_TIME) * time.Millisecond
				}

				elapsedTime := time.Since(startTime)

				loopCount := 0
				dataLen := len(data)

				for loopCount < static.ORDER_ENTRY_COUNT {

					record := data[loopCount%dataLen]

					elapsedTime = time.Since(startTime)

					if packetCount >= static.THROTTLE_VALUE || elapsedTime >= maxInterval {
						if static.THROTTLE_VALUE == 0 {
							fmt.Println("2. Exiting Data Channel as throttle is set to zero and msg seq is", t.Msg_seq)
							t.OrderMap = make(map[uint64]static.OrderResponse)
							t.UiOrderMap = make(map[uint64]static.UiOrderResponse)
							break dataChanLoop
						} else if packetCount >= static.THROTTLE_VALUE && elapsedTime < maxInterval {

							if static.PATTERN_GENERATOR && static.PATTERN_TYPE == static.PEAK_GENERATOR {
								fmt.Printf("Session id %d , sleeping for %v\n", t.Session_id, time.Duration(static.DELAY_TIME)*time.Millisecond)
								time.Sleep(time.Duration(static.DELAY_TIME) * time.Millisecond)
							} else {
								fmt.Printf("Session id %d , sleeping for %v\n", t.Session_id, 1000*time.Millisecond-elapsedTime)
								time.Sleep(time.Second - elapsedTime)
							}
						}
						fmt.Printf("Session id %d sent %d packets.\n", t.Session_id, packetCount)
						startTime = time.Now()
						packetCount = 0
						static.SEND_MOD = true
						t.OrderMap = make(map[uint64]static.OrderResponse)
						t.UiOrderMap = make(map[uint64]static.UiOrderResponse)
					}

					if static.UNIQUE_CLIENT_IDENTIFIER {
						t.ClOrdID++
					}

					if throttleRate != static.THROTTLE_VALUE && static.PATTERN_TYPE != static.PEAK_GENERATOR {
						fmt.Println("1. Throttle value changed, execution will begin from start!")
						t.OrderMap = make(map[uint64]static.OrderResponse)
						t.UiOrderMap = make(map[uint64]static.UiOrderResponse)
						break
					}

					price_factor := record.BidIntrvl
					buy_min_price, buy_max_price, sell_min_price, sell_max_price := utils.GetPriceRange(record.Price, record.MaxPrice, price_factor)
					mid_qty := record.Qty * (1 + record.MaxOrderQty/(record.Qty*2))

					if mid_qty > record.MaxOrderQty {
						mid_qty = record.MaxOrderQty
					}

					if order_type == static.BUY {
						price = utils.GenerateRandomPrice(buy_min_price, buy_max_price, price_factor)
						qty = utils.GenerateRandomQty(record.Qty, mid_qty)
					} else if order_type == static.SELL {
						price = utils.GenerateRandomPrice(sell_min_price, sell_max_price, price_factor)
						qty = utils.GenerateRandomQty(mid_qty, record.MaxOrderQty)
					}

					orderRequest, err := utils.SingleLegLeanOrder(t.MID, t.TID, t.Msg_seq, record.InstrumentId, price, qty, order_type, t.CTX, t.ClOrdID, client_code)
					if err != nil {
						return err
					}

					if err := t.sendDataToServer(orderRequest); err != nil {
						fmt.Printf("Error sending record with session id %d: %v", t.Session_id, err)
					}

					t.Msg_seq++
					packetCount++
					totalPacketCount++
					loopCount++
				}

				t.OrderMap = make(map[uint64]static.OrderResponse)
				t.UiOrderMap = make(map[uint64]static.UiOrderResponse)

				if len(static.RecordsChan) == 0 && !static.SWITCH_CH {
					static.PATTERN_GENERATOR = false
					fmt.Printf("Session id %d took total time: %v and sent packets: %v\n", t.Session_id, time.Since(totalTime), totalPacketCount)
					return nil
				}
			}
		} else {
			static.PATTERN_GENERATOR = false
			fmt.Println("Throttle is 0 and Heartbeat started!")
			if err := t.heartBeatListener(); err != nil {
				return fmt.Errorf("HeartBeat err: %v ", err)
			}
		}
	}
}
