package main

var (
	STANDARD_ORDER            = uint16(10101)
	LEAN_ORDER                = uint16(10102)
	REJECTED                  = uint16(10010)
	HEARTBEAT                 = uint16(10023)
	ORDER_CONFIRMATION        = uint16(10990)
	MODIFY_LEAN_ORDER         = uint16(10108)
	CANCEL_LEAN_ORDER         = uint16(10111)
	IMMEDIATE_EXECUTION       = uint16(10103)
	ORDER_BOOK_EXECUTION      = uint16(10104)
	DELETE_ALL_ORDERS         = uint16(10121)
	DELETE_ALL_ORDERS_NO_HITS = uint16(10124)
	TRADING_SESSION_EVENT     = uint16(10307)
)

var (
	TRADER_COUNT             int
	TOTAL_ORDER_COUNT        int
	TARGET_ENV               int
	MODIFY_PERCENT           int
	CANCEL_PERCENT           int
	THROTTLE_VALUE           int
	FILE_TYPE                int // 0 => All Records , 1 => Limited Records
	TRADERCOUNT              int
	HEARTBEAT_VALUE          int
	UNIQUE_CLIENT_IDENTIFIER bool
	OTR                      int
	TRADE_THROTTLE           int
	DURATION                 int
	SESSIONSETVALUE          string
)

var (
	PATTERN_GENERATOR      bool
	PATTERN                []int
	PATTERN_CHANNEL_LENGTH = 5
	PATTERN_TYPE           int
	RANDOM_WAVE_DETAILS    RandomWaveParmaters
	PEAK_GENERATOR_DETAILS PeakGeneratorParameters
)

var (
	SEND_MOD_COUNT     int
	CANCEL_ORDER_COUNT int
	ORDER_ENTRY_COUNT  int
	MOD_PER_ORDER      int
	SEND_MOD           bool
)

var (
	RECORDS                    []Record
	DELETE_ORDERS_ALL_SESSIONS int
	DELETE_ORDERS_DETAILS      DeleteOrd
	DELAY                      int
	DELAY_TIME                 int
	DELAY_ORDER_DETAILS        DelayDetails
	SLIDING_PRICE              bool
	SWITCH_CH                  = false
)

var (
	DELETE_SPECIFIC_SESSION = 0
	DELETE_SPECIFIC_MEMBER  = 1
	DELETE_ALL              = 2
)

var (
	BUY   = 1
	SELL  = 2
	TRADE = false
)

var (
	CLIENT1 = "CLIENT1"
	CLIENT2 = "CLIENT2"
	CLIENT3 = "CLIENT3"
	CLIENT4 = "CLIENT4"
	CLIENT5 = "CLIENT5"
)

var (
	SQUARE_WAVE    = 1
	SAWTOOTH_WAVE  = 2
	SINE_WAVE      = 3
	STEP_WAVE      = 4
	TRIANGLE_WAVE  = 5
	RANDOM_WAVE    = 6
	PEAK_GENERATOR = 7
)

var (
	UINT64_MAX    = uint64(18446744073709551615)
	DUMMY_NO      = 1234567890
	INTERVAL_TIME = 350
)

var (
	// CONFIG_FILE_PATH = "C:\\Users\\oneture.purushottam\\Desktop\\config\\config.txt"
	CONFIG_FILE_PATH         = "config/config.txt"
	PRODUCTION_RECORDS_FILE  = "C:\\Users\\oneture.purushottam\\Desktop\\config\\prod\\records.txt"
	PRODUCTION_RECORDS1_FILE = "C:\\Users\\oneture.purushottam\\Desktop\\config\\prod\\records1.txt"
	// PRODUCTION_MOCK_RECORDS  = "records/mockRecords.txt"
	PRODUCTION_MOCK_RECORDS = "config/records.txt"
	// PRODUCTION_MOCK_RECORDS = "/home/bse.oneture/server-53/records.txt"
	LAB_RECORDS_FILE = "C:\\Users\\oneture.purushottam\\Desktop\\config\\lab\\records.txt"
	// LAB_RECORDS1_FILE        = "C:\\Users\\oneture.purushottam\\Desktop\\config\\lab\\records1.txt"
	LAB_RECORDS1_FILE       = "records\\recordsLABP2.txt"
	SIMULATION_RECORDS_FILE = "C:\\Users\\oneture.purushottam\\Desktop\\config\\sim\\records.txt"
	// SIMULATION_RECORDS1_FILE = "C:\\Users\\oneture.purushottam\\Desktop\\config\\sim\\records1.txt"
	SIMULATION_RECORDS1_FILE = "records/sim.txt"
)

var (
	PRODUCTION_ENV      = 2
	LAB_ENV             = 1
	SIMULATION_ENV      = 0
	DATA_CHANNEL_LENGTH = 100000
	DUMMY_ID            = 11
	EPOCH_FORMAT        = "2006-01-02 15:04:05.000000000"
)

var SessionIdSets map[string][]Member
