import { Injectable} from '@angular/core';
import { Signal, signal } from '@angular/core';

@Injectable({
  providedIn: 'root'
})
export class StateService {

  constructor() { }

  // Writable signal for marketReplay
  private marketReplaySignal = signal<boolean>(false);

  private burgerMenuState = signal<boolean>(false);

  private loadingSignal = signal<boolean>(false);

  readonly panelOpenState = signal(false);

  togglePanelState() {
    this.panelOpenState.set(!this.panelOpenState());
  }

  setBurgerMenuState(state: boolean) {
    this.burgerMenuState.set(state);
  }

  getBurgerMenuState(): Signal<boolean> {
    return this.burgerMenuState;
  }

  // Getter for the marketReplay signal
  get marketReplay() {
    return this.marketReplaySignal.asReadonly();
  }

  // Method to update the marketReplay signal
  setMarketReplay(value: boolean) {
    this.marketReplaySignal.set(value);
  }

  process_state = signal(false);
  metrics_received = signal(false);

  updateProcessState(state: boolean): void {
    this.process_state.set(state);
  }

  updateMetricsState(state:boolean): void{
    this.metrics_received.set(state);
  }

  // Method to update loading state
  updateLoadingState(state: boolean): void {
    this.loadingSignal.set(state);
  }

  // Getter for loading state signal
  getloading() {
    return this.loadingSignal;
  }
}
