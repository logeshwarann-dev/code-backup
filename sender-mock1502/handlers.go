package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var mu sync.RWMutex

func HandleOPSRate(c *gin.Context) {
	var data OPSRate

	data.Type = 0

	if err := c.ShouldBindJSON(&data); err != nil {
		// fmt.Println("Error in Api", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	if err := ValidateOPSRate(data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation error",
			"details": err.Error(),
		})
		return
	}

	PATTERN_GENERATOR = false

	if data.Type == 1 {
		TRADE_THROTTLE = data.Throttle
	} else {
		THROTTLE_VALUE = data.Throttle

		if err := setOrderPumpingParameters(); err != nil {
			fmt.Printf("Error in changing order pumping parameters: %v", err)

			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to update order pumping parameters",
				"details": err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Throttle value updated successfully",
		"throttle": data.Throttle,
	})
}

func UpdateRecords(c *gin.Context) {

	var data RecordsArray
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Request format is not proper",
		})
		return
	}

	for i, record := range data.Records {
		if err := UpdateRecordsValidationCheck(record); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Validation error at index %d: %s", i, err.Error()),
			})
			return
		}
	}

	for _, record := range data.Records {
		instId := record.InstrumentID
		foundFlag := 0
		mu.Lock()
		for i, r := range RECORDS {
			if r.InstrumentID == instId {
				foundFlag = 1
				RECORDS[i] = record
				break
			}
		}
		if foundFlag == 0 {
			RECORDS = append(RECORDS, record)
		}
		mu.Unlock()
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Data appended successfully",
	})

	go EmptyChannel()
}

func DeleteOrders(c *gin.Context) {

	var data DeleteOrd

	data.InstrumentID = 2147483647

	if err := c.ShouldBindJSON(&data); err != nil {
		fmt.Println("Error in Api", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	if err := DeleteRecordsValidationCheck(data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Validation error : %s", err.Error()),
		})
		return
	}

	DELETE_ORDERS_DETAILS.InstrumentID = data.InstrumentID
	DELETE_ORDERS_DETAILS.ProductID = data.ProductID
	DELETE_ORDERS_DETAILS.Type = data.Type
	DELETE_ORDERS_ALL_SESSIONS++

	if DELETE_ORDERS_DETAILS.Type != DELETE_ALL {
		DELETE_ORDERS_DETAILS.Ids = data.Ids
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Delete Trigger Recevied",
		"Delete Count": DELETE_ORDERS_ALL_SESSIONS,
	})
}

func DeleteRecords(c *gin.Context) {
	var data InstStruct
	if err := c.ShouldBindBodyWithJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Request format is not proper"})
		return
	}

	if len(RECORDS) < 2 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Not enough records present",
		})
		return
	}

	foundFlag := 0

	mu.Lock()
	for i, record := range RECORDS {
		if record.InstrumentID == data.InstId {
			foundFlag = 1
			RECORDS = append(RECORDS[:i], RECORDS[i+1:]...)
			break
		}
	}
	mu.Unlock()

	if foundFlag == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": fmt.Sprintf("No Instrument Id present as %d", data.InstId),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Instrument Id removed successfully %d", data.InstId),
	})

	go EmptyChannel()
}

func GeneratePattern(c *gin.Context) {
	var data GraphPattern

	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	if err := ValidatePattern(data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation error",
			"details": err.Error(),
		})
		return
	}

	PATTERN_GENERATOR = true

	switch data.Type {

	case SQUARE_WAVE:
		PATTERN = GenerateSquareWave(data.Min, data.Max, data.Interval)
	case SAWTOOTH_WAVE:
		PATTERN = GenerateSawtoothWave(data.Min, data.Max, data.Interval)
	case SINE_WAVE:
		PATTERN = GenerateSineWave(data.Min, data.Max)
	case STEP_WAVE:
		PATTERN = GenerateStepWave(data.Min, data.Max, data.Step, data.Interval)
	case TRIANGLE_WAVE:
		PATTERN = GenerateTriangleWave(data.Min, data.Max, data.Interval)
	case RANDOM_WAVE:
		PATTERN = []int{}
		RANDOM_WAVE_DETAILS = RandomWaveParmaters{Min: data.Min, Max: data.Max}
	case PEAK_GENERATOR:
		PATTERN = []int{}
		PEAK_GENERATOR_DETAILS = PeakGeneratorParameters{Min: data.Min, Max: data.Max, DelayMin: data.DelayMin, DelayMax: data.DelayMax}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Throttle pattern updated successfully",
		"pattern": PATTERN,
	})
}

func AddDelay(c *gin.Context) {
	var data DelayDetails

	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	if err := ValidateDelayDetails(data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation error",
			"details": err.Error(),
		})
		return
	}

	DELAY++
	DELAY_ORDER_DETAILS.Min = data.Min
	DELAY_ORDER_DETAILS.Max = data.Max

	c.JSON(http.StatusOK, gin.H{
		"message": "Delay trigger recevied",
	})
}

func AddSlidingPrice(c *gin.Context) {

	var data PriceRangeChangeDetails

	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	if err := ValidatePriceRangeChange(data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation error",
			"details": err.Error(),
		})
		return
	}

	SLIDING_PRICE = false

	c.JSON(http.StatusOK, gin.H{
		"message": "Instruments price range recevied",
	})

	time.Sleep(1500 * time.Millisecond)

	SLIDING_PRICE = true

	go ApplyPriceRangeChange(data)
	go EmptyChannel()

}

func SetConfig(c *gin.Context) {
	var configData Config

	if err := c.ShouldBindJSON(&configData); err != nil {
		fmt.Println("Error in Config Data Format: ", err.Error(), "Request body: ", configData)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "config data not in proper format",
		})
		return
	}

	if configData.TraderCount < 0 || configData.TraderCount > 651 {
		fmt.Println("Error in Config Data Format: Traders must be between 1 and 651")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Traders must be between 1 and 651"})
		return
	}
	if configData.ThrottleValue < 0 {
		fmt.Println("Error in Config Data Format: Throttle must be greater than 0")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Throttle must be greater than 0"})
		return
	}
	if configData.TargetEnv < 0 || configData.TargetEnv > 2 {
		fmt.Println("Error in Config Data Format: Environment must be 0, 1, or 2")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Environment must be 0, 1, or 2"})
		return
	}
	if configData.FileType < 0 || configData.FileType > 1 {
		fmt.Println("Error in Config Data Format: EnvFile must be 0 or 1")
		c.JSON(http.StatusBadRequest, gin.H{"error": "EnvFile must be 0 or 1"})
		return
	}
	if configData.CancelPercent < 0 || configData.CancelPercent > 101 {
		fmt.Println("Error in Config Data Format: CancelPercent must be between 0 and 100")
		c.JSON(http.StatusBadRequest, gin.H{"error": "CancelPercent must be between 0 and 100"})
		return
	}
	if configData.ModifyPercent < 0 || configData.ModifyPercent > 101 {
		fmt.Println("Error in Config Data Format: ModifyPercent must be between 0 and 100")
		c.JSON(http.StatusBadRequest, gin.H{"error": "ModifyPercent must be between 0 and 100"})
		return
	}

	fmt.Println("Config Data: ", configData)

	TRADERCOUNT = configData.TraderCount
	TRADER_COUNT = configData.TraderCount
	THROTTLE_VALUE = configData.ThrottleValue
	TARGET_ENV = configData.TargetEnv
	FILE_TYPE = configData.FileType
	CANCEL_PERCENT = configData.CancelPercent
	MODIFY_PERCENT = configData.ModifyPercent
	HEARTBEAT_VALUE = configData.HeartbeatValue
	UNIQUE_CLIENT_IDENTIFIER = configData.UniqueClientIdentifier
	if configData.TradeFlag == 1 {
		TRADE = true
	} else {
		TRADE = false
	}

	DURATION = configData.Duration

	DATA_CHANNEL_LENGTH = 2 * THROTTLE_VALUE * TRADER_COUNT

	TRADERS = append(TRADERS, configData.Traders...)
	TOTAL_ORDER_COUNT = configData.TraderCount * configData.Duration * configData.ThrottleValue

	fmt.Println("Trader count: ", TRADERCOUNT, "| Throttle value: ", THROTTLE_VALUE, "| Target Env: ", TARGET_ENV, "| file type: ", FILE_TYPE, "| Cancel percent: ", CANCEL_PERCENT, "| Mod Percent: ", MODIFY_PERCENT, "| HEARTBEAT_VALUE: ", HEARTBEAT_VALUE, "| ", "DURATION: ", DURATION, "| TRADERS: ", TRADERS, "| TOTAL_ORDER_COUNT: ", TOTAL_ORDER_COUNT)

	c.JSON(http.StatusOK, gin.H{
		"message": "Configuration updated successfully",
	})

}

func SetRecords(c *gin.Context) {
	var data RecordsArray

	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "config data not in proper format",
		})
		return
	}

	for i, record := range data.Records {
		if err := UpdateRecordsValidationCheck(record); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Validation error at index" + strconv.Itoa(i) + err.Error(),
			})
			return
		}
	}

	RECORDS = append(RECORDS, data.Records...)

	c.JSON(http.StatusOK, gin.H{
		"message": "Records appended successfully",
	})

	go Start()
}

func HealthCheck(c *gin.Context) {
	podName := os.Getenv("POD_NAME")
	statusMsg := fmt.Sprintf("%s is healthy", podName)
	fmt.Println(statusMsg)
	c.JSON(http.StatusOK, gin.H{"pod_status": statusMsg})
}

func CheckOrderPumpingStatus(c *gin.Context) {
	podName := os.Getenv("POD_NAME")
	c.JSON(http.StatusOK, gin.H{"status": ORDERPUMPINGSTATUS, "pod_name": podName})
}
