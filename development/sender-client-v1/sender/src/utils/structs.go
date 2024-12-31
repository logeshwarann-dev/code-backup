package utils

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
	TID           int
	MID           int
	Password      string
	HostIP        string
	Port          string
	Connected     bool
	CGW_Res       GW_Response
	Conn          net.Conn
	Msg_seq       int
	CTX           *aes.CipherContext
	AckCount      int
	ThrottleValue uint32
	TraderChan    chan RedisData
	Session_id    int
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
	// Message      string
	InstrumentId int
	Price        int
	TraderId     int
	ProductId    int
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
	InstrumentID string
	LowerLimit   int
	UpperLimit   int
	MinLot       int
	Product_ID   int
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
}

type RedisData struct {
	TemplateID      int     `json:"template_id"`
	Price           float64 `json:"price"`
	OrderQty        int     `json:"order_qty"`
	MaxShow         int     `json:"maxshow"`
	InstID          int     `json:"instid"`
	AccType         string  `json:"acc_type"`
	TimeInForce     string  `json:"time_in_force"`
	ClientCode      string  `json:"client_code"`
	MsgSeqNum       int     `json:"msg_seq_num"`
	TraderId        int     `json:"trader_id"`
	PartitionId     int     `json:"partition_id"`
	ProductId       int     `json:"product_id"`
	TransactionType int     `json:"transaction_type"`
	TimeStamp       int
	OrderID         float64
}

type CacheEntry struct {
	NewOrderID          int64
	UpdatedActivityTime int64
	Used                bool
}

type RedisDataOrderPumping struct {
	TemplateID  int     `json:"template_id"`
	Price       float64 `json:"price"`
	OrderQty    int     `json:"order_qty"`
	MaxShow     int     `json:"maxshow"`
	InstID      int     `json:"instid"`
	AccType     string  `json:"acc_type"`
	TimeInForce string  `json:"time_in_force"`
	ClientCode  string  `json:"client_code"`
	PartitionId int     `json:"partition_id"`
	ProductId   int     `json:"product_id"`
}
