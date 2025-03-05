package static

type OPSRate struct {
	Pods     []int `json:"pods" binding:"required"`
	Throttle int   `json:"throttle"`
	Type     int   `json:"type"` // 0 => Order Pumping 1=> Trade
}

type RecordsArray struct {
	Records []Record `json:"records" binding:"required"`
}

type DeleteRecordStruct struct {
	InstId int `json:"instId" binding:"required"`
}

type DeleteOrder struct {
	InstrumentID int `json:"instId" binding:"required"`
	ProductID    int `json:"prodId" binding:"required"`
}

type Record struct {
	InstrumentID int `json:"instrument_id"`
	LowerLimit   int `json:"lower_limit"`
	UpperLimit   int `json:"upper_limit"`
	MinLot       int `json:"min_lot"`
	BidInterval  int `json:"bid_interval"`
	MaxTrdQty    int `json:"max_trd_qty"`
	Product_ID   int `json:"product_id"`
}

type Config struct {
	TotalOrderCount        int      `json:"total_order_count"`
	TargetEnv              int      `json:"target_env" binding:"required"`
	ModifyPercent          int      `json:"modify_percent"`
	CancelPercent          int      `json:"cancel_percent"`
	ThrottleValue          int      `json:"throttle_value" binding:"required"`
	FileType               int      `json:"file_type"`
	TraderCount            int      `json:"trader_count" binding:"required"`
	HeartbeatValue         int      `json:"heartbeat_value" binding:"required"`
	UniqueClientIdentifier bool     `json:"unique_identifier"`
	Traders                []Trader `json:"traders"`
	TradeFlag              int      `json:"trade_flag"`
	Duration               int      `json:"duration" binding:"required"`
	Trader_OPS             int      `json:"trader_ops" binding:"required"`
}

type ConfigResponse struct {
	TotalOrderCount        int
	TargetEnv              int
	ModifyPercent          int
	CancelPercent          int
	ThrottleValue          map[string]int
	FileType               int
	TraderCount            int
	HeartbeatValue         int
	UniqueClientIdentifier bool
	ProcessState           int
	Duration               int
}

type Trader struct {
	TID      int    `json:"tid"`
	MID      int    `json:"mid"`
	Password string `json:"password"`
	HostIP   string `json:"host_ip"`
	Port     string `json:"port"`
}

type SseData struct {
	Message string `json:"message" binding:"required"`
	Type    int    `json:"type" binding:"required"`
}

type GraphPattern struct {
	Pods     []int `json:"pods"`
	Min      int   `json:"min" binding:"required"`
	Max      int   `json:"max" binding:"required"`
	Type     int   `json:"type" binding:"required"`
	Step     int
	Interval int
	DelayMin int `json:"delay_min"`
	DelayMax int `json:"delay_max"`
}

type PriceRangeChangeDetails struct {
	Interval    int                         `json:"interval" binding:"required"`
	Instruments map[string]PriceRangeChange `json:"inst_id" binding:"required"`
}

type PriceRangeChange struct {
	Start_min_price int `json:"start_min_price" binding:"required"`
	Start_max_price int `json:"start_max_price" binding:"required"`
	End_min_price   int `json:"end_min_price" binding:"required"`
	End_max_price   int `json:"end_max_price" binding:"required"`
	Min_interval    int
	Max_interval    int
}

type PodResponse struct {
	Address  string `json:"address"`
	Response string `json:"response"`
}

type AddPodCount struct {
	Count int   `json:"count"`
	Pods  []int `json:"pods"`
}

type DeletePodCount struct {
	Pods []int `json:"pods"`
}

type PodActiveResponse struct {
	PodName string `json:"pod_name"`
	Status  bool   `json:"status"`
}
