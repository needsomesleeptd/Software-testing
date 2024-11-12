import os
import json
import numpy as np
import matplotlib.pyplot as plt

from collections import defaultdict

def read_json_files(directory):
    """Reads all JSON files in the given directory and returns them as a list."""
    all_data = []
    for filename in os.listdir(directory):
        if filename.endswith('.json'):  # Adjust if you have a different file extension
            file_path = os.path.join(directory, filename)
            with open(file_path, 'r') as f:
                data = json.load(f)
                all_data.append(data)
    return all_data

def aggregate_data(data_list):
    """Aggregates RAM, CPU, and HTTP data from multiple test runs."""
    aggregated_ram = []
    aggregated_cpu = []
    aggregated_http = []

    for data in data_list:
        # Extract metrics from each run
        ram_usage = data['echo-hello:8080']['go_memstats_alloc_bytes']
        cpu_usage = data['echo-hello:8080'].get('process_cpu_seconds_total', [])
        durations_sum = data['echo-hello:8080']['echo_http_request_duration_seconds_sum']
        counts = data['echo-hello:8080']['echo_http_request_duration_seconds_count']

        # Aggregate RAM usage
        aggregated_ram.extend([int(value) for value in ram_usage])
        
        # Aggregate CPU usage
        aggregated_cpu.extend([float(value) for value in cpu_usage])
        
        # Calculate mean durations and aggregate
        for duration_sum, count in zip(durations_sum, counts):
            if float(count) > 0:
                mean_duration = float(duration_sum) / float(count)
                aggregated_http.append(mean_duration)
            else:
                aggregated_http.append(0)  # Append 0 if count is 0 to avoid division issues

    return aggregated_ram, aggregated_cpu, aggregated_http

def calculate_statistics(data):
    """Calculates statistics for given data."""
    if not data:
        return 0, 0, 0, 0, 0  # Default if no data
    data_array = np.array(data)
    mean_val = np.mean(data_array)
    min_val = np.min(data_array)
    max_val = np.max(data_array)
    quantiles = np.percentile(data_array, [25, 50, 75])  # 25th, 50th, and 75th percentiles
    return mean_val, min_val, max_val, quantiles

def plot_metrics(data, title, ylabel, color):
    """Plots metrics with mean, min, max, and quantiles."""
    mean_val, min_val, max_val, quantiles = calculate_statistics(data)

    plt.figure(figsize=(10, 6))
    
    # Plotting mean, min, max, and quantiles
    plt.bar(['Min', '25th Percentile', 'Mean', 'Median', '75th Percentile', 'Max'], 
            [min_val, quantiles[0], mean_val, quantiles[1], quantiles[2], max_val],
            color=color, alpha=0.6, capsize=5)
    
    plt.title(title)
    plt.ylabel(ylabel)
    plt.grid(axis='y')
    plt.show()

# Main execution flow
directory = '/home/andrew/uni/testing/Software-testing/src/python_scripts/data/echo'  # Update with the correct path
json_data = read_json_files(directory)
aggregated_ram, aggregated_cpu, aggregated_http = aggregate_data(json_data)

# Plot metrics on separate graphs with stats
plot_metrics(aggregated_ram, 'RAM Usage Statistics', 'RAM Usage (bytes)', 'blue')
plot_metrics(aggregated_cpu, 'CPU Usage Statistics', 'CPU Usage (seconds)', 'green')
plot_metrics(aggregated_http, 'HTTP Request Time Statistics', 'Mean HTTP Request Time (seconds)', 'red')
