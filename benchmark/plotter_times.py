import pandas as pd
import matplotlib.pyplot as plt
import glob
import os
import sys
import re

def quick_plot(directory, output_file):
    files = glob.glob(os.path.join(directory, "grpc_rps_*.csv"))
    
    data = []
    for f in sorted(files):
        rps = int(re.search(r'grpc_rps_(\d+)', os.path.basename(f)).group(1))
        df = pd.read_csv(f)
        mean_time = df.iloc[:, 0].mean()  
        data.append((rps, mean_time))
    
    rps, times = zip(*sorted(data))
    
    plt.figure(figsize=(10, 6))
    plt.plot(rps, times, 'bo-', linewidth=2, markersize=6)
    plt.xlabel('RPS')
    plt.ylabel('Среднее время (ms)')
    plt.title('GRPC: Latency(RPS)')
    plt.grid(True)
    plt.savefig(output_file, dpi=300, bbox_inches='tight')
    #plt.show()


quick_plot('./out_grow', sys.argv[1])