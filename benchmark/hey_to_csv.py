import sys
import csv
import pandas as pd
import os

def hey_to_csv(hey_output_file, rps):
    """Конвертирует вывод hey в CSV формат, аналогичный ghz"""
    
    try:
        
        df = pd.read_csv(hey_output_file)
        output_data = []
        for latency in df['response-time']:
            output_data.append({
                'rps': rps,
                'duration (ms)': latency*1000,
                'status': 'OK'  
            })
        

        with open(hey_output_file, 'w', newline='') as f:
            writer = csv.DictWriter(f, fieldnames=['rps', 'duration (ms)', 'status'])
            writer.writeheader()
            writer.writerows(output_data)
            
        print(f"Converted hey output to ghz format: {hey_output_file}")
        print(f"Processed {len(output_data)} requests for RPS {rps}")
        
    except Exception as e:
        print(f"Error converting hey output: {e}")
        with open(hey_output_file, 'w', newline='') as f:
            writer = csv.DictWriter(f, fieldnames=['rps', 'duration (ms)', 'status'])
            writer.writeheader()
        print(f"Created empty template: {hey_output_file}")

if __name__ == "__main__":
    if len(sys.argv) != 3:
        print("Usage: python3 hey_to_csv.py <hey_output_file> <rps>")
        sys.exit(1)
    
    hey_output_file = sys.argv[1]
    rps = int(sys.argv[2])
    hey_to_csv(hey_output_file, rps)
