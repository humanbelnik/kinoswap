import pandas as pd
import matplotlib.pyplot as plt
import sys
import os
from datetime import datetime

def plot_metrics_comparison(grpc_csv_file, http_csv_file, output_plot):
    grpc_df = pd.read_csv(grpc_csv_file)
    print(f"gRPC metrics: {len(grpc_df)} data points from {grpc_csv_file}")
    
    
    http_df = pd.read_csv(http_csv_file)
    print(f"HTTP metrics: {len(http_df)} data points from {http_csv_file}")
    
    if len(grpc_df) == 0 and len(http_df) == 0:
        print("No data found in both CSV files")
        return
    
    
    def convert_timestamp(df):
        if 'timestamp' in df.columns:
            
            df['datetime'] = pd.to_datetime(df['timestamp'], unit='ms')
            df['seconds'] = (df['timestamp'] - df['timestamp'].min()) / 1000
        return df
    
    if len(grpc_df) > 0:
        grpc_df = convert_timestamp(grpc_df)
    if len(http_df) > 0:
        http_df = convert_timestamp(http_df)    

    fig, (ax1, ax2) = plt.subplots(1, 2, figsize=(18, 6))
    fig.suptitle('Docker Metrics Comparison: gRPC vs HTTP', fontsize=16, fontweight='bold')
    
    grpc_cpu_label = 'gRPC'
    http_cpu_label = 'HTTP'
    

    if len(grpc_df) > 0 and 'cpu_percent' in grpc_df.columns:
        grpc_cpu_avg = grpc_df['cpu_percent'].mean()
        grpc_cpu_min = grpc_df['cpu_percent'].min()
        grpc_cpu_max = grpc_df['cpu_percent'].max()
        grpc_cpu_label = f'gRPC (avg={grpc_cpu_avg:.1f}%, min={grpc_cpu_min:.1f}%, max={grpc_cpu_max:.1f}%)'
        
        ax1.plot(grpc_df['seconds'], grpc_df['cpu_percent'], 
                'r-', linewidth=2, label=grpc_cpu_label, alpha=0.8)
    
    if len(http_df) > 0 and 'cpu_percent' in http_df.columns:
        http_cpu_avg = http_df['cpu_percent'].mean()
        http_cpu_min = http_df['cpu_percent'].min()
        http_cpu_max = http_df['cpu_percent'].max()
        http_cpu_label = f'HTTP (avg={http_cpu_avg:.1f}%, min={http_cpu_min:.1f}%, max={http_cpu_max:.1f}%)'
        
        ax1.plot(http_df['seconds'], http_df['cpu_percent'], 
                'b-', linewidth=2, label=http_cpu_label, alpha=0.8)
    
    ax1.set_xlabel('Time (seconds)')
    ax1.set_ylabel('CPU Usage (%)')
    ax1.set_title('CPU Usage Over Time')
    ax1.grid(True, alpha=0.3)
    ax1.legend(fontsize=9)
    

    grpc_mem_label = 'gRPC'
    http_mem_label = 'HTTP'
    
    if len(grpc_df) > 0 and 'memory_percent' in grpc_df.columns:
        grpc_mem_avg = grpc_df['memory_percent'].mean()
        grpc_mem_min = grpc_df['memory_percent'].min()
        grpc_mem_max = grpc_df['memory_percent'].max()
        grpc_mem_label = f'gRPC (avg={grpc_mem_avg:.1f}%, min={grpc_mem_min:.1f}%, max={grpc_mem_max:.1f}%)'
        
        ax2.plot(grpc_df['seconds'], grpc_df['memory_percent'], 
                'r-', linewidth=2, label=grpc_mem_label, alpha=0.8)
    
    if len(http_df) > 0 and 'memory_percent' in http_df.columns:
        http_mem_avg = http_df['memory_percent'].mean()
        http_mem_min = http_df['memory_percent'].min()
        http_mem_max = http_df['memory_percent'].max()
        http_mem_label = f'HTTP (avg={http_mem_avg:.1f}%, min={http_mem_min:.1f}%, max={http_mem_max:.1f}%)'
        
        ax2.plot(http_df['seconds'], http_df['memory_percent'], 
                'b-', linewidth=2, label=http_mem_label, alpha=0.8)
    
    ax2.set_xlabel('Time (seconds)')
    ax2.set_ylabel('Memory Usage (%)')
    ax2.set_title('Memory Usage Over Time')
    ax2.grid(True, alpha=0.3)
    ax2.legend(fontsize=9)
    
    plt.tight_layout()
    
    plt.savefig(output_plot, dpi=300, bbox_inches='tight')
    print(f"Metrics comparison plot saved to: {output_plot}")
    
    if len(grpc_df) > 0:
        print("\ngRPC Metrics:")
        if 'cpu_percent' in grpc_df.columns:
            cpu_stats = grpc_df['cpu_percent'].describe()
            print(f"  CPU Usage: avg={cpu_stats['mean']:.1f}%, min={cpu_stats['min']:.1f}%, max={cpu_stats['max']:.1f}%")
        if 'memory_percent' in grpc_df.columns:
            mem_stats = grpc_df['memory_percent'].describe()
            print(f"  Memory Usage: avg={mem_stats['mean']:.1f}%, min={mem_stats['min']:.1f}%, max={mem_stats['max']:.1f}%")
    
    if len(http_df) > 0:
        print("\nHTTP Metrics:")
        if 'cpu_percent' in http_df.columns:
            cpu_stats = http_df['cpu_percent'].describe()
            print(f"  CPU Usage: avg={cpu_stats['mean']:.1f}%, min={cpu_stats['min']:.1f}%, max={cpu_stats['max']:.1f}%")
        if 'memory_percent' in http_df.columns:
            mem_stats = http_df['memory_percent'].describe()
            print(f"  Memory Usage: avg={mem_stats['mean']:.1f}%, min={mem_stats['min']:.1f}%, max={mem_stats['max']:.1f}%")
    
    if len(grpc_df) > 0 and len(http_df) > 0:
        if 'cpu_percent' in grpc_df.columns and 'cpu_percent' in http_df.columns:
            grpc_cpu_avg = grpc_df['cpu_percent'].mean()
            http_cpu_avg = http_df['cpu_percent'].mean()
            cpu_diff = http_cpu_avg - grpc_cpu_avg
            cpu_diff_percent = (cpu_diff / grpc_cpu_avg) * 100
            print(f"CPU Usage: gRPC={grpc_cpu_avg:.1f}%, HTTP={http_cpu_avg:.1f}%, diff={cpu_diff:+.1f}% ({cpu_diff_percent:+.1f}%)")
        
        if 'memory_percent' in grpc_df.columns and 'memory_percent' in http_df.columns:
            grpc_mem_avg = grpc_df['memory_percent'].mean()
            http_mem_avg = http_df['memory_percent'].mean()
            mem_diff = http_mem_avg - grpc_mem_avg
            mem_diff_percent = (mem_diff / grpc_mem_avg) * 100
            print(f"Memory Usage: gRPC={grpc_mem_avg:.1f}%, HTTP={http_mem_avg:.1f}%, diff={mem_diff:+.1f}% ({mem_diff_percent:+.1f}%)")

def main():
    if len(sys.argv) != 4:
        print("Usage: python3 plot_metrics_comparison.py <grpc_metrics_csv> <http_metrics_csv> <output_png>")
        print("Example: python3 plot_metrics_comparison.py metrics_avg_grpc.csv metrics_avg_http.csv metrics_comparison.png")
        sys.exit(1)
    
    grpc_csv_file = sys.argv[1]
    http_csv_file = sys.argv[2]
    output_plot = sys.argv[3]
    
    if not os.path.exists(grpc_csv_file):
        print(f"gRPC metrics CSV file not found: {grpc_csv_file}")
        sys.exit(1)
    
    if not os.path.exists(http_csv_file):
        print(f"HTTP metrics CSV file not found: {http_csv_file}")
        sys.exit(1)
    
    os.makedirs(os.path.dirname(output_plot) if os.path.dirname(output_plot) else '.', exist_ok=True)
    
    plot_metrics_comparison(grpc_csv_file, http_csv_file, output_plot)

if __name__ == "__main__":
    main()