import pandas as pd
import matplotlib.pyplot as plt
import sys
import os

def plot_latency_comparison(grpc_csv_file, http_csv_file, output_plot):

    
    grpc_df = pd.read_csv(grpc_csv_file)
    print(f"gRPC data: {len(grpc_df)} points from {grpc_csv_file}")
    
    http_df = pd.read_csv(http_csv_file)
    print(f"HTTP data: {len(http_df)} points from {http_csv_file}")
    
    if len(grpc_df) == 0 and len(http_df) == 0:
        print("No data found in both CSV files")
        return
    
    rps_column = grpc_df.columns[0] if len(grpc_df) > 0 else http_df.columns[0]
    latency_column = grpc_df.columns[1] if len(grpc_df) > 0 else http_df.columns[1]
    
    if len(grpc_df) > 0:
        grpc_df = grpc_df.sort_values(rps_column)
    if len(http_df) > 0:
        http_df = http_df.sort_values(rps_column)
    
    plt.figure(figsize=(14, 9))
    
    if len(grpc_df) > 0:
        plt.plot(grpc_df[rps_column], grpc_df[latency_column], 'ro-', 
                 linewidth=3, markersize=10, markerfacecolor='red', 
                 markeredgecolor='darkred', markeredgewidth=1.5,
                 label='gRPC')
        
        for i, (rps, latency) in enumerate(zip(grpc_df[rps_column], grpc_df[latency_column])):
            plt.annotate(f'{latency:.1f}ms', 
                        (rps, latency), 
                        textcoords="offset points", 
                        xytext=(0,15), 
                        ha='center', 
                        fontsize=9,
                        color='red',
                        weight='bold')
    

    if len(http_df) > 0:
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
    
    plt.xlabel('RPS (Requests Per Second)', fontsize=14)
    plt.ylabel('Latency (ms)', fontsize=14)
    plt.title('Latency vs RPS: gRPC vs HTTP Comparison', fontsize=16, fontweight='bold')
    plt.grid(True, alpha=0.3)
    plt.legend(fontsize=12)
    

    all_rps = []
    all_latency = []
    
    if len(grpc_df) > 0:
        all_rps.extend(grpc_df[rps_column].tolist())
        all_latency.extend(grpc_df[latency_column].tolist())
    if len(http_df) > 0:
        all_rps.extend(http_df[rps_column].tolist())
        all_latency.extend(http_df[latency_column].tolist())
    
    if all_rps and all_latency:
        plt.xlim(0, max(all_rps) * 1.05)
        plt.ylim(0, max(all_latency) * 1.15)
    
    plt.tight_layout()
    
    plt.savefig(output_plot, dpi=300, bbox_inches='tight')
    print(f"Comparison plot saved to: {output_plot}")
    
    print("\n=== STATISTICS ===")
    if len(grpc_df) > 0:
        grpc_min = grpc_df[latency_column].min()
        grpc_max = grpc_df[latency_column].max()
        grpc_avg = grpc_df[latency_column].mean()
        print(f"gRPC Latency: {grpc_min:.1f} - {grpc_max:.1f} ms (avg: {grpc_avg:.1f} ms)")
    
    if len(http_df) > 0:
        http_min = http_df[latency_column].min()
        http_max = http_df[latency_column].max()
        http_avg = http_df[latency_column].mean()
        print(f"HTTP Latency: {http_min:.1f} - {http_max:.1f} ms (avg: {http_avg:.1f} ms)")
    

    if len(grpc_df) > 0 and len(http_df) > 0:
        common_rps = set(grpc_df[rps_column]).intersection(set(http_df[rps_column]))
        if common_rps:
            print(f"\nCommon RPS values for comparison: {sorted(common_rps)}")
            for rps in sorted(common_rps):
                grpc_lat = grpc_df[grpc_df[rps_column] == rps][latency_column].iloc[0]
                http_lat = http_df[http_df[rps_column] == rps][latency_column].iloc[0]
                diff = http_lat - grpc_lat
                diff_percent = (diff / grpc_lat) * 100
                print(f"  RPS {rps}: gRPC={grpc_lat:.1f}ms, HTTP={http_lat:.1f}ms, diff={diff:+.1f}ms ({diff_percent:+.1f}%)")

def main():
    if len(sys.argv) != 4:
        print("Usage: python3 plot_latency_comparison.py <grpc_csv> <http_csv> <output_png>")
        print("Example: python3 plot_latency_comparison.py latency_grpc.csv latency_http.csv comparison_plot.png")
        sys.exit(1)
    
    grpc_csv_file = sys.argv[1]
    http_csv_file = sys.argv[2]
    output_plot = sys.argv[3]
    
    if not os.path.exists(grpc_csv_file):
        print(f"gRPC CSV file not found: {grpc_csv_file}")
        sys.exit(1)
    
    if not os.path.exists(http_csv_file):
        print(f"HTTP CSV file not found: {http_csv_file}")
        sys.exit(1)
    
    os.makedirs(os.path.dirname(output_plot) if os.path.dirname(output_plot) else '.', exist_ok=True)
    
    plot_latency_comparison(grpc_csv_file, http_csv_file, output_plot)

if __name__ == "__main__":
    main()