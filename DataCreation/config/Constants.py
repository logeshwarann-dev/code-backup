filePath = r"HistoryDataFile.csv"
sheetName = "Sheet2"
redisDbForData = 5
redisDbForCount = 6

#Variabales for Time Warping
startTimeStr = "01-10-2024 11:33:06"            #input of type "dd:mm:yyyy hh:mm:ss" 
searchTraderId = "*"
searchPartitionId = "*"
searchProductId = "*"
factorOfTimeWarping = 1
duration = 60
redisDbForTimeWarping = 7

#Config File
redisDbForConigFile = 8

#Mapping of template Ids
SingleLegStandardOrder = "10100"
SingleLegLeanOrder = "10125"

#Constants for Parsing Data
partitionId = 1
SingleLegOrderBodyLength = 216
SingleLegTemplateId = 10100
dummyNetMsgId = "AAAAAAAA"
dummyPad2 = "XX"
dummyFiller4 = 13
dummyFiller2 = 1234
stopPx = -9223372036854775808
MaxPricePercentage = -9223372036854775808
senderLocationId = 1234567812345678
expdate = 2147483647
LeanOrderBodyLength = 112
LeanOrderTemplateId = 10125
LeanNetMsgId = "AAAAAAAA"
price = 34000000
locationid = 1234567890123456
clientOrderId = 12345678
side = 1
priceValidityCheck = 0
stpcFlag = 1
ExceInst = 1
AlgoId = ""
ClientCode = "JHJJ" #client id
CpCode = ""
Filler1 = 12345678
regulatoryID = 1234
PartyIDTakeUpTradingFirm = "VVVVV"
PartyIDOrderOriginationFirm = "VVVVVVV"
PartyIDBeneficiary = "VVVVVVVVV"
ApplSeqIndicator = 1
PriceValidityCheckType = 0
ExecInst = 2
RolloverFlag = 0
TradingSessionSubID = 255
TradingCapacity = 1
Account = "A1"
PositionEffect = "C"
PartyIDLocationID = "VV"
CustOrderHandlingInst = "V"
RegulatoryText = ""
AlgoID = ""
FreeText3 = "VV"