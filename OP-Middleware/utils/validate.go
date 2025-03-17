package utils

import (
	"errors"
	"fmt"
	"op-middleware/static"
	// logger "op-middleware/logging"
)

func ValidateChangeOPS(opsRate static.OPSRate) error {
	if len(opsRate.Pods) > 1 {
		return fmt.Errorf("single value must be entered. Don't enter multiple values")
	}
	if opsRate.Type != 0 && opsRate.Type != 1 {
		return fmt.Errorf("invalid type value")
	}

	if opsRate.Throttle < 0 {
		return fmt.Errorf("ops per pod must be atleast 0")
	}

	if opsRate.Throttle > 200000 {
		return fmt.Errorf("ops per pod must be less than 2L")
	}
	// if opsRate.Throttle != 0 {
	// 	if opsRate.Throttle%100 != 0 {
	// 		return fmt.Errorf("ops rate must be divisible by 100")
	// 	}
	// }

	if len(opsRate.Pods) == 1 {
		isPodPresent := CheckPodAvailability(opsRate.Pods[0])
		if !isPodPresent {
			return fmt.Errorf("pod id not available")
		}
	}
	return nil
}

func ValidateRecord(record static.Record) error {
	if record.LowerLimit <= 0 || record.UpperLimit <= 0 {
		fmt.Println("lower_limit and upper_limit must be greater than 0")
		return errors.New("lower_limit and upper_limit must be greater than 0")
	}
	if record.LowerLimit >= record.UpperLimit {
		fmt.Println("lower_limit must be less than upper_limit")
		return errors.New("lower_limit must be less than upper_limit")
	}
	if record.LowerLimit >= 10000000 || record.UpperLimit >= 10000000 {
		fmt.Println("price limit is very high")
		return errors.New("price limit is very high")
	}
	if record.MinLot <= 0 {
		fmt.Println("min_lot must be greater than 0")
		return errors.New("min_lot must be greater than 0")
	}
	if record.MinLot >= 100000 {
		fmt.Println("min_lot is very high")
		return errors.New("min_lot is very high")
	}
	if record.BidInterval <= 0 {
		fmt.Println("bid_interval must be greater than 0")
		return errors.New("bid_interval must be greater than 0")
	}
	if record.BidInterval >= 1000 {
		fmt.Println("bid_interval is very high")
		return errors.New("bid_interval is very high")
	}
	if record.MaxTrdQty <= 0 {
		fmt.Println("max_trd_qty must be greater than 0")
		return errors.New("max_trd_qty must be greater than 0")
	}
	if record.MaxTrdQty >= 100000 {
		fmt.Println("max_trd_qty is very high")
		return errors.New("max_trd_qty is very high")
	}
	if record.Product_ID <= 0 {
		fmt.Println("product_id must be greater than 0")
		return errors.New("product_id must be greater than 0")
	}
	if record.Product_ID >= 10000 {
		fmt.Println("product_id is very high")
		return errors.New("product_id is very high")
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

	if static.ConfigData.TraderCount == 0 {
		return fmt.Errorf("order pumping is not started")
	}
	if len(data.Pods) > 1 {
		return fmt.Errorf("single value must be entered. Don't enter multiple values")
	}

	if len(data.Pods) == 1 {
		if data.Pods[0] > 20 {
			return fmt.Errorf("invalid pod id. Pod id should be less than 20")
		}
		if data.Pods[0] < 0 {
			return fmt.Errorf("pod id cannot be negative")
		}
		isPodPresent := CheckPodAvailability(data.Pods[0])
		if !isPodPresent {
			return fmt.Errorf("pod id not available")
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
		if data.DelayMax > 300000 { // 5mins
			return fmt.Errorf("max delay value is too large")

		}
		if data.DelayMin > 120000 { //2mins
			return fmt.Errorf("min delay value is too large")
		}

		if data.Max > 2000000 {
			return fmt.Errorf("calculated max ops value is too large. reduce base ops & percentage")
		}
		if data.Min > 1500000 {
			return fmt.Errorf("calculated min ops value is too large. reduce base ops & percentage")
		}
		if data.ProcessType < 0 || data.ProcessType > 2 {
			return fmt.Errorf("invalid process type. out of range")
		}
		if data.ProcessType == 1 {
			if len(data.AppSystemVendorName) == 0 {
				return fmt.Errorf("empty vendor name")
			}
			if len(data.AppSystemVersion) == 0 {
				return fmt.Errorf("empty system version")
			}
			if len(data.AppSystemVersion) > 100000 {
				return fmt.Errorf("system version is too large")
			}
		}
		if data.ProcessType == 0 || data.ProcessType == 2 {
			if len(data.AppSystemVendorName) > 0 {
				return fmt.Errorf("vendor name is not required. Default value will be used")
			}
			if len(data.AppSystemVersion) > 0 {
				return fmt.Errorf("system version is not required. Default value will be used")
			}
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

	if static.ConfigData.TraderCount == 0 {
		return fmt.Errorf("order pumping is not started")
	}

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

	if config.TraderCount <= 0 {
		return fmt.Errorf("traders must be greater than 0")
	}
	if config.TraderCount >= 1200 {
		return fmt.Errorf("trader count must be less than 1200")
	}
	if config.TraderCount%100 != 0 {
		return fmt.Errorf("trader count must be multiple of 100")
	}
	if config.ThrottleValue < 0 {
		return fmt.Errorf("ops Rate should not be less than 0")
	}
	if config.TargetEnv == 1 {
		if config.ThrottleValue > 80000 {
			return fmt.Errorf("ops Rate is very high for Lab env. reduce it")
		}
	}
	if config.TargetEnv == 2 {
		if config.ThrottleValue > 1000000 {
			return fmt.Errorf("ops Rate is very high for prod env. reduce it")
		}
	}
	if config.TargetEnv == 0 {
		if config.ThrottleValue > 1000 {
			return fmt.Errorf("ops Rate is very high for simulation env. reduce it")
		}
	}
	if config.ThrottleValue%config.TraderCount != 0 {
		return fmt.Errorf("ops rate must be divisible by trader count")
	}
	if config.ThrottleValue/config.TraderCount > 1000 {
		return fmt.Errorf("ops Rate per pod is very high. increase trader count")
	}
	if config.ThrottleValue%100 != 0 {
		return fmt.Errorf("ops Rate must be multiple of 100")
	}
	if config.Trader_OPS%static.TRADER_CONNECTION_COUNT != 0 {
		return fmt.Errorf("trader ops Rate must be multiple of 100")
	}
	if config.HeartbeatValue <= 0 {
		return fmt.Errorf("heartbeat must be greater than 0")
	}
	if config.HeartbeatValue > 1000 {
		return fmt.Errorf("heartbeat must be less than 1000")
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
	if config.Duration > 36000 {
		return fmt.Errorf("duration should be less than 36000(10hrs)")
	}
	if config.Duration <= 0 {
		return fmt.Errorf("duration should be greater than 0")
	}
	return nil

}
