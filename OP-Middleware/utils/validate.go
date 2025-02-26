package utils

import (
	"errors"
	"fmt"
	"op-middleware/static"
	// logger "op-middleware/logging"
)

func ValidateRecord(record static.Record) error {
	if record.LowerLimit <= 0 || record.UpperLimit <= 0 {
		fmt.Println("lower_limit and upper_limit must be greater than 0")
		return errors.New("lower_limit and upper_limit must be greater than 0")
	}
	if record.LowerLimit >= record.UpperLimit {
		fmt.Println("lower_limit must be less than upper_limit")
		return errors.New("lower_limit must be less than upper_limit")
	}
	if record.MinLot <= 0 {
		fmt.Println("min_lot must be greater than 0")
		return errors.New("min_lot must be greater than 0")
	}
	if record.BidInterval <= 0 {
		fmt.Println("bid_interval must be greater than 0")
		return errors.New("bid_interval must be greater than 0")
	}
	if record.MaxTrdQty <= 0 {
		fmt.Println("max_trd_qty must be greater than 0")
		return errors.New("max_trd_qty must be greater than 0")
	}
	if record.Product_ID <= 0 {
		fmt.Println("product_id must be greater than 0")
		return errors.New("product_id must be greater than 0")
	}
	return nil
}

func DeleteOrdersValidationCheck(record static.DeleteOrder) error {
	if record.InstrumentID <= 0 {
		return fmt.Errorf("invalid instrument_id")
	}
	if record.ProductID <= 0 {
		return fmt.Errorf("product_id must be greater than 0")
	}
	return nil
}

func ValidatePattern(data static.GraphPattern) error {

	if data.Min <= 0 {
		return fmt.Errorf("min must be a positive integer")
	}
	if data.Max <= 0 {
		return fmt.Errorf("max must be a positive integer")
	}
	if data.Min >= data.Max {
		return fmt.Errorf("min must be smaller than Max")
	}
	if data.Type < 1 || data.Type > 7 {
		return fmt.Errorf("type must be between 1 and 7")
	}

	if data.Type == 4 {
		if data.Step == 0 {
			return fmt.Errorf("step must be greater than 0 for STEP_WAVE")
		}
		if data.Min%data.Step != 0 || data.Max%data.Step != 0 {
			return fmt.Errorf("min and max must be multiples of step for STEP_WAVE")
		}
	}

	if data.Type == 5 {
		if data.Interval < 2 {
			return fmt.Errorf("interval must be greater than 1 for TRIANGLE_WAVE")
		}
		if data.Min%data.Interval != 0 || data.Max%data.Interval != 0 {
			return fmt.Errorf("min and max must be multiples of Interval for TRIANGLE_WAVE")
		}
	}

	return nil
}

func ValidatePriceRangeChange(data static.PriceRangeChangeDetails) error {
	if data.Interval <= 1 {
		return fmt.Errorf("interval must be greater than 1")
	}

	return nil
}

func ValidatePrices(PricesMap map[string]static.PriceRangeChange) error {
	for _, details := range PricesMap {
		if details.End_max_price-details.End_min_price < 1000 {
			return fmt.Errorf("price difference is very less in end maximum and end minimum")
		}
		if details.Start_max_price-details.Start_min_price < 1000 {
			return fmt.Errorf("price difference is very less in start maximum and start minimum")
		}
	}
	return nil
}
