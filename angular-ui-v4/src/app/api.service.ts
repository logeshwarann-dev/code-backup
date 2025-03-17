import { HttpClient, HttpErrorResponse, HttpHeaders, HttpResponse } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable, throwError } from 'rxjs';
import { catchError } from 'rxjs/operators';
import { environment } from '../environment';
import { IntervalData } from './models/partition.model';

@Injectable({
  providedIn: 'root',
})
export class ApiService {
  private readonly headers = new HttpHeaders().set('Content-Type', 'application/json');
  private eventSource: EventSource | null = null;

  constructor(private http: HttpClient) {}

  /**
   * Sets configuration via API
   * @param api_data - The configuration data to set
   * @param instId 
   * @returns Observable of the HTTP response
   */
  setConfig(api_data: any): Observable<HttpResponse<any>> {
    return this.http
      .post<any>(`${environment.api_endpoint}/opMiddleware/v1/setConfig`, api_data, {
        headers: this.headers,
        observe: 'response',
      })
      .pipe(catchError(this.handleError));
  }
  
  uploadFile(file: File): Observable<HttpResponse<any>> {
    const formData = new FormData();
    formData.append('file', file);

    return this.http.post<any>(`${environment.api_endpoint}/opMiddleware/v1/fileUpload`, formData, {
      observe: 'response',
    }).pipe(catchError(this.handleError));
  }

  /**
   * Fetches interval data from the API
   * @returns Observable of the interval data
   */
  getIntervalData(): Observable<IntervalData> {
    return this.http
      .get<IntervalData>(`${environment.api_endpoint}/get-interval-data`, { headers: this.headers })
      .pipe(catchError(this.handleError));
  }


  getCompleteDayData() {
    return this.http.get('http://localhost:8080/get-complete-day-data');
  }

  /**
   * Starts pumping orders via API
   * @returns Observable of the HTTP response
   */
  startPumpingOrders(): Observable<HttpResponse<any>> {
    return this.http
      .post<any>(`${environment.api_endpoint}/api/v1/data-processor/sender/trigger`, null, {
        headers: this.headers,
        observe: 'response',
      })
      .pipe(catchError(this.handleError));
  }

  /**
   * Fetches the configuration from the API
   * @returns Observable of the configuration data
   */
  fetchConfig(): Observable<HttpResponse<any>> {
    return this.http
      .get<any>(`${environment.api_endpoint}/opMiddleware/v1/getConfig`, {
        headers: this.headers,
        observe: 'response',
      })
      .pipe(catchError(this.handleError));
  }

  /**
   * Establishes a server-sent events connection
   * @returns EventSource object for receiving events
   */
  connect(): EventSource {
    this.eventSource = new EventSource(`${environment.api_endpoint}/events`);
    return this.eventSource;
  }

  /**
   * Closes the server-sent events connection
   */
  closeEventSource() {
    if (this.eventSource) {
      this.eventSource.close();
      this.eventSource = null;
    }
  }

  addSlidingPriceRange(payload: any): Observable<any> {
    const url = `${environment.api_endpoint}/opMiddleware/v1/addSlidingPriceRange`;
    return this.http.post(url, payload,{
      headers: { 'Content-Type': 'application/json' },
    });
  }

  updateInstrumentRecord(payload: any): Observable<any> {
    const url = `${environment.api_endpoint}/opMiddleware/v1/updateRecords`;
    return this.http.put(url, payload, {
      headers: { 'Content-Type': 'application/json' },
    });
  }   

  // deleteInstrumentRecord(instId: string): Observable<any> {
    
  //   const url = 'http://localhost:8080/api/v1/DeleteInst';
  //   const payload = { instId };
  //   console.log("inside function payload = ", payload)
  //   return this.http.delete(url, {
  //     headers: this.headers,
  //     body: payload,
  //   }).pipe(catchError(this.handleError));
  // }

  deleteInstrumentRecord(instId: number): Observable<any> {
    const url = `${environment.api_endpoint}/opMiddleware/v1/deleteRecords`;
    const payload = { instId: Number(instId) };
    console.log("DeleteRecord Payload = ", JSON.stringify(payload));
    return this.http.delete(url, {
        headers: this.headers,
        body: payload,
    }).pipe(
        catchError((error: HttpErrorResponse) => {
            // Log the error
            console.error("Delete operation failed:", error);

            // Handle specific error scenarios
            if (error.status === 500 && error.error?.error === "Not enough records present") {
                return throwError(() => new Error("Cannot delete: Not enough records present."));
            }

            // Re-throw other errors
            return throwError(() => new Error(error.message));
        })
    );
  }

  

  updateThrottleValue(payload: any): Observable<any> {
    const url = `${environment.api_endpoint}/opMiddleware/v1/changeOPS`;
    return this.http.post(url, payload, {
      headers: { 'Content-Type': 'application/json' },
    });
  }  

  deleteOrders(payload: { instId: number; prodId: number }): Observable<any> {
    const url = `${environment.api_endpoint}/opMiddleware/v1/deleteOrders`;
    return this.http.delete(url, {
      headers: { 'Content-Type': 'application/json' },
      body: payload, // Pass payload as the body for DELETE request
    });
  }

  generatePattern(payload: any): Observable<any> {
    const url = `${environment.api_endpoint}/opMiddleware/v1/generatePattern`;
  
    console.log("Angular Payload Sent:", JSON.stringify(payload));
  
    return this.http.post(url, JSON.stringify(payload), {  
      headers: new HttpHeaders({
        'Content-Type': 'application/json',
        'Accept': 'application/json'  
      }),
    }).pipe(
      catchError((error: HttpErrorResponse) => {
        console.error("Pattern generation failed:", error);
        return throwError(() => new Error(error.message));
      })
    );
  }


  addPods(payload:any): Observable<any>{
    const url = `${environment.api_endpoint}/opMiddleware/v1/addPods`;
    return this.http.post(url, payload, {
      headers: { 'Content-Type': 'application/json' },
    })
    .pipe(catchError(this.handleError));
  } 
  
  deletePods(payload:any): Observable<any>{
    const url = `${environment.api_endpoint}/opMiddleware/v1/deletePods`;
    return this.http.delete(url, {
      headers: { 'Content-Type': 'application/json' },
      body: payload, // Pass payload as the body for DELETE request
    })
    .pipe(catchError(this.handleError));
  } 
  

  stopAllPods(): Observable<any>{
    const url = `${environment.api_endpoint}/opMiddleware/v1/shutdown`;
    return this.http.delete(url, {
      headers: { 'Content-Type': 'application/json' },
    })
    .pipe(catchError(this.handleError));
  } 

  /**
   * Handles HTTP errors
   * @param error - The HttpErrorResponse object
   * @returns Observable that throws an error
   */
  private handleError(error: HttpErrorResponse): Observable<never> {
    console.error('An error occurred:', error.message);
    return throwError(() => new Error(`Error: ${error.message}`));
  }
}