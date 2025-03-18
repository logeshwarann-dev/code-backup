export interface Widget {
  id: string;
  title: string;
  type: string;
  chartType: 'pie' | 'bar' | 'line';
  data: any;
  options: any;
  selectedOption?: string; // For radio button selection
  selectedPartition?: number | null; // For partition selection
}



export interface ChartData {
  labels: string[];
  datasets: {
    data: number[];
    backgroundColor?: string[];
    borderColor?: string[];
    label?: string;
  }[];
}