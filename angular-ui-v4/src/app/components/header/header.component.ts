import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ApiService } from '../../api.service';
import { MatIconModule } from '@angular/material/icon';
import { StateService } from '../../state.service';
import { MatDialog } from '@angular/material/dialog';
import { AllDayStatsModalComponent } from '../all-day-stats-modal/all-day-stats-modal.component';
import { MatMenuModule } from '@angular/material/menu';
import { MatButtonModule } from '@angular/material/button';
import { MessageService } from 'primeng/api';

@Component({
  selector: 'app-header',
  standalone: true,
  imports: [CommonModule, FormsModule, MatIconModule, MatMenuModule, MatButtonModule],
  templateUrl: './header.component.html',
  styleUrls: ['./header.component.css'],
  providers: [MessageService]
})
export class HeaderComponent{
  active: boolean = false;

  selectedMode: string = 'market-replay';
  isActive: any;

  // Holds interval data for widgets
  intervalData: any = null;

  // Loading state for fetching interval data
  loadingIntervalData: boolean = false;

  constructor(
    private apiService: ApiService,
    public stateService: StateService,
    public dialog: MatDialog,
    private messageService: MessageService,
  ) {}

  toggleMenu() {
    this.active = !this.active;
    this.stateService.setBurgerMenuState(this.active);
  }

  openAllDayStats(): void {
    // Make API call to fetch the data
    this.apiService.getCompleteDayData().subscribe({
      next: (response: any) => {
        const allDayData = response; // Assuming the API returns the JSON as described
        this.dialog.open(AllDayStatsModalComponent, {
          width: '600px',
          data: allDayData, // Pass the data to the modal
        });
      },
      error: (error) => {
        console.error('Error fetching all-day stats data:', error);
      },
    });
  }
  
  // startProcess() {
  //   // Update the metrics state
  //   this.stateService.updateMetricsState(!this.stateService.metrics_received);
  
  //   // Call the first API (startPumpingOrders)
  //   // this.apiService.startPumpingOrders().subscribe({
  //   //   next: (response: any) => {
  //   //     if (response.status == 200) {
  //   //       this.stateService.updateProcessState(true);
  //   //       console.log('Response of start pumping:', response);
  //   //     }
  //   //   },
  //   //   error: (error) => {
  //   //     console.error('Error while starting pumping orders:', error);
  //   //   },
  //   // });
  
  //   // Fetch interval data (second API)
  //   this.loadingIntervalData = true; // Start loading
  //   this.apiService.getIntervalData().subscribe({
  //     next: (intervalResponse: any) => {
  //       console.log('Response of get interval data:', intervalResponse);
  //       this.intervalData = intervalResponse.body; // Store interval data for widgets
  //       this.loadingIntervalData = false; // Stop loading
  //     },
  //     error: (error) => {
  //       console.error('Error while fetching interval data:', error);
  //       this.loadingIntervalData = false; // Stop loading even on error
  //     },
  //   });
  // }  


  StopProcess(){

    this.apiService.stopAllPods().subscribe({
      next: response => {
        console.log('Stop Process Response', response.message);
        this.messageService.add({severity: 'success', summary: 'Success', detail: response.message});
        this.stateService.updateLoadingState(false);
      },
      error: error => {
        console.error('Error Stop Process:', error);
        this.messageService.add({severity: 'error', summary: 'Error', detail: 'Stop Process Error: ' +error});
        this.stateService.updateLoadingState(false);
      },
      complete: () => {
        this.stateService.updateLoadingState(false); // Stop loading after completion
        
      }
    });

  }

}
