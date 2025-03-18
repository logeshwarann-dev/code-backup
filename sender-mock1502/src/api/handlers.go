package api

import (
	"fmt"
	"market_replay/src/core"
	"market_replay/src/static"
	"market_replay/src/utils"
	"net/http"
	"os"
	"strconv"

	"time"

	"github.com/gin-gonic/gin"
)

func HandleOPSRate(c *gin.Context) {
	var data static.OPSRate

	if err := c.ShouldBindJSON(&data); err != nil {
		fmt.Println("Error in Api", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	if err := utils.ValidateOPSRate(data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation error",
			"details": err.Error(),
		})
		return
	}

	static.PATTERN_GENERATOR = false

	if data.Type == 1 {
		static.TRADE_THROTTLE = data.Throttle
	} else {
		static.THROTTLE_VALUE = data.Throttle

		if err := utils.SetOrderPumpingParameters(); err != nil {
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

	var data static.RecordsArray
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Request format is not proper",
		})
		return
	}

	for i, record := range data.Records {
		if err := utils.UpdateRecordsValidationCheck(record); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Validation error at index %d: %s", i, err.Error()),
			})
			return
		}
	}

	fmt.Println("Records recevied: ", data.Records)

	for _, record := range data.Records {
		instId := record.InstrumentID
		foundFlag := 0
		static.Mu.Lock()
		for i, r := range static.RECORDS {
			if r.InstrumentID == instId {
				foundFlag = 1
				static.RECORDS[i] = record
				break
			}
		}
		if foundFlag == 0 {
			static.RECORDS = append(static.RECORDS, record)
		}
		static.Mu.Unlock()
	}

	fmt.Println("Updated RECORDS array: ", static.RECORDS)

	c.JSON(http.StatusOK, gin.H{
		"message": "Data appended successfully",
	})

	go utils.EmptyChannel()
}

func DeleteOrders(c *gin.Context) {

	var data static.DeleteOrd

	data.InstrumentID = 2147483647

	if err := c.ShouldBindJSON(&data); err != nil {
		fmt.Println("Error in Api", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	if err := utils.DeleteRecordsValidationCheck(data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Validation error : %s", err.Error()),
		})
		return
	}

	static.DELETE_ORDERS_DETAILS.InstrumentID = data.InstrumentID
	static.DELETE_ORDERS_DETAILS.ProductID = data.ProductID
	static.DELETE_ORDERS_DETAILS.Type = data.Type
	static.DELETE_ORDERS_ALL_SESSIONS++

	if static.DELETE_ORDERS_DETAILS.Type != static.DELETE_ALL {
		static.DELETE_ORDERS_DETAILS.Ids = data.Ids
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Delete Trigger Recevied",
		"Delete Count": static.DELETE_ORDERS_ALL_SESSIONS,
	})
}

func DeleteRecords(c *gin.Context) {
	var data static.InstStruct
	if err := c.ShouldBindBodyWithJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Request format is not proper"})
		return
	}

	if len(static.RECORDS) < 2 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Not enough records present",
		})
		return
	}

	foundFlag := 0

	static.Mu.Lock()
	for i, record := range static.RECORDS {
		if record.InstrumentID == data.InstId {
			foundFlag = 1
			static.RECORDS = append(static.RECORDS[:i], static.RECORDS[i+1:]...)
			break
		}
	}
	static.Mu.Unlock()

	if foundFlag == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": fmt.Sprintf("No Instrument Id present as %d", data.InstId),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Instrument Id removed successfully %d", data.InstId),
	})

	go utils.EmptyChannel()
}

func GeneratePattern(c *gin.Context) {
	var data static.GraphPattern

	if err := c.ShouldBindJSON(&data); err != nil {
		fmt.Println("Error binding: ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	if err := utils.ValidatePattern(data); err != nil {
		fmt.Println("error validation pattern: ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation error",
			"details": err.Error(),
		})
		return
	}

	fmt.Println("Payload: ", data)
	fmt.Println("TYpe: ", data.Type)

	static.PATTERN_GENERATOR = true
	static.PATTERN_TYPE = data.Type

	switch data.Type {

	case static.SQUARE_WAVE:
		static.PATTERN = utils.GenerateSquareWave(data.Min, data.Max, data.Interval)
	case static.SAWTOOTH_WAVE:
		static.PATTERN = utils.GenerateSawtoothWave(data.Min, data.Max, data.Interval)
	case static.SINE_WAVE:
		static.PATTERN = utils.GenerateSineWave(data.Min, data.Max)
	case static.STEP_WAVE:
		static.PATTERN = utils.GenerateStepWave(data.Min, data.Max, data.Step, data.Interval)
	case static.TRIANGLE_WAVE:
		static.PATTERN = utils.GenerateTriangleWave(data.Min, data.Max, data.Interval)
	case static.RANDOM_WAVE:
		static.PATTERN = []int{}
		static.RANDOM_WAVE_DETAILS = static.RandomWaveParmaters{Min: data.Min, Max: data.Max}
	case static.PEAK_GENERATOR:
		static.PATTERN = []int{}
		static.PEAK_GENERATOR_DETAILS = static.PeakGeneratorParameters{Min: data.Min, Max: data.Max, DelayMin: data.DelayMin, DelayMax: data.DelayMax}
		fmt.Println("Peak Generator Details: ", static.PEAK_GENERATOR_DETAILS)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Throttle pattern updated successfully",
		"pattern": static.PATTERN,
	})

	if static.PATTERN_TYPE == static.PEAK_GENERATOR {
		go utils.ChangeDelay()
	}
}

func AddDelay(c *gin.Context) {
	var data static.DelayDetails

	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	if err := utils.ValidateDelayDetails(data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation error",
			"details": err.Error(),
		})
		return
	}

	static.DELAY++
	static.DELAY_ORDER_DETAILS.Min = data.Min
	static.DELAY_ORDER_DETAILS.Max = data.Max

	c.JSON(http.StatusOK, gin.H{
		"message": "Delay trigger recevied",
	})
}

func AddSlidingPrice(c *gin.Context) {

	var data static.PriceRangeChangeDetails

	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	if err := utils.ValidatePriceRangeChange(data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation error",
			"details": err.Error(),
		})
		return
	}

	static.SLIDING_PRICE = false

	c.JSON(http.StatusOK, gin.H{
		"message": "Instruments price range recevied",
	})

	time.Sleep(1500 * time.Millisecond)

	static.SLIDING_PRICE = true

	go utils.ApplyPriceRangeChange(data)
	go utils.EmptyChannel()

}

func SetConfig(c *gin.Context) {
	var configData static.Config

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

	if configData.ProcessType < 0 || configData.ProcessType > 2 {
		fmt.Println("Error in Config Data Format: order pumping type must be between 0 and 2")
		c.JSON(http.StatusBadRequest, gin.H{"error": "order pumping type must be between 0 and 2"})
		return
	}

	fmt.Println("UniqClID: ", configData.UniqueClientIdentifier)
	fmt.Println("Config Data: ", configData)

	static.TRADERCOUNT = configData.TraderCount
	static.TRADER_COUNT = configData.TraderCount
	static.THROTTLE_VALUE = configData.ThrottleValue
	static.TRADE_THROTTLE = configData.ThrottleValue
	static.TARGET_ENV = configData.TargetEnv
	static.FILE_TYPE = configData.FileType
	static.CANCEL_PERCENT = configData.CancelPercent
	static.MODIFY_PERCENT = configData.ModifyPercent
	static.HEARTBEAT_VALUE = configData.HeartbeatValue
	static.UNIQUE_CLIENT_IDENTIFIER = configData.UniqueClientIdentifier

	switch configData.ProcessType {
	case static.SYNC_ORDER_PUMPING_TYPE:
		static.ORDER_PUMPING_TYPE = static.SYNC_ORDER_PUMPING_TYPE
	case static.ASYNC_ORDER_PUMPING_OEOMOC_TYPE:
		static.ORDER_PUMPING_TYPE = static.ASYNC_ORDER_PUMPING_OEOMOC_TYPE
		static.DISABLED_ACTIVITY_SYSTEM_VENDOR = configData.AppSystemVendorName
		if static.TARGET_ENV == static.LAB_ENV {
			static.LAB_DISABLED_ACTIVITY_SYSTEM_VERSION = configData.AppSystemVersion
		} else if static.TARGET_ENV == static.PRODUCTION_ENV {
			static.PROD_DISABLED_ACTIVITY_SYSTEM_VERSION = configData.AppSystemVersion
		}
	case static.ASYNC_ORDER_PUMPING_IOC_TYPE:
		static.ORDER_PUMPING_TYPE = static.ASYNC_ORDER_PUMPING_IOC_TYPE
	}

	if configData.TradeFlag == 1 {
		static.TRADE = true
	} else {
		static.TRADE = false
	}

	static.DURATION = configData.Duration

	static.DATA_CHANNEL_LENGTH = 2 * static.THROTTLE_VALUE * static.TRADER_COUNT

	static.TRADERS = append(static.TRADERS, configData.Traders...)
	static.TOTAL_ORDER_COUNT = configData.TraderCount * configData.Duration * configData.ThrottleValue

	fmt.Println("Trader count: ", static.TRADERCOUNT, "| Throttle value: ", static.THROTTLE_VALUE, "| Target Env: ", static.TARGET_ENV, "| file type: ", static.FILE_TYPE, "| Cancel percent: ", static.CANCEL_PERCENT, "| Mod Percent: ", static.MODIFY_PERCENT, "| HEARTBEAT_VALUE: ", static.HEARTBEAT_VALUE, "| ", "DURATION: ", static.DURATION, "| TRADERS: ", static.TRADERS, "| TOTAL_ORDER_COUNT: ", static.TOTAL_ORDER_COUNT)

	c.JSON(http.StatusOK, gin.H{
		"message": "Configuration updated successfully",
	})

}

func SetRecords(c *gin.Context) {
	var data static.RecordsArray

	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "config data not in proper format",
		})
		return
	}

	for i, record := range data.Records {
		if err := utils.UpdateRecordsValidationCheck(record); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Validation error at index" + strconv.Itoa(i) + err.Error(),
			})
			return
		}
	}

	fmt.Println("Records recevied: ", data.Records)

	static.RECORDS = append(static.RECORDS, data.Records...)

	fmt.Println("Updated RECORDS array: ", static.RECORDS)

	c.JSON(http.StatusOK, gin.H{
		"message": "Records appended successfully",
	})

	if !static.ORDERPUMPINGSTATUS {
		go core.Start()
	} else {
		fmt.Println("Orderpumping already active.")
	}

}

func HealthCheck(c *gin.Context) {
	podName := os.Getenv("POD_NAME")
	statusMsg := fmt.Sprintf("%s is healthy", podName)
	fmt.Println(statusMsg)
	c.JSON(http.StatusOK, gin.H{"pod_status": statusMsg})
}

func CheckOrderPumpingStatus(c *gin.Context) {
	podName := os.Getenv("POD_NAME")
	c.JSON(http.StatusOK, gin.H{"status": static.ORDERPUMPINGSTATUS, "pod_name": podName})
}
