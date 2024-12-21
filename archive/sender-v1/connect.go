// Sync Response
package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"market_replay/aes"
	"math"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	dataChan       = make(chan redisData, DATA_CHANNEL_LENGTH)
	connCheckChan  chan int
	totalTCPConn   int
	members        []Member
	recordFile     string
	cache          = make(map[string]CacheEntry)
	cacheMu        sync.Mutex
	traderMap      = make(map[int]*Trader)
	orderPump_flag = true
)

const (
	retryConn = 5
)

func Start() {

	err := ReadConfig(CONFIG_FILE_PATH)
	if err != nil {
		fmt.Println("[ERR] Error reading config file.")
	}

	switch TARGET_ENV {
	case PRODUCTION_ENV:
		members = PROD_ENV_IDS_WIN
		recordFile = PRODUCTION_RECORDS1_FILE
		if FILE_TYPE == 0 {
			recordFile = PRODUCTION_RECORDS_FILE
		}
		fmt.Println("Environment : PRODUCTION")
	case LAB_ENV:
		members = LAB_ENV_IDS_WIN
		recordFile = LAB_RECORDS1_FILE
		if FILE_TYPE == 0 {
			recordFile = LAB_RECORDS_FILE
		}
		fmt.Println("Environment : LAB")
	case SIMULATION_ENV:
		members = SIMULATION_ENV_ID
		recordFile = SIMULATION_RECORDS1_FILE
		if FILE_TYPE == 0 {
			recordFile = SIMULATION_RECORDS_FILE
		}
		fmt.Println("Environment : SIMULATION")
	}

	var wg sync.WaitGroup
	connCheckChan = make(chan int, 1000)
	totalTCPConn = TRADERCOUNT

	for i := range members {
		for j := range members[i].Traders {
			wg.Add(1)
			trader := &members[i].Traders[j]
			t_session := (trader.MID * 100000) + trader.TID
			traderMap[t_session] = trader
			trader.traderChan = make(chan redisData, DATA_CHANNEL_LENGTH)

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

	// fmt.Println("ORDER PUMPING WILL START IN 2 Mins...")
	// time.Sleep(120*time.Second)
	// fmt.Println("ORDER PUMPING STARTED...")
	// orderPump_flag = false

	go PumpData()
	go PumpTradersData()

	wg.Wait()
}

func PumpData() {
	ticker := time.NewTicker(1 * time.Second)
	second := 0

loop:
	for {
		select {
		case <-ticker.C:
			fmt.Println("Ticker triggered for second:", second)
			if second == TOTAL_RUN_TIME+1 {
				ticker.Stop()
				break loop
			}
			second++
			select {
			case data := <-mainChannel:
				// fmt.Println(data, "IN MAIN CHANNEL LISTNER!!!", len(data))
				if len(data) != 0 {

					jsonData := ProcessValues(data)

					for _, order := range jsonData {
						dataChan <- order
					}

					fmt.Println("Orders pushed in data channel")

				}

			default:
				log.Println("No data available in mainChannel")
			}
		}
	}

	fmt.Println("ORDER PUMPING COMPLETED!!")
}

func PumpTradersData() {
	for data := range dataChan {
		traderID := data.TraderId
		fmt.Printf("Fetched data for session %d from the Data Channel : %v \n", traderID, data)

		trader, ok := traderMap[traderID]
		if !ok {
			fmt.Printf("Trader with ID %d not found\n", traderID)
			continue
		}

		trader.traderChan <- data
	}
}

func (t *Trader) establishConnection(attempt int) error {
	attempt++
	t.session_id = (t.MID * 100000) + t.TID
	t.msg_seq = 1
	// sslConn, err := t.createConnectionTLS()
	// if err != nil {
	// 	return err
	// }

	// Step 1: Connection Request/Response
	// if err := t.handleConnectionRequest(sslConn); err != nil {
	// 	return err
	// }

	// sslConn.Close()
	// fmt.Println("[INFO] SSL Connection Closed for session id :", t.session_id)

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
	// if err := t.handleSessionRegistration(); err != nil {
	// 	return err
	// }
	// t.msg_seq++

	// Step 3: Session Login
	// if err := t.handleSessionLogon(); err != nil {
	// 	return err
	// }
	// t.msg_seq++

	// Step 4: User Login
	// if err := t.handleUserLogin(); err != nil {
	// 	return err
	// }
	// t.msg_seq++

	// go func(){
	// 	 err= t.heartBeatListener()
	// 	 if err != nil {
	// 		// return err
	// 		fmt.Println(err)
	// 	}
	// }()

	if err := t.handleOrderEntry(); err != nil {
		return err
	}

	t.msg_seq++

	// time.Sleep(600*time.Second)
	time.Sleep(10000 * time.Millisecond)

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
				err = SetSocketBuffer(tcpConn, sendBufferSize, recvBufferSize)
				if err != nil {
					return sslConn, fmt.Errorf("Failed to set socket buffer sizes: %v", err)
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

	address := net.JoinHostPort(t.CGW_Res.GW1, strconv.FormatUint(uint64(t.CGW_Res.Port1), 10))

	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect: %v", err)
	}

	tcpConn, ok := conn.(*net.TCPConn)
	if ok {
		err = SetSocketBuffer(tcpConn, sendBufferSize, recvBufferSize)
		if err != nil {
			return conn, fmt.Errorf("Failed to set socket buffer sizes: %v", err)
		}
	} else {
		fmt.Println("[ERR] Failed to get TCP Conn.")
	}
	return conn, nil
}

func (t *Trader) handleConnectionRequest(sslConn *tls.Conn) error {

	gatewayRequest, err := CreateGatewayConnectionRequest(t.MID, t.TID, t.Password)
	if err != nil {
		return fmt.Errorf("Failed to create gateway connection request with session id: %v %v", t.session_id, err)
	}

	_, err = sslConn.Write(gatewayRequest)
	if err != nil {
		return fmt.Errorf("Failed to send tls connection request with session id: %v %v", t.session_id, err)
	}

	buffer := make([]byte, 2048)
	n, err := sslConn.Read(buffer)
	if err != nil {
		return fmt.Errorf("Error reading tls connection response for session id: %v %v", t.session_id, err)
	}

	response := buffer[:n]

	val, err := GetTemplateID(response)
	if err != nil {
		fmt.Println("[ERR] Error in template id parser in connection gateway response :", err)
		return nil
	}

	fmt.Println("TemplateID: ", val)

	if val == REJECTED {
		rejectedMsg, err := ReadRejectedMsgNonDecrypted(response)
		if err != nil {
			fmt.Println("[ERR] Error in rejected message parser connection gateway response:", err)
			return nil
		}
		fmt.Printf("[ERR] Connection Gateway Response for session id %v  has been rejected.\n", t.session_id)
		return fmt.Errorf("Reason: %v", rejectedMsg)
	}

	gatewayResponse, err := ConnectionGatewayResponse(response)
	if err != nil {
		return fmt.Errorf("Error parsing connection gateway response: %v", err)
	}

	t.CGW_Res = *gatewayResponse

	fmt.Println("[INFO] Gateway Connected Successfully for session id: ", t.session_id)
	return nil
}

func (t *Trader) handleSessionRegistration() error {

	sessionRegRequest, err := CreateSessionRegistrationRequest(t.MID, t.TID, t.msg_seq)
	if err != nil {
		return fmt.Errorf("Failed to create session registration request with session id: %v %v", t.session_id, err)
	}

	_, err = t.conn.Write(sessionRegRequest)
	if err != nil {
		return fmt.Errorf("Failed to send session registration request with session id: %v %v", t.session_id, err)
	}

	buffer := make([]byte, 2048)
	n, err := t.conn.Read(buffer)
	if err != nil {
		return fmt.Errorf("Error reading session registration response for session id: %v %v", t.session_id, err)
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
		return fmt.Errorf("Reason: %v", rejectedMsg)
	}

	fmt.Println("[INFO] Session Registration Completed Successfully for session id:", t.session_id)
	return nil
}

func (t *Trader) handleSessionLogon() error {

	sessionLogonRequest, err := CreateSessionLogonRequest(t.MID, t.TID, t.msg_seq, t.Password, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
	if err != nil {
		return fmt.Errorf("Failed to create session logon request with session id: %v %v", t.session_id, err)
	}

	_, err = t.conn.Write(sessionLogonRequest)
	if err != nil {
		return fmt.Errorf("Failed to send session logon request with session id: %v %v", t.session_id, err)
	}

	buffer := make([]byte, 2048)
	n, err := t.conn.Read(buffer)
	if err != nil {
		return fmt.Errorf("Error reading session logon response for session id: %v %v", t.session_id, err)
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
		return fmt.Errorf("Reason: %v", rejectedMsg)
	}

	session_res, err := DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
	if err != nil {
		fmt.Println("[ERR] Error in decrypting session logon response :", err)
		return nil
	}

	sessionLogonResponse, err := SessionLogonResponse(session_res)
	if err != nil {
		return fmt.Errorf("Error parsing session logon response: %v", err)
	}
	t.throttle = sessionLogonResponse

	fmt.Println("[INFO] Session Logon Completed Successfully for Session id:", t.session_id)
	return nil
}

func (t *Trader) handleUserLogin() error {

	userLogonRequest, err := CreateUserLogonRequest(t.MID, t.TID, t.msg_seq, t.Password, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
	if err != nil {
		return fmt.Errorf("Failed to create user logon request with session id: %v %v", t.session_id, err)
	}

	_, err = t.conn.Write(userLogonRequest)
	if err != nil {
		return fmt.Errorf("Failed to send user logon request with session id: %v %v", t.session_id, err)
	}

	buffer := make([]byte, 2048)
	n, err := t.conn.Read(buffer)
	if err != nil {
		return fmt.Errorf("Error reading user logon response with session id: %v %v", t.session_id, err)
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
		return fmt.Errorf("Reason: %v", rejectedMsg)
	}

	_, err = DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
	if err != nil {
		fmt.Println("[ERR] Error in response user logon parser :", err)
		return nil
	}

	fmt.Println("[INFO] User Logon Completed Successfully for session id:", t.session_id)
	return nil
}

func (t *Trader) handleOrderEntry() error {

	startTime := time.Now()
	totalTime := time.Now()
	packetCount := 0
	totalPacketCount := 0
	// batchSize := int(t.throttle)
	batchSize := THROTTLE_VALUE

	sendMod := false
	if MODIFY_PERCENT != 0 {
		sendMod = true
	}
	sendModCount := int(math.Floor(float64(MODIFY_PERCENT*batchSize) / 100))

	cancelOrders := int(math.Floor(float64(CANCEL_PERCENT*batchSize) / 100))
	currentCancelCount := 0
	fmt.Println(sendMod, currentCancelCount)

	fmt.Printf("MODIFY ORDER count/sec for session id  %v : %v \n", t.session_id, sendModCount)
	fmt.Printf("CANCEL ORDER count/sec for session id  %v : %v \n", t.session_id, cancelOrders)

	for data := range t.traderChan {

		// Check if we've reached the batch size
		fmt.Printf("\nData received for sessionid %v: %v \n", t.session_id, data)
		elapsedTime := time.Since(startTime)
		if packetCount >= batchSize && elapsedTime < 1000*time.Millisecond {
			time.Sleep(1000*time.Millisecond - elapsedTime)
			fmt.Printf("Session id %d sent %d packets, sleeping for %v\n", t.session_id, packetCount, 1000*time.Millisecond-elapsedTime)
			startTime = time.Now()
			// totalPacketCount = totalPacketCount + packetCount
			packetCount = 0
			currentCancelCount = 0
			sendMod = true
		}

		if MARKET_REPLAY {
			switch data.TransactionType {
			case ORDER_ENTRY:

				orderRequest, err := SingleLegLeanOrder(t.MID, t.TID, t.msg_seq, data.InstID, int(data.Price), t.CTX)
				if err != nil {
					return err
				}

				err = t.sendDataToServer(orderRequest)
				if err != nil {
					fmt.Printf("Error sending data with session id %d: %v", t.session_id, err)
				}

				order_id, act_time, order_template, err := t.OrderResponseListner(data.InstID)
				if err != nil {
					fmt.Printf("Error in reading received response for session id %d: %v", t.session_id, err)
					return err
				}

				if order_template == HEARTBEAT || order_template == REJECTED {
					break
				}

				orderID_str := strconv.FormatFloat(data.orderID, 'f', -1, 64)

				cache[orderID_str] = CacheEntry{
					NewOrderID:          int64(order_id),
					UpdatedActivityTime: int64(act_time),
					Used:                false,
				}

				fmt.Println(cache, "====== CACHE ======", len(cache))

				// case ORDER_MODIFY:
				// 	orderID_str := strconv.FormatFloat(data.orderID, 'f', -1, 64)
				// 	order_id := uint64(cache[orderID_str].NewOrderID)
				// 	act_time := uint64(cache[orderID_str].UpdatedActivityTime)

				// 	order_id, act_time, order_template, err := t.sendOrderModify(data.InstID, int(data.Price), data.OrderQty, order_id, act_time)
				// 	if err != nil {
				// 		fmt.Printf("Error in order modification for session id %d: %v", t.session_id, err)
				// 	}

				// 	if order_template == HEARTBEAT || order_template == REJECTED {
				// 		break
				// 	}

				// 	cache[orderID_str] = CacheEntry{
				// 		NewOrderID:          int64(order_id),
				// 		UpdatedActivityTime: int64(act_time),
				// 		Used:                false,
				// 	}

				// case ORDER_CANCEL:
				// 	orderID_str := strconv.FormatFloat(data.orderID, 'f', -1, 64)
				// 	order_id := uint64(cache[orderID_str].NewOrderID)

				// 	err := t.sendOrderCancel(data.InstID, data.ProductId, order_id)
				// 	if err != nil {
				// 		fmt.Printf("Error in order cancellation for session id %d: %v", t.session_id, err)
				// 	}
			}

			packetCount++
			t.msg_seq++
			totalPacketCount++
		} else {

			orderRequest, err := SingleLegLeanOrder(t.MID, t.TID, t.msg_seq, data.InstID, int(data.Price), t.CTX)
			if err != nil {
				return err
			}

			err = t.sendDataToServer(orderRequest)
			if err != nil {
				fmt.Printf("Error sending data with session id %d: %v", t.session_id, err)
			}

			_, _, _, err = t.OrderResponseListner(data.InstID)
			if err != nil {
				fmt.Printf("Error in reading received response for session id %d: %v", t.session_id, err)
				return err
			}
			t.msg_seq++

			// if order_template == LEAN_ORDER && sendMod {
			// 	for i := 0; i < sendModCount-packetCount; i++ {

			// 		order_id, act_time, order_template, err = t.sendOrderModify(data.InstID, int(data.Price), data.OrderQty, order_id, act_time)
			// 		if err != nil {
			// 			fmt.Printf("Error in order modification for session id %d: %v", t.session_id, err)
			// 		}

			// 		packetCount++
			// 		t.msg_seq++
			// 		totalPacketCount++

			// 		if order_template == HEARTBEAT {
			// 			break
			// 		}

			// 	}

			// 	if packetCount >= sendModCount {
			// 		sendMod = false
			// 	}

			// }

			// if order_template == LEAN_ORDER && CANCEL_PERCENT != 0 && currentCancelCount <= cancelOrders {

			// 	err = t.sendOrderCancel(data.InstID, data.ProductId, order_id)
			// 	if err != nil {
			// 		fmt.Printf("Error in order cancellation for session id %d: %v", t.session_id, err)
			// 	}

			// 	currentCancelCount++
			// 	t.msg_seq++

			// }

			packetCount++
			// t.msg_seq++
			totalPacketCount++
		}

		if packetCount >= batchSize {
			packetCount = 0
			currentCancelCount = 0
			sendMod = true
		}

		if len(dataChan) == 0 && MARKET_REPLAY == false {
			fmt.Printf("Session id %d took total time: %v and sent packets: %v \n", t.session_id, time.Since(totalTime), totalPacketCount)
			return nil
		}
	}

	return nil
}

func (t *Trader) sendDataToServer(order []byte) error {

	_, err := t.conn.Write(order)
	if err != nil {
		fmt.Printf("[ERR] Error sending data with session id %d: %v", t.session_id, err)
	}

	return nil
}

func (t *Trader) OrderResponseListner(inst_id int) (uint64, uint64, uint16, error) {

	for {
		reader := bufio.NewReader(t.conn)
		buffer := make([]byte, 2048)

		n, err := reader.Read(buffer)

		if err != nil {
			if err == io.EOF {
				fmt.Printf("[INFO] Connection closed for session id %d: %v\n", t.session_id, err)
				return 0, 0, 0, err
			}
			fmt.Printf("[ERR] Error reading responses for session id %d: %v", t.session_id, err)
			return 0, 0, 0, nil
		}

		response := buffer[:n]
		fmt.Println("[INFO] Response Received: ", response)
		act_time := uint64(time.Now().UnixNano())
		order_id := uint64(1111)
		return order_id, act_time, LEAN_ORDER, nil

		// val, err := GetTemplateID(response)
		// if err != nil {
		// 	fmt.Println("[ERR] Error in template id parser :", err)
		// 	// return
		// }

		// read_response, err := t.readReceivedResponse(val, inst_id, response)
		// if err != nil {
		// 	fmt.Println("[ERR] Error in reading received response :", err)
		// }

		// if read_response != ORDER_CONFIRMATION {

		// 	if read_response == LEAN_ORDER {
		// 		order_res, err := DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
		// 		if err != nil {
		// 			fmt.Println("[ERR] Error in decryption for order response parser :", err)
		// 			// return nil
		// 			// return
		// 		}

		// 		_, err1 := GetTimestamps(order_res)
		// 		if err1 != nil {
		// 			fmt.Println("[ERR] Error while getting timestamps")
		// 		}

		// 		order_id, act_time, err := SingleLegLeanOrderResponse(order_res, t.CTX)
		// 		if err != nil {
		// 			fmt.Println("[ERR] Error in lean order response parser :", err)
		// 			// return nil
		// 			// return
		// 		}

		// 		return order_id, act_time, LEAN_ORDER, nil
		// 	}

		// 	if read_response == CANCEL_LEAN_ORDER {

		// 		_, err := DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
		// 		if err != nil {
		// 			fmt.Println("[ERR] Error in decryption for order response parser :", err)
		// 		}

		// 		return 0, 0, CANCEL_LEAN_ORDER, nil
		// 	}

		// 	if read_response == MODIFY_LEAN_ORDER {

		// 		order_res, err := DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
		// 		if err != nil {
		// 			fmt.Println("[ERR] Error in decryption for order response parser :", err)
		// 			// return nil
		// 			// return
		// 		}
		// 		_, err1 := GetTimestamps(order_res)
		// 		if err1 != nil {
		// 			fmt.Println("[ERR] Error while getting timestamps")
		// 		}

		// 		order_id, act_time, err := SingleLegLeanOrderModifiedResponse(order_res, t.CTX)
		// 		if err != nil {
		// 			fmt.Println("[ERR] Error in lean order response parser :", err)
		// 			// return nil
		// 			// return
		// 		}

		// 		return order_id, act_time, MODIFY_LEAN_ORDER, nil
		// 	}

		// 	if read_response == HEARTBEAT {
		// 		return 0, 0, HEARTBEAT, nil
		// 	}

		// 	if read_response == REJECTED {
		// 		return 0, 0, REJECTED, nil
		// 	}

		// }
		// fmt.Println("[PASS] Order Confirmation Received for session id:", t.session_id)
		// continue
	}
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
		fmt.Println("Reason:", rejectedMsg)
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
		fmt.Printf("[PASS] Order Modifed successfully by session id %v for instrument id - %v (Lean Order).\n", t.session_id, inst_id)
		t.AckCount++
		return MODIFY_LEAN_ORDER, nil

	case CANCEL_LEAN_ORDER:
		fmt.Printf("[PASS] Order Cancelled successfully by session id %v for instrument id - %v (Lean Order).\n", t.session_id, inst_id)
		t.AckCount++
		return CANCEL_LEAN_ORDER, nil

	default:
		return 0, nil
	}
}

func (t *Trader) sendOrderModify(inst_id, price, qty int, order_id, act_time uint64) (uint64, uint64, uint16, error) {

	modified_data, err := ModifyLeanOrder(t.MID, t.TID, t.msg_seq, inst_id, price, qty, order_id, act_time, t.CTX)
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

	order_id, act_time, order_template, err := t.OrderResponseListner(inst_id)
	fmt.Println("[INFO] Mod Order Request RTT: ", time.Since(modOrderTime))
	if err != nil {
		fmt.Println("[ERR] Error in reading received response :", err)
	}

	return order_id, act_time, order_template, nil
}

func (t *Trader) sendOrderCancel(inst_id, product_id int, order_id uint64) error {

	cancelOrder, err := CancelSingleLegOrderRequest(t.MID, t.TID, t.msg_seq, inst_id, product_id, order_id, t.CTX)
	if err != nil {
		fmt.Println("[ERR] Error in cancel order", err)
	}

	cancelOrderTime := time.Now()
	err = t.sendDataToServer(cancelOrder)
	if err != nil {
		fmt.Printf("[ERR] Error sending cancel order for session id %d: %v", t.session_id, err)
		return err
	}

	_, _, _, err = t.OrderResponseListner(inst_id)
	fmt.Println("[INFO] Cancel Order Request RTT: ", time.Since(cancelOrderTime))
	if err != nil {
		fmt.Println("[ERR] Error in reading received response :", err)
	}

	return nil
}

func (t *Trader) heartBeatListener() error {

	for {

		if orderPump_flag == false {
			break
		}
		reader := bufio.NewReader(t.conn)
		buffer := make([]byte, 2048)

		if len(connCheckChan) >= (totalTCPConn) {
			fmt.Println("[INFO] All connections are established. Starting to pump.")
			return nil
		}

		n, err := reader.Read(buffer)

		if err != nil {
			if err == io.EOF {
				fmt.Println("[ERR] Received EOF")
				return err
			}
			fmt.Println("[ERR] Error reading:", err)
			return err
		}

		response := buffer[:n]

		val, err := GetTemplateID(response)
		if err != nil {
			fmt.Println("[ERR] Error in template id parser :", err)
			// return
		}

		read_response, err := t.readReceivedResponse(val, 0, response)
		if err != nil {
			fmt.Println("[ERR] Error in reading received response :", err)
		}

		if read_response == HEARTBEAT {
			heartbeat, err := HeartBeatRequest()
			if err != nil {
				return err
			}

			t.sendDataToServer(heartbeat)
			//return nil
		}
	}

	return nil
}
