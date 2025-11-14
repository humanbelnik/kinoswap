import json
import matplotlib.pyplot as plt
import numpy as np

def plot_ghz_comparison(grpc_json_file, http_json_file, rps, output_image_path=None):
    with open(grpc_json_file, 'r') as f:
        grpc_data = json.load(f)
    
    with open(http_json_file, 'r') as f:
        http_data = json.load(f)
    
    fig, (ax1, ax2) = plt.subplots(1, 2, figsize=(16, 6))
    
    
    if 'fixed_bins_300ms' in grpc_data and 'fixed_bins_300ms' in http_data:
        grpc_fixed_bins = grpc_data['fixed_bins_300ms']
        http_fixed_bins = http_data['fixed_bins_300ms']
        
        labels = grpc_fixed_bins['labels']
        grpc_counts = grpc_fixed_bins['counts']
        http_counts = http_fixed_bins['counts']
        
        x_pos = np.arange(len(labels))
        bar_width = 0.35
        
        grpc_bars = ax1.bar(x_pos - bar_width/2, grpc_counts, bar_width, 
                           color='red', edgecolor='darkred', alpha=0.7, 
                           label='gRPC')
        http_bars = ax1.bar(x_pos + bar_width/2, http_counts, bar_width, 
                           color='blue', edgecolor='darkblue', alpha=0.7, 
                           label='HTTP')
        
        ax1.set_xlabel('Время ответа (мс)')
        ax1.set_ylabel('Количество запросов')
        ax1.set_title('Сравнение гистограмм времени ответа\n(бины по 150мс)')
        ax1.set_xticks(x_pos)
        ax1.set_xticklabels(labels, rotation=45, ha='right')
        ax1.legend()
        
        for bars, counts in [(grpc_bars, grpc_counts), (http_bars, http_counts)]:
            for bar, count in zip(bars, counts):
                height = bar.get_height()
                if height > 0:
                    ax1.text(bar.get_x() + bar.get_width()/2., height + 0.1,
                            f'{count}', ha='center', va='bottom', fontsize=7)
        
        ax1.grid(True, alpha=0.3, axis='y')
    

    if 'latency_percentiles' in grpc_data and 'latency_percentiles' in http_data:
        grpc_percentiles = grpc_data['latency_percentiles']
        http_percentiles = http_data['latency_percentiles']
        
        percentiles = list(grpc_percentiles.keys())
        grpc_latency = [grpc_percentiles[p] for p in percentiles]
        http_latency = [http_percentiles[p] for p in percentiles]
        
        x_pos = np.arange(len(percentiles))
        bar_width = 0.35
        
        grpc_bars = ax2.bar(x_pos - bar_width/2, grpc_latency, bar_width,
                           color='red', edgecolor='darkred', alpha=0.7,
                           label='gRPC')
        http_bars = ax2.bar(x_pos + bar_width/2, http_latency, bar_width,
                           color='blue', edgecolor='darkblue', alpha=0.7,
                           label='HTTP')
        
        
        ax2.set_xlabel('Персентиль (%)')
        ax2.set_ylabel('Время ответа (мс)')
        ax2.set_title('Сравнение персентилей времени ответа')
        ax2.grid(True, alpha=0.3, axis='y')
        ax2.legend()
        

        ax2.set_xticks(x_pos)
        ax2.set_xticklabels([f'p{p}' for p in percentiles])
        
        for bars, latency_values in [(grpc_bars, grpc_latency), (http_bars, http_latency)]:
            for bar, value in zip(bars, latency_values):
                height = bar.get_height()
                ax2.text(bar.get_x() + bar.get_width()/2., height + 0.1,
                        f'{value:.1f}ms', ha='center', va='bottom', fontsize=8)
        

        max_latency = max(max(grpc_latency), max(http_latency))
        ax2.set_ylim(0, max_latency * 1.15)  
    

    plt.suptitle(f'gRPC vs HTTP (RPS {rps})', 
                fontsize=16, fontweight='bold')
    plt.tight_layout()
    
    if output_image_path:
        plt.savefig(output_image_path, dpi=300, bbox_inches='tight')
        print(f"Graph saved to: {output_image_path}")
    else:
        plt.show()

def main():
    import sys
    
    if len(sys.argv) != 4:
        sys.exit(1)
    
    grpc_json_file = sys.argv[1]
    http_json_file = sys.argv[2]
    output_image = sys.argv[3]

    rps = "500"     
    try:
        plot_ghz_comparison(grpc_json_file, http_json_file, rps, output_image)
    except Exception as e:
        print(f"Error: {e}")
        sys.exit(1)

if __name__ == "__main__":
    main()
