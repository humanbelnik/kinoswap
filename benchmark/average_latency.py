import pandas as pd
import glob
import os
import sys
import re

def average_latency_results(out_dir, n_runs, cores, proto):
    
    print(f"Starting processing: out_dir={out_dir}, n_runs={n_runs}, cores={cores}")
    
    all_files = glob.glob(os.path.join(out_dir, proto, "rps_*.csv"))
    print(f"Found files: {len(all_files)}")
    for f in all_files:
        print(f"  {f}")
    
    rps_values = set()
    
    for file_path in all_files:
        
        match = re.search(r'rps_(\d+)_n_(\d+)_cores_', os.path.basename(file_path))
        if match:
            rps = int(match.group(1))
            rps_values.add(rps)
            print(f"Found RPS: {rps} in file {os.path.basename(file_path)}")
    
    if not rps_values:
        print("No RPS values found! Checking file patterns...")
        for file_path in all_files[:3]:  
            print(f"File: {os.path.basename(file_path)}")
    
    rps_values = sorted(rps_values)
    print(f"RPS values to process: {rps_values}")
    

    combined_avg_file = os.path.join('processed', f'{proto}_latency_cores_{cores}.csv')
    os.makedirs(os.path.dirname(combined_avg_file), exist_ok=True)
    
    with open(combined_avg_file, 'w') as f:
        f.write("rps,average_latency_ms,status\n")
    

    for rps in rps_values:
        print(f"\nAveraging results for RPS: {rps}")
        
        all_runs_data = []
        

        for run in range(1, n_runs + 1):
            file_path = os.path.join(out_dir, proto, f"rps_{rps}_n_{run}_cores_{cores}.csv")
            print(f"Looking for file: {file_path}")
            
            if os.path.exists(file_path):
                try:
                    print(f"  Processing file: {os.path.basename(file_path)}")
                    df = pd.read_csv(file_path)
                    print(f"  File shape: {df.shape}")
                    print(f"  Columns: {df.columns.tolist()}")
                    
                    if len(df) > 0:
                        if 'status' in df.columns:
                            successful_requests = df[df['status'] == 'OK']
                            print(f"  Total requests: {len(df)}, Successful: {len(successful_requests)}")
                            
                            if len(successful_requests) > 0:
                                all_runs_data.append(successful_requests)
                                print(f"  Run {run}: {len(successful_requests)} successful requests")
                            else:
                                print(f"  Run {run}: no successful requests (all failed)")
                        else:
                            print(f"  Run {run}: no 'status' column in file")
                            all_runs_data.append(df)
                    else:
                        print(f"  Run {run}: empty file")
                        
                except Exception as e:
                    print(f"  Error processing {file_path}: {e}")
            else:
                print(f"  File not found: {file_path}")
        
        if not all_runs_data:
            print(f"  No valid data for RPS {rps}")
            continue
        

        combined_df = pd.concat(all_runs_data)
        print(f"  Combined data shape: {combined_df.shape}")
        

        if 'duration (ms)' in combined_df.columns:
            avg_latency = combined_df['duration (ms)'].mean()
            print(f"  Average latency: {avg_latency:.2f} ms")
            
            with open(combined_avg_file, 'a') as f:
                
                f.write(f"{rps},{avg_latency:.2f},OK\n")
        else:
            print(f"  No 'duration (ms)' column found. Available columns: {combined_df.columns.tolist()}")
    
    print(f"\nAveraged latency results saved to: {combined_avg_file}")

if __name__ == "__main__":
    if len(sys.argv) != 5:
        print("Usage: python3 average_latency_results.py <out_dir> <n_runs> <cores>")
        print("Example: python3 average_latency_results.py increase_rps_out 2 1.0")
        sys.exit(1)
    
    out_dir = sys.argv[1]
    n_runs = int(sys.argv[2])
    cores = float(sys.argv[3])
    proto = sys.argv[4]
    
    print(f"Script started with: out_dir={out_dir}, n_runs={n_runs}, cores={cores}")
    average_latency_results(out_dir, n_runs, cores, proto)