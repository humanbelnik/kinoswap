import pandas as pd
import matplotlib.pyplot as plt
import sys
import os
from datetime import datetime

def plot_metrics(grpc_csv_file, output_plot):
    grpc_df = pd.read_csv(grpc_csv_file)
    print(f"gRPC metrics: {len(grpc_df)} data points from {grpc_csv_file}")
    
    if len(grpc_df) == 0:
        print("No data found in CSV file")
        return
    
    def convert_timestamp(df):
        if 'timestamp' in df.columns:
            df['datetime'] = pd.to_datetime(df['timestamp'], unit='ms')
            df['seconds'] = (df['timestamp'] - df['timestamp'].min()) / 1000
        return df
    
    grpc_df = convert_timestamp(grpc_df)

    fig, (ax1, ax2) = plt.subplots(1, 2, figsize=(18, 6))
    fig.suptitle('Docker Metrics: gRPC', fontsize=16, fontweight='bold')
    
    # CPU Usage plot
    if 'cpu_percent' in grpc_df.columns:
        grpc_cpu_avg = grpc_df['cpu_percent'].mean()
        grpc_cpu_min = grpc_df['cpu_percent'].min()
        grpc_cpu_max = grpc_df['cpu_percent'].max()
        grpc_cpu_label = f'gRPC (avg={grpc_cpu_avg:.1f}%, min={grpc_cpu_min:.1f}%, max={grpc_cpu_max:.1f}%)'
        
        ax1.plot(grpc_df['seconds'], grpc_df['cpu_percent'], 
                'r-', linewidth=2, label=grpc_cpu_label, alpha=0.8)
    
    ax1.set_xlabel('Time (seconds)')
    ax1.set_ylabel('CPU Usage (%)')
    ax1.set_title('CPU Usage Over Time')
    ax1.grid(True, alpha=0.3)
    ax1.legend(fontsize=9)
    
    # Memory Usage plot (in MB)
    if 'memory_percent' in grpc_df.columns:
        # Конвертируем проценты в мегабайты
        CONTAINER_TOTAL_MEMORY_MB = 940  # Используем ваше значение
        grpc_df['memory_mb'] = (grpc_df['memory_percent'] / 100) * CONTAINER_TOTAL_MEMORY_MB
        
        grpc_mem_avg_mb = grpc_df['memory_mb'].mean()
        grpc_mem_min_mb = grpc_df['memory_mb'].min()
        grpc_mem_max_mb = grpc_df['memory_mb'].max()
        grpc_mem_label = f'gRPC (avg={grpc_mem_avg_mb:.1f}MB, min={grpc_mem_min_mb:.1f}MB, max={grpc_mem_max_mb:.1f}MB)'
        
        ax2.plot(grpc_df['seconds'], grpc_df['memory_mb'], 
                'r-', linewidth=2, label=grpc_mem_label, alpha=0.8)
    
    ax2.set_xlabel('Time (seconds)')
    ax2.set_ylabel('Memory Usage (MB)')
    ax2.set_title('Memory Usage Over Time')
    ax2.grid(True, alpha=0.3)
    ax2.legend(fontsize=9)
    
    plt.tight_layout()
    
    plt.savefig(output_plot, dpi=300, bbox_inches='tight')
    print(f"Metrics plot saved to: {output_plot}")
    
    # Print statistics
    print("\ngRPC Metrics:")
    if 'cpu_percent' in grpc_df.columns:
        cpu_stats = grpc_df['cpu_percent'].describe()
        print(f"  CPU Usage: avg={cpu_stats['mean']:.1f}%, min={cpu_stats['min']:.1f}%, max={cpu_stats['max']:.1f}%")
    if 'memory_mb' in grpc_df.columns:
        mem_stats = grpc_df['memory_mb'].describe()
        print(f"  Memory Usage: avg={mem_stats['mean']:.1f}MB, min={mem_stats['min']:.1f}MB, max={mem_stats['max']:.1f}MB")

def main():
    if len(sys.argv) != 3:
        print("Usage: python3 plot_metrics.py <grpc_metrics_csv> <output_png>")
        print("Example: python3 plot_metrics.py metrics_avg_grpc.csv metrics_plot.png")
        sys.exit(1)
    
    grpc_csv_file = sys.argv[1]
    output_plot = sys.argv[2]
    
    if not os.path.exists(grpc_csv_file):
        print(f"gRPC metrics CSV file not found: {grpc_csv_file}")
        sys.exit(1)
    
    os.makedirs(os.path.dirname(output_plot) if os.path.dirname(output_plot) else '.', exist_ok=True)
    
    plot_metrics(grpc_csv_file, output_plot)

if __name__ == "__main__":
    main()
