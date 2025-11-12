import json
import matplotlib.pyplot as plt
import numpy as np

def plot_ghz_summary(json_file_path, rps, output_image_path=None):
    """
    Строит два графика на одном plot:
    - слева: гистограмма по fixed_bins_300ms
    - справа: latency percentiles
    """
    
    # Читаем данные из JSON файла
    with open(json_file_path, 'r') as f:
        data = json.load(f)
    
    # Создаем фигуру с двумя подграфиками
    fig, (ax1, ax2) = plt.subplots(1, 2, figsize=(15, 6))
    
    # Первый график: гистограмма fixed_bins_300ms
    if 'fixed_bins_300ms' in data:
        fixed_bins = data['fixed_bins_300ms']
        labels = fixed_bins['labels']
        counts = fixed_bins['counts']
        
        # Создаем гистограмму
        x_pos = np.arange(len(labels))
        bars = ax1.bar(x_pos, counts, color='skyblue', edgecolor='black', alpha=0.7)
        
        # Настраиваем график
        ax1.set_xlabel('Время ответа (мс)')
        ax1.set_ylabel('Количество запросов')
        ax1.set_title('Гистограмма времени ответа\n(бины по 300мс)')
        ax1.set_xticks(x_pos)
        ax1.set_xticklabels(labels, rotation=45, ha='right')
        
        # Добавляем значения над столбцами
        for bar, count in zip(bars, counts):
            height = bar.get_height()
            if height > 0:  # Подписываем только ненулевые значения
                ax1.text(bar.get_x() + bar.get_width()/2., height + 0.1,
                        f'{count:.1f}', ha='center', va='bottom', fontsize=8)
        
        # Настраиваем сетку
        ax1.grid(True, alpha=0.3, axis='y')
    
    # Второй график: latency percentiles
    if 'latency_percentiles' in data:
        percentiles_data = data['latency_percentiles']
        
        # Подготавливаем данные для графика
        percentiles = list(percentiles_data.keys())
        latency_values = list(percentiles_data.values())
        
        # Создаем гистограмму (столбчатую диаграмму)
        x_pos = np.arange(len(percentiles))
        bars = ax2.bar(x_pos, latency_values, color='lightgreen', 
                      edgecolor='darkgreen', alpha=0.7, width=0.6)
        
        # Настраиваем график
        ax2.set_xlabel('Персентиль (%)')
        ax2.set_ylabel('Время ответа (мс)')
        ax2.set_title('Персентили времени ответа')
        ax2.grid(True, alpha=0.3, axis='y')
        
        # Устанавливаем подписи по оси X
        ax2.set_xticks(x_pos)
        ax2.set_xticklabels([f'p{p}' for p in percentiles])
        
        # Добавляем подписи значений над столбцами
        for bar, value in zip(bars, latency_values):
            height = bar.get_height()
            ax2.text(bar.get_x() + bar.get_width()/2., height + 0.1,
                    f'{value:.1f}ms', ha='center', va='bottom', fontsize=9)
        
        # Автоматически настраиваем пределы по Y для лучшего отображения
        max_latency = max(latency_values)
        ax2.set_ylim(0, max_latency * 1.1)  # +10% от максимального значения
    # Настраиваем общий вид
    plt.suptitle(f'RPS {rps}', fontsize=16, fontweight='bold')
    plt.tight_layout()
    
    # Сохраняем или показываем график
    if output_image_path:
        plt.savefig(output_image_path, dpi=300, bbox_inches='tight')
        print(f"Graph saved to: {output_image_path}")
    else:
        plt.show()

def main():
    import sys
    
    if len(sys.argv) < 2:
        print("Usage: python3 plot_ghz_summary.py <json_file> [output_image]")
        print("Example: python3 plot_ghz_summary.py ghz_reduced.json results_plot.png")
        sys.exit(1)
    
    json_file = sys.argv[1]
    output_image = sys.argv[2] if len(sys.argv) > 2 else None
    
    rps = sys.argv[3]
    try:
        plot_ghz_summary(json_file, rps, output_image)
    except Exception as e:
        print(f"Error: {e}")
        sys.exit(1)

if __name__ == "__main__":
    main()