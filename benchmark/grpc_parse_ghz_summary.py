import matplotlib.pyplot as plt
import numpy as np
import re
import sys
import json

def parse_ghz_summary(file_path):
    with open(file_path, 'r') as f:
        content = f.read()
    
    data = {
        'histogram_bins': [],
        'histogram_counts': [],
        'latency_percentiles': {}
    }
    
    lines = content.split('\n')
    in_histogram = False
    for line in lines:
        if 'Response time histogram:' in line:
            in_histogram = True
            continue
        if in_histogram and '|' in line:
            parts = line.strip().split('|')
            if len(parts) >= 1:
                left_part = parts[0].strip()
                match = re.search(r'([\d.]+)\s+\[(\d+)\]', left_part)
                if match:
                    bin_value = float(match.group(1))
                    count = int(match.group(2))
                    data['histogram_bins'].append(bin_value)
                    data['histogram_counts'].append(count)
        elif in_histogram and not line.strip():
            in_histogram = False
    
    for line in lines:
        if '% in' in line:
            match = re.search(r'(\d+) % in ([\d.]+)', line)
            if match:
                percentile = int(match.group(1))
                latency = float(match.group(2))
                data['latency_percentiles'][percentile] = latency
    
    return data

def create_fixed_bins(data, bin_step_ms=150):
    if not data['histogram_bins'] or not data['histogram_counts']:
        return [], []

    min_latency = 0
    max_latency = 2000
    
    # Создаем фиксированные бины с шагом 300ms
    max_bin = ((int(max_latency) // bin_step_ms) + 1) * bin_step_ms
    fixed_bins = list(range(0, int(max_bin) + bin_step_ms, bin_step_ms))
    
    # Инициализируем счетчики для фиксированных бинов
    fixed_counts = [0] * (len(fixed_bins) - 1)
    
    for bin_value, count in zip(data['histogram_bins'], data['histogram_counts']):
        bin_index = int(bin_value // bin_step_ms)
        if bin_index < len(fixed_counts):
            fixed_counts[bin_index] += count
        else:

            fixed_counts[-1] += count
    
    bin_labels = [f"{fixed_bins[i]}-{fixed_bins[i+1]}" for i in range(len(fixed_bins)-1)]
    
    return bin_labels, fixed_counts

def main():
    if len(sys.argv) < 3:
        print("Usage: python3 parse_ghz_summary.py <file_in> <file_out>")
        sys.exit(1)
    
    file_path = sys.argv[1]
    output_path = sys.argv[2]
    
    
    parsed_data = parse_ghz_summary(file_path)
    fixed_bin_labels, fixed_bin_counts = create_fixed_bins(parsed_data, 150)
    parsed_data['fixed_bins_300ms'] = {
        'labels': fixed_bin_labels,
        'counts': fixed_bin_counts
    }
    
    with open(output_path, 'w') as f:
        json.dump(parsed_data, f, indent=2)
    
    print(f"Data successfully parsed and saved to {output_path}")

if __name__ == "__main__":
    main()