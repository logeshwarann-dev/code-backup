package static

var SenderPodIPMap map[string]string

var TraderPodIPMap map[string]string
var LabSessionIdSets map[string][]Trader
var ProdSessionIdSets map[string][]Trader
var SimSessionIdSets map[string][]Trader
var MemberPodMap map[string]string
var SyncPeakPodIPMap map[string]string
var AsyncIOCPeakPodIPMap map[string]string
var AsyncEMCPeakPodIPMap map[string]string

const (
	ServicePort             = "8080"
	MaxRetry                = 3
	FilePath                = "Records/mockRecords.txt"
	SENDER_CONNECTION_COUNT = 100
	TRADER_CONNECTION_COUNT = 100
	SENDER_POD_LABEL        = "mock-sender"
	TRADER_POD_LABEL        = "mock-trader"
	// PEAK_POD_LABEL          = "peak-generator"
	TRADER_POD_COUNT = 1
	K8S_NAMESPACE    = "default"
	TRADER_MEMBERID  = "1046"
	// PEAK_POD_MEMBERID       = "1045"
	TOTAL_AVL_SESSIONS   = 1300
	TRADER_POD_MAGIC_NUM = 18
)

var (
	RECORDS_FILE_UPLOADED    = false
	PROCESS_STATE            = 0
	SESSIONID_INDEX          = 0
	SYNC_PEAK_POD_LIMIT      = 1
	ASYNC_IOC_PEAK_POD_LIMIT = 1
	ASYNC_EMC_PEAK_POD_LIMIT = 1
)

var (
	SYNC_ORDER_PUMPING_TYPE         = 0
	ASYNC_ORDER_PUMPING_OEOMOC_TYPE = 1
	ASYNC_ORDER_PUMPING_IOC_TYPE    = 2
)

var (
	DEFAULT_SYSTEM_VERSION                = "A"
	DEFAULT_SYSTEM_VENDOR                 = "A"
	LAB_DISABLED_ACTIVITY_SYSTEM_VERSION  = "8016"
	DISABLED_ACTIVITY_SYSTEM_VENDOR       = "TWSIML"
	PROD_DISABLED_ACTIVITY_SYSTEM_VERSION = "8026"
)

var EventChannel chan (SseData)

var ConfigDetails ConfigResponse

var ConfigData Config

var RECORDS []Record

var TARGETENV int

var CalculatedSenderThrottle int
