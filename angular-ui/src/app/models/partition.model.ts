export interface PartitionData {
    id: number;
    orderStats: {
      orderEntry: number;
      orderModify: number;
      orderCancel: number;
    };
    opsRate: number[];
    activeConnections: number[];
    topStocks: {
      instId: string;
      volume: number;
      trades: number;
      value: number;
    }[];
    topSessions: {
      sessionId: string;
      volume: number;
      trades: number;
    }[];
  }
  
  export interface ConfigSettings {
    numTraders: number;
    marketReplay: boolean;
    throttle: number;
    batchSize: number;
    startTime: string;
    endTime: string;
    envFile: number; 
    environment: number; 
    // instruments: string;
    modifyPercentage: number; 
    cancelPercentage: number; 
    heartbeat: number; 
    duration: number;
    trade_throttle: number;
    [key: string]: any;
    unique_identifier:boolean;
  }
  

  export interface IntervalData {
    ops_interval: Record<string, Record<string, number>>;
    otc_partition: Record<string, { order_entry: number }>;
    sender_connections: Record<string, number[]>;
    top_inst_ids: Record<string, Record<string, number>>;
    top_session_ids: Record<string, Record<string, { order_cancellation: number; order_entry: number; order_modification: number }>>;
  }
  