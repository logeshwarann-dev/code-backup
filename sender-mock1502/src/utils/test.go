package utils

// func (t *Trader) heartBeatListener() error {

// 	for {
// 		// reader := bufio.NewReader(t.conn)
// 		// buffer := make([]byte, 16*1024*1024)

// 		// n, err := reader.Read(buffer)

// 		// if err != nil {
// 		// 	if err == io.EOF {
// 		// 		utils.Printf(static.LOG_FLAG, "[ERR] Received EOF")
// 		// 		return err
// 		// 	}
// 		// 	utils.Printf(static.LOG_FLAG, "[ERR] Error reading:", err)
// 		// 	return err
// 		// }

// 		// response := buffer[:n]

// 		// utils.Printf(static.LOG_FLAG, "From Socket :", response, "length of response :", len(response))

// 		// val, err := GetTemplateID(response)
// 		// if err != nil {
// 		// 	utils.Printf(static.LOG_FLAG, "[ERR] Error in template id parser :", err)
// 		// 	// return
// 		// }

// 		// read_response, err := t.readReceivedResponse(val, 0, response)
// 		// if err != nil {
// 		// 	utils.Printf(static.LOG_FLAG, "[ERR] Error in reading received response :", err)
// 		// }

// 		// _, err := DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
// 		// if err != nil {
// 		// 	utils.Printf(static.LOG_FLAG, "[ERR] Error in decryption for order response parser :", err)
// 		// }

// 		_, _, order_template, err := t.NewOrderResponseListner(11)
// 		// utils.Printf(static.LOG_FLAG, "[INFO] Delete Order Request RTT: ", time.Since(deleteOrderTime))
// 		if err != nil {
// 			utils.Printf(static.LOG_FLAG, "[ERR] Error in reading received response :", err)
// 		}

// 		if order_template == HEARTBEAT {
// 			heartbeat, err := HeartBeatRequest()
// 			if err != nil {
// 				return err
// 			}

// 			t.sendDataToServer(heartbeat)

// 			if THROTTLE_VALUE > 0 {
// 				utils.Printf(static.LOG_FLAG, "[INFO] Throttle value is updated. Exiting heartbeat case ")
// 				return nil
// 			}
// 			//return nil
// 		}
// 	}

// }

// func (t *Trader) NewCheckSocketBuffer(inst_id int) (uint64, uint64, uint16, error) {
// 	utils.Printf(static.LOG_FLAG, fmt.Sprintf("For session ID %v, Checking socket in checkSocketBuffer()\n", t.session_id)
// 	order_id := uint64(0)
// 	act_time := uint64(0)
// 	template_type := uint16(0)
// 	reader := bufio.NewReader(t.conn)
// 	if reader.Buffered() == 0 {
// 		utils.Printf(static.LOG_FLAG, "In socket, size is 0")
// 		return order_id, act_time, template_type, nil
// 	}
// 	utils.Printf(static.LOG_FLAG, "In socket, some data is present!")
// 	buffer := make([]byte, 16*1024*1024)
// 	n, err := reader.Read(buffer)
// 	if err != nil {
// 		if err == io.EOF {
// 			utils.Printf(static.LOG_FLAG, fmt.Sprintf("[INFO] Connection closed for session id %d: %v\n", t.session_id, err)
// 			return 0, 0, 0, err
// 		}
// 		utils.Printf(static.LOG_FLAG, fmt.Sprintf("[ERR] Error reading responses for session id %d: %v\n", t.session_id, err)
// 		return 0, 0, 0, nil
// 	}
// 	responseBuf := buffer[:n]
// 	responseLength := len(responseBuf)
// 	if responseLength == 0 {
// 		utils.Printf(static.LOG_FLAG, "In CheckSocketBuffer(), No response is present in Socket buffer.")
// 		return order_id, act_time, 0, nil
// 	}
// 	var responseArr [][]byte
// 	count := 0
// 	for len(responseBuf) > 0 {
// 		count++
// 		if len(responseBuf) < 4 {
// 			utils.Printf(static.LOG_FLAG, "[ERR] Response buffer too short to read size")
// 			return 0, 0, 0, fmt.Errorf("response buffer too short")
// 		}
// 		size := int(binary.LittleEndian.Uint32(responseBuf[:4]))
// 		utils.Printf(static.LOG_FLAG, "size: ", size)
// 		if len(responseBuf) < size {
// 			utils.Printf(static.LOG_FLAG, "[ERR] Response buffer too short for the indicated size")
// 			return 0, 0, 0, fmt.Errorf("response buffer too short for indicated size")
// 		}
// 		responseArr = append(responseArr, responseBuf[:size])
// 		responseBuf = responseBuf[size:]
// 	}
// 	utils.Printf(static.LOG_FLAG, "Len of response array: ", len(responseArr))
// 	for _, response := range responseArr {
// 		if len(response) == 0 {
// 			break
// 		}
// 		val, err := GetTemplateID(response)
// 		if err != nil {
// 			utils.Printf(static.LOG_FLAG, "[ERR] Error in template id parser :", err)
// 		}
// 		utils.Printf(static.LOG_FLAG, fmt.Sprintf("In CheckSockerBuffer(), for Session ID %v, TEMPLATE ID: %v ", t.session_id, val)
// 		read_response, err := t.readReceivedResponse(val, inst_id, response)
// 		if err != nil {
// 			utils.Printf(static.LOG_FLAG, "[ERR] Error in reading received response :", err)
// 		}
// 		switch read_response {
// 		case LEAN_ORDER:
// 			order_res, err := DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
// 			if err != nil {
// 				utils.Printf(static.LOG_FLAG, "[ERR] Error in decryption for order response parser :", err)
// 			}
// 			server_res, err1 := GetTimestamps(order_res)
// 			if err1 != nil {
// 				utils.Printf(static.LOG_FLAG, "[ERR] Error while getting timestamps")
// 			} else {
// 				utils.Printf(static.LOG_FLAG, fmt.Sprintf("[INFO] Order Entry Server Processing Time(T6-T1):  %v microseconds\n", (server_res.GW_Res_Out-server_res.GW_Req_In)/1000)
// 			}
// 			order_id, act_time, err = SingleLegLeanOrderResponse(order_res, t.CTX)
// 			if err != nil {
// 				utils.Printf(static.LOG_FLAG, "[ERR] Error in lean order response parser :", err)
// 			}
// 			template_type = LEAN_ORDER
// 		case CANCEL_LEAN_ORDER:
// 			order_res, err := DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
// 			if err != nil {
// 				utils.Printf(static.LOG_FLAG, "[ERR] Error in decryption for order response parser :", err)
// 			}
// 			server_res, err1 := GetTimestamps(order_res)
// 			if err1 != nil {
// 				utils.Printf(static.LOG_FLAG, "[ERR] Error while getting timestamps")
// 			} else {
// 				utils.Printf(static.LOG_FLAG, fmt.Sprintf("[INFO] Order Cancel Server Processing Time(T6-T1):  %v microseconds\n", (server_res.GW_Res_Out-server_res.GW_Req_In)/1000)
// 			}
// 			order_id = uint64(0)
// 			act_time = uint64(0)
// 			template_type = CANCEL_LEAN_ORDER
// 		case MODIFY_LEAN_ORDER:
// 			order_res, err := DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
// 			if err != nil {
// 				utils.Printf(static.LOG_FLAG, "[ERR] Error in decryption for order response parser :", err)
// 			}
// 			server_res, err1 := GetTimestamps(order_res)
// 			if err1 != nil {
// 				utils.Printf(static.LOG_FLAG, "[ERR] Error while getting timestamps")
// 			} else {
// 				utils.Printf(static.LOG_FLAG, fmt.Sprintf("[INFO] Order Mod Server Processing Time(T6-T1):  %v microseconds\n", (server_res.GW_Res_Out-server_res.GW_Req_In)/1000)
// 			}
// 			order_id, act_time, err = SingleLegLeanOrderModifiedResponse(order_res, t.CTX)
// 			if err != nil {
// 				utils.Printf(static.LOG_FLAG, "[ERR] Error in lean order response parser :", err)
// 			}
// 			template_type = MODIFY_LEAN_ORDER
// 		case DELETE_ALL_ORDERS:
// 			order_res, err := DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
// 			if err != nil {
// 				utils.Printf(static.LOG_FLAG, "[ERR] Error in decryption for order response parser :", err)
// 			}
// 			server_res, err1 := GetTimestamps(order_res)
// 			if err1 != nil {
// 				utils.Printf(static.LOG_FLAG, "[ERR] Error while getting timestamps")
// 			} else {
// 				utils.Printf(static.LOG_FLAG, fmt.Sprintf("[INFO] Order Delete All Orders Server Processing Time(T6-T1):  %v microseconds\n", (server_res.GW_Res_Out-server_res.GW_Req_In)/1000)
// 			}
// 			order_id = uint64(0)
// 			act_time = uint64(0)
// 			template_type = read_response
// 		case DELETE_ALL_ORDERS_NO_HITS:
// 			order_res, err := DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
// 			if err != nil {
// 				utils.Printf(static.LOG_FLAG, "[ERR] Error in decryption for order response parser :", err)
// 			}
// 			server_res, err1 := GetTimestamps(order_res)
// 			if err1 != nil {
// 				utils.Printf(static.LOG_FLAG, "[ERR] Error while getting timestamps")
// 			} else {
// 				utils.Printf(static.LOG_FLAG, fmt.Sprintf("[INFO] Order Delete All Orders Server Processing Time(T6-T1):  %v microseconds\n", (server_res.GW_Res_Out-server_res.GW_Req_In)/1000)
// 			}
// 			order_id = uint64(0)
// 			act_time = uint64(0)
// 			template_type = read_response
// 		case HEARTBEAT:
// 			_, err := DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
// 			if err != nil {
// 				utils.Printf(static.LOG_FLAG, "[ERR] Error in decryption for order response parser :", err)
// 			}
// 			order_id = uint64(0)
// 			act_time = uint64(0)
// 			template_type = HEARTBEAT
// 		case ORDER_CONFIRMATION:
// 			order_id = uint64(0)
// 			act_time = uint64(0)
// 			template_type = ORDER_CONFIRMATION
// 			utils.Printf(static.LOG_FLAG, "[PASS] Order Confirmation Received for session id:", t.session_id)
// 		case REJECTED:
// 			order_id = uint64(0)
// 			act_time = uint64(0)
// 			template_type = REJECTED
// 		case IMMEDIATE_EXECUTION:
// 			_, err := DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
// 			if err != nil {
// 				utils.Printf(static.LOG_FLAG, "[ERR] Error in decryption for order response parser :", err)
// 			}
// 			order_id = uint64(0)
// 			act_time = uint64(0)
// 			template_type = IMMEDIATE_EXECUTION
// 		case ORDER_BOOK_EXECUTION:
// 			_, err := DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
// 			if err != nil {
// 				utils.Printf(static.LOG_FLAG, "[ERR] Error in decryption for order response parser :", err)
// 			}
// 			order_id = uint64(0)
// 			act_time = uint64(0)
// 			template_type = ORDER_BOOK_EXECUTION
// 		default:
// 			order_id = uint64(0)
// 			act_time = uint64(0)
// 			template_type = uint16(0)
// 		}
// 	}
// 	utils.Printf(static.LOG_FLAG, "Response of checkbuffer => Order Id: ", order_id, "Act time:", act_time, "Template Id:", template_type)
// 	return order_id, act_time, template_type, nil
// }

// func (t *Trader) CheckSocketBuffer(inst_id int) (uint64, uint64, uint16, error) {
// 	utils.Printf(static.LOG_FLAG, fmt.Sprintf("FOr session ID %v, Checking socket in checkSocketBuffer()\n", t.session_id)
// 	order_id := uint64(0)
// 	act_time := uint64(0)
// 	template_type := uint16(0)
// 	reader := bufio.NewReader(t.conn)

// 	if reader.Buffered() == 0 {

// 		utils.Printf(static.LOG_FLAG, "In socket, size is 0")
// 		return order_id, act_time, template_type, nil
// 	}

// 	utils.Printf(static.LOG_FLAG, "In socket, some data is present!")

// 	// t.conn.SetReadDeadline(time.Now().Add(100* time.Millisecond))
// 	buffer := make([]byte, 16*1024*1024)

// 	n, err := reader.Read(buffer)

// 	if err != nil {
// 		if err == io.EOF {
// 			utils.Printf(static.LOG_FLAG, fmt.Sprintf("[INFO] Connection closed for session id %d: %v\n", t.session_id, err)
// 			return 0, 0, 0, err
// 		}
// 		utils.Printf(static.LOG_FLAG, fmt.Sprintf("[ERR] Error reading responses for session id %d: %v\n", t.session_id, err)
// 		return 0, 0, 0, nil
// 	}

// 	responseBuf := buffer[:n]

// 	responseLength := len(responseBuf)

// 	// utils.Printf(static.LOG_FLAG, "Len of Response: ", responseLength)
// 	if responseLength == 0 {
// 		utils.Printf(static.LOG_FLAG, "In CheckSocketBuffer(), No response is present in Socket buffer.")
// 		return order_id, act_time, 0, nil
// 	}

// 	var responseArr [][]byte
// 	count := 0
// 	for len(responseBuf) > 0 {
// 		count++
// 		if len(responseBuf) < 4 {
// 			utils.Printf(static.LOG_FLAG, "[ERR] Response buffer too short to read size")
// 			return 0, 0, 0, fmt.Errorf("response buffer too short")
// 		}
// 		size := int(binary.LittleEndian.Uint32(responseBuf[:4]))
// 		utils.Printf(static.LOG_FLAG, "size: ", size)
// 		if len(responseBuf) < size {
// 			utils.Printf(static.LOG_FLAG, "[ERR] Response buffer too short for the indicated size")
// 			return 0, 0, 0, fmt.Errorf("response buffer too short for indicated size")
// 		}

// 		responseArr = append(responseArr, responseBuf[:size])

// 		responseBuf = responseBuf[size:]
// 		// utils.Printf(static.LOG_FLAG, "COunt: ", count)
// 	}

// 	// utils.Printf(static.LOG_FLAG, "Len of response buf: ", len(responseArr))

// 	for _, response := range responseArr {

// 		if len(response) == 0 {
// 			break
// 		}

// 		val, err := GetTemplateID(response)
// 		if err != nil {
// 			utils.Printf(static.LOG_FLAG, "[ERR] Error in template id parser :", err)
// 			// return
// 		}

// 		utils.Printf(static.LOG_FLAG, fmt.Sprintf("In CheckSockerBuffer(), for Session ID %v, TEMPLATE ID: %v ", t.session_id, val)
// 		read_response, err := t.readReceivedResponse(val, inst_id, response)
// 		if err != nil {
// 			utils.Printf(static.LOG_FLAG, "[ERR] Error in reading received response :", err)
// 		}

// 		if read_response == LEAN_ORDER {
// 			order_res, err := DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
// 			if err != nil {
// 				utils.Printf(static.LOG_FLAG, "[ERR] Error in decryption for order response parser :", err)
// 				// return nil
// 				// return
// 			}

// 			server_res, err1 := GetTimestamps(order_res)
// 			if err1 != nil {
// 				utils.Printf(static.LOG_FLAG, "[ERR] Error while getting timestamps")
// 			} else {
// 				utils.Printf(static.LOG_FLAG, fmt.Sprintf("[INFO] Order Entry Server Processing Time(T6-T1):  %v microseconds\n", (server_res.GW_Res_Out-server_res.GW_Req_In)/1000)
// 			}
// 			// utils.Printf(static.LOG_FLAG, "[INFO] TimeStamps: ", timeStamps)

// 			order_id, act_time, err = SingleLegLeanOrderResponse(order_res, t.CTX)
// 			if err != nil {
// 				utils.Printf(static.LOG_FLAG, "[ERR] Error in lean order response parser :", err)
// 				// return nil
// 				// return
// 			}

// 			template_type = LEAN_ORDER
// 			// return order_id, act_time, LEAN_ORDER, nil
// 		}

// 		if read_response == CANCEL_LEAN_ORDER {

// 			order_res, err := DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
// 			if err != nil {
// 				utils.Printf(static.LOG_FLAG, "[ERR] Error in decryption for order response parser :", err)
// 			}

// 			server_res, err1 := GetTimestamps(order_res)
// 			if err1 != nil {
// 				utils.Printf(static.LOG_FLAG, "[ERR] Error while getting timestamps")
// 			} else {
// 				utils.Printf(static.LOG_FLAG, fmt.Sprintf("[INFO] Order Cancel Server Processing Time(T6-T1):  %v microseconds\n", (server_res.GW_Res_Out-server_res.GW_Req_In)/1000)
// 			}

// 			order_id = uint64(0)
// 			act_time = uint64(0)

// 			template_type = CANCEL_LEAN_ORDER
// 			// return 0, 0, CANCEL_LEAN_ORDER, nil

// 		}

// 		if read_response == MODIFY_LEAN_ORDER {

// 			order_res, err := DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
// 			if err != nil {
// 				utils.Printf(static.LOG_FLAG, "[ERR] Error in decryption for order response parser :", err)
// 				// return nil
// 				// return
// 			}
// 			server_res, err1 := GetTimestamps(order_res)
// 			if err1 != nil {
// 				utils.Printf(static.LOG_FLAG, "[ERR] Error while getting timestamps")
// 			} else {
// 				utils.Printf(static.LOG_FLAG, fmt.Sprintf("[INFO] Order Mod Server Processing Time(T6-T1):  %v microseconds\n", (server_res.GW_Res_Out-server_res.GW_Req_In)/1000)
// 			}
// 			// utils.Printf(static.LOG_FLAG, "[INFO] TimeStamps: ", timeStamps)

// 			order_id, act_time, err = SingleLegLeanOrderModifiedResponse(order_res, t.CTX)
// 			if err != nil {
// 				utils.Printf(static.LOG_FLAG, "[ERR] Error in lean order response parser :", err)
// 				// return nil
// 				// return
// 			}

// 			template_type = MODIFY_LEAN_ORDER

// 			// return order_id, act_time, MODIFY_LEAN_ORDER, nil

// 		}

// 		if read_response == DELETE_ALL_ORDERS || read_response == DELETE_ALL_ORDERS_NO_HITS {
// 			order_res, err := DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
// 			if err != nil {
// 				utils.Printf(static.LOG_FLAG, "[ERR] Error in decryption for order response parser :", err)
// 				// return nil
// 				// return
// 			}

// 			server_res, err1 := GetTimestamps(order_res)
// 			if err1 != nil {
// 				utils.Printf(static.LOG_FLAG, "[ERR] Error while getting timestamps")
// 			} else {
// 				utils.Printf(static.LOG_FLAG, fmt.Sprintf("[INFO] Order Delete All Orders Server Processing Time(T6-T1):  %v microseconds\n", (server_res.GW_Res_Out-server_res.GW_Req_In)/1000)
// 			}

// 			order_id = uint64(0)
// 			act_time = uint64(0)
// 			template_type = read_response
// 			// return 0, 0, read_response, nil
// 		}

// 		if read_response == HEARTBEAT {
// 			// return 0, 0, HEARTBEAT, nil
// 			order_id = uint64(0)
// 			act_time = uint64(0)
// 			template_type = HEARTBEAT
// 		}

// 		if read_response == REJECTED {
// 			// return 0, 0, REJECTED, nil
// 			order_id = uint64(0)
// 			act_time = uint64(0)
// 			template_type = REJECTED
// 		}

// 		if read_response == IMMEDIATE_EXECUTION {

// 			utils.Printf(static.LOG_FLAG, "Received IMMEDIATE_EXECUTION")

// 			_, err := DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
// 			if err != nil {
// 				utils.Printf(static.LOG_FLAG, "[ERR] Error in decryption for order response parser :", err)
// 			}

// 			// return 0, 0, IMMEDIATE_EXECUTION, nil
// 			order_id = uint64(0)
// 			act_time = uint64(0)
// 			template_type = IMMEDIATE_EXECUTION

// 			utils.Printf(static.LOG_FLAG, "IMMEDIATE_EXECUTION, Template type: ", template_type)

// 		}

// 		if read_response == ORDER_BOOK_EXECUTION {

// 			utils.Printf(static.LOG_FLAG, "Received ORDER_BOOK_EXECUTION")
// 			decytped_res, err := DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
// 			if err != nil {
// 				utils.Printf(static.LOG_FLAG, "[ERR] Error in decryption for order response parser :", err)
// 			}

// 			utils.Printf(static.LOG_FLAG, "BOok execution res: ", decytped_res)

// 			// return 0, 0, ORDER_BOOK_EXECUTION, nil
// 			order_id = uint64(0)
// 			act_time = uint64(0)
// 			template_type = ORDER_BOOK_EXECUTION

// 			utils.Printf(static.LOG_FLAG, "ORDER_BOOK_EXECUTION, Template type: ", template_type)

// 		}

// 	}

// 	utils.Printf(static.LOG_FLAG, "Response of checkbuffer: ", order_id, act_time, template_type)

// 	return order_id, act_time, template_type, nil

// }

// func (t *Trader) NewHandleOrderEntry() error {

// 	startTime := time.Now()
// 	totalTime := time.Now()
// 	packetCount := 0
// 	totalPacketCount := 0
// 	batchSize := THROTTLE_VALUE

// 	// Calculate initial parameters
// 	sendModCount := int(math.Floor(float64(MODIFY_PERCENT*batchSize) / 100))
// 	cancelOrders := int(math.Floor(float64(CANCEL_PERCENT*batchSize) / 100))
// 	orderEntry := batchSize - sendModCount - cancelOrders
// 	modPerOrder := int(math.Floor(float64(sendModCount) / float64(orderEntry)))
// 	sendMod := MODIFY_PERCENT != 0

// 	utils.Printf(static.LOG_FLAG, "Batch Size :", batchSize)
// 	utils.Printf(static.LOG_FLAG, "Total Order Entry Count :", orderEntry)
// 	utils.Printf(static.LOG_FLAG, "Total Mod Count :", sendModCount, "Mods/order count :", modPerOrder)
// 	utils.Printf(static.LOG_FLAG, "Total Cancel Count: ", cancelOrders)

// 	count := 0

// 	for data := range dataChan {

// 		count++
// 		utils.Printf(static.LOG_FLAG, "Packet No. (Iteration) in data channel :", count)

// 	outerLoop:
// 		for j := 0; j < orderEntry; j++ {

// 			utils.Printf(static.LOG_FLAG, "Outer for loop :", j, "Iteration order :", count)

// 			elapsedTime := time.Since(startTime)
// 			utils.Printf(static.LOG_FLAG, "Elapsed Time :", elapsedTime)
// 			if packetCount >= batchSize || elapsedTime >= time.Second {
// 				utils.Printf(static.LOG_FLAG, "Initial Case matched!")
// 				if packetCount >= batchSize && elapsedTime < time.Second {
// 					utils.Printf(static.LOG_FLAG, "In Sleep Case!!")
// 					utils.Printf(static.LOG_FLAG, fmt.Sprintf("Session id %d , sleeping for %v\n", t.session_id, 1000*time.Millisecond-elapsedTime)
// 					time.Sleep(time.Second - elapsedTime)
// 				}
// 				utils.Printf(static.LOG_FLAG, fmt.Sprintf("Session id %d sent %d packets.\n", t.session_id, packetCount)
// 				startTime = time.Now()
// 				packetCount = 0
// 				sendMod = true
// 			}

// 			if UNIQUE_CLIENT_IDENTIFIER {
// 				t.ClOrdID++
// 			}

// 			price_factor := 5
// 			price := data.Price

// 			mid_price := (data.Price + data.MaxPrice) / 2
// 			mid_price = ((mid_price + price_factor/2) / price_factor) * price_factor

// 			client_code := "CLIENT3"

// 			order_type := 1 // Buy
// 			if t.session_id%2 == 0 {
// 				order_type = 2 //sell
// 			}

// 			utils.Printf(static.LOG_FLAG, "OE => Price :", price, "Qty :", data.Qty, "Order Type :", order_type)

// 			orderRequest, err := SingleLegLeanOrder(t.MID, t.TID, t.msg_seq, data.InstrumentId, price, data.Qty, order_type, t.CTX, t.ClOrdID, client_code)
// 			if err != nil {
// 				return err
// 			}

// 			if err := t.sendDataToServer(orderRequest); err != nil {
// 				utils.Printf(static.LOG_FLAG, fmt.Sprintf("Error sending data with session id %d: %v\n", t.session_id, err)
// 			}

// 			orderID, actTime, template, err := t.OrderResponseListner(data.InstrumentId)
// 			if err != nil {
// 				return fmt.Errorf("response error for session id %d: %v\n", t.session_id, err)
// 			}
// 			t.msg_seq++
// 			packetCount++
// 			totalPacketCount++

// 			utils.Printf(static.LOG_FLAG, "Packet Count after LEAN ORDER :", packetCount)

// 			if template == LEAN_ORDER && sendMod {
// 				for i := 0; i < modPerOrder; i++ {
// 					utils.Printf(static.LOG_FLAG, "Mod order iteration count :", i)
// 					if UNIQUE_CLIENT_IDENTIFIER {
// 						t.OrigClOrdID = t.ClOrdID
// 						t.ClOrdID++
// 					}

// 					if order_type == 1 {
// 						price = price_factor*i + data.Price

// 						if price > mid_price {
// 							price = data.Price
// 						}
// 					} else if order_type == 2 {
// 						price = price_factor + i + mid_price

// 						if price > data.MaxPrice {
// 							price = mid_price + price_factor
// 						}
// 					}

// 					utils.Printf(static.LOG_FLAG, "OM => Price :", price, "Qty :", data.Qty, "Order Type :", order_type)

// 					orderID, actTime, template, err = t.sendOrderModify(data.InstrumentId, price, data.Qty, order_type, orderID, actTime)
// 					if err != nil {
// 						utils.Printf(static.LOG_FLAG, fmt.Sprintf("Modify error for session id %d: %v\n", t.session_id, err)
// 					}

// 					packetCount++
// 					t.msg_seq++
// 					totalPacketCount++

// 					if template != MODIFY_LEAN_ORDER {
// 						utils.Printf(static.LOG_FLAG, "Breaking from Mod loop :", template)
// 						break
// 					}

// 					utils.Printf(static.LOG_FLAG, "MODIFY ORDER iteration count success :", i, "Packet count :", packetCount)

// 					elapsedTime = time.Since(startTime)
// 					utils.Printf(static.LOG_FLAG, "Elapsed Time in Modify :", elapsedTime)
// 					if packetCount >= batchSize || elapsedTime >= time.Second {
// 						utils.Printf(static.LOG_FLAG, "Initial Case matched in modify!")
// 						if packetCount >= batchSize && elapsedTime < time.Second {
// 							utils.Printf(static.LOG_FLAG, "In modify Sleep Case!!")
// 							utils.Printf(static.LOG_FLAG, fmt.Sprintf("Session id %d , sleeping for %v\n", t.session_id, 1000*time.Millisecond-elapsedTime)
// 							time.Sleep(time.Second - elapsedTime)
// 						}
// 						utils.Printf(static.LOG_FLAG, fmt.Sprintf("Session id %d sent %d packets.\n", t.session_id, packetCount)
// 						startTime = time.Now()
// 						packetCount = 0
// 						sendMod = true
// 						utils.Printf(static.LOG_FLAG, "Exiting for loop!")
// 						break outerLoop
// 					}
// 				}
// 			}

// 			if template == LEAN_ORDER || template == MODIFY_LEAN_ORDER && CANCEL_PERCENT != 0 {

// 				utils.Printf(static.LOG_FLAG, "In cancel order Case!!")
// 				if UNIQUE_CLIENT_IDENTIFIER {
// 					t.OrigClOrdID = t.ClOrdID
// 					t.ClOrdID++
// 				}

// 				if err := t.sendOrderCancel(data.InstrumentId, data.ProductId, orderID); err != nil {
// 					utils.Printf(static.LOG_FLAG, fmt.Sprintf("Cancel error for session id %d: %v\n", t.session_id, err)
// 				}

// 				t.msg_seq++
// 				packetCount++
// 				totalPacketCount++

// 				utils.Printf(static.LOG_FLAG, "Packet count after CANCEL ORDER success :", packetCount)

// 				orderID = 0
// 				actTime = 0
// 				template = 0
// 			}

// 		}

// 		if len(dataChan) == 0 {
// 			utils.Printf(static.LOG_FLAG, fmt.Sprintf("Session id %d took total time: %v and sent packets: %v\n",
// 				t.session_id, time.Since(totalTime), totalPacketCount)
// 			return nil
// 		}

// 	}

// 	return nil
// }

// func (t *Trader) OrderResponseListner(inst_id int) (uint64, uint64, uint16, error) {

// 	for {
// 		reader := bufio.NewReader(t.conn)
// 		buffer := make([]byte, 16*1024*1024)

// 		n, err := reader.Read(buffer)

// 		if err != nil {
// 			if err == io.EOF {
// 				utils.Printf(static.LOG_FLAG, fmt.Sprintf("[INFO] Connection closed for session id %d: %v\n", t.session_id, err)
// 				return 0, 0, 0, err
// 			}
// 			utils.Printf(static.LOG_FLAG, fmt.Sprintf("[ERR] Error reading responses for session id %d: %v\n", t.session_id, err)
// 			return 0, 0, 0, nil
// 		}

// 		responseBuf := buffer[:n]

// 		responseLength := len(responseBuf)

// 		if responseLength == 0 {
// 			utils.Printf(static.LOG_FLAG, "Response len is 0.")
// 			return 0, 0, 0, nil
// 		}

// 		utils.Printf(static.LOG_FLAG, "Len of Response: ", responseLength)

// 		var responseArr [][]byte
// 		count := 0
// 		for len(responseBuf) > 0 {
// 			count++
// 			if len(responseBuf) < 4 {
// 				utils.Printf(static.LOG_FLAG, "[ERR] Response buffer too short to read size")
// 				return 0, 0, 0, fmt.Errorf("response buffer too short")
// 			}
// 			size := int(binary.LittleEndian.Uint32(responseBuf[:4]))
// 			utils.Printf(static.LOG_FLAG, "size: ", size)
// 			if len(responseBuf) < size {
// 				utils.Printf(static.LOG_FLAG, "Response Buf: ", responseBuf)
// 				utils.Printf(static.LOG_FLAG, "[ERR] Response buffer too short for the indicated size")
// 				return 0, 0, 0, fmt.Errorf("response buffer too short for indicated size")
// 			}

// 			responseArr = append(responseArr, responseBuf[:size])

// 			responseBuf = responseBuf[size:]
// 			utils.Printf(static.LOG_FLAG, "COunt: ", count)
// 		}

// 		utils.Printf(static.LOG_FLAG, "Len of response buf: ", len(responseArr))

// 		for _, response := range responseArr {

// 			if len(response) == 0 {
// 				break
// 			}

// 			val, err := GetTemplateID(response)
// 			if err != nil {
// 				utils.Printf(static.LOG_FLAG, "[ERR] Error in template id parser :", err)
// 				// return
// 			}

// 			utils.Printf(static.LOG_FLAG, fmt.Sprintf("For session Id %v, TEMPLATE ID: %v\n", t.session_id, val)
// 			read_response, err := t.readReceivedResponse(val, inst_id, response)
// 			if err != nil {
// 				utils.Printf(static.LOG_FLAG, "[ERR] Error in reading received response :", err)
// 			}

// 			switch {
// 			case read_response == ORDER_CONFIRMATION:
// 				utils.Printf(static.LOG_FLAG, "[PASS] Order Confirmation Received for session id:", t.session_id)

// 			case read_response != ORDER_CONFIRMATION:

// 				if read_response == LEAN_ORDER {
// 					order_res, err := DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
// 					if err != nil {
// 						utils.Printf(static.LOG_FLAG, "[ERR] Error in decryption for order response parser :", err)
// 						// return nil
// 						// return
// 					}

// 					server_res, err1 := GetTimestamps(order_res)
// 					if err1 != nil {
// 						utils.Printf(static.LOG_FLAG, "[ERR] Error while getting timestamps")
// 					} else {
// 						utils.Printf(static.LOG_FLAG, fmt.Sprintf("[INFO] Order Entry Server Processing Time(T6-T1):  %v microseconds\n", (server_res.GW_Res_Out-server_res.GW_Req_In)/1000)
// 					}
// 					// utils.Printf(static.LOG_FLAG, "[INFO] TimeStamps: ", timeStamps)

// 					order_id, act_time, err := SingleLegLeanOrderResponse(order_res, t.CTX)
// 					if err != nil {
// 						utils.Printf(static.LOG_FLAG, "[ERR] Error in lean order response parser :", err)
// 						// return nil
// 						// return
// 					}

// 					return order_id, act_time, LEAN_ORDER, nil
// 				}

// 				if read_response == CANCEL_LEAN_ORDER {

// 					order_res, err := DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
// 					if err != nil {
// 						utils.Printf(static.LOG_FLAG, "[ERR] Error in decryption for order response parser :", err)
// 					}

// 					server_res, err1 := GetTimestamps(order_res)
// 					if err1 != nil {
// 						utils.Printf(static.LOG_FLAG, "[ERR] Error while getting timestamps")
// 					} else {
// 						utils.Printf(static.LOG_FLAG, fmt.Sprintf("[INFO] Order Cancel Server Processing Time(T6-T1):  %v microseconds\n", (server_res.GW_Res_Out-server_res.GW_Req_In)/1000)
// 					}

// 					return 0, 0, CANCEL_LEAN_ORDER, nil

// 				}

// 				if read_response == MODIFY_LEAN_ORDER {

// 					order_res, err := DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
// 					if err != nil {
// 						utils.Printf(static.LOG_FLAG, "[ERR] Error in decryption for order response parser :", err)
// 						// return nil
// 						// return
// 					}
// 					server_res, err1 := GetTimestamps(order_res)
// 					if err1 != nil {
// 						utils.Printf(static.LOG_FLAG, "[ERR] Error while getting timestamps")
// 					} else {
// 						utils.Printf(static.LOG_FLAG, fmt.Sprintf("[INFO] Order Mod Server Processing Time(T6-T1):  %v microseconds\n", (server_res.GW_Res_Out-server_res.GW_Req_In)/1000)
// 					}
// 					// utils.Printf(static.LOG_FLAG, "[INFO] TimeStamps: ", timeStamps)

// 					order_id, act_time, err := SingleLegLeanOrderModifiedResponse(order_res, t.CTX)
// 					if err != nil {
// 						utils.Printf(static.LOG_FLAG, "[ERR] Error in lean order response parser :", err)
// 						// return nil
// 						// return
// 					}

// 					return order_id, act_time, MODIFY_LEAN_ORDER, nil

// 				}

// 				if read_response == HEARTBEAT {
// 					return 0, 0, HEARTBEAT, nil
// 				}

// 				if read_response == REJECTED {
// 					return 0, 0, REJECTED, nil
// 				}

// 				if read_response == IMMEDIATE_EXECUTION {

// 					_, err := DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
// 					if err != nil {
// 						utils.Printf(static.LOG_FLAG, "[ERR] Error in decryption for order response parser :", err)
// 					}

// 					return 0, 0, IMMEDIATE_EXECUTION, nil

// 				}

// 				if read_response == ORDER_BOOK_EXECUTION {

// 					decytped_res, err := DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
// 					if err != nil {
// 						utils.Printf(static.LOG_FLAG, "[ERR] Error in decryption for order response parser :", err)
// 					}

// 					utils.Printf(static.LOG_FLAG, "BOok execution res: ", decytped_res)

// 					return 0, 0, ORDER_BOOK_EXECUTION, nil

// 				}

// 				if read_response == TRADING_SESSION_EVENT {

// 					decytped_res, err := DecryptData(response, t.CGW_Res.SecKey, t.CGW_Res.IV, t.CTX)
// 					if err != nil {
// 						utils.Printf(static.LOG_FLAG, "[ERR] Error in decryption for order response parser :", err)
// 					}

// 					utils.Printf(static.LOG_FLAG, "TRADING_SESSION_EVENT: ", decytped_res)

// 					return 0, 0, TRADING_SESSION_EVENT, nil

// 				}

// 			default:
// 				utils.Printf(static.LOG_FLAG, "Other Template ID: ", read_response)

// 			}

// 		}

// 		// continue

// 	}

// }

// func CreateNewChannelInstance()  {
// 	mu.Lock()
// 	defer mu.Unlock()

// 	defer func() { // Recover in case of panic
// 		if r := recover(); r != nil {
// 			utils.Printf(static.LOG_FLAG, "Recovered from panic:", r)
// 		}
// 	}()

// 	SWITCH_CH =  true

// 	oldDataChan := dataChan
// 	oldRecordsChan := recordsChan

// 	newDataChan := make(chan DataPacket, DATA_CHANNEL_LENGTH)
// 	newRecordsChan := make(chan []DataPacket, DATA_CHANNEL_LENGTH)

// 	dataChan = newDataChan
// 	recordsChan = newRecordsChan

// 	utils.Printf(static.LOG_FLAG, "Instances set before delete")

// 	err1 := safeClose(oldDataChan)
// 	err2 := safeClose(oldRecordsChan)

// 	if err1 != nil || err2 != nil {
// 		utils.Printf(static.LOG_FLAG, "error closing old channels")
// 	}

// 	SWITCH_CH = false
// 	utils.Printf(static.LOG_FLAG, "Switching channel instances!")

// 	// return nil
// }

// func safeClose(ch interface{}) error {
// 	defer func() {
// 		if r := recover(); r != nil {
// 			utils.Printf(static.LOG_FLAG, "Recovered in safeClose:", r)
// 		}
// 	}()

// 	switch c := ch.(type) {
// 	case chan DataPacket:
// 		for range c {
// 		}
// 		close(c)
// 	case chan []DataPacket:
// 		for range c {
// 		}
// 		close(c)
// 	default:
// 		utils.Printf(static.LOG_FLAG, "invalid channel type")
// 		return nil
// 	}

// 	return nil
// }
