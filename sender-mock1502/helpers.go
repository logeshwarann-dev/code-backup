package main

import (
	"bufio"
	"bytes"
	"fmt"
	"market_replay/aes"
	"math"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
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
			fmt.Println("[ERR] Invalid line:", line)
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

		AssignSessionMap()

		if err0 != nil || err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil || err6 != nil || err7 != nil || err8 != nil || err9 != nil {
			fmt.Println("[ERR] Error parsing limits: ")
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

		TRADER_COUNT = configs.TraderCount
		TARGET_ENV = configs.TargetEnv
		TOTAL_ORDER_COUNT = configs.TraderCount * configs.ThrottleLimit * configs.RunTimeLimit
		THROTTLE_VALUE = configs.ThrottleLimit
		FILE_TYPE = configs.FileType
		MODIFY_PERCENT = configs.OrderModPercent
		CANCEL_PERCENT = configs.OrderCancelPercent
		TRADERCOUNT = configs.TraderCount
		HEARTBEAT_VALUE = configs.HeartBeat
		DURATION = configs.RunTimeLimit
		SESSIONSETVALUE = configs.SessionSetValue
		TRADE_THROTTLE = configs.TradeThrottle

		if configs.UniqueIdentifier == 1 {
			UNIQUE_CLIENT_IDENTIFIER = true
		} else {
			UNIQUE_CLIENT_IDENTIFIER = false
		}

		DATA_CHANNEL_LENGTH = 2 * THROTTLE_VALUE * TRADER_COUNT

		fmt.Println("[INFO] Total No. of Traders :", TRADER_COUNT)
		fmt.Println("[INFO] Total No. of Orders :", TOTAL_ORDER_COUNT)
		fmt.Println("[INFO] Throttle Value:", THROTTLE_VALUE)
		fmt.Println("[INFO] Run time (seconds) :", configs.RunTimeLimit)
		fmt.Println("[INFO] Modify Order Percentage :", MODIFY_PERCENT)
		fmt.Println("[INFO] Cancel Order Percentage :", CANCEL_PERCENT)
		fmt.Println("[INFO] HeartBeat value :", HEARTBEAT_VALUE)
		fmt.Println("[INFO] Data Channel Length :", DATA_CHANNEL_LENGTH)
		fmt.Println("[INFO] Duration :", DURATION)
		fmt.Println("[INFO] Trade Throttle :", TRADE_THROTTLE)
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

	var records []Record
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		entries := strings.Split(line, ",")

		if len(entries) < 9 {
			fmt.Println("[ERR] Invalid line:", line)
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

		fmt.Printf("[INFO] Instrument ID: %v | Bid Interval: %v | MaxTrdQty: %v |  MinLot value: %v | Product ID: %v | Partition ID : %v\n", instrumentID, bidIntrvl, maxTrdQty, minLot, product_id, partition_id)

		if err0 != nil || err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil || err6 != nil || err7 != nil {
			fmt.Println("[ERR] Error parsing limits:", err0, err1, err2, err3, err4, err5, err6, err7)
			continue
		}

		records = append(records, Record{
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

	RECORDS = records

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

	return res, nil
}

// func convertTimestamp(timeValue uint64) string {
// 	epoch := time.Unix(0, int64(timeValue))
// 	return epoch.Format(EPOCH_FORMAT)

// }

func setOrderPumpingParameters() error {

	SEND_MOD_COUNT = int(math.Floor(float64(MODIFY_PERCENT*THROTTLE_VALUE) / 100))
	CANCEL_ORDER_COUNT = int(math.Floor(float64(CANCEL_PERCENT*THROTTLE_VALUE) / 100))
	ORDER_ENTRY_COUNT = THROTTLE_VALUE - SEND_MOD_COUNT - CANCEL_ORDER_COUNT
	MOD_PER_ORDER = int(math.Floor(float64(SEND_MOD_COUNT) / float64(ORDER_ENTRY_COUNT)))
	SEND_MOD = MODIFY_PERCENT != 0

	DATA_CHANNEL_LENGTH = 2 * THROTTLE_VALUE * TRADER_COUNT

	fmt.Println("Batch Size :", THROTTLE_VALUE)
	fmt.Println("Total Order Entry Count :", ORDER_ENTRY_COUNT)
	fmt.Println("Total Mod Count :", SEND_MOD_COUNT, "Mods/order count :", MOD_PER_ORDER)
	fmt.Println("Total Cancel Count: ", CANCEL_ORDER_COUNT)
	fmt.Println("Data Channel Length :", DATA_CHANNEL_LENGTH)

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

	fmt.Println("Square Wave Pattern:", pattern)
	result := CalculatePatternThrottle(pattern)

	return result
}

func GenerateSawtoothWave(min, max, interval int) []int {
	pattern := make([]int, interval)

	step := (max - min) / (interval - 1)

	for i := 0; i < interval; i++ {
		pattern[i] = min + i*step
	}

	fmt.Println("Sawtooth Wave Pattern:", pattern)
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

	fmt.Println("Sine Wave Pattern:", pattern)
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

	fmt.Println("Step Wave Pattern:", pattern)
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

	fmt.Println("Triangle Wave Pattern:", pattern[:len(pattern)-1])
	result := CalculatePatternThrottle(pattern)

	return result
}

func CalculatePatternThrottle(arr []int) []int {

	result := make([]int, len(arr))
	for i, val := range arr {
		result[i] = val / TRADER_COUNT
	}

	return result
}

func PatternThrottle() {
	for {
		if PATTERN_GENERATOR {
			for i := 0; i < TOTAL_ORDER_COUNT; i++ {
				var value int
				if len(PATTERN) > 0 {
					value = PATTERN[i%len(PATTERN)]
				} else {
					switch PATTERN_TYPE {
					case RANDOM_WAVE:
						value = (rand.Intn(RANDOM_WAVE_DETAILS.Max-RANDOM_WAVE_DETAILS.Min+1) + RANDOM_WAVE_DETAILS.Min) / TRADER_COUNT
					case PEAK_GENERATOR:
						value = (rand.Intn(PEAK_GENERATOR_DETAILS.Max-PEAK_GENERATOR_DETAILS.Min+1) + PEAK_GENERATOR_DETAILS.Min) / TRADER_COUNT
						fmt.Println("In Pattern Throttle(), PEAK_GENERATOR_DETAILS: ", PEAK_GENERATOR_DETAILS, "| TRADER_COUNT: ", TRADER_COUNT, "| PEAK GENERATOR THROTTLE VALUE:", value)
					}
				}

				if !PATTERN_GENERATOR {
					break
				}
				connThrottleChan <- value
			}
		}
		time.Sleep(1000 * time.Millisecond)
	}
}

func ChangeThrottle() {
	ticker := time.NewTicker(1 * time.Second)

loop:
	for {
		// if PATTERN_GENERATOR {
		// 	select {
		// 	case <-ticker.C:
		// 		select {
		// 		case data := <-connThrottleChan:
		// 			THROTTLE_VALUE = data
		// 			fmt.Println("Throttle Value in pattern generator :", THROTTLE_VALUE)

		// 			err := setOrderPumpingParameters()
		// 			if err != nil {
		// 				fmt.Println("Error in setting Order Pumping Parameters.")
		// 			}
		// 		default:
		// 			fmt.Println("No data available in connThrottleChan")
		// 		}

		// 		if !PATTERN_GENERATOR {
		// 			fmt.Println("Breaking out of Pattern Generator")
		// 			break loop
		// 		}
		// 	}
		// }
		// time.Sleep(1000 * time.Millisecond)

		if PATTERN_GENERATOR {

			<-ticker.C

			data, ok := <-connThrottleChan
			if ok {
				THROTTLE_VALUE = data
				fmt.Println("Throttle Value in pattern generator :", THROTTLE_VALUE)

				if err := setOrderPumpingParameters(); err != nil {
					fmt.Println("Error in setting Order Pumping Parameters.")

				}
			} else {
				fmt.Println("No data available in connThrottleChan")
			}

			if !PATTERN_GENERATOR {
				fmt.Println("Breaking out of Pattern Generator")
				break loop
			}
		}
		time.Sleep(1 * time.Second)
	}
}

func ApplyPriceRangeChange(data PriceRangeChangeDetails) {

	for key, value := range data.Instruments {
		value.Min_interval = getFactor(value.Start_min_price, value.End_min_price, data.Interval)
		value.Max_interval = getFactor(value.Start_max_price, value.End_max_price, data.Interval)
		data.Instruments[key] = value

		if !SLIDING_PRICE {
			break
		}
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	fmt.Println("Sliding price range update started!")

	for second := 0; second < data.Interval; second++ {
		if !SLIDING_PRICE {
			break
		}

		<-ticker.C

		mu.Lock()
		for i := range RECORDS {
			if instrument, exists := data.Instruments[strconv.Itoa(RECORDS[i].InstrumentID)]; exists {
				rec := RECORDS[i]
				rec.LowerLimit = instrument.Start_min_price + second*instrument.Min_interval
				rec.UpperLimit = instrument.Start_max_price + second*instrument.Max_interval

				RECORDS[i] = rec
			}
		}
		mu.Unlock()
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
	SWITCH_CH = true

	if TRADE {
		for range dataChan {
			count++
			if count >= DATA_CHANNEL_LENGTH {
				break
			}
		}

	} else {
		for range recordsChan {
			count++
			if count >= DATA_CHANNEL_LENGTH {
				break
			}
		}
	}

	fmt.Println("Total items removed from channel :", count)
	SWITCH_CH = false
}

func ChangeDelay() {

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	if PATTERN_GENERATOR && PATTERN_TYPE == PEAK_GENERATOR {
		for {
			<-ticker.C
			DELAY_TIME = (rand.Intn(PEAK_GENERATOR_DETAILS.DelayMax-PEAK_GENERATOR_DETAILS.DelayMin+1) + PEAK_GENERATOR_DETAILS.DelayMin)
			fmt.Println("DELAY TIME :", DELAY_TIME)
			if !PATTERN_GENERATOR || PATTERN_TYPE != PEAK_GENERATOR {
				break
			}
		}
	}
}
