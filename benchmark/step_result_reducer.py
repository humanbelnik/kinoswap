import json
import glob
import re
import numpy as np
from collections import defaultdict
import sys

def read_ghz_files(pattern):
    files = glob.glob(pattern)
    if not files:
        raise ValueError(f"No files found matching pattern: {pattern}")
    
    data_list = []
    for file_path in sorted(files):
        try:
            with open(file_path, 'r') as f:
                data = json.load(f)
                data_list.append(data)
                print(f"Successfully read: {file_path}")
        except Exception as e:
            print(f"Error reading {file_path}: {e}")
    
    return data_list

def calculate_average_data(data_list):
    if not data_list:
        return {}
    
    all_histogram_bins = []
    all_histogram_counts = []
    all_percentiles = defaultdict(list)
    all_fixed_bins_labels = []
    all_fixed_bins_counts = []
    
    # Собираем данные из всех файлов
    for data in data_list:
        # Гистограмма bins (должны быть одинаковыми во всех файлах)
        if 'histogram_bins' in data:
            all_histogram_bins.append(data['histogram_bins'])
        
        # Гистограмма counts
        if 'histogram_counts' in data:
            all_histogram_counts.append(data['histogram_counts'])
        
        # Персентили
        if 'latency_percentiles' in data:
            for percentile, value in data['latency_percentiles'].items():
                all_percentiles[percentile].append(value)
        
        # Фиксированные бины
        if 'fixed_bins_300ms' in data:
            fixed_data = data['fixed_bins_300ms']
            if 'labels' in fixed_data:
                all_fixed_bins_labels.append(fixed_data['labels'])
            if 'counts' in fixed_data:
                all_fixed_bins_counts.append(fixed_data['counts'])
    
    # Вычисляем средние значения
    result = {}
    
    # Средние для гистограммы bins (берем первый, так как они должны быть одинаковыми)
    if all_histogram_bins:
        result['histogram_bins'] = all_histogram_bins[0]
    
    # Средние для гистограммы counts
    if all_histogram_counts:
        # Преобразуем в numpy array для удобства вычислений
        counts_array = np.array(all_histogram_counts)
        avg_counts = np.mean(counts_array, axis=0).tolist()
        result['histogram_counts'] = [int(x) for x in avg_counts]
    
    # Средние для персентилей
    if all_percentiles:
        result['latency_percentiles'] = {}
        for percentile, values in all_percentiles.items():
            avg_value = np.mean(values)
            result['latency_percentiles'][percentile] = round(avg_value, 2)
    
    # Средние для фиксированных бинов
    if all_fixed_bins_counts:
        result['fixed_bins_300ms'] = {}
        
        # Метки берем из первого файла (должны быть одинаковыми)
        if all_fixed_bins_labels:
            result['fixed_bins_300ms']['labels'] = all_fixed_bins_labels[0]
        
        # Средние для счетчиков фиксированных бинов
        fixed_counts_array = np.array(all_fixed_bins_counts)
        avg_fixed_counts = np.mean(fixed_counts_array, axis=0).tolist()
        result['fixed_bins_300ms']['counts'] = [int(x) for x in avg_fixed_counts]
    
    # Добавляем метаинформацию
    result['meta'] = {
        'files_processed': len(data_list),
        'description': 'Average values from all ghz run files'
    }
    
    return result

def main():
    try:
        rps_value = sys.argv[2]
        prefix = sys.argv[1] # http or grpc
        out = sys.argv[3]
        pattern = f"./out/{prefix}_rps_{rps_value}_*.json"
        data_list = read_ghz_files(pattern)
        
        print(f"Processing {len(data_list)} files...")
        
        # Вычисляем средние значения
        reduced_data = calculate_average_data(data_list)
        
        # Сохраняем результат
        output_file = out
        with open(output_file, 'w') as f:
            json.dump(reduced_data, f, indent=2)
        
        print(f"Successfully created {output_file} with averaged data from {len(data_list)} files")
        
        # Выводим краткую статистику
        if 'meta' in reduced_data:
            print(f"Files processed: {reduced_data['meta']['files_processed']}")
        if 'latency_percentiles' in reduced_data:
            print("Average percentiles:")
            for p, v in reduced_data['latency_percentiles'].items():
                print(f"  p{p}: {v}ms")
                
    except Exception as e:
        print(f"Error: {e}")
        sys.exit(1)

if __name__ == "__main__":
    main()