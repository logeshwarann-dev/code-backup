package utils

var (
	BATCH_SIZE     = int64(10)
	START_TIME_STR = "01-10-2024 11:33:07"
	END_TIME_STR   = "01-10-2024 11:49:46"
	LAYOUT         = "02-01-2006 15:04:05"
)

var (
	Traders_DB = []DB{
		// {Set: 0, Addr: "", session_ids: MEMBER_SET["set0"]},
		// {Set: 2, Addr: "192.168.52.8:6379", session_ids: []int{103501001, 103501002}},
	}
)

var (
	MEMBER_SET    = make(map[string]map[string]string)
	INSTRUMENTS   = make(map[string]string)
	MARKET_REPLAY bool
	// CONFIG           = make(map[string]string)
	SET_1_SESSION_ID = []int{103401001, 103401002, 103401003, 103401004, 103401005, 103401006, 103401007, 103401008, 103401009, 103401010, 103401011, 103401012, 103401013, 103401014, 103401015, 103401016, 103401017, 103401018, 103401019, 103401020, 103401021, 103401022, 103401023, 103401024, 103401025, 103401026, 103401027, 103401028, 103401029, 103401030, 103401031, 103401032, 103401033, 103401034, 103401035, 103401036, 103401037, 103401038, 103401039, 103401040, 103401041, 103401042, 103401043, 103401044, 103401045, 103401046, 103401047, 103401048, 103401049, 103401050}
)

var (
	KAFKA_BROKER = "44.204.236.162:9092"
	KAFKA_TOPIC  = "SET1"
)

var CONFIG_VALUES = Config{
	CancelPercent:            5,
	Duration:                 120,
	EndTime:                  "01-10-2024 11:33:18",
	EnvFile:                  1,
	Environment:              1,
	Heartbeat:                1000,
	MarketReplay:             1,
	ModifyPercent:            90,
	RedisTimestampsBatchSize: 10,
	StartTime:                "01-10-2024 11:33:06",
	Throttle:                 1000,
	Traders:                  50,
	Instruments:              []int{838992, 837698, 834497},
}
