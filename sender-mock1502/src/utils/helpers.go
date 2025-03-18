package utils

import (
	"bufio"
	"bytes"
	"fmt"
	"market_replay/src/encryption/aes"
	"market_replay/src/static"
	"math"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"
)

func ReadConfig(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	configs := &static.ProcessConfig{}
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		entries := strings.Split(line, ",")
		if len(entries) < 1 {
			Printf(static.LOG_FLAG, "[ERR] Invalid line:"+line)
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
		unique_identifier, err8 := strconv.Atoi(entries[8])
		sessionIdSetNum := entries[9]
		tradeThrottle, err9 := strconv.Atoi(entries[10])
		orderPumpingType, err10 := strconv.Atoi(entries[11])

		static.AssignSessionMap()

		if err0 != nil || err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil || err6 != nil || err7 != nil || err8 != nil || err9 != nil || err10 != nil {
			Printf(static.LOG_FLAG, "[ERR] Error parsing limits: ")
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
		configs.UniqueIdentifier = unique_identifier
		configs.SessionSetValue = sessionIdSetNum
		configs.TradeThrottle = tradeThrottle
		configs.OrderPumpingType = orderPumpingType

		static.TRADER_COUNT = configs.TraderCount
		static.TARGET_ENV = configs.TargetEnv
		static.TOTAL_ORDER_COUNT = configs.TraderCount * configs.ThrottleLimit * configs.RunTimeLimit
		static.THROTTLE_VALUE = configs.ThrottleLimit
		static.FILE_TYPE = configs.FileType
		static.MODIFY_PERCENT = configs.OrderModPercent
		static.CANCEL_PERCENT = configs.OrderCancelPercent
		static.TRADERCOUNT = configs.TraderCount
		static.HEARTBEAT_VALUE = configs.HeartBeat
		static.DURATION = configs.RunTimeLimit
		static.SESSIONSETVALUE = configs.SessionSetValue
		static.TRADE_THROTTLE = configs.TradeThrottle
		static.ORDER_PUMPING_TYPE = configs.OrderPumpingType

		if configs.UniqueIdentifier == 1 {
			static.UNIQUE_CLIENT_IDENTIFIER = true
		} else {
			static.UNIQUE_CLIENT_IDENTIFIER = false
		}

		if static.ORDER_PUMPING_TYPE == static.ASYNC_ORDER_PUMPING_OEOMOC_TYPE {
			static.UNIQUE_CLIENT_IDENTIFIER = true
		}

		static.DATA_CHANNEL_LENGTH = 2 * static.THROTTLE_VALUE * static.TRADER_COUNT

		Printf(static.LOG_FLAG, fmt.Sprintf("[INFO] Total No. of Traders :%v", static.TRADER_COUNT))
		Printf(static.LOG_FLAG, fmt.Sprintf("[INFO] Total No. of Orders :%v", static.TOTAL_ORDER_COUNT))
		Printf(static.LOG_FLAG, fmt.Sprintf("[INFO] Throttle Value:%v", static.THROTTLE_VALUE))
		Printf(static.LOG_FLAG, fmt.Sprintf("[INFO] Run time (seconds) :%v", configs.RunTimeLimit))
		Printf(static.LOG_FLAG, fmt.Sprintf("[INFO] Modify Order Percentage :%v", static.MODIFY_PERCENT))
		Printf(static.LOG_FLAG, fmt.Sprintf("[INFO] Cancel Order Percentage :%v", static.CANCEL_PERCENT))
		Printf(static.LOG_FLAG, fmt.Sprintf("[INFO] HeartBeat value :%v", static.HEARTBEAT_VALUE))
		Printf(static.LOG_FLAG, fmt.Sprintf("[INFO] Data Channel Length :%v", static.DATA_CHANNEL_LENGTH))
		Printf(static.LOG_FLAG, fmt.Sprintf("[INFO] static.DURATION :%v", static.DURATION))
		Printf(static.LOG_FLAG, fmt.Sprintf("[INFO] Trade Throttle :%v", static.TRADE_THROTTLE))
		Printf(static.LOG_FLAG, fmt.Sprintf("[INFO] Order Pumping Type :%v", static.ORDER_PUMPING_TYPE))
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func ReadRecords(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	var records []static.Record
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		entries := strings.Split(line, ",")

		if len(entries) < 9 {
			Printf(static.LOG_FLAG, "[ERR] Invalid line:"+line)
			continue
		}

		instrumentID, err0 := strconv.Atoi(entries[0])
		lowerLimit, err1 := strconv.Atoi(entries[5])
		upperLimit, err2 := strconv.Atoi(entries[6])
		bidIntrvl, err3 := strconv.Atoi(entries[7])
		maxTrdQty, err4 := strconv.Atoi(entries[8])
		minLot, err5 := strconv.Atoi(entries[9])
		product_id, err6 := strconv.Atoi(entries[10])
		partition_id, err7 := strconv.Atoi(entries[11])

		Printf(static.LOG_FLAG, fmt.Sprintf("[INFO] Instrument ID: %v | Bid Interval: %v | MaxTrdQty: %v |  MinLot value: %v | Product ID: %v | Partition ID : %v\n", instrumentID, bidIntrvl, maxTrdQty, minLot, product_id, partition_id))

		if err0 != nil || err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil || err6 != nil || err7 != nil {
			// Printf(static.LOG_FLAG, "[ERR] Error parsing limits:%v", err0, err1, err2, err3, err4, err5, err6, err7)
			continue
		}

		records = append(records, static.Record{
			InstrumentID: instrumentID,
			LowerLimit:   lowerLimit,
			UpperLimit:   upperLimit,
			BidInterval:  bidIntrvl,
			MaxTrdQty:    maxTrdQty,
			MinLot:       minLot,
			Product_ID:   product_id,
		})
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	static.RECORDS = records

	return nil

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

func GetTimestamps(data []byte) (*static.TimeStamps, error) {

	res := &static.TimeStamps{}

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

	return res, nil
}

func SetOrderPumpingParameters() error {

	static.SEND_MOD_COUNT = int(math.Floor(float64(static.MODIFY_PERCENT*static.THROTTLE_VALUE) / 100))
	static.CANCEL_ORDER_COUNT = int(math.Floor(float64(static.CANCEL_PERCENT*static.THROTTLE_VALUE) / 100))
	static.ORDER_ENTRY_COUNT = static.THROTTLE_VALUE - static.SEND_MOD_COUNT - static.CANCEL_ORDER_COUNT
	static.MOD_PER_ORDER = int(math.Floor(float64(static.SEND_MOD_COUNT) / float64(static.ORDER_ENTRY_COUNT)))
	static.SEND_MOD = static.MODIFY_PERCENT != 0

	static.DATA_CHANNEL_LENGTH = 2 * static.THROTTLE_VALUE * static.TRADER_COUNT

	Printf(static.LOG_FLAG, fmt.Sprintf("Batch Size : %v", static.THROTTLE_VALUE))
	Printf(static.LOG_FLAG, fmt.Sprintf("Total Order Entry Count : %v", static.ORDER_ENTRY_COUNT))
	Printf(static.LOG_FLAG, fmt.Sprintf("Total Mod Count : %v Mods/order count :%v", static.SEND_MOD_COUNT, static.MOD_PER_ORDER))
	Printf(static.LOG_FLAG, fmt.Sprintf("Total Cancel Count: %v ", static.CANCEL_ORDER_COUNT))
	Printf(static.LOG_FLAG, fmt.Sprintf("Data Channel Length : %v", static.DATA_CHANNEL_LENGTH))

	return nil
}

func GenerateRandomPrice(start, end, interval int) int {
	start = ((start + interval - 1) / interval) * interval
	end = (end / interval) * interval

	if start > end {
		return start
	}

	numPossibilities := ((end - start) / interval) + 1
	randomOffset := rand.Intn(numPossibilities)

	return start + (randomOffset * interval)
}

func GenerateRandomQty(min, max int) int {
	if min == 0 {
		return 0
	}

	numMultiples := max/min - 1

	if numMultiples <= 0 {
		return min
	}

	multiplier := rand.Intn(numMultiples) + 1

	return min * multiplier
}

func GetPriceRange(min, max, factor int) (int, int, int, int) {

	mid_price := (min + max) / 2
	mid_price = ((mid_price + factor/2) / factor) * factor
	buy_min_price := min
	buy_max_price := mid_price - (factor * 55)
	sell_min_price := mid_price + (factor * 55)
	sell_max_price := max

	return buy_min_price, buy_max_price, sell_min_price, sell_max_price
}

func GenerateSquareWave(min, max, interval int) []int {
	pattern := make([]int, 2*interval)

	for i := 0; i < interval; i++ {
		pattern[i] = min
		pattern[i+interval] = max
	}

	Printf(static.LOG_FLAG, fmt.Sprintf("Square Wave Pattern: %v", pattern))
	result := CalculatePatternThrottle(pattern)

	return result
}

func GenerateSawtoothWave(min, max, interval int) []int {
	pattern := make([]int, interval)

	step := (max - min) / (interval - 1)

	for i := 0; i < interval; i++ {
		pattern[i] = min + i*step
	}

	Printf(static.LOG_FLAG, fmt.Sprintf("Sawtooth Wave Pattern: %v", pattern))
	result := CalculatePatternThrottle(pattern)

	return result
}

func GenerateSineWave(min, max int) []int {
	pattern := make([]int, 8)

	mid := (min + max) / 2
	step := (max - mid) / 2
	values := []int{mid, mid + step, max, mid + step, mid, mid - step, min, mid - step}

	for i := 0; i < len(pattern); i++ {
		pattern[i] = values[i]
	}

	Printf(static.LOG_FLAG, fmt.Sprintf("Sine Wave Pattern: %v", pattern))
	result := CalculatePatternThrottle(pattern)

	return result
}

func GenerateStepWave(min, max, step, interval int) []int {

	var pattern []int

	for val := min; val <= max; val += step {
		for i := 0; i < interval; i++ {
			pattern = append(pattern, val)
		}
	}

	for val := max - step; val > min; val -= step {
		for i := 0; i < interval; i++ {
			pattern = append(pattern, val)
		}
	}

	Printf(static.LOG_FLAG, fmt.Sprintf("Step Wave Pattern: %v", pattern))
	result := CalculatePatternThrottle(pattern)

	return result
}

func GenerateTriangleWave(min, max, interval int) []int {
	pattern := make([]int, 2*interval-1)

	step := (max - min) / (interval - 1)

	for i := 0; i < interval; i++ {
		pattern[i] = min + i*step
	}

	for i := 1; i < interval; i++ {
		pattern[interval+i-1] = max - i*step
	}

	Printf(static.LOG_FLAG, fmt.Sprintf("Triangle Wave Pattern: %v", pattern[:len(pattern)-1]))
	result := CalculatePatternThrottle(pattern)

	return result
}

func CalculatePatternThrottle(arr []int) []int {

	result := make([]int, len(arr))
	for i, val := range arr {
		result[i] = val / static.TRADER_COUNT
	}

	return result
}

func PatternThrottle() {
	for {
		if static.PATTERN_GENERATOR {
			for i := 0; i < static.TOTAL_ORDER_COUNT; i++ {
				var value int
				if len(static.PATTERN) > 0 {
					value = static.PATTERN[i%len(static.PATTERN)]
				} else {
					switch static.PATTERN_TYPE {
					case static.RANDOM_WAVE:
						value = (rand.Intn(static.RANDOM_WAVE_DETAILS.Max-static.RANDOM_WAVE_DETAILS.Min+1) + static.RANDOM_WAVE_DETAILS.Min) / static.TRADER_COUNT
					case static.PEAK_GENERATOR:
						value = (rand.Intn(static.PEAK_GENERATOR_DETAILS.Max-static.PEAK_GENERATOR_DETAILS.Min+1) + static.PEAK_GENERATOR_DETAILS.Min) / static.TRADER_COUNT
						Printf(static.LOG_FLAG, fmt.Sprintf("PEAK GENERATOR THROTTLE VALUE: %v", value))
					}
				}

				if !static.PATTERN_GENERATOR {
					break
				}
				static.ConnThrottleChan <- value
			}
		}
		time.Sleep(1000 * time.Millisecond)
	}
}

func ChangeThrottle() {
	ticker := time.NewTicker(1 * time.Second)

loop:
	for {
		// if static.PATTERN_GENERATOR {
		// 	select {
		// 	case <-ticker.C:
		// 		select {
		// 		case data := <-static.ConnThrottleChan:
		// 			static.THROTTLE_VALUE = data
		// 			Printf(static.LOG_FLAG, "Throttle Value in pattern generator :%v", static.THROTTLE_VALUE)

		// 			err := setOrderPumpingParameters()
		// 			if err != nil {
		// 				Printf(static.LOG_FLAG, "Error in setting Order Pumping Parameters.")
		// 			}
		// 		default:
		// 			Printf(static.LOG_FLAG, "No data available in static.ConnThrottleChan")
		// 		}

		// 		if !static.PATTERN_GENERATOR {
		// 			Printf(static.LOG_FLAG, "Breaking out of Pattern Generator")
		// 			break loop
		// 		}
		// 	}
		// }
		// time.Sleep(1000 * time.Millisecond)

		if static.PATTERN_GENERATOR {

			<-ticker.C

			data, ok := <-static.ConnThrottleChan
			if ok {
				static.THROTTLE_VALUE = data
				Printf(static.LOG_FLAG, fmt.Sprintf("Throttle Value in pattern generator: %v", static.THROTTLE_VALUE))

				if err := SetOrderPumpingParameters(); err != nil {
					Printf(static.LOG_FLAG, "Error in setting Order Pumping Parameters.")

				}
			} else {
				Printf(static.LOG_FLAG, "No data available in static.ConnThrottleChan")
			}

			if !static.PATTERN_GENERATOR {
				Printf(static.LOG_FLAG, "Breaking out of Pattern Generator")
				break loop
			}
		}
		time.Sleep(1 * time.Second)
	}
}

func ApplyPriceRangeChange(data static.PriceRangeChangeDetails) {

	for key, value := range data.Instruments {
		value.Min_interval = getFactor(value.Start_min_price, value.End_min_price, data.Interval)
		value.Max_interval = getFactor(value.Start_max_price, value.End_max_price, data.Interval)
		data.Instruments[key] = value

		if !static.SLIDING_PRICE {
			break
		}
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	Printf(static.LOG_FLAG, "Sliding price range update started!")

	for second := 0; second < data.Interval; second++ {
		if !static.SLIDING_PRICE {
			break
		}

		<-ticker.C

		static.Mu.Lock()
		for i := range static.RECORDS {
			if instrument, exists := data.Instruments[strconv.Itoa(static.RECORDS[i].InstrumentID)]; exists {
				rec := static.RECORDS[i]
				rec.LowerLimit = instrument.Start_min_price + second*instrument.Min_interval
				rec.UpperLimit = instrument.Start_max_price + second*instrument.Max_interval

				static.RECORDS[i] = rec
			}
		}
		static.Mu.Unlock()
	}
}

func getFactor(start, end, interval int) int {
	if interval <= 1 {
		return start
	}

	return (end - start) / (interval - 1)
}

func EmptyChannel() {
	count := 0
	static.SWITCH_CH = true

	if static.TRADE {
		for range static.DataChan {
			count++
			if count >= static.DATA_CHANNEL_LENGTH {
				break
			}
		}

	} else {
		for range static.RecordsChan {
			count++
			if count >= static.DATA_CHANNEL_LENGTH {
				break
			}
		}
	}

	Printf(static.LOG_FLAG, fmt.Sprintf("Total items removed from channel : %v", count))
	static.SWITCH_CH = false
}

func ChangeDelay() {

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	if static.PATTERN_GENERATOR && static.PATTERN_TYPE == static.PEAK_GENERATOR {
		for {
			<-ticker.C
			static.DELAY_TIME = (rand.Intn(static.PEAK_GENERATOR_DETAILS.DelayMax-static.PEAK_GENERATOR_DETAILS.DelayMin+1) + static.PEAK_GENERATOR_DETAILS.DelayMin)
			Printf(static.LOG_FLAG, fmt.Sprintf("DELAY TIME : %v", static.DELAY_TIME))
			if !static.PATTERN_GENERATOR || static.PATTERN_TYPE != static.PEAK_GENERATOR {
				break
			}
		}
	}
}

func CheckRejectedSessions() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		<-ticker.C
		Printf(static.LOG_FLAG, fmt.Sprintf("Total no. of sessions facing rejection : %v", len(static.ORDERS_REJECTION)))
	}
}

func CheckSessionLogon(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range 100 {
		<-ticker.C
		length := len(static.SessionConn)
		fmt.Println("[INFO] Session Logon: ", length)
		if length >= 100 { // after all sessions are connected, print session count every 2 minutes.
			// Change interval to 2 minutes
			ticker.Stop()
			CheckSessionLogon(2 * time.Minute)
			return
		}
	}

}

func VerifyRejectedMsg(msg string, sessionId string, instId string) {
	isValid := func(string) bool {
		for _, r := range msg {
			if !unicode.IsPrint(r) || unicode.IsSymbol(r) || unicode.IsPunct(r) {
				return false
			}
		}
		return true
	}(msg)

	if isValid {
		fmt.Printf("[ERR] ORDER REJECTED|%v|instId:%v|msg:%v\n", sessionId, instId, msg)
	}
}
