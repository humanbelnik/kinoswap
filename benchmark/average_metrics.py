import pandas as pd
import glob
import os
import sys
import numpy as np

def convert_memory_to_mb(memory_str):
    """Конвертирует строки с памятью в мегабайты"""
    if pd.isna(memory_str) or memory_str == '0':
        return 0
    
    memory_str = str(memory_str).strip()
    try:
        if 'GiB' in memory_str:
            return float(memory_str.replace('GiB', '').strip()) * 1024
        elif 'MiB' in memory_str:
            return float(memory_str.replace('MiB', '').strip())
        elif 'KiB' in memory_str:
            return float(memory_str.replace('KiB', '').strip()) / 1024
        else:
            return float(memory_str)
    except (ValueError, TypeError):
        return 0

def convert_memory_back_to_gib(mb_value):
    """Конвертирует мегабайты обратно в GiB для сохранения формата"""
    return f"{mb_value / 1024:.3f}GiB"

def average_metrics(out_dir, n_runs, cores, proto):
    """Усредняет метрики ресурсов построчно по всем прогонам"""
    
    all_dataframes = []
    min_length = float('inf')
    
    print("Loading metrics files...")
    for run in range(1, n_runs + 1):
        metrics_file = os.path.join(out_dir, proto, f"metrics_n_{run}_cores_{cores}.csv")
        if os.path.exists(metrics_file):
            try:
                df = pd.read_csv(metrics_file)
                if len(df) > 0:

                    df['memory_usage_mb'] = df['memory_usage'].apply(convert_memory_to_mb)
                    df['memory_limit_mb'] = df['memory_limit'].apply(convert_memory_to_mb)
                    
                    df['time_seconds'] = (df['timestamp'] - df['timestamp'].min()) / 1000.0
                    
                    all_dataframes.append(df)
                    min_length = min(min_length, len(df))
                    print(f"  Run {run}: {len(df)} records")
                else:
                    print(f"  Run {run}: file is empty")
            except Exception as e:
                print(f"  Error processing {metrics_file}: {e}")
        else:
            print(f"  Run {run}: file not found")
    
    if not all_dataframes:
        print("No metrics data found")
        return
    
    print(f"\nFound {len(all_dataframes)} files with metrics")
    print(f"Minimum records per file: {min_length}")
    

    averaged_data = []
    
    for line_idx in range(min_length):
        cpu_values = []
        memory_percent_values = []
        memory_usage_mb_values = []
        memory_limit_mb_values = []
        timestamps = []
        
        for df in all_dataframes:
            if line_idx < len(df):
                cpu_values.append(df.iloc[line_idx]['cpu_percent'])
                memory_percent_values.append(df.iloc[line_idx]['memory_percent'])
                memory_usage_mb_values.append(df.iloc[line_idx]['memory_usage_mb'])
                memory_limit_mb_values.append(df.iloc[line_idx]['memory_limit_mb'])
                timestamps.append(df.iloc[line_idx]['timestamp'])
        

        if cpu_values: 
            avg_cpu = np.mean(cpu_values)
            avg_memory_percent = np.mean(memory_percent_values)
            avg_memory_usage_mb = np.mean(memory_usage_mb_values)
            avg_memory_limit_mb = np.mean(memory_limit_mb_values)
            avg_timestamp = np.mean(timestamps)
            
            avg_memory_usage_gib = convert_memory_back_to_gib(avg_memory_usage_mb)
            avg_memory_limit_gib = convert_memory_back_to_gib(avg_memory_limit_mb)
            
            averaged_data.append({
                'timestamp': int(avg_timestamp),
                'cpu_percent': round(avg_cpu, 2),
                'memory_usage': avg_memory_usage_gib,
                'memory_limit': avg_memory_limit_gib,
                'memory_percent': round(avg_memory_percent, 2),
                'time_seconds': round(line_idx, 2) 
            })
    
    avg_df = pd.DataFrame(averaged_data)
    
    avg_metrics_file = os.path.join("processed", f'{proto}_metrics_cores_{cores}.csv')
    os.makedirs(os.path.dirname(avg_metrics_file), exist_ok=True)
    
    avg_df[['timestamp', 'cpu_percent', 'memory_usage', 'memory_limit', 'memory_percent']].to_csv(
        avg_metrics_file, index=False)
    
    print(f"\nAveraged metrics saved to: {avg_metrics_file}")
    print(f"Total averaged records: {len(avg_df)}")
    
    print(f"CPU: min={avg_df['cpu_percent'].min():.2f}%, max={avg_df['cpu_percent'].max():.2f}%, avg={avg_df['cpu_percent'].mean():.2f}%")
    print(f"Memory: min={avg_df['memory_percent'].min():.2f}%, max={avg_df['memory_percent'].max():.2f}%, avg={avg_df['memory_percent'].mean():.2f}%")
    

    print(f"CPU values range: {avg_df['cpu_percent'].min():.2f} - {avg_df['cpu_percent'].max():.2f}")
    print(f"Memory values range: {avg_df['memory_percent'].min():.2f} - {avg_df['memory_percent'].max():.2f}")
    
    if avg_df['cpu_percent'].min() == avg_df['cpu_percent'].max():
        print("WARNING: All CPU values are the same!")
    
    if avg_df['memory_percent'].min() == avg_df['memory_percent'].max():
        print("WARNING: All memory values are the same!")

def main():
    if len(sys.argv) != 5:
        sys.exit(1)
    
    out_dir = sys.argv[1]
    n_runs = int(sys.argv[2])
    cores = float(sys.argv[3])
    proto = sys.argv[4]
    average_metrics(out_dir, n_runs, cores,proto)

if __name__ == "__main__":
    main()