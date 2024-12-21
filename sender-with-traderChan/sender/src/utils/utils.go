package utils

import (
	"bytes"
	"fmt"
	"market_replay/src/encryption/aes"
)

func CreateGatewayConnectionRequest(memberId, traderId int, password string) ([]byte, error) {

	buf := new(bytes.Buffer)

	// 1. BodyLen: 4-byte unsigned integer
	bodyLen := uint32(96) // example value
	if err := ParseUnsignedInt(buf, uint64(bodyLen), 4); err != nil {
		return nil, fmt.Errorf("error parsing BodyLen: %w", err)
	}

	// 2. TemplateID: 2-byte unsigned integer
	templateID := uint16(10020)
	if err := ParseUnsignedInt(buf, uint64(templateID), 2); err != nil {
		return nil, fmt.Errorf("error parsing TemplateID: %w", err)
	}

	// 3. NetworkMsgID: 8-byte FixedString, not used (fill with null bytes)
	if err := ParseFixedString(buf, "AAAAAAAA", 8); err != nil {
		return nil, fmt.Errorf("error parsing NetworkMsgID: %w", err)
	}

	// 4. Pad2: 2-byte FixedString, not used (fill with null bytes)
	if err := ParseFixedString(buf, "XX", 2); err != nil {
		return nil, fmt.Errorf("error parsing Pad2: %w", err)
	}

	// 5. MsgSeqNum: 4-byte unsigned integer, must be set to 1
	msgSeqNum := uint32(1)
	if err := ParseUnsignedInt(buf, uint64(msgSeqNum), 4); err != nil {
		return nil, fmt.Errorf("error parsing MsgSeqNum: %w", err)
	}

	// 6. SenderSubID: 4-byte unsigned integer, not used (fill with zero)
	if err := ParseUnsignedInt(buf, 0, 4); err != nil {
		return nil, fmt.Errorf("error parsing SenderSubID: %w", err)
	}

	// 7. PartyIDSessionID: 4-byte unsigned integer, session ID
	partyIDSessionID := uint32((memberId * 100000) + traderId) // example session ID
	if err := ParseUnsignedInt(buf, uint64(partyIDSessionID), 4); err != nil {
		return nil, fmt.Errorf("error parsing PartyIDSessionID: %w", err)
	}

	// 8. DefaultCstmApplVerID: 30-byte FixedString (0-terminable)
	defaultCstmApplVerID := "2.4"
	if err := ParseFixedStringTerminable(buf, defaultCstmApplVerID, 30); err != nil {
		return nil, fmt.Errorf("error parsing DefaultCstmApplVerID: %w", err)
	}

	// 9. Password: 32-byte FixedString (0-terminable)
	pwd := password
	if err := ParseFixedStringTerminable(buf, pwd, 32); err != nil {
		return nil, fmt.Errorf("error parsing Password: %w", err)
	}

	// 10. Pad6: 6-byte FixedString, not used (fill with null bytes)
	if err := ParseFixedString(buf, "XXXXXX", 6); err != nil {
		return nil, fmt.Errorf("error parsing Pad6: %w", err)
	}

	// The byte buffer now contains the entire message to be sent.
	byteData := buf.Bytes()

	return byteData, nil
}

func ConnectionGatewayResponse(data []byte) (*GW_Response, error) {

	res := &GW_Response{}

	buff := bytes.NewBuffer(data)
	buff.Next(32)

	if val, err := ByteToIP(buff.Next(4)); err != nil {
		return nil, fmt.Errorf("error reading GW 1: %w", err)
	} else {
		res.GW1 = val
		// fmt.Println(val, "Primary gateway IP")
	}

	if val, err := ReadUnsignedInt(buff, 4, true); err != nil {
		return nil, fmt.Errorf("error reading Port 1: %w", err)
	} else {
		res.Port1 = uint32(val)
		// fmt.Println(val, "Primary gateway Port")
	}

	if val, err := ByteToIP(buff.Next(4)); err != nil {
		return nil, fmt.Errorf("error reading GW 2: %w", err)
	} else {
		res.GW2 = val
		// fmt.Println(val, "Secondary gateway IP")
	}

	if val, err := ReadUnsignedInt(buff, 4, true); err != nil {
		return nil, fmt.Errorf("error reading Port 2: %w", err)
	} else {
		res.Port2 = uint32(val)
		// fmt.Println(val, "Secondary gateway Port")
	}

	buff.Next(2)

	if sec_key, err := ReadFixedString(buff, 32); err != nil {
		return nil, fmt.Errorf("error reading Sec Key: %w", err)
	} else {
		res.SecKey = sec_key
		// fmt.Println(sec_key, "Key")
	}

	if iv, err := ReadFixedString(buff, 16); err != nil {
		return nil, fmt.Errorf("error reading IV: %w", err)
	} else {
		res.IV = iv
		// fmt.Println(iv, "IV")
	}

	return res, nil
}

func CreateSessionRegistrationRequest(memberId, traderId, msg_seq int) ([]byte, error) {

	buf := new(bytes.Buffer)

	// 1. BodyLen: 4-byte unsigned integer
	bodyLen := uint32(40) // example value
	if err := ParseUnsignedInt(buf, uint64(bodyLen), 4); err != nil {
		return nil, fmt.Errorf("error parsing BodyLen: %w", err)
	}

	// 2. TemplateID: 2-byte unsigned integer
	templateID := uint16(10053)
	if err := ParseUnsignedInt(buf, uint64(templateID), 2); err != nil {
		return nil, fmt.Errorf("error parsing TemplateID: %w", err)
	}

	// 3. NetworkMsgID: 8-byte FixedString, not used (fill with null bytes)
	if err := ParseFixedString(buf, "88888888", 8); err != nil {
		return nil, fmt.Errorf("error parsing NetworkMsgID: %w", err)
	}

	// 4. Pad2: 2-byte FixedString, not used (fill with null bytes)
	if err := ParseFixedString(buf, "XX", 2); err != nil {
		return nil, fmt.Errorf("error parsing Pad2: %w", err)
	}

	// 5. MsgSeqNum: 4-byte unsigned integer, must be set to 1
	msgSeqNum := uint32(msg_seq)
	if err := ParseUnsignedInt(buf, uint64(msgSeqNum), 4); err != nil {
		return nil, fmt.Errorf("error parsing MsgSeqNum: %w", err)
	}

	// 6. SenderSubID: 4-byte unsigned integer, not used (fill with zero)
	if err := ParseUnsignedInt(buf, 0, 4); err != nil {
		return nil, fmt.Errorf("error parsing SenderSubID: %w", err)
	}

	// 7. PartyIDSessionID: 4-byte unsigned integer, session ID
	partyIDSessionID := uint32((memberId * 100000) + traderId) // example session ID
	if err := ParseUnsignedInt(buf, uint64(partyIDSessionID), 4); err != nil {
		return nil, fmt.Errorf("error parsing PartyIDSessionID: %w", err)
	}

	// 8. Pad4: 4-byte FixedString, not used (fill with null bytes)
	if err := ParseFixedString(buf, "XX", 4); err != nil {
		return nil, fmt.Errorf("error parsing Pad6: %w", err)
	}

	// 9. filler1: 8-byte unsigned integer
	filler1 := uint32(0) // example value
	if err := ParseUnsignedInt(buf, uint64(filler1), 8); err != nil {
		return nil, fmt.Errorf("error parsing BodyLen: %w", err)
	}

	byteData := buf.Bytes()

	return byteData, nil
}

func CreateSessionLogonRequest(memberId, traderId, msg_seq int, password, key, iv string, CTX *aes.CipherContext) ([]byte, error) {
	buf := new(bytes.Buffer)

	// 1. BodyLen: 4-byte unsigned integer
	bodyLen := uint32(280) // example value
	if err := ParseUnsignedInt(buf, uint64(bodyLen), 4); err != nil {
		return nil, fmt.Errorf("error parsing BodyLen: %w", err)
	}

	// 2. TemplateID: 2-byte unsigned integer
	templateID := uint16(10000)
	if err := ParseUnsignedInt(buf, uint64(templateID), 2); err != nil {
		return nil, fmt.Errorf("error parsing TemplateID: %w", err)
	}

	// 3. NetworkMsgID: 8-byte FixedString, not used (fill with null bytes)
	if err := ParseFixedString(buf, "88888888", 8); err != nil {
		return nil, fmt.Errorf("error parsing NetworkMsgID: %w", err)
	}

	// 4. Pad2: 2-byte FixedString, not used (fill with null bytes)
	if err := ParseFixedString(buf, "XX", 2); err != nil {
		return nil, fmt.Errorf("error parsing Pad2: %w", err)
	}

	startOffset := buf.Len()

	// 5. MsgSeqNum: 4-byte unsigned integer, must be set to 1
	msgSeqNum := uint32(msg_seq)
	if err := ParseUnsignedInt(buf, uint64(msgSeqNum), 4); err != nil {
		return nil, fmt.Errorf("error parsing MsgSeqNum: %w", err)
	}

	// 6. SenderSubID: 4-byte unsigned integer, not used (fill with zero)
	if err := ParseUnsignedInt(buf, 0, 4); err != nil {
		return nil, fmt.Errorf("error parsing SenderSubID: %w", err)
	}

	// 7. heartBeat: 4-byte unsigned integer
	// heartBeat := uint32(1000)
	fmt.Println("HeartBeat value is set to - ", HEARTBEAT_VALUE)
	if err := ParseUnsignedInt(buf, uint64(HEARTBEAT_VALUE), 4); err != nil {
		return nil, fmt.Errorf("error parsing BodyLen: %w", err)
	}

	// 8. PartyIDSessionID: 4-byte unsigned integer, session ID
	partyIDSessionID := uint32((memberId * 100000) + traderId) // example session ID
	if err := ParseUnsignedInt(buf, uint64(partyIDSessionID), 4); err != nil {
		return nil, fmt.Errorf("error parsing PartyIDSessionID: %w", err)
	}

	// 9. DefaultCstmApplVerID: 30-byte FixedString (0-terminable)
	defaultCstmApplVerID := "2.4"
	if err := ParseFixedStringTerminable(buf, defaultCstmApplVerID, 30); err != nil {
		return nil, fmt.Errorf("error parsing DefaultCstmApplVerID: %w", err)
	}

	// 10. Password: 32-byte FixedString (0-terminable)
	pwd := password
	if err := ParseFixedStringTerminable(buf, pwd, 32); err != nil {
		return nil, fmt.Errorf("error parsing Password: %w", err)
	}

	// 11. ApplUsageOrders : 1-byte char
	applUsageOrders := "M"
	if err := ParseChar(buf, applUsageOrders); err != nil {
		return nil, fmt.Errorf("error parsing applUsageOrders: %w", err)
	}

	// 12. ApplUsageQuotes : 1-byte char
	applUsageQuotes := "N"
	if err := ParseChar(buf, applUsageQuotes); err != nil {
		return nil, fmt.Errorf("error parsing applUsageQuotes: %w", err)
	}

	// 13. OrderRoutingIndicator : 1-byte char
	orderRoutingIndicator := "N"
	if err := ParseChar(buf, orderRoutingIndicator); err != nil {
		return nil, fmt.Errorf("error parsing orderRoutingIndicator: %w", err)
	}

	// 14. FIXEngineName 30-byte FixedString (0-terminable)
	fixEngineName := "A"
	if err := ParseFixedStringTerminable(buf, fixEngineName, 30); err != nil {
		return nil, fmt.Errorf("error parsing fixEngineName: %w", err)
	}

	// 15. FIXEngineVersion 30-byte FixedString (0-terminable)
	fixEngineVersion := "A"
	if err := ParseFixedStringTerminable(buf, fixEngineVersion, 30); err != nil {
		return nil, fmt.Errorf("error parsing fixEngineVersion: %w", err)
	}

	// 16. FIXEngineVendor 30-byte FixedString (0-terminable)
	fixEngineVendor := "A"
	if err := ParseFixedStringTerminable(buf, fixEngineVendor, 30); err != nil {
		return nil, fmt.Errorf("error parsing fixEngineVendor: %w", err)
	}

	// 17. ApplicationSystemName 30-byte FixedString (0-terminable)
	applicationSystemName := "A"
	if err := ParseFixedStringTerminable(buf, applicationSystemName, 30); err != nil {
		return nil, fmt.Errorf("error parsing applicationSystemName: %w", err)
	}

	// 18. ApplicationSystemVersion 30-byte FixedString (0-terminable)
	applicationSystemVersion := "A"
	if err := ParseFixedStringTerminable(buf, applicationSystemVersion, 30); err != nil {
		return nil, fmt.Errorf("error parsing applicationSystemVersion: %w", err)
	}

	// 19. ApplicationSystemVendor 30-byte FixedString (0-terminable)
	applicationSystemVendor := "A"
	if err := ParseFixedStringTerminable(buf, applicationSystemVendor, 30); err != nil {
		return nil, fmt.Errorf("error parsing applicationSystemVendor: %w", err)
	}

	// 20. Pad3: 3-byte FixedString, not used (fill with null bytes)
	if err := ParseFixedString(buf, "AAA", 3); err != nil {
		return nil, fmt.Errorf("error parsing Pad2: %w", err)
	}

	endOffset := buf.Len()
	data2enc := buf.Bytes()[startOffset:endOffset]

	ciphertext, err := CTX.EncryptUpdate(data2enc)
	if err != nil {
		return nil, fmt.Errorf("encryption failed: %w", err)
	}

	copy(buf.Bytes()[startOffset:endOffset], ciphertext)
	byteData := buf.Bytes()

	return byteData, nil
}

func SessionLogonResponse(data []byte) (uint32, error) {

	buff := bytes.NewBuffer(data)

	buff.Next(44)

	val, err := ReadUnsignedInt(buff, 4, true)
	if err != nil {
		return 0, fmt.Errorf("error reading throttle value: %w", err)
	}

	return uint32(val), nil
}

func CreateUserLogonRequest(memberId, traderId, msg_seq int, password, key, iv string, CTX *aes.CipherContext) ([]byte, error) {

	buf := new(bytes.Buffer)

	// 1. BodyLen: 4-byte unsigned integer
	bodyLen := uint32(64) // example value
	if err := ParseUnsignedInt(buf, uint64(bodyLen), 4); err != nil {
		return nil, fmt.Errorf("error parsing BodyLen: %w", err)
	}

	// 2. TemplateID: 2-byte unsigned integer
	templateID := uint16(10018)
	if err := ParseUnsignedInt(buf, uint64(templateID), 2); err != nil {
		return nil, fmt.Errorf("error parsing TemplateID: %w", err)
	}

	// 3. NetworkMsgID: 8-byte FixedString, not used (fill with null bytes)
	if err := ParseFixedString(buf, "AAAAAAAA", 8); err != nil {
		return nil, fmt.Errorf("error parsing NetworkMsgID: %w", err)
	}

	// 4. Pad2: 2-byte FixedString, not used (fill with null bytes)
	if err := ParseFixedString(buf, "XX", 2); err != nil {
		return nil, fmt.Errorf("error parsing Pad2: %w", err)
	}

	startOffset := buf.Len()

	// 5. MsgSeqNum: 4-byte unsigned integer, must be set to 1
	msgSeqNum := uint32(msg_seq)
	if err := ParseUnsignedInt(buf, uint64(msgSeqNum), 4); err != nil {
		return nil, fmt.Errorf("error parsing MsgSeqNum: %w", err)
	}

	// 6. SenderSubID: 4-byte unsigned integer, not used (fill with zero)
	if err := ParseUnsignedInt(buf, 0, 4); err != nil {
		return nil, fmt.Errorf("error parsing SenderSubID: %w", err)
	}

	// 7. PartyIDSessionID: 4-byte unsigned integer, session ID
	partyIDSessionID := uint32((memberId * 100000) + traderId) // example session ID
	if err := ParseUnsignedInt(buf, uint64(partyIDSessionID), 4); err != nil {
		return nil, fmt.Errorf("error parsing PartyIDSessionID: %w", err)
	}

	// 8. Password: 32-byte FixedString (0-terminable)
	pwd := password
	if err := ParseFixedStringTerminable(buf, pwd, 32); err != nil {
		return nil, fmt.Errorf("error parsing Password: %w", err)
	}

	// 9. Pad4: 4-byte FixedString, not used (fill with null bytes)
	if err := ParseFixedString(buf, "XXXX", 4); err != nil {
		return nil, fmt.Errorf("error parsing Pad4: %w", err)
	}

	endOffset := buf.Len()
	data2enc := buf.Bytes()[startOffset:endOffset]

	ciphertext, err := CTX.EncryptUpdate(data2enc)
	if err != nil {
		return nil, fmt.Errorf("encryption failed: %w", err)
	}

	copy(buf.Bytes()[startOffset:endOffset], ciphertext)
	byteData := buf.Bytes()

	return byteData, nil
}

func SingleLegOrder(memberId, traderId, msg_seq int, CTX *aes.CipherContext) ([]byte, error) {

	buf := new(bytes.Buffer)

	// 1. BodyLen: 4-byte unsigned integer
	bodyLen := uint32(216) // example value
	if err := ParseUnsignedInt(buf, uint64(bodyLen), 4); err != nil {
		return nil, fmt.Errorf("error parsing BodyLen: %w", err)
	}

	// 2. TemplateID: 2-byte unsigned integer
	templateID := uint16(10100)
	if err := ParseUnsignedInt(buf, uint64(templateID), 2); err != nil {
		return nil, fmt.Errorf("error parsing TemplateID: %w", err)
	}

	// 3. NetworkMsgID: 8-byte FixedString, not used (fill with null bytes)
	if err := ParseFixedString(buf, "AAAAAAAA", 8); err != nil {
		return nil, fmt.Errorf("error parsing NetworkMsgID: %w", err)
	}

	// 4. Pad2: 2-byte FixedString, not used (fill with null bytes)
	if err := ParseFixedString(buf, "XX", 2); err != nil {
		return nil, fmt.Errorf("error parsing Pad2: %w", err)
	}

	startOffset := buf.Len()

	// 5. MsgSeqNum: 4-byte unsigned integer, must be set to 1
	msgSeqNum := uint32(msg_seq)
	if err := ParseUnsignedInt(buf, uint64(msgSeqNum), 4); err != nil {
		return nil, fmt.Errorf("error parsing MsgSeqNum: %w", err)
	}

	// 6. SenderSubID: 4-byte unsigned integer, not used (fill with zero)
	userId := uint32((memberId * 100000) + traderId)
	if err := ParseUnsignedInt(buf, uint64(userId), 4); err != nil {
		return nil, fmt.Errorf("error parsing SenderSubID: %w", err)
	}

	//7. Price: 8-byte pricetype
	pricetype := 5708000000
	if err := ParsePriceType(buf, int64(pricetype)); err != nil {
		return nil, fmt.Errorf("error parsing pricetype: %w", err)
	}

	//8. StopPx: 8-byte pricetype
	stopPX := int64(-9223372036854775808)
	if err := ParsePriceType(buf, int64(stopPX)); err != nil {
		return nil, fmt.Errorf("error parsing StopPx: %w", err)
	}

	//9. MaxPricePercentage: 8-byte pricetype
	maxPricePercentage := int64(-9223372036854775808)
	if err := ParsePriceType(buf, int64(maxPricePercentage)); err != nil {
		return nil, fmt.Errorf("error parsing StopPx: %w", err)
	}

	//10. SenderLocationID: 8-byte unsigned integer
	if err := ParseUnsignedInt(buf, 1234567812345678, 8); err != nil {
		return nil, fmt.Errorf("error parsing SenderLocationID: %w", err)
	}

	//11. ClOrdID: 8-byte unsigned integer
	if err := ParseUnsignedInt(buf, 12345678, 8); err != nil {
		return nil, fmt.Errorf("error parsing ClOrdID: %w", err)
	}

	//12. Filler1: 8-byte unsigned integer
	if err := ParseUnsignedInt(buf, 12345678, 8); err != nil {
		return nil, fmt.Errorf("error parsing Filler1: %w", err)
	}

	//13. Filler2: 4-byte unsigned integer
	if err := ParseUnsignedInt(buf, 9876, 4); err != nil {
		return nil, fmt.Errorf("error parsing Filler1: %w", err)
	}

	//14. MessageTag: 4-byte signed int
	messageTag := int32(1234)
	if err := ParseSignedInt(buf, int64(messageTag), 4); err != nil {
		return nil, fmt.Errorf("error parsing MessageTag: %w", err)
	}

	//15. OrderQty -4 byte qty
	orderQty := int32(100)
	if err := ParseQty(buf, orderQty); err != nil {
		return nil, fmt.Errorf("error parsing OrderQty: %w", err)
	}

	//16. MaxShow -4 byte qty
	maxShow := int32(0)
	if err := ParseQty(buf, maxShow); err != nil {
		return nil, fmt.Errorf("error parsing MaxShow: %w", err)
	}

	//17. ExpireDate -4 byte LocalMktDate
	expireDate := -1
	if err := ParseLocalMktDate(buf, expireDate); err != nil {
		return nil, fmt.Errorf("error parsing ExpireDate: %w", err)
	}

	//18. MarketSegmentID: 4-byte signed int
	marketSegmentID := int32(247)
	if err := ParseSignedInt(buf, int64(marketSegmentID), 4); err != nil {
		return nil, fmt.Errorf("error parsing MarketSegmentID: %w", err)
	}

	// 19. SimpleSecurityID: 4-byte unsigned integer
	// simpleSecurityID := uint32(542641)
	simpleSecurityID := uint32(542669)
	if err := ParseUnsignedInt(buf, uint64(simpleSecurityID), 4); err != nil {
		return nil, fmt.Errorf("error parsing SimpleSecurityID: %w", err)
	}

	//20. RegulatoryID: 4-byte unsigned integer
	regulatoryID := uint32(1234)
	if err := ParseUnsignedInt(buf, uint64(regulatoryID), 4); err != nil {
		return nil, fmt.Errorf("error parsing RegulatoryID: %w", err)
	}

	//21. Filler4: 2-byte unsigned integer
	filler4 := uint16(12)
	if err := ParseUnsignedInt(buf, uint64(filler4), 2); err != nil {
		return nil, fmt.Errorf("error parsing Filler4: %w", err)
	}

	//22. PartyIDTakeUpTradingFirm: 5-byte FixedString
	if err := ParseFixedString(buf, "VVVVV", 5); err != nil {
		return nil, fmt.Errorf("error parsing PartyIDTakeUpTradingFirm: %w", err)
	}

	//23. PartyIDOrderOriginationFirm: 7-byte FixedString
	if err := ParseFixedString(buf, "VVVVVVV", 7); err != nil {
		return nil, fmt.Errorf("error parsing PartyIDOrderOriginationFirm: %w", err)
	}

	//24. PartyIDBeneficiary: 9-byte FixedString
	if err := ParseFixedString(buf, "VVVVVVVVV", 9); err != nil {
		return nil, fmt.Errorf("error parsing PartyIDBeneficiary: %w", err)
	}

	//25. AccountType: 1-byte unsigned integer
	accountType := uint8(30)
	if err := ParseUnsignedInt(buf, uint64(accountType), 1); err != nil {
		return nil, fmt.Errorf("error parsing AccountType: %w", err)
	}

	//26. ApplSeqIndicator: 1-byte unsigned integer
	applSeqIndicator := uint8(1)
	if err := ParseUnsignedInt(buf, uint64(applSeqIndicator), 1); err != nil {
		return nil, fmt.Errorf("error parsing ApplSeqIndicator: %w", err)
	}

	//27. Side: 1-byte unsigned integer
	side := uint8(1)
	if err := ParseUnsignedInt(buf, uint64(side), 1); err != nil {
		return nil, fmt.Errorf("error parsing Side: %w", err)
	}

	//28. OrdType: 1-byte unsigned integer
	ordType := uint8(2)
	if err := ParseUnsignedInt(buf, uint64(ordType), 1); err != nil {
		return nil, fmt.Errorf("error parsing OrdType: %w", err)
	}

	//29. PriceValidityCheckType: 1-byte unsigned integer
	priceValidityCheckType := uint8(0)
	if err := ParseUnsignedInt(buf, uint64(priceValidityCheckType), 1); err != nil {
		return nil, fmt.Errorf("error parsing PriceValidityCheckType: %w", err)
	}

	//30. TimeInForce: 1-byte unsigned integer
	timeInForce := uint8(7)
	if err := ParseUnsignedInt(buf, uint64(timeInForce), 1); err != nil {
		return nil, fmt.Errorf("error parsing TimeInForce: %w", err)
	}

	//31. ExecInst: 1-byte unsigned integer
	execInst := uint8(2)
	if err := ParseUnsignedInt(buf, uint64(execInst), 1); err != nil {
		return nil, fmt.Errorf("error parsing ExecInst: %w", err)
	}

	//32. STPCFlag: 1-byte unsigned integer
	STPCFlag := uint8(0)
	if err := ParseUnsignedInt(buf, uint64(STPCFlag), 1); err != nil {
		return nil, fmt.Errorf("error parsing STPCFlag: %w", err)
	}

	//33. RolloverFlag: 1-byte unsigned integer
	rolloverFlag := uint8(0)
	if err := ParseUnsignedInt(buf, uint64(rolloverFlag), 1); err != nil {
		return nil, fmt.Errorf("error parsing RolloverFlag: %w", err)
	}

	//34. TradingSessionSubID: 1-byte unsigned integer
	tradingSessionSubID := uint8(255)
	if err := ParseUnsignedInt(buf, uint64(tradingSessionSubID), 1); err != nil {
		return nil, fmt.Errorf("error parsing TradingSessionSubID: %w", err)
	}

	//35. TradingCapacity: 1-byte unsigned integer
	tradingCapacity := uint8(1)
	if err := ParseUnsignedInt(buf, uint64(tradingCapacity), 1); err != nil {
		return nil, fmt.Errorf("error parsing TradingCapacity: %w", err)
	}

	//36. Account: 2-byte FixedString
	if err := ParseFixedString(buf, "A1", 2); err != nil {
		return nil, fmt.Errorf("error parsing Account: %w", err)
	}

	//37. PositionEffect: 1-byte char
	if err := ParseChar(buf, "C"); err != nil {
		return nil, fmt.Errorf("error parsing Account: %w", err)
	}

	//38. PartyIDLocationID: 2-byte FixedString
	if err := ParseFixedString(buf, "VV", 2); err != nil {
		return nil, fmt.Errorf("error parsing PartyIDLocationID: %w", err)
	}

	//39. CustOrderHandlingInst: 1-byte FixedString
	if err := ParseFixedString(buf, "V", 1); err != nil {
		return nil, fmt.Errorf("error parsing CustOrderHandlingInst: %w", err)
	}

	//40. RegulatoryText: 2-byte FixedString
	regulatoryText := ""
	if err := ParseFixedString(buf, regulatoryText, 20); err != nil {
		return nil, fmt.Errorf("error parsing RegulatoryText: %w", err)
	}

	//41. AlgoID: 16-byte FixedString (0-terminable)
	algoID := ""
	if err := ParseFixedStringTerminable(buf, algoID, 16); err != nil {
		return nil, fmt.Errorf("error parsing AlgoID: %w", err)
	}

	//42. ClientCode: 16-byte FixedString (0-terminable)
	clientCode := "JHJJ"
	if err := ParseFixedStringTerminable(buf, clientCode, 12); err != nil {
		return nil, fmt.Errorf("error parsing ClientCode: %w", err)
	}

	//43. CPCode: 12-byte FixedString (0-terminable)
	CPCode := ""
	if err := ParseCharText(buf, CPCode, 12); err != nil {
		return nil, fmt.Errorf("error parsing CPCode: %w", err)
	}

	//44. FreeText3: 12-byte FixedString (0-terminable)
	freeText3 := "VV"
	if err := ParseCharText(buf, freeText3, 12); err != nil {
		return nil, fmt.Errorf("error parsing FreeText3: %w", err)
	}

	endOffset := buf.Len()
	data2enc := buf.Bytes()[startOffset:endOffset]

	ciphertext, err := CTX.EncryptUpdate(data2enc)
	if err != nil {
		return nil, fmt.Errorf("encryption failed: %w", err)
	}

	copy(buf.Bytes()[startOffset:endOffset], ciphertext)
	byteData := buf.Bytes()

	return byteData, nil
}

func SingleLegLeanOrder(memberId, traderId, msg_seq, inst_id, price int, CTX *aes.CipherContext) ([]byte, error) {

	buf := new(bytes.Buffer)

	// 1. BodyLen: 4-byte unsigned integer
	bodyLen := uint32(112) // example value
	if err := ParseUnsignedInt(buf, uint64(bodyLen), 4); err != nil {
		return nil, fmt.Errorf("error parsing BodyLen: %w", err)
	}

	// 2. TemplateID: 2-byte unsigned integer
	templateID := uint16(10125)
	if err := ParseUnsignedInt(buf, uint64(templateID), 2); err != nil {
		return nil, fmt.Errorf("error parsing TemplateID: %w", err)
	}

	// 3. NetworkMsgID: 8-byte FixedString, not used (fill with null bytes)
	if err := ParseFixedString(buf, "AAAAAAAA", 8); err != nil {
		return nil, fmt.Errorf("error parsing NetworkMsgID: %w", err)
	}

	// 4. Pad2: 2-byte FixedString, not used (fill with null bytes)
	if err := ParseFixedString(buf, "XX", 2); err != nil {
		return nil, fmt.Errorf("error parsing Pad2: %w", err)
	}

	startOffset := buf.Len()

	// 5. MsgSeqNum: 4-byte unsigned integer, must be set to 1
	msgSeqNum := uint32(msg_seq)
	if err := ParseUnsignedInt(buf, uint64(msgSeqNum), 4); err != nil {
		return nil, fmt.Errorf("error parsing MsgSeqNum: %w", err)
	}

	// 6. SenderSubID: 4-byte unsigned integer, not used (fill with zero)
	userId := uint32((memberId * 100000) + traderId)
	if err := ParseUnsignedInt(buf, uint64(userId), 4); err != nil {
		return nil, fmt.Errorf("error parsing SenderSubID: %w", err)
	}

	//7. Price: 8-byte pricetype
	// pricetype := 34000000
	// pricetype := 690000000
	pricetype := price * 1000000
	if err := ParsePriceType(buf, int64(pricetype)); err != nil {
		return nil, fmt.Errorf("error parsing pricetype: %w", err)
	}

	// 8. SenderLocationID: 8-byte unsigned integer
	location_id := 1234567890123456
	if err := ParseUnsignedInt(buf, uint64(location_id), 8); err != nil {
		return nil, fmt.Errorf("error parsing SenderLocationID: %w", err)
	}

	// 9. ClOrdID: 8-byte unsigned integer
	if err := ParseUnsignedInt(buf, 18446744073709551615, 8); err != nil {
		return nil, fmt.Errorf("error parsing ClOrdID: %w", err)
	}

	// 10. OrderQty -4 byte qty
	orderQty := int32(10)
	// orderQty := int32(2625)
	if err := ParseQty(buf, orderQty); err != nil {
		return nil, fmt.Errorf("error parsing OrderQty: %w", err)
	}

	// 11. MaxShow -4 byte qty
	maxShow := int32(10)
	// maxShow := int32(2625)
	if err := ParseQty(buf, maxShow); err != nil {
		return nil, fmt.Errorf("error parsing MaxShow: %w", err)
	}

	// 12. SimpleSecurityID: 4-byte unsigned integer
	// simpleSecurityID := uint32(542641)
	// simpleSecurityID := uint32(542579)
	simpleSecurityID := uint32(inst_id)
	if err := ParseUnsignedInt(buf, uint64(simpleSecurityID), 4); err != nil {
		return nil, fmt.Errorf("error parsing SimpleSecurityID: %w", err)
	}

	// 13. Filler2: 4-byte unsigned integer
	if err := ParseUnsignedInt(buf, 1234, 4); err != nil {
		return nil, fmt.Errorf("error parsing Filler1: %w", err)
	}

	// 14. Filler4: 4-byte unsigned integer
	filler4 := uint32(13)
	if err := ParseUnsignedInt(buf, uint64(filler4), 2); err != nil {
		return nil, fmt.Errorf("error parsing Filler4: %w", err)
	}

	// 15. AccountType: 1-byte unsigned integer
	accountType := uint32(30)
	if err := ParseUnsignedInt(buf, uint64(accountType), 1); err != nil {
		return nil, fmt.Errorf("error parsing AccountType: %w", err)
	}

	// 16. Side: 1-byte unsigned integer
	side := uint16(1)
	if err := ParseUnsignedInt(buf, uint64(side), 1); err != nil {
		return nil, fmt.Errorf("error parsing Side: %w", err)
	}

	// 17. PriceValidityCheckType: 1-byte unsigned integer
	priceValidityCheckType := uint8(0)
	if err := ParseUnsignedInt(buf, uint64(priceValidityCheckType), 1); err != nil {
		return nil, fmt.Errorf("error parsing PriceValidityCheckType: %w", err)
	}

	// 18. TimeInForce: 1-byte unsigned integer
	timeInForce := uint8(7)
	if err := ParseUnsignedInt(buf, uint64(timeInForce), 1); err != nil {
		return nil, fmt.Errorf("error parsing TimeInForce: %w", err)
	}

	// 19. STPCFlag: 1-byte unsigned integer
	STPCFlag := uint16(1)
	if err := ParseUnsignedInt(buf, uint64(STPCFlag), 1); err != nil {
		return nil, fmt.Errorf("error parsing STPCFlag: %w", err)
	}

	// 20. ExecInst: 1-byte unsigned integer
	execInst := uint16(1)
	if err := ParseUnsignedInt(buf, uint64(execInst), 1); err != nil {
		return nil, fmt.Errorf("error parsing ExecInst: %w", err)
	}

	// 21. AlgoID: 16-byte FixedString (0-terminable)
	algoID := ""
	if err := ParseFixedStringTerminable(buf, algoID, 16); err != nil {
		return nil, fmt.Errorf("error parsing AlgoID: %w", err)
	}

	// 22. ClientCode: 16-byte FixedString (0-terminable)
	clientCode := "CLIENT3"
	if err := ParseFixedStringTerminable(buf, clientCode, 12); err != nil {
		return nil, fmt.Errorf("error parsing ClientCode: %w", err)
	}

	// 23. CPCode: 12-byte FixedString (0-terminable)
	CPCode := ""
	if err := ParseCharText(buf, CPCode, 12); err != nil {
		return nil, fmt.Errorf("error parsing CPCode: %w", err)
	}

	endOffset := buf.Len()
	data2enc := buf.Bytes()[startOffset:endOffset]

	ciphertext, err := CTX.EncryptUpdate(data2enc)
	if err != nil {
		return nil, fmt.Errorf("encryption failed: %w", err)
	}

	copy(buf.Bytes()[startOffset:endOffset], ciphertext)
	byteData := buf.Bytes()

	return byteData, nil
}

func SingleLegLeanOrderResponse(data []byte, CTX *aes.CipherContext) (uint64, uint64, error) {

	buff := bytes.NewBuffer(data)

	buff.Next(48)

	// MsgSeq
	msg_seq, err := ReadUnsignedInt(buff, 4, true)
	if err != nil {
		return 0, 0, fmt.Errorf("MSG SEQ: %w", err)
	}

	fmt.Println("[INFO] Message Sequence Number: ", msg_seq)

	buff.Next(4)

	// OrderId
	order_id, err := ReadUnsignedInt(buff, 8, true)
	if err != nil {
		return 0, 0, fmt.Errorf("Order Id: %w", err)
	}

	// fmt.Println(order_id, "ORDER ID LEAN ORDER")

	buff.Next(48)
	activity_time, err := ReadUTCTimestamp(buff, true)
	if err != nil {
		return 0, 0, fmt.Errorf("Order Id: %w", err)
	}

	// fmt.Println(activity_time, "ACT TIME")

	return order_id, activity_time, nil
}

func ModifyLeanOrder(memberId, traderId, msg_seq, inst_id, price, qty int, order_id, act_time uint64, CTX *aes.CipherContext) ([]byte, error) {

	buf := new(bytes.Buffer)

	// 1. BodyLen: 4-byte unsigned integer
	bodyLen := uint32(136)
	if err := ParseUnsignedInt(buf, uint64(bodyLen), 4); err != nil {
		return nil, fmt.Errorf("error parsing BodyLen: %w", err)
	}

	// 2. TemplateID: 2-byte unsigned integer
	templateID := uint16(10126)
	if err := ParseUnsignedInt(buf, uint64(templateID), 2); err != nil {
		return nil, fmt.Errorf("error parsing TemplateID: %w", err)
	}

	// 3. NetworkMsgID: 8-byte FixedString, not used (fill with null bytes)
	if err := ParseFixedString(buf, "AAAAAAAA", 8); err != nil {
		return nil, fmt.Errorf("error parsing NetworkMsgID: %w", err)
	}

	// 4. Pad2: 2-byte FixedString, not used (fill with null bytes)
	if err := ParseFixedString(buf, "XX", 2); err != nil {
		return nil, fmt.Errorf("error parsing Pad2: %w", err)
	}

	startOffset := buf.Len()

	// 5. MsgSeqNum: 4-byte unsigned integer, must be set to 1
	msgSeqNum := uint32(msg_seq)
	if err := ParseUnsignedInt(buf, uint64(msgSeqNum), 4); err != nil {
		return nil, fmt.Errorf("error parsing MsgSeqNum: %w", err)
	}

	// 6. SenderSubID: 4-byte unsigned integer, not used (fill with zero)
	userId := uint32((memberId * 100000) + traderId)
	if err := ParseUnsignedInt(buf, uint64(userId), 4); err != nil {
		return nil, fmt.Errorf("error parsing SenderSubID: %w", err)
	}

	// 7. OrderID: 8-byte unsigned integer
	orderId := order_id
	if err := ParseUnsignedInt(buf, uint64(orderId), 8); err != nil {
		return nil, fmt.Errorf("error parsing OrderID: %w", err)
	}

	//8. ClOrdID: 8-byte unsigned integer
	// clOrdID := 12345678
	if err := ParseUnsignedInt(buf, uint64(18446744073709551615), 8); err != nil {
		return nil, fmt.Errorf("error parsing ClOrdID: %w", err)
	}

	//9. OrderID: 8-byte unsigned integer
	// origClOrdID := 12345678
	if err := ParseUnsignedInt(buf, uint64(18446744073709551615), 8); err != nil {
		return nil, fmt.Errorf("error parsing OrigClOrdID: %w", err)
	}

	//10. Price: 8-byte pricetype
	pricetype := price * 1000000
	if err := ParsePriceType(buf, int64(pricetype)); err != nil {
		return nil, fmt.Errorf("error parsing pricetype: %w", err)
	}

	// 11. SenderLocationID: 8-byte unsigned integer
	location_id := 1234567890123456
	if err := ParseUnsignedInt(buf, uint64(location_id), 8); err != nil {
		return nil, fmt.Errorf("error parsing SenderLocationID: %w", err)
	}

	// 12. ActivityTime: 8-byte UTCTimestamp
	if err := ParseUTCTimestamp(buf, uint64(act_time)); err != nil {
		return nil, fmt.Errorf("error parsing ActivityTime: %w", err)
	}

	// 13. OrderQty -4 byte qty
	orderQty := int32(10)
	if err := ParseQty(buf, orderQty); err != nil {
		return nil, fmt.Errorf("error parsing OrderQty: %w", err)
	}

	// 14. MaxShow -4 byte qty
	maxShow := int32(10)
	if err := ParseQty(buf, maxShow); err != nil {
		return nil, fmt.Errorf("error parsing MaxShow: %w", err)
	}

	// 15. SimpleSecurityID: 4-byte unsigned integer
	// simpleSecurityID := uint32(542641)
	// simpleSecurityID := uint32(542579)
	simpleSecurityID := uint32(inst_id)
	if err := ParseUnsignedInt(buf, uint64(simpleSecurityID), 4); err != nil {
		return nil, fmt.Errorf("error parsing SimpleSecurityID: %w", err)
	}

	// 13. Filler2: 4-byte unsigned integer
	if err := ParseUnsignedInt(buf, 1234, 4); err != nil {
		return nil, fmt.Errorf("error parsing Filler1: %w", err)
	}

	// 14. Filler4: 4-byte unsigned integer
	filler4 := uint32(13)
	if err := ParseUnsignedInt(buf, uint64(filler4), 2); err != nil {
		return nil, fmt.Errorf("error parsing Filler4: %w", err)
	}

	// 15. AccountType: 1-byte unsigned integer
	accountType := uint32(30)
	if err := ParseUnsignedInt(buf, uint64(accountType), 1); err != nil {
		return nil, fmt.Errorf("error parsing AccountType: %w", err)
	}

	// 16. Side: 1-byte unsigned integer
	side := uint16(1)
	if err := ParseUnsignedInt(buf, uint64(side), 1); err != nil {
		return nil, fmt.Errorf("error parsing Side: %w", err)
	}

	// 17. PriceValidityCheckType: 1-byte unsigned integer
	priceValidityCheckType := uint8(0)
	if err := ParseUnsignedInt(buf, uint64(priceValidityCheckType), 1); err != nil {
		return nil, fmt.Errorf("error parsing PriceValidityCheckType: %w", err)
	}

	// 18. TimeInForce: 1-byte unsigned integer
	timeInForce := uint8(7)
	if err := ParseUnsignedInt(buf, uint64(timeInForce), 1); err != nil {
		return nil, fmt.Errorf("error parsing TimeInForce: %w", err)
	}

	// 19. ExecInst: 1-byte unsigned integer
	execInst := uint16(1)
	if err := ParseUnsignedInt(buf, uint64(execInst), 1); err != nil {
		return nil, fmt.Errorf("error parsing ExecInst: %w", err)
	}

	// 20. AlgoID: 16-byte FixedString (0-terminable)
	algoID := ""
	if err := ParseFixedStringTerminable(buf, algoID, 16); err != nil {
		return nil, fmt.Errorf("error parsing AlgoID: %w", err)
	}

	// 21. ClientCode: 16-byte FixedString (0-terminable)
	clientCode := "JHJJ"
	if err := ParseFixedStringTerminable(buf, clientCode, 12); err != nil {
		return nil, fmt.Errorf("error parsing ClientCode: %w", err)
	}

	// 22. CPCode: 12-byte FixedString (0-terminable)
	CPCode := ""
	if err := ParseCharText(buf, CPCode, 12); err != nil {
		return nil, fmt.Errorf("error parsing CPCode: %w", err)
	}

	if err := ParseFixedString(buf, "X", 1); err != nil {
		return nil, fmt.Errorf("error parsing CPCode: %w", err)
	}

	endOffset := buf.Len()
	data2enc := buf.Bytes()[startOffset:endOffset]

	ciphertext, err := CTX.EncryptUpdate(data2enc)
	if err != nil {
		return nil, fmt.Errorf("encryption failed: %w", err)
	}

	copy(buf.Bytes()[startOffset:endOffset], ciphertext)
	byteData := buf.Bytes()

	return byteData, nil
}

func SingleLegLeanOrderModifiedResponse(data []byte, CTX *aes.CipherContext) (uint64, uint64, error) {

	buff := bytes.NewBuffer(data)

	buff.Next(48)

	// MsgSeq
	msg_seq, err := ReadUnsignedInt(buff, 4, true)
	if err != nil {
		return 0, 0, fmt.Errorf("MSG SEQ: %w", err)
	}

	fmt.Println("[INFO] Message Sequence Number: ", msg_seq)

	buff.Next(4)

	// OrderId
	order_id, err := ReadUnsignedInt(buff, 8, true)
	if err != nil {
		return 0, 0, fmt.Errorf("Order Id: %w", err)
	}

	// fmt.Println(order_id, "ORDER ID")

	buff.Next(56)
	activity_time, err := ReadUTCTimestamp(buff, true)
	if err != nil {
		return 0, 0, fmt.Errorf("Order Id: %w", err)
	}

	// fmt.Println(activity_time, "ACT TIME")

	return order_id, activity_time, nil
}

func CancelSingleLegOrderRequest(memberId, traderId, msg_seq, inst_id, product_id int, order_id uint64, CTX *aes.CipherContext) ([]byte, error) {
	buf := new(bytes.Buffer)

	// 1. BodyLen: 4-byte unsigned integer
	bodyLen := uint32(96) // example value
	if err := ParseUnsignedInt(buf, uint64(bodyLen), 4); err != nil {
		return nil, fmt.Errorf("error parsing BodyLen: %w", err)
	}

	// 2. TemplateID: 2-byte unsigned integer
	templateID := uint16(10109)
	if err := ParseUnsignedInt(buf, uint64(templateID), 2); err != nil {
		return nil, fmt.Errorf("error parsing TemplateID: %w", err)
	}

	// 3. NetworkMsgID: 8-byte FixedString, not used (fill with null bytes)
	if err := ParseFixedString(buf, "AAAAAAAA", 8); err != nil {
		return nil, fmt.Errorf("error parsing NetworkMsgID: %w", err)
	}

	// 4. Pad2: 2-byte FixedString, not used (fill with null bytes)
	if err := ParseFixedString(buf, "XX", 2); err != nil {
		return nil, fmt.Errorf("error parsing Pad2: %w", err)
	}

	startOffset := buf.Len()

	// 5. MsgSeqNum: 4-byte unsigned integer, must be set to 1
	msgSeqNum := uint32(msg_seq)
	if err := ParseUnsignedInt(buf, uint64(msgSeqNum), 4); err != nil {
		return nil, fmt.Errorf("error parsing MsgSeqNum: %w", err)
	}

	// 6. SenderSubID: 4-byte unsigned integer, not used (fill with zero)
	userId := uint32((memberId * 100000) + traderId)
	if err := ParseUnsignedInt(buf, uint64(userId), 4); err != nil {
		return nil, fmt.Errorf("error parsing SenderSubID: %w", err)
	}

	// 7. OrderID: 8-byte unsigned integer
	orderId := order_id
	if err := ParseUnsignedInt(buf, uint64(orderId), 8); err != nil {
		return nil, fmt.Errorf("error parsing OrderID: %w", err)
	}

	//8. ClOrdID: 8-byte unsigned integer
	// clOrdID := 18446744073709551615
	if err := ParseUnsignedInt(buf, uint64(18446744073709551615), 8); err != nil {
		return nil, fmt.Errorf("error parsing ClOrdID: %w", err)
	}

	//9. OrgiOrderID: 8-byte unsigned integer
	// origClOrdID := 18446744073709551615
	if err := ParseUnsignedInt(buf, uint64(18446744073709551615), 8); err != nil {
		return nil, fmt.Errorf("error parsing OrigClOrdID: %w", err)
	}

	// 10. Filler1: 8-byte unsigned integer
	if err := ParseUnsignedInt(buf, 1234, 8); err != nil {
		return nil, fmt.Errorf("error parsing Filler1: %w", err)
	}

	// 11. MessageTag: 4-byte unsigned integer
	if err := ParseSignedInt(buf, int64(1234), 4); err != nil {
		fmt.Println("Error : ", err)
		return nil, err
	}

	// 12. MarketSegmentID: 4-byte signed integer
	marketSegmentID := product_id
	marketSegmentID = -2147483648
	if err := ParseSignedInt(buf, int64(marketSegmentID), 4); err != nil {
		fmt.Println("Error : ", err)
		return nil, err
	}

	// 13. SimpleSecurityID: 4-byte unsigned integer
	simpleSecurityID := uint32(inst_id)
	if err := ParseUnsignedInt(buf, uint64(simpleSecurityID), 4); err != nil {
		return nil, fmt.Errorf("error parsing SimpleSecurityID: %w", err)
	}

	// 14. TargetPartyIDSessionID: 4-byte unsigned integer
	targetPartyIdSessionId := uint32((memberId * 100000) + traderId)
	if err := ParseUnsignedInt(buf, uint64(targetPartyIdSessionId), 4); err != nil {
		fmt.Println("error : ", err)
		return nil, err
	}

	// 15. RegulatoryID: 4-byte unsigned integer
	regulatoryID := 1234
	if err := ParseUnsignedInt(buf, uint64(regulatoryID), 4); err != nil {
		fmt.Println("Error ", err)
		return nil, err
	}

	// 16. AlgoID: 16-byte FixedString (0-terminable)
	algoID := ""
	if err := ParseFixedStringTerminable(buf, algoID, 16); err != nil {
		return nil, fmt.Errorf("error parsing AlgoID: %w", err)
	}

	// 17. Pad4: 4-byte FixedString, not used (fill with null bytes)
	if err := ParseFixedString(buf, "XXXX", 4); err != nil {
		return nil, fmt.Errorf("error parsing Pad2: %w", err)
	}

	endOffset := buf.Len()
	data2enc := buf.Bytes()[startOffset:endOffset]

	ciphertext, err := CTX.EncryptUpdate(data2enc)
	if err != nil {
		return nil, fmt.Errorf("encryption failed: %w", err)
	}

	copy(buf.Bytes()[startOffset:endOffset], ciphertext)
	byteData := buf.Bytes()

	return byteData, nil
}

func CancelSingleLegOrderResponse(data []byte, CTX *aes.CipherContext) (uint64, uint64, string, error) {

	buff := bytes.NewBuffer(data)

	buff.Next(48)

	// MsgSeq
	msg_seq, err := ReadUnsignedInt(buff, 4, true)
	if err != nil {
		return 0, 0, "", fmt.Errorf("MSG SEQ: %w", err)
	}

	fmt.Println("[INFO] Message Sequence Number: ", msg_seq)

	buff.Next(4)

	// OrderId
	order_id, err := ReadUnsignedInt(buff, 8, true)
	if err != nil {
		return 0, 0, "", fmt.Errorf("Order Id: %w", err)
	}

	buff.Next(16)

	//InstId
	instId, err := ReadSignedInt(buff, 8, true)
	if err != nil {
		return 0, 0, "", fmt.Errorf("Instrument Id: %w", err)
	}

	buff.Next(16)

	//OrderStatus
	OrdStatus, err := ReadChar(buff)
	if err != nil {
		return 0, 0, "", fmt.Errorf("Order Status: %w", err)
	}

	return order_id, uint64(instId), OrdStatus, nil
}

func HeartBeatRequest() ([]byte, error) {
	buf := new(bytes.Buffer)

	// 1. BodyLen: 4-byte unsigned integer
	bodyLen := uint32(16) // example value
	if err := ParseUnsignedInt(buf, uint64(bodyLen), 4); err != nil {
		return nil, fmt.Errorf("error parsing BodyLen: %w", err)
	}

	// 2. TemplateID: 2-byte unsigned integer
	templateID := uint16(10011)
	if err := ParseUnsignedInt(buf, uint64(templateID), 2); err != nil {
		return nil, fmt.Errorf("error parsing TemplateID: %w", err)
	}

	// 3. NetworkMsgID: 8-byte FixedString, not used (fill with null bytes)
	if err := ParseFixedString(buf, "AAAAAAAA", 8); err != nil {
		return nil, fmt.Errorf("error parsing NetworkMsgID: %w", err)
	}

	// 4. Pad2: 2-byte FixedString, not used (fill with null bytes)
	if err := ParseFixedString(buf, "XX", 2); err != nil {
		return nil, fmt.Errorf("error parsing Pad2: %w", err)
	}

	byteData := buf.Bytes()
	return byteData, nil
}
