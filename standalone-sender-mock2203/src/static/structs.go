package static

import (
	"market_replay/src/encryption/aes"
	"net"
)

type Member struct {
	ID       int
	Password string
	Status   string
	Traders  []Trader
}

type Trader struct {
	TID       int    `json:"tid"`
	MID       int    `json:"mid"`
	Password  string `json:"password"`
	HostIP    string `json:"host_ip"`
	Port      string `json:"port"`
	Connected bool
	CGW_Res   GW_Response
	Conn      net.Conn
	Msg_seq   int
	CTX       *aes.CipherContext
	AckCount  int
	Throttle  uint32
	// dataChan     chan DataPacket
	Session_id   int
	OrigClOrdID  uint64
	ClOrdID      uint64
	Delete_order int
	OrderMap     map[uint64]OrderResponse
	// setDelay     int
	UiOrderMap map[uint64]UiOrderResponse
}

type GW_Response struct {
	GW1    string
	Port1  uint32
	GW2    string
	Port2  uint32
	SecKey string
	IV     string
}

type DataPacket struct {
	InstrumentId int
	Price        int
	TraderId     int
	ProductId    int
	Qty          int
	MaxPrice     int
	BidIntrvl    int
	MaxOrderQty  int
}

type TimeStamps struct {
	GW_Req_In  uint64
	GW_ME_Req  uint64
	ME_Req_In  uint64
	ME_Res_Out uint64
	GW_ME_Res  uint64
	GW_Res_Out uint64
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

type ProcessConfig struct {
	TraderCount        int
	ThrottleLimit      int
	RunTimeLimit       int
	TargetEnv          int
	FileType           int
	OrderModPercent    int
	OrderCancelPercent int
	HeartBeat          int
	UniqueIdentifier   int
	SessionSetValue    string
	TradeThrottle      int
	OrderPumpingType   int
	LogFlag            int
	TradeFlag          string
}

type OrderResponse struct {
	OrderID      uint64
	ActTime      uint64
	InstrumentID int
	ProductID    int
	OrderDetails OrderDetails
}

type OrderDetails struct {
	Price_factor   int
	Buy_min_price  int
	Buy_max_price  int
	Sell_min_price int
	Sell_max_price int
	Mid_qty        int
	Min_qty        int
	Max_qty        int
}

type UiOrderResponse struct {
	OrigClOrdID  uint64
	ActTime      uint64
	InstrumentID int
	ProductID    int
	OrderDetails OrderDetails
}

type OPSRate struct {
	Throttle int `json:"throttle"`
	Type     int `json:"type"` // 0 => Order Pumping 1=> Trade
}

type RecordsArray struct {
	Records []Record `json:"records" binding:"required"`
}

type InstStruct struct {
	InstId int `json:"instId" binding:"required"`
}

type DeleteOrd struct {
	InstrumentID int `json:"instId" binding:"required"`
	ProductID    int `json:"prodId" binding:"required"`
	Type         int `json:"type"`
	Ids          map[int]int
}

type GraphPattern struct {
	Min      int `json:"min" binding:"required"`
	Max      int `json:"max" binding:"required"`
	Type     int `json:"type" binding:"required"`
	Step     int
	Interval int
	DelayMin int `json:"delay_min"`
	DelayMax int `json:"delay_max"`
}

type RandomWaveParmaters struct {
	Min int
	Max int
}

type PeakGeneratorParameters struct {
	Min      int
	Max      int
	DelayMin int
	DelayMax int
}

type DelayDetails struct {
	Min int `json:"min" binding:"required"`
	Max int `json:"max" binding:"required"`
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

type Config struct {
	TotalOrderCount        int      `json:"total_order_count"`
	TargetEnv              int      `json:"target_env" binding:"required"`
	ModifyPercent          int      `json:"modify_percent" `
	CancelPercent          int      `json:"cancel_percent" `
	ThrottleValue          int      `json:"throttle_value" `
	FileType               int      `json:"file_type"`
	TraderCount            int      `json:"trader_count" `
	HeartbeatValue         int      `json:"heartbeat_value" `
	UniqueClientIdentifier bool     `json:"unique_identifier"`
	Traders                []Trader `json:"traders" binding:"required"`
	Duration               int      `json:"duration" binding:"required"`
	TradeFlag              int      `json:"trade_flag"`
	ProcessType            int      `json:"process_type"`
	AppSystemVendorName    string   `json:"app_system_vendor"`
	AppSystemVersion       string   `json:"app_system_version"`
	LogFlag                bool     `json:"log_flag"`
}
