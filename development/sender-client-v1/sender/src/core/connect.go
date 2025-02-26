// Sync Response
package core

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	utils "market_replay/src/utils"
	"math"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Traders struct {
	*utils.Trader
}

var (
	dataChan         = make(chan utils.RedisData, utils.DATA_CHANNEL_LENGTH)
	connCheckChan    chan int
	totalTCPConn     int
	members          []utils.Member
	recordFile       string
	cache            = make(map[string]utils.CacheEntry)
	cacheMu          sync.Mutex
	traderMap        = make(map[int]Traders)
	sendOrders       = false
	connectedTraders = 0
)

const (
	retryConn = 5
)

func Start() {
	time.Sleep(60 * time.Second)
	var wg sync.WaitGroup
	connCheckChan = make(chan int, 1000)
	totalTCPConn = utils.TRADERCOUNT

	members = utils.TRADER_IDS
	for i := range members {
		for j := range members[i].Traders {
			wg.Add(1)
			trader := Traders{Trader: &members[i].Traders[j]}
			Session_id := (trader.MID * 100000) + trader.TID
			traderMap[Session_id] = trader
			trader.TraderChan = make(chan utils.RedisData, utils.DATA_CHANNEL_LENGTH)

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

	go checkConnectedTraders()
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
			// fmt.Println("Ticker triggered for second:", second)
			if second == utils.TOTAL_RUN_TIME+1 {
				ticker.Stop()
				break loop
			}
			second++
			select {
			case data := <-utils.MainChannel:
				if len(data) != 0 {

					jsonData := utils.ProcessValues(data)

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
		// fmt.Println("Fetched data from the Data Channel : ", data)
		traderID := data.TraderId

		trader, ok := traderMap[traderID]
		if !ok {
			fmt.Printf("Trader with ID %d not found\n", traderID)
			continue
		}
		// fmt.Println("here", data)
		trader.TraderChan <- data
	}

}

func (t Traders) establishConnection(attempt int) error {
	attempt++
	t.Session_id = (t.MID * 100000) + t.TID
	t.Msg_seq = 1
	// sslConn, err := t.createConnectionTLS()
	// if err != nil {
	// 	return err
	// }

	// // Step 1: Connection Request/Response
	// if err := t.handleConnectionRequest(sslConn); err != nil {
	// 	return err
	// }

	// sslConn.Close()
	// fmt.Println("[INFO] SSL Connection Closed for session id :", t.Session_id)

	conn, err := t.createConnection()
	if err != nil {
		return err
	}
	t.Conn = conn

	// // Initialize the CipherContext
	// cipherCtx, err := aes.NewCipherContext([]byte(t.CGW_Res.SecKey), []byte(t.CGW_Res.IV))
	// if err != nil {
	// 	fmt.Printf("[ERR] Error initializing CipherContext: %v", err)
	// }
	// t.CTX = cipherCtx

	// // Step 2: Session Registration
	// if err := t.handleSessionRegistration(); err != nil {
	// 	return err
	// }
	// t.Msg_seq++

	// // Step 3: Session Login
	// if err := t.handleSessionLogon(); err != nil {
	// 	return err
	// }
	// t.Msg_seq++

	// // Step 4: User Login
	// if err := t.handleUserLogin(); err != nil {
	// 	return err
	// }
	// t.Msg_seq++
	connectedTraders++

	// if err := t.SendHeartBeat(); err != nil {
	// 	return err
	// }

	if utils.MARKET_REPLAY {
		if err := t.handleOrderEntryMR(); err != nil {
			return err
		}
	} else {
		if err := t.handleOrderEntryOP(); err != nil {
			return err
		}
	}

	t.Msg_seq++

	// time.Sleep(600*time.Second)
	time.Sleep(10000 * time.Millisecond)

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
				err = utils.SetSocketBuffer(tcpConn, sendBufferSize, recvBufferSize)
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

func (t Traders) createConnection() (net.Conn, error) {

	sendBufferSize := 64 * 1024 * 1024
	recvBufferSize := 64 * 1024 * 1024

	// address := net.JoinHostPort(t.CGW_Res.GW1, strconv.FormatUint(uint64(t.CGW_Res.Port1), 10))
	address := net.JoinHostPort(t.HostIP, t.Port)

	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect: %v", err)
	}

	tcpConn, ok := conn.(*net.TCPConn)
	if ok {
		err = utils.SetSocketBuffer(tcpConn, sendBufferSize, recvBufferSize)
		if err != nil {
			return conn, fmt.Errorf("Failed to set socket buffer sizes: %v", err)
		}
	} else {
		fmt.Println("[ERR] Failed to get TCP Conn.")
	}
	return conn, nil
}

func (t Traders) handleConnectionRequest(sslConn *tls.Conn) error {

	gatewayRequest, err := utils.CreateGatewayConnectionRequest(t.MID, t.TID, t.Password)
	if err != nil {
		return fmt.Errorf("Failed to create gateway connection request with session id: %v %v", t.Session_id, err)
	}

	_, err = sslConn.Write(gatewayRequest)
	if err != nil {
		return fmt.Errorf("Failed to send tls connection request with session id: %v %v", t.Session_id, err)
	}

	buffer := make([]byte, 2048)
	n, err := sslConn.Read(buffer)
	if err != nil {
		return fmt.Errorf("Error reading tls connection response for session id: %v %v", t.Session_id, err)
	}

	response := buffer[:n]

	val, err := utils.GetTemplateID(response)
	if err != nil {
		fmt.Println("[ERR] Error in template id parser in connection gateway response :", err)
		return nil
	}

	fmt.Println("TemplateID: ", val)

	if val == utils.REJECTED {
		rejectedMsg, err := utils.ReadRejectedMsgNonDecrypted(response)
		if err != nil {
			fmt.Println("[ERR] Error in rejected message parser connection gateway response:", err)
			return nil
		}
		fmt.Printf("[ERR] Connection Gateway Response for session id %v  has been rejected.\n", t.Session_id)
		return fmt.Errorf("Reason: %v", rejectedMsg)
	}

	gatewayResponse, err := utils.ConnectionGatewayResponse(response)
	if err != nil {
		return fmt.Errorf("Error parsing connection gateway response: %v", err)
	}

	t.CGW_Res = *gatewayResponse

	fmt.Println("[INFO] Gateway Connected Successfully for session id: ", t.Session_id)
	return nil
}

func (t Traders) handleSessionRegistration() error {

	sessionRegRequest, err := utils.CreateSessionRegistrationRequest(t.MID, t.TID, t.Msg_seq)
	if err != nil {
		return fmt.Errorf("Failed to create session registration request with session id: %v %v", t.Session_id, err)
	}

	_, err = t.Conn.Write(sessionRegRequest)
	if err != nil {
		return fmt.Errorf("Failed to send session registration request with session id: %v %v", t.Session_id, err)
	}

	buffer := make([]byte, 2048)
	n, err := t.Conn.Read(buffer)
	if err != nil {
		return fmt.Errorf("Error reading session registration response for session id: %v %v", t.Session_id, err)
	}

	response := buffer[:n]

	val, err := utils.GetTemplateID(response)
	if err != nil {
		fmt.Println("[ERR] Error in template id parser in session registration :", err)
		return nil
	}

	if val == utils.REJECTED {
		rejectedMsg, err := utils.ReadRejectedMsgNonDecrypted(response)
		if err != nil {
			fmt.Println("[ERR] Error in rejected message parser:", err)
			return nil
		}
		fmt.Printf("[ERR] Session Registration for session id %v  has been rejected.\n", t.Session_id)
		return fmt.Errorf("Reason: %v", rejectedMsg)
	}

	fmt.Println("[INFO] Session Registration Completed Successfully for session id:", t.Session_id)
	return nil
}

func (t Traders) handleSessionLogon() error {

	sessionLogonRequest, err := utils.CreateSessionLogonRequest(t.MID, t.TID, t.Msg_seq, t.Password, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
	if err != nil {
		return fmt.Errorf("Failed to create session logon request with session id: %v %v", t.Session_id, err)
	}

	_, err = t.Conn.Write(sessionLogonRequest)
	if err != nil {
		return fmt.Errorf("Failed to send session logon request with session id: %v %v", t.Session_id, err)
	}

	buffer := make([]byte, 2048)
	n, err := t.Conn.Read(buffer)
	if err != nil {
		return fmt.Errorf("Error reading session logon response for session id: %v %v", t.Session_id, err)
	}

	response := buffer[:n]

	val, err := utils.GetTemplateID(response)
	if err != nil {
		fmt.Println("[ERR] Error in template id parser :", err)
		return nil
	}

	if val == utils.REJECTED {
		rejectedMsg, err := utils.ReadRejectedMsg(response, t.CTX)
		if err != nil {
			fmt.Println("[ERR] Error in rejected message parser in session logon:", err)
			return nil
		}
		fmt.Printf("[ERR] Session Logon for session id %v  has been rejected.\n", t.Session_id)
		return fmt.Errorf("Reason: %v", rejectedMsg)
	}

	session_res, err := utils.DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
	if err != nil {
		fmt.Println("[ERR] Error in decrypting session logon response :", err)
		return nil
	}

	sessionLogonResponse, err := utils.SessionLogonResponse(session_res)
	if err != nil {
		return fmt.Errorf("Error parsing session logon response: %v", err)
	}
	t.ThrottleValue = sessionLogonResponse

	fmt.Println("[INFO] Session Logon Completed Successfully for Session id:", t.Session_id)
	return nil
}

func (t Traders) handleUserLogin() error {

	userLogonRequest, err := utils.CreateUserLogonRequest(t.MID, t.TID, t.Msg_seq, t.Password, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
	if err != nil {
		return fmt.Errorf("Failed to create user logon request with session id: %v %v", t.Session_id, err)
	}

	_, err = t.Conn.Write(userLogonRequest)
	if err != nil {
		return fmt.Errorf("Failed to send user logon request with session id: %v %v", t.Session_id, err)
	}

	buffer := make([]byte, 2048)
	n, err := t.Conn.Read(buffer)
	if err != nil {
		return fmt.Errorf("Error reading user logon response with session id: %v %v", t.Session_id, err)
	}

	response := buffer[:n]
	val, err := utils.GetTemplateID(response)
	if err != nil {
		fmt.Println("[ERR] Error in template id parser :", err)
		return nil
	}

	if val == utils.REJECTED {
		rejectedMsg, err := utils.ReadRejectedMsg(response, t.CTX)
		if err != nil {
			fmt.Println("[ERR] Error in rejected message parser:", err)
			return nil
		}
		fmt.Printf("[ERR] User Logon for session id %v  has been rejected.\n", t.Session_id)
		return fmt.Errorf("Reason: %v", rejectedMsg)
	}

	_, err = utils.DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
	if err != nil {
		fmt.Println("[ERR] Error in response user logon parser :", err)
		return nil
	}

	fmt.Println("[INFO] User Logon Completed Successfully for session if:", t.Session_id)
	return nil
}

func (t Traders) handleOrderEntryMR() error {

	startTime := time.Now()
	totalTime := time.Now()
	packetCount := 0
	totalPacketCount := 0
	// batchSize := int(t.throttle)
	batchSize := utils.THROTTLE_VALUE

	for {
		select {
		case data, ok := <-t.TraderChan:

			if !ok {
				return fmt.Errorf("[INFO] Trader channel closed.")
			}

			// fmt.Println("\nData: ", data)
			elapsedTime := time.Since(startTime)
			if packetCount >= batchSize && elapsedTime < 1000*time.Millisecond {
				time.Sleep(1000*time.Millisecond - elapsedTime)
				fmt.Printf("Session id %d sent %d packets, sleeping for %v\n", t.Session_id, packetCount, 1000*time.Millisecond-elapsedTime)
				startTime = time.Now()
				// totalPacketCount = totalPacketCount + packetCount
				packetCount = 0
			}

			switch data.TransactionType {
			case utils.ORDER_ENTRY:

				order_id, act_time, order_template, err := t.RetryOrderEntry(data.InstID, int(data.Price), 1)
				if err != nil {
					fmt.Printf("Error in reading received response for session id %d: %v", t.Session_id, err)
					break
				}

				if order_template == utils.REJECTED {
					break
				}

				orderID_str := strconv.FormatFloat(data.OrderID, 'f', -1, 64)

				cacheMu.Lock()
				cache[orderID_str] = utils.CacheEntry{
					NewOrderID:          int64(order_id),
					UpdatedActivityTime: int64(act_time),
					Used:                false,
				}

				// fmt.Println(cache, "====== CACHE")
				cacheMu.Unlock()

				// case utils.ORDER_MODIFY:
				// 	orderID_str := strconv.FormatFloat(data.OrderID, 'f', -1, 64)
				// 	order_id := uint64(cache[orderID_str].NewOrderID)
				// 	act_time := uint64(cache[orderID_str].UpdatedActivityTime)

				// 	order_id, act_time, order_template, err := t.sendOrderModify(data.InstID, int(data.Price), data.OrderQty, order_id, act_time)
				// 	if err != nil {
				// 		fmt.Printf("Error in order modification for session id %d: %v", t.Session_id, err)
				// 	}

				// 	if order_template == utils.HEARTBEAT || order_template == utils.REJECTED {
				// 		break
				// 	}

				// cacheMu.Lock()
				// 	cache[orderID_str] = utils.CacheEntry{
				// 		NewOrderID:          int64(order_id),
				// 		UpdatedActivityTime: int64(act_time),
				// 		Used:                false,
				// 	}
				// cacheMu.Unlock()

				// case utils.ORDER_CANCEL:
				// 	orderID_str := strconv.FormatFloat(data.OrderID, 'f', -1, 64)
				// 	order_id := uint64(cache[orderID_str].NewOrderID)

				// 	err := t.sendOrderCancel(data.InstID, data.ProductId, order_id)
				// 	if err != nil {
				// 		fmt.Printf("Error in order cancellation for session id %d: %v", t.Session_id, err)
				// 	}
			}

			packetCount++
			t.Msg_seq++
			totalPacketCount++

			if packetCount >= batchSize {
				packetCount = 0
			}

			if len(dataChan) == 0 && utils.MARKET_REPLAY == false {
				fmt.Printf("Session id %d took total time: %v and sent packets: %v \n", t.Session_id, time.Since(totalTime), totalPacketCount)
				return nil
			}

			// case <-time.After(1 * time.Second):
			// 	// Timeout after 1 second, send a heartbeat if no data was received
			// 	fmt.Printf("[INFO] Sending heartbeat for Session_id %v ... \n", t.Session_id)
			// 	heartbeat, err := utils.HeartBeatRequest()
			// 	if err != nil {
			// 		fmt.Println("[ERR] Failed to create heartbeat:", err)
			// 		continue
			// 	}

			// 	err = t.sendDataToServer(heartbeat)
			// 	if err != nil {
			// 		fmt.Println("[ERR] Failed to send heartbeat:", err)
			// 	}

		}
	}
}

func (t Traders) handleOrderEntryOP() error {

	startTime := time.Now()
	totalTime := time.Now()
	packetCount := 0
	totalPacketCount := 0
	// batchSize := int(t.throttle)
	batchSize := utils.THROTTLE_VALUE

	sendMod := false
	if utils.MODIFY_PERCENT != 0 {
		sendMod = true
	}
	sendModCount := int(math.Floor(float64(utils.MODIFY_PERCENT*batchSize) / 100))

	cancelOrders := int(math.Floor(float64(utils.CANCEL_PERCENT*batchSize) / 100))
	currentCancelCount := 0
	fmt.Println(sendMod, currentCancelCount)

	fmt.Printf("MODIFY ORDER count/sec for session id  %v : %v", sendModCount, t.Session_id)
	fmt.Printf("CANCEL ORDER count/sec for session id  %v : %v", cancelOrders, t.Session_id)

	for data := range dataChan {

		// Check if we've reached the batch size
		fmt.Println("\nData: ", data)
		elapsedTime := time.Since(startTime)
		if packetCount >= batchSize && elapsedTime < 1000*time.Millisecond {
			time.Sleep(1000*time.Millisecond - elapsedTime)
			fmt.Printf("Session id %d sent %d packets, sleeping for %v\n", t.Session_id, packetCount, 1000*time.Millisecond-elapsedTime)
			startTime = time.Now()
			// totalPacketCount = totalPacketCount + packetCount
			packetCount = 0
			currentCancelCount = 0
			sendMod = true
		}

		orderRequest, err := utils.SingleLegLeanOrder(t.MID, t.TID, t.Msg_seq, data.InstID, int(data.Price), "" /*t.CTX*/)
		if err != nil {
			return err
		}

		err = t.sendDataToServer(orderRequest)
		if err != nil {
			fmt.Printf("Error sending data with session id %d: %v", t.Session_id, err)
		}

		order_id, act_time, order_template, err := t.OrderResponseListner(data.InstID)
		if err != nil {
			fmt.Printf("Error in reading received response for session id %d: %v", t.Session_id, err)
			return err
		}
		t.Msg_seq++

		if order_template == utils.LEAN_ORDER && sendMod {
			for i := 0; i < sendModCount-packetCount; i++ {

				order_id, act_time, order_template, err = t.sendOrderModify(data.InstID, int(data.Price), data.OrderQty, order_id, act_time)
				if err != nil {
					fmt.Printf("Error in order modification for session id %d: %v", t.Session_id, err)
				}

				packetCount++
				t.Msg_seq++
				totalPacketCount++

				if order_template == utils.HEARTBEAT {
					break
				}

			}

			if packetCount >= sendModCount {
				sendMod = false
			}

		}

		if order_template == utils.LEAN_ORDER && utils.CANCEL_PERCENT != 0 && currentCancelCount <= cancelOrders {

			err = t.sendOrderCancel(data.InstID, data.ProductId, order_id)
			if err != nil {
				fmt.Printf("Error in order cancellation for session id %d: %v", t.Session_id, err)
			}

			currentCancelCount++
			t.Msg_seq++

		}

		packetCount++
		// t.Msg_seq++
		totalPacketCount++

		if packetCount >= batchSize {
			packetCount = 0
			currentCancelCount = 0
			sendMod = true
		}

		if len(dataChan) == 0 {
			fmt.Printf("Session id %d took total time: %v and sent packets: %v \n", t.Session_id, time.Since(totalTime), totalPacketCount)
			return nil
		}
	}

	return nil
}

func (t Traders) sendDataToServer(order []byte) error {

	_, err := t.Conn.Write(order)
	if err != nil {
		fmt.Printf("[ERR] Error sending data with session id %d: %v", t.Session_id, err)
	}

	return nil
}

func (t Traders) OrderResponseListner(inst_id int) (uint64, uint64, uint16, error) {

	for {
		reader := bufio.NewReader(t.Conn)
		buffer := make([]byte, 2048)

		n, err := reader.Read(buffer)

		if err != nil {
			if err == io.EOF {
				fmt.Printf("[INFO] Connection closed for session id %d: %v\n", t.Session_id, err)
				return 0, 0, 0, err
			}
			fmt.Printf("[ERR] Error reading responses for session id %d: %v", t.Session_id, err)
			return 0, 0, 0, nil
		}

		response := buffer[:n]

		// =====For Mock===========

		if len(response) > 0 {
			fmt.Printf("Response received for Trader %v: %v", t.Session_id, string(response[:]))
			return 0, 0, 0, nil
		} else {
			fmt.Printf("Error response: %v", string(response[:]))
			return 0, 0, 0, nil
		}

		//=====================

		val, err := utils.GetTemplateID(response)
		if err != nil {
			fmt.Println("[ERR] Error in template id parser :", err)
			// return
		}

		read_response, err := t.readReceivedResponse(val, inst_id, response)
		if err != nil {
			fmt.Println("[ERR] Error in reading received response :", err)
		}

		if read_response != utils.ORDER_CONFIRMATION {

			if read_response == utils.LEAN_ORDER {
				order_res, err := utils.DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
				if err != nil {
					fmt.Println("[ERR] Error in decryption for order response parser :", err)
					// return nil
					// return
				}

				_, err1 := utils.GetTimestamps(order_res)
				if err1 != nil {
					fmt.Println("[ERR] Error while getting timestamps")
				}

				order_id, act_time, err := utils.SingleLegLeanOrderResponse(order_res, t.CTX)
				if err != nil {
					fmt.Println("[ERR] Error in lean order response parser :", err)
					// return nil
					// return
				}

				return order_id, act_time, utils.LEAN_ORDER, nil
			}

			if read_response == utils.CANCEL_LEAN_ORDER {

				_, err := utils.DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
				if err != nil {
					fmt.Println("[ERR] Error in decryption for order response parser :", err)
				}

				return 0, 0, utils.CANCEL_LEAN_ORDER, nil
			}

			if read_response == utils.MODIFY_LEAN_ORDER {

				order_res, err := utils.DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
				if err != nil {
					fmt.Println("[ERR] Error in decryption for order response parser :", err)
					// return nil
					// return
				}
				_, err1 := utils.GetTimestamps(order_res)
				if err1 != nil {
					fmt.Println("[ERR] Error while getting timestamps")
				}

				order_id, act_time, err := utils.SingleLegLeanOrderModifiedResponse(order_res, t.CTX)
				if err != nil {
					fmt.Println("[ERR] Error in lean order response parser :", err)
					// return nil
					// return
				}

				return order_id, act_time, utils.MODIFY_LEAN_ORDER, nil
			}

			if read_response == utils.HEARTBEAT {
				return 0, 0, utils.HEARTBEAT, nil
			}

			if read_response == utils.REJECTED {
				return 0, 0, utils.REJECTED, nil
			}

		}
		fmt.Println("[PASS] Order Confirmation Received for session id:", t.Session_id)
		continue
	}
}

func (t Traders) readReceivedResponse(val uint16, inst_id int, response []byte) (uint16, error) {

	switch val {
	case utils.REJECTED:
		rejectedMsg, err := utils.ReadRejectedMsg(response, t.CTX)
		if err != nil {
			fmt.Println("[ERR] Error in rejected message parser:", err)
			return 0, err
		}
		t.AckCount++
		fmt.Printf("[ERR] Order placed by session id %v has been rejected for instrument id - %v.\n", t.Session_id, inst_id)
		fmt.Println("Reason:", rejectedMsg)
		return utils.REJECTED, nil

	case utils.STANDARD_ORDER, utils.LEAN_ORDER:
		orderType := "Standard Order"
		if val == utils.LEAN_ORDER {
			orderType = "Lean Order"
		}
		fmt.Printf("[PASS] Order placed successfully by session id %v for instrument id - %v (%s).\n", t.Session_id, inst_id, orderType)
		t.AckCount++
		return utils.LEAN_ORDER, nil

	case utils.HEARTBEAT:
		fmt.Println("[INFO] HeartBeat Received for session id:", t.Session_id)
		return utils.HEARTBEAT, nil

	case utils.ORDER_CONFIRMATION:
		_, err := utils.DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
		if err != nil {
			fmt.Println("[ERR] Error in decrypting order confirmation :", err)
			return 0, err
		}

		return utils.ORDER_CONFIRMATION, nil

	case utils.MODIFY_LEAN_ORDER:
		fmt.Printf("[PASS] Order Modifed successfully by session id %v for instrument id - %v (Lean Order).\n", t.Session_id, inst_id)
		t.AckCount++
		return utils.MODIFY_LEAN_ORDER, nil

	case utils.CANCEL_LEAN_ORDER:
		fmt.Printf("[PASS] Order Cancelled successfully by session id %v for instrument id - %v (Lean Order).\n", t.Session_id, inst_id)
		t.AckCount++
		return utils.CANCEL_LEAN_ORDER, nil

	default:
		return 0, nil
	}
}

func (t Traders) sendOrderModify(inst_id, price, qty int, order_id, act_time uint64) (uint64, uint64, uint16, error) {

	modified_data, err := utils.ModifyLeanOrder(t.MID, t.TID, t.Msg_seq, inst_id, price, qty, order_id, act_time, t.CTX)
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

	order_id, act_time, order_template, err := t.OrderResponseListner(inst_id)
	fmt.Println("[INFO] Mod Order Request RTT: ", time.Since(modOrderTime))
	if err != nil {
		fmt.Println("[ERR] Error in reading received response :", err)
	}

	return order_id, act_time, order_template, nil
}

func (t Traders) sendOrderCancel(inst_id, product_id int, order_id uint64) error {

	cancelOrder, err := utils.CancelSingleLegOrderRequest(t.MID, t.TID, t.Msg_seq, inst_id, product_id, order_id, t.CTX)
	if err != nil {
		fmt.Println("[ERR] Error in cancel order", err)
	}

	cancelOrderTime := time.Now()
	err = t.sendDataToServer(cancelOrder)
	if err != nil {
		fmt.Printf("[ERR] Error sending cancel order for session id %d: %v", t.Session_id, err)
		return err
	}

	_, _, _, err = t.OrderResponseListner(inst_id)
	fmt.Println("[INFO] Cancel Order Request RTT: ", time.Since(cancelOrderTime))
	if err != nil {
		fmt.Println("[ERR] Error in reading received response :", err)
	}

	return nil
}

func (t Traders) heartBeatListener() error {

	for {
		reader := bufio.NewReader(t.Conn)
		buffer := make([]byte, 2048)

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

		val, err := utils.GetTemplateID(response)
		if err != nil {
			fmt.Println("[ERR] Error in template id parser :", err)
			// return
		}

		if val == utils.HEARTBEAT {
			heartbeat, err := utils.HeartBeatRequest()
			if err != nil {
				return err
			}

			err = t.sendDataToServer(heartbeat)
			if err != nil {
				fmt.Println("error in sending heartnbeat to server")
				return err
			}
		}
	}
}

func (t Traders) RetryOrderEntry(inst_id, price, retryCount int) (uint64, uint64, uint16, error) {

	if retryCount >= 5 {
		t.Msg_seq--
		return 0, 0, 0, fmt.Errorf("Retry limit reached!")
	}

	orderRequest, err := utils.SingleLegLeanOrder(t.MID, t.TID, t.Msg_seq, inst_id, price, "" /*t.CTX*/)
	if err != nil {
		return 0, 0, 0, err
	}

	err = t.sendDataToServer(orderRequest)
	if err != nil {
		fmt.Printf("Error sending data with session id %d: %v", t.Session_id, err)
	}

	order_id, act_time, order_template, err := t.OrderResponseListner(inst_id)
	if err != nil {
		fmt.Printf("Error in reading received response for session id %d: %v", t.Session_id, err)
		return 0, 0, 0, err
	}

	if order_template == utils.HEARTBEAT {
		t.Msg_seq++
		fmt.Println("OE RETRY COUNT :", retryCount)
		order_id, act_time, order_template, err = t.RetryOrderEntry(inst_id, price, retryCount+1)
	}

	return order_id, act_time, order_template, nil
}

func (t Traders) SendHeartBeat() error {

	for {
		if sendOrders {
			fmt.Println("Paused sending heartbeat message for session id:", t.Session_id)
			return nil
		}

		heartbeat, err := utils.HeartBeatRequest()
		if err != nil {
			fmt.Println("Error in heartbeat request parser")
		}

		err = t.sendDataToServer(heartbeat)
		if err != nil {
			fmt.Println("error in sending heartnbeat to server")
		}

		time.Sleep(time.Duration(utils.HEARTBEAT_VALUE/2) * time.Millisecond)
	}
}

func checkConnectedTraders() error {
	for {
		count := int(math.Floor(float64(utils.TRADERCOUNT*90) / 100))

		if count >= connectedTraders {
			sendOrders = true
			return nil
		}

		time.Sleep(1 * time.Second)
	}
}

// ================ARCHIVE================
