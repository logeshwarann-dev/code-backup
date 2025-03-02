package main

import (
	"fmt"
	"net/http"
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
	PATTERN_TYPE = data.Type

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

	if PATTERN_TYPE == PEAK_GENERATOR {
		go ChangeDelay()
	}
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
