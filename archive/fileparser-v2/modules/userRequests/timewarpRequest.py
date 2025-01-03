class TimeWarp:
    def __init__(self, start_time_str, end_time_str, search_trader_id, search_partition_id, search_product_id, factor_of_time_warping ):
        self.start_time_str = start_time_str
        self.end_time_str = end_time_str
        self.search_trader_id = search_trader_id
        self.search_partition_id = search_partition_id
        self.search_product_id = search_product_id
        self.factor_of_time_warping = factor_of_time_warping

    @staticmethod
    def from_json(timewarp_request_body):
        
        if not timewarp_request_body or 'start_time_str' not in timewarp_request_body or 'end_time_str' not in timewarp_request_body:
            raise ValueError("Invalid data. 'start_time_str' and 'end_time_str' are required.")
        return TimeWarp(start_time_str=timewarp_request_body['start_time_str'], end_time_str=timewarp_request_body['end_time_str'], search_trader_id=timewarp_request_body['search_trader_id'], search_partition_id=timewarp_request_body['search_partition_id'], search_product_id=timewarp_request_body['search_product_id'], factor_of_time_warping=timewarp_request_body['factor_of_time_warping'] )

    def to_dict(self):
        return {"start_time_str": self.start_time_str, "search_trader_id": self.search_trader_id, "end_time_str": self.search_partition_id, "search_partition_id": self.search_product_id,"search_product_id": self.end_time_str, "factor_of_time_warping": self.factor_of_time_warping}
