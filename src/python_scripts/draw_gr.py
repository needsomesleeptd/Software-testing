import os
import json
import numpy as np
import matplotlib.pyplot as plt

def read_json_files(directory):
    """Reads all JSON files in the given directory and returns them as a list."""
    all_data = []
    for filename in os.listdir(directory):
        if filename.endswith('.json'):  # Assuming your files have a .json extension
            file_path = os.path.join(directory, filename)
            with open(file_path, 'r') as f:
                data = json.load(f)
                all_data.append(data)
    return all_data

def aggregate_data(data_list, instance):
    """Aggregates RAM, CPU, and HTTP data from multiple test runs."""
    aggregated_ram = []
    aggregated_cpu = []
    aggregated_http = []
    aggregated_cpu_percent = []  # For storing CPU usage percentage

    for data in data_list:
        # Extract metrics from each run
        ram_usage = data[instance]['go_memstats_alloc_bytes']
        cpu_usage = data[instance].get('process_cpu_seconds_total', [])
        durations_sum = data[instance]['echo_http_request_duration_seconds_sum']
        counts = data[instance]['echo_http_request_duration_seconds_count']

        # Aggregate RAM usage
        aggregated_ram.append([int(value) for value in ram_usage])

        # Aggregate CPU usage
        aggregated_cpu.append([float(value) for value in cpu_usage])

        # Calculate CPU usage percentage
        cpu_percent = []
        for i in range(len(cpu_usage)):
            if i == 0:
                cpu_time = 0#float(cpu_usage[i])  # First entry
                cpu_percent.append((cpu_time / 1) * 100)  # Assuming sampling interval of 1 second
            else:
                # Calculate usage as the difference between current and previous
                cpu_time = float(cpu_usage[i]) - float(cpu_usage[i - 1])  # CPU seconds used since the last measurement
                cpu_percent.append((cpu_time / 1) * 100)  # Assuming 1 second intervals

        # Aggregate CPU usage percentages
        aggregated_cpu_percent.append(cpu_percent)

        # Calculate mean durations and aggregate
        http_metrics = []
        for duration_sum, count in zip(durations_sum, counts):
            if float(count) > 0:
                mean_duration = float(duration_sum) / float(count)
                http_metrics.append(mean_duration)
            else:
                http_metrics.append(0)  # Append 0 if count is 0 to avoid division issues
                  
        aggregated_http.append(http_metrics)  # Keep per-file HTTP metrics

    return aggregated_ram, aggregated_cpu, aggregated_http, aggregated_cpu_percent

def flatten_stats(l):
    """Flattens the list of lists and calculates means and standard deviations."""
    # Convert the list of lists into a NumPy array
    # Use np.vstack to ensure that inner lists are arranged properly
    min_size = min([len(li) for li in l])
    res_l = [li[:min_size] for li in l]
    data_array = np.vstack(res_l)

    # Calculate the means and standard deviations across the rows
    means = np.mean(data_array, axis=0).tolist()  # Mean of each column
    stddevs = np.std(data_array, axis=0).tolist()  # Stddev of each column

    return means, stddevs

def plot_metrics(data, title, ylabel, color):
    """Plots metrics with mean and standard deviation."""
    means, stdev = flatten_stats(data)

    plt.figure(figsize=(10, 6))
    x_labels = range(1, len(means) + 1)  # Create x labels for plot

    # Plotting means and standard deviations
    plt.errorbar(x_labels, means, yerr=stdev, fmt='o-', label='Mean ± stddev', color=color)
    plt.title(title)
    plt.ylabel(ylabel)
    plt.xlabel('Metric Index')
    plt.xticks(x_labels)  # Show metric indices on x-axis
    plt.grid(axis='y')
    plt.legend()
    plt.show()

# Main execution flow
dirs = ['echo', 'gin']
instances = ['echo-hello:8080', 'gin-hello:8080']

for i in range(len(dirs)):
    directory = '/home/andrew/uni/testing/Software-testing/src/python_scripts/data/' + dirs[i]  # Update with the correct path
    json_data = read_json_files(directory)

    aggregated_ram, aggregated_cpu, aggregated_http, aggregated_cpu_percent = aggregate_data(json_data, instance=instances[i])

    # Plot metrics with stats
    plot_metrics(aggregated_ram, f'{instances[i]} RAM Usage Statistics', 'RAM Usage (bytes)', 'blue')
    plot_metrics(aggregated_cpu, f'{instances[i]} CPU Usage Statistics', 'CPU Usage (seconds)', 'green')
    plot_metrics(aggregated_http, f'{instances[i]} HTTP Request Time Statistics', 'Mean HTTP Request Time (seconds)', 'red')
    plot_metrics(aggregated_cpu_percent, f'{instances[i]} CPU Usage Percentage', 'CPU Usage (%)', 'orange')  # Plot CPU percentage