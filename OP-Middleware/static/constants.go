package static

var SenderPodIPMap map[string]string
var PeakPodIPMap map[string]string
var TraderPodIPMap map[string]string
var LabSessionIdSets map[string][]Trader
var ProdSessionIdSets map[string][]Trader
var SimSessionIdSets map[string][]Trader
var MemberPodMap map[string]string

const (
	ServicePort             = "8080"
	MaxRetry                = 3
	FilePath                = "Records/mockRecords.txt"
	SENDER_CONNECTION_COUNT = 100
	TRADER_CONNECTION_COUNT = 100
	SENDER_POD_LABEL        = "mock-sender"
	TRADER_POD_LABEL        = "mock-trader"
	PEAK_POD_LABEL          = "peak-generator"
	TRADER_POD_COUNT        = 1
	K8S_NAMESPACE           = "default"
	TRADER_MEMBERID         = "1046"
	PEAK_POD_MEMBERID       = "1045"
)

var (
	PROCESS_STATE   = 0
	SESSIONID_INDEX = 0
	PEAK_POD_COUNT  = 1
)

var EventChannel chan (SseData)

var ConfigDetails ConfigResponse

var ConfigData Config

var RECORDS []Record

var TARGETENV int

var CalculatedSenderThrottle int
