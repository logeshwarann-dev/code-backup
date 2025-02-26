package main

import "fmt"

func ValidateOPSRate(data OPSRate) error {
	if data.Type != 0 && data.Type != 1 {
		return fmt.Errorf("invalid value for Type: %d (allowed values: 0 or 1)", data.Type)
	}

	if data.Throttle < 0 {
		return fmt.Errorf("invalid value for Throttle: %d (Throttle must be a non-negative integer)", data.Throttle)
	}

	return nil
}

func UpdateRecordsValidationCheck(record Record) error {
	if record.InstrumentID <= 0 {
		return fmt.Errorf("instrument_id is required")
	}
	if record.LowerLimit <= 0 || record.UpperLimit <= 0 {
		return fmt.Errorf("lower_limit and upper_limit must be greater than 0")
	}
	if record.LowerLimit >= record.UpperLimit {
		return fmt.Errorf("lower_limit must be less than upper_limit")
	}
	if record.MinLot <= 0 {
		return fmt.Errorf("min_lot must be greater than 0")
	}
	if record.BidInterval <= 0 {
		return fmt.Errorf("bid_interval must be greater than 0")
	}
	if record.MaxTrdQty <= 0 {
		return fmt.Errorf("max_trd_qty must be greater than 0")
	}
	if record.Product_ID <= 0 {
		return fmt.Errorf("product_id must be greater than 0")
	}
	return nil
}

func DeleteRecordsValidationCheck(record DeleteOrd) error {
	if record.InstrumentID <= 0 {
		return fmt.Errorf("invalid instrument_id")
	}
	if record.ProductID <= 0 {
		return fmt.Errorf("product_id must be greater than 0")
	}
	return nil
}

func ValidatePattern(data GraphPattern) error {

	if data.Min <= 0 {
		return fmt.Errorf("min must be a positive integer")
	}
	if data.Max <= 0 {
		return fmt.Errorf("max must be a positive integer")
	}
	if data.Min >= data.Max {
		return fmt.Errorf("min must be smaller than Max")
	}
	if data.Type < 1 || data.Type > 6 {
		return fmt.Errorf("type must be between 1 and 6")
	}

	if data.Type == 4 {
		if data.Step == 0 {
			return fmt.Errorf("step must be greater than 0 for STEP_WAVE")
		}
		if data.Min%data.Step != 0 || data.Max%data.Step != 0 {
			return fmt.Errorf("min and Max must be multiples of Step for STEP_WAVE")
		}
	}

	if data.Type == 5 {
		if data.Interval < 2 {
			return fmt.Errorf("interval must be greater than 1 for TRIANGLE_WAVE")
		}
		if data.Min%data.Interval != 0 || data.Max%data.Interval != 0 {
			return fmt.Errorf("min and Max must be multiples of Interval for TRIANGLE_WAVE")
		}
	}

	return nil
}

func ValidateDelayDetails(data DelayDetails) error {
	if data.Min < 1 {
		return fmt.Errorf("min must be at least 1")
	}
	if data.Max > 1000 {
		return fmt.Errorf("max must be at most 1000")
	}
	if data.Min >= data.Max {
		return fmt.Errorf("min must be smaller than Max")
	}

	return nil
}

func ValidatePriceRangeChange(data PriceRangeChangeDetails) error {

	if data.Interval <= 1 {
		return fmt.Errorf("interval must be greater than 1")
	}

	return nil
}
