import { Component, Output, EventEmitter, OnInit} from '@angular/core';
import { CommonModule, JsonPipe } from '@angular/common';
import { FormsModule, AbstractControl, ValidationErrors } from '@angular/forms';
import { ReactiveFormsModule, FormBuilder, FormGroup, Validators } from '@angular/forms';
import { ApiService } from '../../api.service';
import { ConfigSettings } from '../../models/partition.model';
import { MatCheckboxModule } from '@angular/material/checkbox';
import { StateService } from '../../state.service';
import { MatTabsModule } from '@angular/material/tabs';
import { MatExpansionModule } from '@angular/material/expansion';
import { MatIconModule } from '@angular/material/icon';
import { MatRadioModule } from '@angular/material/radio';
import { MatSelectModule } from '@angular/material/select';
import { MatOptionModule } from '@angular/material/core'
import { ToastModule } from 'primeng/toast'
import { MessageService } from 'primeng/api';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatTooltipModule } from '@angular/material/tooltip';

@Component({
  selector: 'app-config-panel',
  standalone: true,
  imports: [CommonModule, FormsModule, MatCheckboxModule, MatTabsModule, MatExpansionModule, ReactiveFormsModule, MatIconModule, MatRadioModule, MatSelectModule, MatOptionModule, ToastModule, MatProgressSpinnerModule, MatTooltipModule],
  templateUrl: './sidebar.component.html',
  styleUrls: ['./sidebar.component.css'],
  providers: [MessageService]
})
export class ConfigPanelComponent implements OnInit {
  recordForm!: FormGroup;
  @Output() configChange = new EventEmitter<ConfigSettings>();

  addRecordForm!: FormGroup;
  peakGeneratorForm!: FormGroup
  waveForm!: FormGroup;
  selectedAction: string = 'add';
  deleteInstrumentId: number = 1234;
  deleteAllInstrumentId: string = '';
  deleteAllProductId: string = '';
  deleteAllId:number=1
  throttleValue: number = 100;
  selectedRateType: number = 0;
  podIdString: string = '';
  podIds: Number[] = [];
  patternpodIds: Number[] = [];
  selectedPattern: string = 'random';
  interval: number = 1;
  isButtonDisabled = false;
  deleteOrderType = 2
  podCount = 1
  managePodSelection = 1
  // pumpingConfigs: any[] = [];

  // Step Wave properties
  stepwaveMin: number | undefined;
  stepwaveMax: number | undefined;
  stepwaveStep: number | undefined;
  stepwaveInterval: number | undefined;

  // Triangular Wave properties
  triangularMin: number | undefined;
  triangularMax: number | undefined;
  triangularInterval: number | undefined;

  // Sawtooth Wave properties
  sawtoothMin: number | undefined;
  sawtoothMax: number | undefined;
  sawtoothInterval: number | undefined;

  // Square Wave properties
  squareMin: number | undefined;
  squareMax: number | undefined;
  squareInterval: number | undefined;

  // Sine Wave properties
  sineMin: number | undefined;
  sineMax: number | undefined;

  // Random Wave properties
  randomMin: number | undefined;
  randomMax: number | undefined;

  config: ConfigSettings = {
    numTraders: 1,
    marketReplay: false,
    throttle: 100,
    batchSize: 10,
    startTime: '',
    endTime: '',
    envFile: 1,
    environment: 1,
    modifyPercentage:90,
    cancelPercentage: 5,
    heartbeat: 60000,
    duration:10000,
    trade_throttle: 100, 
  };

  throttleValues = [100, 500, 1000, 1500, 2000];
  isSectionVisible: Record<string, boolean> = { general: true, time: false };
  errors: Record<string, string> = {};
  recommendedBatchSize = 100;
  isLoading = false;
  previewData: { startTime: string; endTime: string } | null = null;

  constructor( 
    private apiService: ApiService, 
    public stateService: StateService, 
    private fb: FormBuilder,
    private messageService: MessageService,
  ) {}

  ngOnInit() {
    this.getConfig();
  
    this.addRecordForm = this.fb.group({
      instrument_id: ['', [Validators.required, Validators.min(1)]],
      lower_limit: ['',  [Validators.required, Validators.min(1)]],
      upper_limit: ['', [Validators.required, Validators.min(1)]],
      min_lot: ['', [Validators.required, Validators.min(1)]],
      bid_interval: ['', [Validators.required, Validators.min(1)]],
      max_trd_qty: ['', [Validators.required, Validators.min(1)]],
      product_id: ['', [Validators.required, Validators.min(1)]]
    });

    this.peakGeneratorForm = this.fb.group({
      base_ops: ['',  [Validators.required, Validators.min(1)]],
      up_percentage: ['',  [Validators.required, Validators.min(1)]],
      down_percentage: ['', [Validators.required, Validators.min(1)]],
      min_delay: ['', [Validators.required, Validators.min(1)]],
      max_delay: ['', [Validators.required, Validators.min(1)]],
    });

    this.waveForm = this.fb.group(
      {
        min: [null, [Validators.required, Validators.min(1)]],
        max: [null, [Validators.required, Validators.min(1)]],
        interval: [null],
        step: [null],
        custom_pattern: [null]  // Add this for the custom wave pattern
      },
      { validators: this.customValidations() }
    );    

    this.updateWaveFormValidators(); // Ensure it's called at the start
  }

  updatePodIds() {
    this.podIds = this.podIdString
      .split(',')
      .map(id => id.trim())
      .filter(id => id !== '')  // Remove empty values
      .map(id => Number(id))  // Convert to numbers
      .filter(id => !isNaN(id)); // Remove invalid numbers
  }

  updatePatternPods() {
    this.patternpodIds = this.podIdString
    .split(',')
    .map(id => id.trim())
    .filter(id => id !== '')  // Remove empty values
    .map(id => Number(id))  // Convert to numbers
    .filter(id => !isNaN(id)); // Remove invalid numbers
  }
  

  onPatternChange(event: any) {
    console.log("Pattern selected:", event.value);
    this.updateWaveFormValidators(); // Trigger validators update here
  }

  private minMaxValidator(group: FormGroup) {
    const min = group.get('min')?.value;
    const max = group.get('max')?.value;
    return min !== null && max !== null && min >= max ? { minMaxError: true } : null;
  }

  customValidations() {
    return (form: AbstractControl): ValidationErrors | null => {
      const min = form.get('min')?.value;
      const max = form.get('max')?.value;
      const type = this.getPatternType(this.selectedPattern);
      const step = form.get('step')?.value;
      const interval = form.get('interval')?.value;

      if (min !== null && max !== null && min >= max) {
        return { minMaxError: true };
      }

      if (type === 4) {
        if (!step || step <= 0) {
          return { stepError: true };
        }
        if (min % step !== 0 || max % step !== 0) {
          return { stepMultipleError: true };
        }
      }

      if (type === 5) {
        if (!interval || interval <= 0) {
          return { intervalError: true };
        }
        if (min % interval !== 0 || max % interval !== 0) {
          return { intervalMultipleError: true };
        }
      }

      return null;
    };
  }

  updateWaveFormValidators() {
    console.log('Running updateWaveFormValidators for pattern:', this.selectedPattern);
    
    const intervalControl = this.waveForm.get('interval');
    const stepControl = this.waveForm.get('step');
  
    // Reset validators
    intervalControl?.clearValidators();
    stepControl?.clearValidators();
  
    // Apply required validators based on the selected wave type
    if (['square', 'sawtooth', 'triangular', 'stepwave'].includes(this.selectedPattern)) {
      intervalControl?.setValidators([Validators.required]);
      console.log('Interval control set to required');
    } else {
      console.log('Interval control NOT required');
    }
  
    if (this.selectedPattern === 'stepwave') {
      stepControl?.setValidators([Validators.required]);
      console.log('Step control set to required');
    } else {
      console.log('Step control NOT required');
    }
  
    // Update form state
    intervalControl?.updateValueAndValidity();
    stepControl?.updateValueAndValidity();
  
    // Debug current control status
    console.log('Interval Control Valid:', intervalControl?.valid);
    console.log('Step Control Valid:', stepControl?.valid);
  }
   
  pumpingConfigs = [
    { instId: '', startMinPrice: 1, startMaxPrice: 1001, endMinPrice: 1, endMaxPrice: 1001 } // Default sub-panel
  ];

  addPumpingConfig() {
    this.pumpingConfigs.push({ instId: '', startMinPrice: 1, startMaxPrice: 1001, endMinPrice: 1, endMaxPrice: 1001 });
  }

  removePumpingConfig(index: number) {
    this.pumpingConfigs.splice(index, 1);
  }

  submitPumpingConfigs() {
    this.stateService.updateLoadingState(true);
    const instruments: { [key: number]: any } = {}; // Use number instead of string
  
    this.pumpingConfigs.forEach(config => {
      if (config.instId) {
        const instId = Number(config.instId); // Convert to number
        if (!isNaN(instId)) { // Ensure it's a valid number
          instruments[instId] = {
            start_min_price: config.startMinPrice,
            start_max_price: config.startMaxPrice,
            end_min_price: config.endMinPrice,
            end_max_price: config.endMaxPrice,
            // min_interval: 0,
            // max_interval: 0
          };
        }
      }
    });
  
    const payload = {
      interval: this.interval,
      inst_id: instruments
    };
  
    console.log("Final Payload:", JSON.stringify(payload, null, 2));
  
    this.apiService.addSlidingPriceRange(payload).subscribe({
      next: (response) => {
        console.log('Response from API:', response);
        this.messageService.add({severity: 'success', summary: 'Success', detail: response.message});
        this.stateService.updateLoadingState(false);
      },
      error: (error) => {
        console.error('Error submitting price range:', error);
        this.messageService.add({severity: 'error', summary: 'Error', detail: error});
        // alert('Error submitting price range. Check console for details.');
        this.stateService.updateLoadingState(false);
      }
    });
  }
  
  

  submitAddRecord() {
    this.stateService.updateLoadingState(true); // Start loading
  
    if (this.addRecordForm.invalid) {
      this.messageService.add({severity: 'warn', summary: 'Warning', detail: 'Please fill all required fields.'});
      this.stateService.updateLoadingState(false); // Stop loading if form is invalid
      return;
    }
  
    const formData = {
      records: [{
        ...this.addRecordForm.value,
        instrument_id: parseInt(this.addRecordForm.value.instrument_id, 10) // Convert to integer
      }]
    };
    console.log(formData);
    this.apiService.updateInstrumentRecord(formData).subscribe({
      next: response => {
        console.log('Record Updated:', response.message);
        this.messageService.add({severity: 'success', summary: 'Success', detail: response.message});
        this.stateService.updateLoadingState(false);
      },
      error: error => {
        console.error('Error Updating Record:', error);
        this.messageService.add({severity: 'error', summary: 'Error', detail: 'Failed to update record.'});
        this.stateService.updateLoadingState(false);
      },
      complete: () => {
        console.log('Update Record Request Completed.');
        this.stateService.updateLoadingState(false); // Stop loading after completion
      }
    });
  }
  
  /** Delete a Specific Record */
  deleteRecord() {
    this.stateService.updateLoadingState(true); // Start loading
  
    if (!this.deleteInstrumentId) {
      this.messageService.add({severity: 'warn', summary: 'Warning', detail: 'Please enter an Instrument ID.'});
      this.stateService.updateLoadingState(false); // Stop loading if ID is missing
      return;
    }
    
    this.apiService.deleteInstrumentRecord(this.deleteInstrumentId).subscribe({
      next: response => {
        console.log('Record Deleted:', response);
        this.messageService.add({severity: 'success', summary: 'Success', detail: 'Record deleted successfully.'});
        this.stateService.updateLoadingState(false);
      },
      error: error => {
        console.error('Error Deleting Record:', error);
        this.messageService.add({severity: 'error', summary: 'Error', detail: 'Failed to delete record.'});
        this.stateService.updateLoadingState(false);
      },
      complete: () => {
        console.log('Delete Record Request Completed.');
        this.stateService.updateLoadingState(false); // Stop loading after completion
      }
    });
  }
  
  /** Delete All Orders */
  deleteAllOrders() {
    this.stateService.updateLoadingState(true); // Start loading

    if (!this.deleteAllInstrumentId || !this.deleteAllProductId || (this.deleteOrderType !== 2 && !this.deleteAllId)) {
      const detailMessage = !this.deleteAllInstrumentId || !this.deleteAllProductId
        ? 'Please enter both Instrument ID and Product ID.'
        : `Please enter ${this.deleteOrderType == 1 ? "Member" : "Session"} Id`;
    
      this.messageService.add({ severity: 'warn', summary: 'Warning', detail: detailMessage });
      this.stateService.updateLoadingState(false); // Stop loading if fields are missing
      return;
    }
    
    let payload = {
      instId: Number(this.deleteAllInstrumentId), 
      prodId: Number(this.deleteAllProductId),
      type: this.deleteOrderType,
      ...(this.deleteOrderType != 2 && { ids: { [this.deleteAllId]: 1 } }) 
    };

    this.apiService.deleteOrders(payload).subscribe({
      next: response => {
        console.log('All Orders Deleted:', response);
        this.messageService.add({severity: 'success', summary: 'Success', detail: 'All orders deleted successfully.'});
        this.stateService.updateLoadingState(false);
      },
      error: error => {
        console.error('Error Deleting Orders:', error);
        this.messageService.add({severity: 'error', summary: 'Error', detail: 'Failed to delete orders.'});
        this.stateService.updateLoadingState(false);
      },
      complete: () => {
        console.log('Delete All Orders Request Completed.');
        this.stateService.updateLoadingState(false); // Stop loading after completion
      }
    });
  }
  
  /** Update Throttle Value */
  updateThrottle() {
    this.stateService.updateLoadingState(true); // Start loading
  
    if (!this.throttleValue) {
      this.messageService.add({severity: 'warn', summary: 'Warning', detail: 'Please enter a Throttle Value.'});
      this.stateService.updateLoadingState(false); // Stop loading if no throttle value
      return;
    }

    const payload = { 
      throttle: Number(this.throttleValue),
      pods: this.podIds, 
      type: Number(this.selectedRateType),
    };
    
  
    this.apiService.updateThrottleValue(payload).subscribe({
      next: response => {
        console.log('Throttle Updated:', response);
        this.messageService.add({severity: 'success', summary: 'Success', detail: 'Throttle value updated successfully.'});
        this.stateService.updateLoadingState(false);
      },
      error: error => {
        console.error('Error Updating Throttle:', error);
        this.messageService.add({severity: 'error', summary: 'Error', detail: 'Failed to update throttle value.'});
        this.stateService.updateLoadingState(false);
      },
      complete: () => {
        console.log('Update Throttle Request Completed.');
        this.stateService.updateLoadingState(false); // Stop loading after completion
      }
    });
  }
  
  submitPattern() {
    this.stateService.updateLoadingState(true); // Start loading
  
    if (this.waveForm.invalid) {
      console.warn('Invalid waveForm:', this.waveForm.value);
      
      Object.keys(this.waveForm.controls).forEach(key => {
        const control = this.waveForm.get(key);
        if (control?.invalid) {
          console.warn(`Control "${key}" is invalid. Errors:`, control.errors);
        }
      });
  
      this.messageService.add({
        severity: 'warn',
        summary: 'Warning',
        detail: 'Please fill all required fields.'
      });
  
      this.stateService.updateLoadingState(false); // Stop loading if form is invalid
      return;
    }
  
    const { min, max, interval, step } = this.waveForm.value;
    let payload: any = {
      min: Number(min),
      max: Number(max),
      type: this.getPatternType(this.selectedPattern), 
      pods: this.patternpodIds  
    };
  
    if (!payload.type) {
      console.error('Invalid Pattern Selection');
      this.messageService.add({
        severity: 'error',
        summary: 'Error',
        detail: 'Invalid pattern selection.'
      });
      this.stateService.updateLoadingState(false);
      return;
    }
  
    // Add required parameters based on the selected wave type
    if (['square', 'sawtooth', 'triangular', 'stepwave'].includes(this.selectedPattern)) {
      payload.interval = Number(interval);
    }
    if (this.selectedPattern === 'stepwave') {
      payload.step = Number(step);
    }
  
    this.apiService.generatePattern(payload).subscribe(this.handleResponse);
  }

  isFormValid(): boolean {
    return this.pumpingConfigs.every(config =>
      config.instId && config.instId.trim() !== '' &&
      config.startMaxPrice >= 1 && config.startMinPrice >= 1 &&
      config.endMaxPrice >= 1 && config.endMinPrice >= 1 &&
      (config.startMaxPrice - config.startMinPrice) >= 1000 &&
      (config.endMaxPrice - config.endMinPrice) >= 1000  && this.interval >= 1
    );
  }

  
  
  /** Maps selected pattern to corresponding type number */
  private getPatternType(pattern: string): number | null {
    const patternMap: { [key: string]: number } = {
      'square': 1,
      'sawtooth': 2,
      'sine': 3,
      'stepwave': 4,
      'triangular': 5,
      'random': 6
    };
    return patternMap[pattern] || null;
  }
  
  /** Common Response Handler for Wave Generation */
  private handleResponse = {
    next: (response: any) => {
      console.log('Wave Generated:', response);
      this.messageService.add({
        severity: 'success',
        summary: 'Success',
        detail: 'Wave generated successfully.'
      });
      this.stateService.updateLoadingState(false);
    },
    error: (error: any) => {
      console.error('Error Generating Wave:', error);
      this.messageService.add({
        severity: 'error',
        summary: 'Error',
        detail: error?.error?.details || 'Failed to generate wave.'
      });
      this.stateService.updateLoadingState(false);
    },
    complete: () => {
      console.log('Wave Generation Request Completed.');
      this.stateService.updateLoadingState(false);
    }
  };
  




  onMarketReplayChange(event: Event) {
    const isChecked = (event.target as HTMLInputElement).checked;
    this.stateService.setMarketReplay(isChecked);
  }

  toggleSection(section: string): void {
    this.isSectionVisible[section] = !this.isSectionVisible[section];
  }


  validateInput(field: keyof ConfigSettings): void {
    const validationMap: Record<string, (value: number) => string | undefined> = {
      numTraders: (value) => (value < 1 || value > 1200 ? 'Number of traders must be between 1 and 1200.' : undefined)
    };
  
    const error = validationMap[field]?.(this.config[field]);
    if (error) {
      this.errors[field] = error;
    } else {
      delete this.errors[field];
    }
  }
  
  updateRecommendations(): void {
    this.recommendedBatchSize = this.config.numTraders > 500 ? 300 : 100;
  }

  previewConfig(): void {
    this.isLoading = true;
    setTimeout(() => {
      this.previewData = { startTime: this.config.startTime, endTime: this.config.endTime };
      this.isLoading = false;
    }, 1000);
  }

  applyConfig(): void {
    this.isButtonDisabled = true;
    this.isLoading = true;

    const dataToSend = this.prepareConfigData();
    console.log(JSON.stringify(dataToSend));

    this.apiService.setConfig(dataToSend).subscribe((response: any) => {
      if (response.status == 200){
        console.log('Response of set config', response);
        this.isLoading = false;
      }
    });
  }

  getConfig(): void{
    this.apiService.fetchConfig().subscribe((response : any)=>{
      const configData = response.body.config
      console.log("Get Config Data = ", JSON.stringify(response.body.config));

      this.config = {
        numTraders: configData.TraderCount || 1,
        marketReplay: false, 
        throttle: typeof configData.ThrottleValue === 'number' ? configData.ThrottleValue : 100, // Default value
        batchSize: 10,
        startTime: '',
        endTime: '',
        envFile: configData.FileType || 1,
        environment: configData.TargetEnv || 1,
        modifyPercentage: configData.ModifyPercent || 90,
        cancelPercentage: configData.CancelPercent || 5,
        heartbeat: configData.HeartbeatValue || 60000,
        duration: 10000,
        trade_throttle: 1234
      };
    })
  }


  prepareConfigData() {
    return {
      trader_count: this.config.numTraders ? this.config.numTraders : 1, // Required
      target_env: Number(this.config.environment), // Required
      throttle_value: this.config.throttle ? this.config.throttle: 100, // Required
      cancel_percent: this.config.cancelPercentage ? this.config.cancelPercentage : 5, // Required
      modify_percent: this.config.modifyPercentage ? this.config.modifyPercentage :90, // Required
      heartbeat_value: this.config.heartbeat ? this.config.heartbeat : 60000, // Required
      file_type: this.config.envFile, 
      duration: this.config.duration? this.config.duration: 10000,
      trader_ops: this.config.trade_throttle ? this.config.trade_throttle : 100,
    };
  }


  addPods(){
    this.stateService.updateLoadingState(true);

    const payload = { 
      count: Number(this.podCount),
      pods: this.podIds, 
    };
  
    this.apiService.addPods(payload).subscribe({
      next: response => {
        console.log('Add pods response', response);
        this.messageService.add({severity: 'success', summary: 'Success', detail: response.message});
        this.stateService.updateLoadingState(false);
      },
      error: error => {
        console.error('Error Add Pods:', error);
        this.messageService.add({severity: 'error', summary: 'Error', detail: 'Failed to send add pod request.'});
        this.stateService.updateLoadingState(false);
      },
      complete: () => {
        console.log('Add Pods Request Completed.');
        this.stateService.updateLoadingState(false); // Stop loading after completion
      }
    });


  }

  deletePods(){
    this.stateService.updateLoadingState(true);

    const payload = { 
      pods: this.podIds, 
    };

    this.apiService.deletePods(payload).subscribe({
      next: response => {
        console.log('Delete pods response', response);
        this.messageService.add({severity: 'success', summary: 'Success', detail: response.message});
        this.stateService.updateLoadingState(false);
      },
      error: error => {
        console.error('Error Delete Pods:', error);
        this.messageService.add({severity: 'error', summary: 'Error', detail: 'Failed to send delete pod request.'});
        this.stateService.updateLoadingState(false);
      },
      complete: () => {
        console.log('Delete Pods Request Completed.');
        this.stateService.updateLoadingState(false); // Stop loading after completion
      }
    });
  }


  generatePeak() {
    this.stateService.updateLoadingState(true); // Start loading
  
    if (this.peakGeneratorForm.invalid) {
      this.messageService.add({severity: 'warn', summary: 'Warning', detail: 'Please fill all required fields.'});
      this.stateService.updateLoadingState(false); // Stop loading if form is invalid
      return;
    }

    let min_th =  this.peakGeneratorForm.value.base_ops - (this.peakGeneratorForm.value.base_ops * this.peakGeneratorForm.value.down_percentage * 0.01)
    let max_th =  this.peakGeneratorForm.value.base_ops + (this.peakGeneratorForm.value.base_ops * this.peakGeneratorForm.value.up_percentage * 0.01)

    const payload = {
      type: 7, 
      min: min_th,
      max: max_th,
      delay_min : this.peakGeneratorForm.value.min_delay, 
      delay_max : this.peakGeneratorForm.value.max_delay, 
    };
  

    console.log(payload);

    this.apiService.generatePattern(payload).subscribe({
      next: response => {
        console.log('Peak Generator Response', response.message);
        this.messageService.add({severity: 'success', summary: 'Success', detail: response.message});
        this.stateService.updateLoadingState(false);
      },
      error: error => {
        console.error('Error Peak Generator:', error);
        this.messageService.add({severity: 'error', summary: 'Error', detail: 'Peak Generator Error: ' +error});
        this.stateService.updateLoadingState(false);
      },
      complete: () => {
        this.stateService.updateLoadingState(false); // Stop loading after completion
      }
    });
  }

}


