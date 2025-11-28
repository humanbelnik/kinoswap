import pandas as pd
import matplotlib.pyplot as plt
import sys
import os

def plot_http_latency(http_csv_file, output_plot):
    
    http_df = pd.read_csv(http_csv_file)
    print(f"HTTP data: {len(http_df)} points from {http_csv_file}")
    
    if len(http_df) == 0:
        print("No data found in HTTP CSV file")
        return
    
    rps_column = http_df.columns[0]
    latency_column = http_df.columns[1]
    
    http_df = http_df.sort_values(rps_column)
    
    plt.figure(figsize=(14, 9))
    
    plt.plot(http_df[rps_column], http_df[latency_column], 'bo-', 
             linewidth=3, markersize=10, markerfacecolor='blue', 
             markeredgecolor='darkblue', markeredgewidth=1.5,
             label='HTTP')
    
    for i, (rps, latency) in enumerate(zip(http_df[rps_column], http_df[latency_column])):
        plt.annotate(f'{latency:.1f}ms', 
                    (rps, latency), 
                    textcoords="offset points", 
                    xytext=(0,-20), 
                    ha='center', 
                    fontsize=9,
                    color='blue',
                    weight='bold')
    
    plt.xlabel('Частота (запрос/сек)', fontsize=14)
    plt.ylabel('Время (мс)', fontsize=14)
    plt.title('Зависимость времени запрос-ответ от частоты запросов в секунду', fontsize=16, fontweight='bold')
    plt.grid(True, alpha=0.3)
    plt.legend(fontsize=12)
    
    all_rps = http_df[rps_column].tolist()
    all_latency = http_df[latency_column].tolist()
    
    if all_rps and all_latency:
        plt.xlim(0, max(all_rps) * 1.05)
        plt.ylim(0, max(all_latency) * 1.15)
    
    plt.tight_layout()
    
    plt.savefig(output_plot, dpi=300, bbox_inches='tight')
    print(f"HTTP latency plot saved to: {output_plot}")
    
    print("\n=== HTTP STATISTICS ===")
    http_min = http_df[latency_column].min()
    http_max = http_df[latency_column].max()
    http_avg = http_df[latency_column].mean()
    print(f"HTTP Latency: {http_min:.1f} - {http_max:.1f} ms (avg: {http_avg:.1f} ms)")
    
    print(f"\nAll data points:")
    for i, row in http_df.iterrows():
        print(f"  RPS {row[rps_column]}: {row[latency_column]:.1f}ms")

def main():
    http_csv_file = sys.argv[1]
    output_plot = sys.argv[2]
    
    if not os.path.exists(http_csv_file):
        print(f"HTTP CSV file not found: {http_csv_file}")
        sys.exit(1)
    
    os.makedirs(os.path.dirname(output_plot) if os.path.dirname(output_plot) else '.', exist_ok=True)
    
    plot_http_latency(http_csv_file, output_plot)

if __name__ == "__main__":
    main()
    