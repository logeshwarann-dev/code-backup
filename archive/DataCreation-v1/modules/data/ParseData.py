from formatter import *
import random
from config.Constants import *
from modules.structs import RedisData
from datetime import datetime, timezone

def generateNewSingleLegOrder(buf,msg_seq,row):
    # row = {}

    messageSeqNum = msg_seq

    templateId = SingleLegTemplateId
    
    # NetworkMessageID
    try:
        networkmsgid = row["NetworkMessageID"]
        if networkmsgid == "":
            print("NetWork Message Id is empty")
            return None,None,None
    except KeyError:  
        print("NetWork Message Id not found")
        return None,None,None
    

    # SenderSubId / UserId(TraderId)
    try:
        traderid = int(row["Trading Member Id"])*100000 + int(row["Trader Id"])
        if traderid == "":
            print("Error UserId was null")
            return None,None,None
    except KeyError:  
        print("Error UserId was Not found")
        return None,None,None
    
    try:
        dateTimestamp = row["Transaction Timestamp"]
        if dateTimestamp == "":
            print("Transaction Timestamp is empty")
            return None,None,None
        transactionTimestamp = int(datetime.strptime(dateTimestamp, '%Y-%m-%d %H:%M:%S').replace(tzinfo=timezone.utc).timestamp())
    except KeyError:  
        print("Transaction Timestamp not found")
        return None,None,None
    
    try:
        productID = row["Exchange Trading  Product Id"]
        if productID == "":
            productID = 247
            # print("Exchange Trading  Product Id is empty")
            # return None,None,None
    except KeyError:  
        print("Exchange Trading  Product Id not found")
        return None,None,None
    
    try:
        memberId = row["Trading Member Id"]
        if memberId == "":
            print("Trading Member Id is empty")
            return None,None,None
    except KeyError:  
        print("Trading Member Id not found")
        return None,None,None
        
    try:
        actualTraderId = row["Trader Id"]
        if actualTraderId == "":
            print("Trader Id is empty")
            return None,None,None
    except KeyError:  
        print("Trader Id not found")
        return None,None,None
    

    # Price
    try:
        price = row["Order Price"]
        if price == "":
            print("Price empty")
            return None,None,None
    except KeyError:
        print("Price not found")
        return None,None,None

    # Order Quantity
    try:
        ordqty = row["Order Quantity"]
        if ordqty == "":
            print("order Quantity empty")
            return None,None,None
        else:
            parse_qty(buf,ordqty)
    except KeyError:  
        print("Order Quantity not found")
        return None,None,None

    # MaxShow
    try:
        maxShow = row["Disclosed Quantity"]
        if maxShow == "":
            print("Max show is empty")
            return None,None,None
    except KeyError:
        print("MAX show not found")
        return None,None,None
                    
    # MarketSegmentID / Product ID
    productId = 247
    try:
        productId = row["Exchange Trading  Product Id"]
        if productId == "":
            print("product id is empty")
            return None,None,None
    except KeyError:
        print("Product Id not found")
        return None,None,None
    
    try:
        Partition_Id = row["Partition Id"]
        if Partition_Id == "":
            Partition_Id = 0
            # print("product id is empty")
            return None,None,None
    except KeyError:
        Partition_Id = 0
        #print("Product Id not found")
        # return None,None,None

    # SimpleSecurityID
    try :
        instid = row["Instrument ID"]
        if instid == "":
            print("Instd id is empty")
            return None,None,None
    except KeyError:
        print("Inst id not found")
        return None,None,None

    # AccountType 
    try:
        accountType = row["CA Category"]
        if accountType == "":
            print("Account Type is empty")
            return None,None,None
    except KeyError:
        print("Account type not found")
        return None,None,None

    # TimeInForce 
    try:
        timeInForce = row["Retention"]
        if timeInForce == "" :
            print("Time In force is empty")
            return None,None,None
    except KeyError:
        print("Time In force not found")
        return None,None,None  

    try:
        transactionType = row["Transaction Type"]
        if transactionType == "":
            print("Error in Transaction Type")
            return None,None,None
    except KeyError:
        print("Transaction Type not Found")
        return None,None,None
    

    #Order Number
    try:
        orderNumber = row["Order Number"]
        if orderNumber == "":
            print("Error in Order Number")
            return None,None,None
    except KeyError:
        print("Order Number not Found")
        return None,None,None
    
    data = RedisData(
        template_id=templateId,
        price=price,
        order_qty=ordqty,
        maxshow=maxShow,
        acc_type=accountType,
        instid=instid,
        time_in_force=timeInForce,
        clientCode=ClientCode,
        MsgSeqNum = messageSeqNum,
        transactionType= transactionType,
        OrderNumber=float(orderNumber),
        TransactionTimeStamp=transactionTimestamp,
        TradingProductId=productID,
        memberID=memberId,
        actualTraderID=actualTraderId,
        PartitionID=Partition_Id
    )

    return data, traderid, productId






def singleLegLeanOrder(msg_seq,row):
                  
    # row = {}

    #Template id fix to 10125
    templateId = LeanOrderTemplateId

    #Message Sequence Number
    messageSeqNum = msg_seq

    #SenderSubID 
    try:
        traderid = int(row["Trading Member Id"])*100000 + int(row["Trader Id"])
        if traderid == "":
            print("User Id empty")
            return None,None, None
    except KeyError:  
        print("User Id not found")
        return None,None, None
    
    try:
        dateTimestamp = row["Transaction Timestamp"]
        if dateTimestamp == "":
            print("Transaction Timestamp is empty")
            return None,None,None
        transactionTimestamp = int(datetime.strptime(dateTimestamp, '%Y-%m-%d %H:%M:%S').replace(tzinfo=timezone.utc).timestamp())
    except KeyError:  
        print("Transaction Timestamp not found")
        return None,None,None

    try:
        productID = row["Exchange Trading  Product Id"]
        if productID == "":
            productID=-2147483648
            # print("Exchange Trading  Product Id is empty")
            # return None,None,None
    except KeyError:  
        print("Exchange Trading  Product Id not found")
        return None,None,None
    
    try:
        memberId = row["Trading Member Id"]
        if memberId == "":
            print("Trading Member Id is empty")
            return None,None,None
    except KeyError:  
        print("Trading Member Id not found")
        return None,None,None
        
    try:
        actualTraderId = row["Trader Id"]
        if actualTraderId == "":
            print("Trader Id is empty")
            return None,None,None
    except KeyError:  
        print("Trader Id not found")
        return None,None,None

    try:
        Partition_Id = row["Partition Id"]
        if Partition_Id == "":
            Partition_Id = 0
            # print("product id is empty")
            return None,None,None
    except KeyError:
        Partition_Id = 0
        #print("Product Id not found")
        # return None,None,None
    
    #Order Price
    try:
        price = row["Order Price"]
        if price == "":
            print("Price is empty")
            return None,None, None
    except KeyError:
        print("Price not found")
        return None,None, None
    
    #total order quantity
    try:
        ordqty = row["Order Quantity"]
        if ordqty == "":
            print("order Quantity empty")
            return None,None , None
    except KeyError:  
        print("Order Quantity not found")
        return None,None, None

    #maxshow
    try:
        maxShow = row["Disclosed Quantity"]
        if maxShow == "":
            print("Max show is empty")
            return None,None, None
    except KeyError:
        print("MAX show not found")
        return None,None, None
            

    #simpleSecurityID
    try :
        instid = row["Instrument ID"]
        if instid == "":
            print("Instd id is empty")
            return None,None, None
    except KeyError:
        print("Inst id not found")
        return None,None, None

    #AccountType
    try:
        accountType = row["CA Category"]
        if accountType == "":
            print("Account Type is empty")
            return None,None, None
    except KeyError:
        print("Account type not found")
        return None,None, None

    
    #TimeInforce
    try:
        timeInForce = row["Retention"]
        if timeInForce == "" :
            print("Time In force is empty")
            return None,None, None
    except KeyError:
        print("Time In force not found")
        return None,None, None

    #ClientCode
    try:
        ClientCode = row["Client Id"]
        if ClientCode == "":
            print("Client Code is empty")
            return None,None, None
    except KeyError:
        print("Client Code Not found")
        return None,None, None
    
    #transaction Type
    try:
        transactionType = row["Transaction Type"]
        if transactionType == "":
            print("Error in Transaction Type")
            return None,None,None
    except KeyError:
        print("Transaction Type not Found")
        return None,None,None


    #Order Number
    try:
        orderNumber = row["Order Number"]
        if orderNumber == "":
            print("Error in Order Number")
            return None,None,None
    except KeyError:
        print("Order Number not Found")
        return None,None,None
    
    data = RedisData(
        template_id=templateId,
        price=price,
        order_qty=ordqty,
        maxshow=maxShow,
        acc_type=accountType,
        instid=instid,
        time_in_force=timeInForce,
        clientCode=ClientCode,
        MsgSeqNum = messageSeqNum,
        transactionType= transactionType,
        OrderNumber=float(orderNumber),
        TransactionTimeStamp=transactionTimestamp,
        TradingProductId=productID,
        memberID=memberId,
        actualTraderID=actualTraderId,        
        PartitionID=Partition_Id
    )

    return data ,traderid, -2147483648