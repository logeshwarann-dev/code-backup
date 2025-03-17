import { Component, Input, OnChanges, ChangeDetectorRef, AfterViewInit, OnInit, SimpleChanges } from '@angular/core';
import { CdkDragDrop, DragDropModule, moveItemInArray } from '@angular/cdk/drag-drop';
import { Widget } from '../../../models/widget.model';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ApiService } from '../../../api.service';
import { IntervalData } from '../../../models/partition.model';
import { Chart, registerables } from 'chart.js';

Chart.register(...registerables);

@Component({
  selector: 'app-chart-widget',
  templateUrl: './chart-widget.component.html',
  styleUrls: ['./chart-widget.component.css'],
  standalone: true,
  imports: [CommonModule, FormsModule, DragDropModule]
})
export class ChartWidgetComponent implements OnInit, OnChanges, AfterViewInit {
  widgets: Widget[] = [];
  @Input() intervalData: IntervalData | null = null;

  chartInstances: Record<string, Chart | null> = {};
  selectedOption: string = 'allPartitions';
  selectedPartition: number | null = null;
  isSelectPartitionDisabled: boolean = true;

  partitions: number[] = Array.from({ length: 11 }, (_, i) => i + 1);

  constructor(private apiService: ApiService, private cdr: ChangeDetectorRef) {}

  ngOnInit() {
    this.loadDefaultWidgets();
  }

  ngOnChanges(changes: SimpleChanges) {
    console.log("--------NGONCHANGES--------")
    if (changes['intervalData'] && this.intervalData) {
      console.log('Interval Data Changed:', this.intervalData);
      this.updateAllWidgets();
    }
  }

  ngAfterViewInit() {
    this.createAllCharts();
  }

  // Load default widgets
  loadDefaultWidgets() {
    this.widgets = [
        {
            id: '1',
            title: 'Market Distribution',
            type: 'Pie Chart',
            chartType: 'pie',
            data: {
                labels: ['Order Entry', 'Order Modify', 'Order Cancel'],
                datasets: [
                    {
                        data: [0, 0, 0],
                        backgroundColor: ['#00ff9d', '#00cc7d', '#009f63'], // Shades of green for pie slices
                        borderColor: '#ffffff', // White border for pie slices
                        borderWidth: 2,
                    },
                ],
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        display: true, // Enable legend
                        position: 'bottom', // Position legend at the top
                        labels: {
                            color: '#ffffff', // White text for visibility
                            font: {
                                size: 14, // Adjust font size
                            },
                        },
                    },
                },
            },
            selectedOption: 'allPartitions',
        },
        {
            id: '2',
            title: 'Top 5 Traders (SessionIDs)',
            type: 'Bar Chart',
            chartType: 'bar',
            data: {
                labels: ['Trader 1', 'Trader 2', 'Trader 3', 'Trader 4', 'Trader 5'],
                datasets: [
                    {
                        label: 'Total Trades', // Legend label
                        data: [0],
                        backgroundColor: '#00ff9d', // Green color for bars
                        borderColor: '#00cc7d', // Darker green border
                        borderWidth: 2,
                    },
                ],
            },
            options: {
                responsive: true,
                maintainAspectRatio: false, // Disable aspect ratio to use full height
                plugins: {
                    legend: {
                        display: true,
                        position: 'bottom',
                        labels: {
                            color: '#ffffff',
                            font: {
                                size: 14,
                            },
                        },
                    },
                },
                scales: {
                    y: { beginAtZero: true },
                },
            },
            selectedOption: 'allPartitions',
        },
        {
            id: '3',
            title: 'Order Rate Over Time',
            type: 'Line Graph',
            chartType: 'line',
            data: {
                labels: ['Time 1', 'Time 2', 'Time 3', 'Time 4', 'Time 5'],
                datasets: [
                    {
                        label: 'Order Rate', // Legend label
                        data: [0],
                        borderColor: '#00cc7d', // Green line for the graph
                        backgroundColor: 'rgba(0, 204, 125, 0.1)', // Light green fill for the area under the line
                        borderWidth: 2,
                    },
                ],
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        display: true,
                        position: 'bottom',
                        labels: {
                            color: '#ffffff',
                            font: {
                                size: 14,
                            },
                        },
                    },
                },
            },
            selectedOption: 'allPartitions',
        },
        {
            id: '4',
            title: 'Execution Rate Over Time',
            type: 'Line Graph',
            chartType: 'line',
            data: {
                labels: ['Time 1', 'Time 2', 'Time 3', 'Time 4', 'Time 5'],
                datasets: [
                    {
                        label: 'Execution Rate', // Legend label
                        data: [0],
                        borderColor: '#00cc7d', // Green line for the graph
                        backgroundColor: 'rgba(0, 204, 125, 0.1)', // Light green fill for the area under the line
                        borderWidth: 2,
                    },
                ],
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        display: true,
                        position: 'bottom',
                        labels: {
                            color: '#ffffff',
                            font: {
                                size: 14,
                            },
                        },
                    },
                },
            },
            selectedOption: 'allPartitions',
        },
    ];
  }

  // Create charts for all widgets
  createAllCharts() {
    this.widgets.forEach((widget) => {
      const canvas = document.getElementById(`chart-${widget.id}`) as HTMLCanvasElement;
      if (canvas) {
        if (this.chartInstances[widget.id]) {
          this.chartInstances[widget.id]?.destroy();
        }
        this.chartInstances[widget.id] = new Chart(canvas, {
          type: widget.chartType,
          data: widget.data,
          options: widget.options,
        });
      }
    });
  }

  // Update data for all widgets
  updateAllWidgets() {
    this.widgets.forEach((widget) => {
      this.updateWidgetData(widget);
    });
  }

  // Method to handle changes in radio button selection
  onOptionChange(widget: Widget): void {
    this.isSelectPartitionDisabled = widget.selectedOption !== 'selectPartition';
    if (widget.selectedOption === 'allPartitions') {
      widget.selectedPartition = null;
      this.applyChanges(widget);
    }
  }
  

  applyChanges(widget: Widget): void {
    if (this.intervalData) {
      this.updateWidgetData(widget);
    }
  }
  
  

  // Update specific widget data
  updateWidgetData(widget: Widget) {
    console.log("--------UPDATE WIDGET -", widget.chartType, "--------");

    if (!this.intervalData) {
        console.warn(`Interval data is not available for widget: ${widget.id}`);
        widget.data.datasets[0].data = [0];
        widget.data.labels = [];
        this.cdr.detectChanges();
        this.chartInstances[widget.id]?.update();
        return;
    }

    const selectedPartitions =
    widget.selectedPartition === null || widget.selectedPartition === undefined
      ? Object.keys(this.intervalData.otc_partition || {}).map(Number) // Convert keys to numbers
      : [widget.selectedPartition];

    // Update logic for Pie Chart
    const updatePieChart = () => {
        const aggregatedData = selectedPartitions.reduce(
            (acc, partition) => {
                const partitionData = this.intervalData?.otc_partition?.[partition] || { order_entry: 0 };
                acc.order_entry += partitionData.order_entry;
                return acc;
            },
            { order_entry: 0 } // Default initial value
        );

        widget.data = {
            ...widget.data,
            datasets: [{ ...widget.data.datasets[0], data: [aggregatedData.order_entry] }],
        };
    };

    // Update logic for Bar Chart
    const updateBarChart = () => {
        const sessionDataMap: Record<string, number> = {};

        selectedPartitions.forEach((partition) => {
            const partitionData = this.intervalData?.top_session_ids?.[partition] || {};

            Object.entries(partitionData).forEach(([sessionId, metrics]) => {
                if (metrics && typeof metrics === "object") {
                    const totalOrders =
                        (metrics.order_entry || 0) +
                        (metrics.order_cancellation || 0) +
                        (metrics.order_modification || 0);

                    sessionDataMap[sessionId] = (sessionDataMap[sessionId] || 0) + totalOrders;
                }
            });
        });

        const topSessionData = Object.entries(sessionDataMap)
            .sort(([, valueA], [, valueB]) => valueB - valueA)
            .slice(0, 5);

        widget.data = {
            ...widget.data,
            datasets: [{ ...widget.data.datasets[0], data: topSessionData.map(([, value]) => value) }],
            labels: topSessionData.map(([sessionId]) => sessionId),
        };
    };

    // Determine widget type and update data accordingly
    switch (widget.id) {
        case '1': // Pie Chart
            updatePieChart();
            break;
        case '2': // Bar Chart
            updateBarChart();
            break;
        default:
            console.warn(`Unsupported widget type or ID: ${widget.id}`);
    }

    // Refresh chart instance
    if (this.chartInstances[widget.id]) {
        this.chartInstances[widget.id]!.data = { ...widget.data }; // Use non-null assertion
        this.chartInstances[widget.id]!.options = { ...widget.options }; // Update options if necessary
        this.chartInstances[widget.id]!.update(); // Trigger chart refresh
    } else {
        console.warn(`Chart instance not found for widget: ${widget.id}`);
    }

    // Trigger Angular's change detection to ensure UI updates
    this.cdr.detectChanges();
  }





  // Handle drag-and-drop reordering of widgets
  onDrop(event: CdkDragDrop<Widget[]>) {
    moveItemInArray(this.widgets, event.previousIndex, event.currentIndex);
    this.cdr.detectChanges();
  }

  ngOnDestroy() {
    Object.values(this.chartInstances).forEach((instance) => instance?.destroy());
  }
}
