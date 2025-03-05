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

	if len(data.Pods) > 1 {
		return fmt.Errorf("single value must be entered. Don't enter multiple values")
	}

	if len(data.Pods) > 0 {
		if data.Pods[0] > 20 {
			return fmt.Errorf("invalid pod id. Pod id should be less than 20")
		}
		if data.Pods[0] < 0 {
			return fmt.Errorf("pod id cannot be negative")
		}
	}

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

	if data.Type == 7 {
		if data.DelayMax > 300000 {
			return fmt.Errorf("max delay value is too large")

		}
		if data.DelayMin > 120000 {
			return fmt.Errorf("min delay value is too large")
		}
		if data.Max > 2000000 {
			return fmt.Errorf("calculated max ops value is too large. reduce base ops & percentage")
		}
		if data.Min > 1500000 {
			return fmt.Errorf("calculated min ops value is too large. reduce base ops & percentage")
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

func ValidateAddPod(data static.AddPodCount) error {

	if data.Count < 1 {
		return fmt.Errorf("pod count must be greater than 1")
	}
	if data.Count > 20 {
		return fmt.Errorf("pod count must be less than 20")
	}

	return nil
}

func ValidateDeletePod(data static.DeletePodCount) error {

	if len(data.Pods) == 0 {
		return nil
	}

	if len(data.Pods) > 1 {
		return fmt.Errorf("single value must be entered. Don't enter multiple values")
	}

	if len(data.Pods) > 0 {
		if data.Pods[0] > 20 {
			return fmt.Errorf("invalid pod id. Pod id should be less than 20")
		}
		if data.Pods[0] < 0 {
			return fmt.Errorf("pod id cannot be negative")
		}

	}

	return nil

}

func ValidateConfig(config static.Config) error {

	if config.TraderCount < 0 {
		return fmt.Errorf("traders must be greater than 0")
	}
	if config.ThrottleValue <= 0 {
		return fmt.Errorf("ops Rate must be greater than 0")
	}
	if config.HeartbeatValue <= 0 {
		return fmt.Errorf("heartbeat must be greater than 0")
	}
	if config.TargetEnv < 0 || config.TargetEnv > 2 {
		return fmt.Errorf("environment must be 0, 1, or 2")
	}
	if config.FileType < 0 || config.FileType > 1 {
		return fmt.Errorf("environment File must be 0 or 1")
	}
	if config.CancelPercent < 0 || config.CancelPercent > 101 {
		return fmt.Errorf("cancel Percentage must be between 0 and 100")
	}
	if config.ModifyPercent < 0 || config.ModifyPercent > 101 {
		return fmt.Errorf("modify Percentage must be between 0 and 100")
	}
	return nil

}
