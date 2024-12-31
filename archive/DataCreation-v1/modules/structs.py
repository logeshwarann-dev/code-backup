class RedisData:
    
    def __init__(self,template_id,price,order_qty,maxshow,instid,acc_type,time_in_force,clientCode,MsgSeqNum,transactionType,OrderNumber,memberID, actualTraderID, TransactionTimeStamp,TradingProductId,PartitionID, trader_id = 0,partition_id = 0,product_id = 0,timestamp = 0):
        self.template_id = template_id
        self.instid = instid
        self.price = price
        self.order_qty = order_qty
        self.maxshow = maxshow
        self.acc_type = acc_type
        self.time_in_force = time_in_force
        self.clientCode = clientCode
        self.MsgSeqNum = MsgSeqNum
        self.trader_id = trader_id
        self.product_id = product_id
        self.timestamp = timestamp
        self.transactionType = transactionType
        self.OrderNumber = OrderNumber
        self.TransactionTimeStamp = TransactionTimeStamp
        self.TradingProductID = TradingProductId
        self.memberID = memberID
        self.actualTraderID = actualTraderID
        self.PartitionID = PartitionID
        self.partition_id = partition_id
        
        

    def to_dict(self):
        return {
            "template_id": self.template_id,
            "instid": self.instid,
            "price": self.price,
            "order_qty": self.order_qty,
            "maxshow": self.maxshow,
            "acc_type": self.acc_type,
            "time_in_force": self.time_in_force,
            "clientCode": self.clientCode,
            "MsgSeqNum" : self.MsgSeqNum,
            "transactionType":self.transactionType,
            "OrderNumber":self.OrderNumber,
            "TransactionTimeStamp": self.TransactionTimeStamp,
            "TradingProductID": self.TradingProductID,
            "member_id": self.memberID,
            "trader_id": self.actualTraderID,
            "PartitionID": self.PartitionID
        }
    
    def toCommaSeparated(self):
        return ",".join([
            str(self.template_id),
            str(self.instid),
            str(self.price),
            str(self.order_qty),
            str(self.maxshow),
            str(self.acc_type),
            str(self.time_in_force),
            str(self.clientCode),
            str(self.MsgSeqNum),
            str(self.transactionType),
            str(self.OrderNumber),
            str(self.TransactionTimeStamp),
            str(self.TradingProductID),
            str(self.memberID),
            str(self.actualTraderID),
            str(self.PartitionID)
        ])



class Trader:
    def __init__(self, tid, host_ip, port, mid, password):
        self.tid = tid
        self.host_ip = host_ip
        self.port = port
        self.mid = mid
        self.password = password

    def to_dict(self):
        return {
            "TID": self.tid,
            "HostIP": self.host_ip,
            "Port": self.port,
            "MID": self.mid,
            "Password": self.password,
        }

    @staticmethod
    def from_dict(data):
        return Trader(
            tid=data["TID"],
            host_ip=data["HostIP"],
            port=data["Port"],
            mid=data["MID"],
            password=data["Password"],
        )
