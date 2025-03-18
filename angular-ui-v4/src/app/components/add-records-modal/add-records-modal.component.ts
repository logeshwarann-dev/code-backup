import { Component, ViewEncapsulation, Inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ReactiveFormsModule, FormBuilder, FormGroup, Validators } from '@angular/forms';
import { MatDialogModule, MatDialogRef,MAT_DIALOG_DATA } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { MatRadioModule } from '@angular/material/radio'; // Import Radio Module
import { ApiService } from '../../api.service';
import { FormsModule } from '@angular/forms';
import { MatSnackBar } from '@angular/material/snack-bar';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';

@Component({
  selector: 'app-add-records-modal',
  standalone: true,
  encapsulation: ViewEncapsulation.None,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    MatDialogModule,
    MatFormFieldModule,
    MatInputModule,
    MatButtonModule,
    MatRadioModule,
    FormsModule,
    MatProgressSpinnerModule,
  ],
  templateUrl: './add-records-modal.component.html',
  styleUrls: ['./add-records-modal.component.css']
})
export class AddRecordsModalComponent {
  recordForm!: FormGroup;
  loading: boolean = false;
  mode: string = 'add'; // Default mode is 'add'

  constructor(
    private fb: FormBuilder,
    private dialogRef: MatDialogRef<AddRecordsModalComponent>,
    private apiService: ApiService,
    private snackBar: MatSnackBar,
    @Inject(MAT_DIALOG_DATA) public data: any
  ) {
    this.initializeForm(data?.mode);
  }

  initializeForm(mode: string): void {
    if (mode === 'option1') {
      // Form for Delete Order with Product ID
      this.recordForm = this.fb.group({
        instrument_id: ["", Validators.required],
        product_id: ["", Validators.required],
      });
    } else if (mode === 'option2') {
      // Form for Change Throttle with Throttle Value
      this.recordForm = this.fb.group({
        throttle_value: [0, Validators.required],
      });
    } else {
      // Default mode with Add/Delete radio button
      this.mode = 'add'; // Initialize default as 'add'
      this.recordForm = this.fb.group({
        instrument_id: ["", Validators.required],
        lower_limit: [0, Validators.required],
        upper_limit: [0, Validators.required],
        min_lot: [0, Validators.required],
        bid_interval: [0, Validators.required],
        max_trd_qty: [0, Validators.required],
        product_id: [0, Validators.required]
      });
    }
  }
  
  onModeChange(event: Event): void {
    const target = event.target as HTMLInputElement; // Safely cast the target
    const mode = target.value;
    this.mode = mode;
  
    if (mode === 'delete') {
      // Simplified form for delete operation in default mode (only Instrument ID)
      this.recordForm = this.fb.group({
        instrument_id: ["", Validators.required],
      });
    } else {
      // Full form for add operation in default mode
      this.recordForm = this.fb.group({
        instrument_id: ["", Validators.required],
        lower_limit: [0, Validators.required],
        upper_limit: [0, Validators.required],
        min_lot: [0, Validators.required],
        bid_interval: [0, Validators.required],
        max_trd_qty: [0, Validators.required],
        product_id: [0, Validators.required],
      });
    }
  }
 

  showToast(message: string, duration: number = 3000, type: 'success' | 'error' = 'success'): void {
    const panelClass = type === 'success' ? 'custom-snackbar-success' : 'custom-snackbar-error';
  
    this.snackBar.open(message, 'Close', {
      duration: 50000000000, // Duration in milliseconds
      verticalPosition: 'bottom', // Position the toast at the top
      horizontalPosition: 'center', // Center horizontally
      panelClass, // Add dynamic class based on the message type
    });
  }
  
  getSubmitButtonText(): string {
    if (this.data?.mode === 'option1') {
      return 'Delete Order';
    } else if (this.data?.mode === 'option2') {
      return 'Change Throttle';
    } else if (this.mode === 'add') {
      return 'Add Record';
    } else {
      return 'Delete Record';
    }
  }
  

  onCancel(): void {
    this.dialogRef.close();
  }

  onSubmit(): void {
    if (this.recordForm.valid) {
      this.loading = true;
      if (this.data?.mode === 'default' && this.mode === 'add') {
        // Add Record logic
        const payload = {
          records: [
            {
              instrument_id: this.recordForm.value.instrument_id,
              lower_limit: +this.recordForm.value.lower_limit,
              upper_limit: +this.recordForm.value.upper_limit,
              min_lot: +this.recordForm.value.min_lot,
              bid_interval: +this.recordForm.value.bid_interval,
              max_trd_qty: +this.recordForm.value.max_trd_qty,
              product_id: +this.recordForm.value.product_id,
            },
          ],
        };
  
        // Call the API to add record
        this.apiService.updateInstrumentRecord(payload).subscribe({
          next: (response) => {
            // console.log('Record added successfully:', response);
            this.loading = false;
            this.showToast(response?.message || 'Record added successfully', 5000, 'success');
            this.dialogRef.close();
          },
          error: (error) => {
            // console.error('Error while adding the record:', error);
            this.loading = false;
            this.showToast(
              error?.error?.message || 'Failed to add record',
              5000, 'error'
            ); // Show detailed error for 5 seconds
          },
        });
      } else if (this.data?.mode === 'default' && this.mode === 'delete') {
        // Delete Record logic
        const instrument_id = this.recordForm.value.instrument_id;
  
        // Call the API to delete the record
        console.log('Instrument ID: ', instrument_id);
        this.apiService.deleteInstrumentRecord(instrument_id).subscribe({
          next: (response) => {
            // console.log('Record deleted successfully:', response);
            this.loading = false;
            this.showToast(response?.message || 'Record deleted successfully', 5000, 'success');
            this.dialogRef.close();
          },
          error: (error) => {
            // console.error('Error while deleting the record:', error);
            this.loading = false;
            this.showToast(
              error?.error?.message || 'Failed to delete record',
              5000, 'error'
            ); // Show detailed error for 5 seconds
          },
        });
      } else if (this.data?.mode === 'option2') {
        // Throttle Change logic
        const throttleValue = this.recordForm.value.throttle_value;
  
        // Call the API to change throttle
        const payload = {
          throttle: +this.recordForm.value.throttle_value, // Ensure it's a number
        };
  
        this.apiService.updateThrottleValue(payload).subscribe({
          next: (response) => {
            // console.log('Throttle value changed successfully:', response);
            this.loading = false;
            this.showToast(response?.message || 'Throttle value updated successfully', 5000, 'success');
            this.dialogRef.close();
          },
          error: (error) => {
            // console.error('Error while updating throttle value:', error);
            this.loading = false;
            this.showToast(
              error?.error?.message || 'Failed to update throttle value',
              5000, 'error'
            ); // Show detailed error for 5 seconds
          },
        });
      } else if (this.data?.mode === 'option1') {
        const product_id = this.recordForm.value.product_id;
        const instrument_id = this.recordForm.value.instrument_id;
      
        const payload = {
          instId: +this.recordForm.value.instrument_id, // Convert to number
          prodId: +this.recordForm.value.product_id,    // Convert to number
        };
      
        console.log("DeleteOrder Payload:", payload);
      
        // Call the API to delete orders
        this.apiService.deleteOrders(payload).subscribe({
          next: (response) => {
            // console.log('Orders deleted successfully:', response);
            this.loading = false;
            this.showToast(response?.message || 'Orders deleted successfully', 5000, 'success');
            this.dialogRef.close();
          },
          error: (error) => {
            // console.error('Error while deleting orders:', error);
            this.loading = false;
            this.showToast(
              error?.error?.message || 'Failed to delete orders',
              5000,
              'error'
            ); // Show detailed error for 5 seconds
          },
        });
      }      
    }
  }
  
}
