from metrics import *
def set_member_ids_for_order_pumping(redis_Client):
    session_id_data = {}
    for Mids in range(1034, 1047):
        for Tids in range(1001, 1051):
            sessionId = (Mids * 100000) + Tids
            session_id_data[sessionId] = {
                "MemberId": Mids,
                "TraderId": Tids,
                "SessionId": sessionId
            }

    store_member_set_to_redis(session_id_data,redis_Client)
    