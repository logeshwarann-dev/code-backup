import { Component, OnInit } from '@angular/core';
import { DashboardComponent } from './components/dashboard/dashboard.component';
import { HeaderComponent } from './components/header/header.component';
import { FooterComponent } from './components/footer/footer.component';
import { ApiService } from './api.service';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { CommonModule } from '@angular/common';
import { StateService } from './state.service';
import { ToastModule } from 'primeng/toast'
import { MessageService } from 'primeng/api';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [DashboardComponent, HeaderComponent, FooterComponent, MatProgressSpinnerModule, CommonModule, ToastModule],
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.css'],
  providers: [MessageService],
})
export class App implements OnInit {
  name = 'Angular Dashboard';

  private eventSource: EventSource | null = null;

  constructor(private apiService: ApiService, public stateService: StateService, private messageService: MessageService,) { }

  ngOnInit(): void {
    this.eventSource = this.apiService.connect();

    this.eventSource.onmessage = (event) => {
      console.log('Raw event received:', event); // Check if any data is received

      try {
        const data = JSON.parse(event.data);
        console.log('Parsed data:', data);

        if (data.type === 1) {
          // Show error message if type is 0
          this.messageService.add({
            severity: 'error',
            summary: 'Error',
            detail: data.message,
          });

        } else if (data.type === 0) {
          // Show success message if type is 1
          this.messageService.add({
            severity: 'success',
            summary: 'Success',
            detail: data.message,
          });

          // Handle any additional logic for success
          this.stateService.updateProcessState(true);
        } else {
          console.warn("Received unknown event type:", data.type);
        }

      } catch (error) {
        console.error('Error parsing SSE data:', error);
      }
    };

    // this.eventSource.onerror = () => {
    //   console.error('An error occurred while listening for events.', this.eventSource);
    //   this.messageService.add({
    //     severity: 'error',
    //     summary: 'Error',
    //     detail: this.eventSource ? `SSE Error: ${this.eventSource.url}` : 'SSE connection is null',
    //   });
    //   this.closeConnection(); // Clean up on error
    // };
    this.eventSource.onerror = () => {
      console.error('An error occurred while listening for events.', this.eventSource);
      console.error('SSE connection lost, attempting to reconnect...');
      this.closeConnection();

      setTimeout(() => {
        console.log("Reconnecting SSE...");
        this.apiService.connect();// Retry connection
      }, 5000); // Retry after 5 seconds
    };


  }

  ngOnDestroy() {
    this.closeConnection();
  }

  private closeConnection() {
    if (this.eventSource) {
      this.eventSource.close();
      this.eventSource = null;
    }
  }
}
