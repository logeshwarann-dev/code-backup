import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ChartWidgetComponent } from '../widgets/chart-widget/chart-widget.component';
import { ConfigPanelComponent } from '../sidebar/sidebar.component';
import { TopStocksComponent } from '../topstocks/topstocks.component';
import { GrafanaDashboardComponent } from '../grafana-dashboard/grafana-dashboard.component';
import { ApiService } from '../../api.service';
import { MatIconModule } from '@angular/material/icon';
import { StateService } from '../../state.service';
import { MatDialog } from '@angular/material/dialog';
import { IntervalData } from '../../models/partition.model'; // Assuming you have the interface defined here
import { AddRecordsModalComponent } from '../add-records-modal/add-records-modal.component';
import { MatMenuModule } from '@angular/material/menu';
import { MatButtonModule } from '@angular/material/button';

@Component({
  selector: 'app-dashboard',
  standalone: true,
  imports: [CommonModule, ChartWidgetComponent, ConfigPanelComponent, TopStocksComponent, MatIconModule, GrafanaDashboardComponent, MatMenuModule, MatButtonModule],
  templateUrl: './dashboard.component.html',
  styleUrls: ['./dashboard.component.css']
})
export class DashboardComponent {
  metricsReceived: any;
  selectedMode: string = 'market-replay';
  isActive: any;

  // Holds interval data for widgets
  intervalData: IntervalData | null = null;

  // Loading state for fetching interval data
  loadingIntervalData: boolean = false;
  
  isSidebarOpen: boolean = false;

  constructor(
    private apiService: ApiService,
    public stateService: StateService,
    public dialog: MatDialog
  ) {}

  startProcess() {
    // Update the metrics state
    this.stateService.updateMetricsState(!this.stateService.metrics_received);

    // Call the first API (startPumpingOrders)
    // this.apiService.startPumpingOrders().subscribe({
    //   next: (response: any) => {
    //     if (response.status == 200) {
    //       this.stateService.updateProcessState(true);
    //       console.log('Response of start pumping:', response);
    //     }
    //   },
    //   error: (error) => {
    //     console.error('Error while starting pumping orders:', error);
    //   },
    // });

    // Fetch interval data (second API)
    this.loadingIntervalData = true; // Start loading
    this.apiService.getIntervalData().subscribe({
      next: (intervalResponse: any) => {
        console.log('Response of get interval data:', intervalResponse);
        this.intervalData = intervalResponse as IntervalData; // Store interval data for widgets
        this.loadingIntervalData = false; // Stop loading
        console.log('Binded Response: ', this.intervalData);
      },
      error: (error) => {
        console.error('Error while fetching interval data:', error);
        this.loadingIntervalData = false; // Stop loading even on error
      },
    });
  }

  onConfigChange(config: any) {
    console.log('Configuration updated:', config);
  }

  openAddRecordsModal(mode: string) {
    // Open the AddRecordsModalComponent using MatDialog
    this.dialog.open(AddRecordsModalComponent, {
      width: '400px',
      data: { mode: mode }, // Pass mode as dialog data
    });
  }
}
