class OrderPumpConfig:
    def __init__(self, user_template_id, user_account_type, user_time_in_force, user_client_code, user_client_order_id ):
        self.user_template_id = user_template_id
        self.user_account_type = user_account_type
        self.user_time_in_force = user_time_in_force
        self.user_client_code = user_client_code
        self.user_client_order_id = user_client_order_id

    @staticmethod
    def from_json(orderpumpconfig_request_body):
        
        if not orderpumpconfig_request_body or 'user_template_id' not in orderpumpconfig_request_body or 'user_account_type' not in orderpumpconfig_request_body:
            raise ValueError("Invalid data. 'user_template_id' and 'user_account_type' are required.")
        return OrderPumpConfig(user_template_id=orderpumpconfig_request_body['user_template_id'], user_account_type=orderpumpconfig_request_body['user_account_type'], user_time_in_force=orderpumpconfig_request_body['user_time_in_force'], user_client_code=orderpumpconfig_request_body['user_client_code'], user_client_order_id=orderpumpconfig_request_body['user_client_order_id'])

    def to_dict(self):
        return {"user_template_id": self.user_template_id, "user_time_in_force": self.user_time_in_force, "user_account_type": self.user_client_code, "user_client_code": self.user_client_order_id,"user_client_order_id": self.user_account_type}
