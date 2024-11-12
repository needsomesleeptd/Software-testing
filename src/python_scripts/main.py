import threading
import time
from typing import List, Dict
import json
import os

import httpx
from prometheus_client import CollectorRegistry, Gauge, generate_latest
from datetime import datetime
import pytz

# --- Load Testing Scripts ---

def load_test(url: str, num_users: int, requests_per_user: int, duration: int):
    """
    Runs a load test with the specified number of users and requests per user for a given duration.

    Args:
        url: The URL to target.
        num_users: The number of simulated users.
        requests_per_user: The number of requests each user will make.
        duration: The test duration in seconds.
    """

    def user_task(user_id: int):
        with httpx.Client() as client:
            for _ in range(requests_per_user):
                res = client.get(url)
        print(f'client {user_id} finished')
                #print(f'client_{user_id}: {res.status_code}, {res.text[:100]}')  # Print status code and first 100 chars

    start_time = time.time()
    
    threads = [threading.Thread(target=user_task, args=(i,)) for i in range(num_users)]

    for thread in threads:
        thread.start()

    for thread in threads:
        thread.join()

    end_time = time.time()
    
    # Convert to datetime objects
    moscow_tz = pytz.timezone('Europe/Moscow')
    start_time_dt = datetime.fromtimestamp(start_time, moscow_tz)
    end_time_dt = datetime.fromtimestamp(end_time, moscow_tz)

    elapsed_time = end_time - start_time

    print(f"Load test completed in {elapsed_time:.2f} seconds.")
    return start_time_dt, end_time_dt
  # Wait for the specified duration, even if the load test finishes early
  #time.sleep(duration - elapsed_time) # Ensure the test runs for the full duration


def get_prometheus_data(host: str, port: int, metrics: List[str], start_time: str, end_time: str,instance) -> Dict[str, Dict[str, List[float]]]:
  """
  Fetches data from a Prometheus server for a specific time range, split by instance.

  Args:
    host: The Prometheus server hostname.
    port: The Prometheus server port.
    metrics: A list of metric names to retrieve.
    start_time: The start time of the query in Prometheus' time format (e.g., "2023-10-26T10:00:00Z").
    end_time: The end time of the query in Prometheus' time format (e.g., "2023-10-26T12:00:00Z").

  Returns:
    A dictionary with instances as keys, each containing a dictionary mapping metric names to lists of data points.
  """
  data = {}
  for metric in metrics:
    for instance in get_instances(host, port,[instance]):
      if instance not in data:
        data[instance] = {}
      metric_data = _fetch_metric_data(host, port, metric, start_time, end_time, instance)
      if metric_data:
        data[instance][metric] = metric_data
  return data

def get_instances(host: str, port: int,instance: str) -> List[str]:
  """
  Retrieves a list of unique instances from Prometheus.
  """
  url = f"http://{host}:{port}/api/v1/query"
  params = {'query': 'up'}
  try:
    response = httpx.get(url, params=params)
    response.raise_for_status()
    data = response.json()
    #print(data,instance)
    instances = [result['metric']['instance'] for result in data['data']['result'] if  result['metric']['instance'] in instance]
    return instances
  except Exception as e:
    print(f"Error fetching instances: {e}")
    return []


def _fetch_metric_data(host: str, port: int, metric: str, start_time: str, end_time: str, instance: str) -> List[float]:
  """
  Fetches data for a single metric from a specific instance.
  """
  registry = CollectorRegistry()
  url = f"http://{host}:{port}/api/v1/query_range"
  params = {
    'query': f'{metric}{{instance="{instance}"}}',
    'start': start_time,
    'end': end_time,
    'step': '1s' # Adjust step for desired sampling resolution (e.g., 1s, 5m, 1h)
  }
  try:
    response = httpx.get(url, params=params)
    response.raise_for_status()
    data = response.json()
    # Process data
    results = data.get('data', {}).get('result', [])
    if results:
      values = [value[1] for value in results[0].get('values', [])]
      return values
    else:
      return []
  except Exception as e:
    print(f"Error fetching metric data: {e}")
    return []

# --- Save Data to Local FS ---

def save_data_to_file(data: Dict[str, List[float]], filename: str):
  """
  Saves the data to a JSON file.

  Args:
    data: The data to be saved.
    filename: The name of the output file.
  """
  try:
    os.makedirs(os.path.dirname(filename), exist_ok=True)
    with open(filename, 'w') as f:
      json.dump(data, f, indent=4)
    print(f"Data saved to {filename}")
  except Exception as e:
    print(f"Error saving data to file: {e}")

# --- Example Usage ---

def main():
  tests_url = ["http://localhost:8081/","http://localhost:8082/"]
  num_users = 100
  requests_per_user = 200
  test_duration = 30
  prometheus_host = "localhost"
  prometheus_port = 9090
  metrics_to_fetch = ["go_memstats_alloc_bytes","echo_http_request_duration_seconds_sum","echo_http_request_duration_seconds_count","process_cpu_seconds_total"]
  instances  = ['echo','gin']
  tests_num = 20
  
  for i,test_url in enumerate(tests_url):
    for _ in range(tests_num):
      instance = instances[i]
      
    # 1. Run the Load Test
      start_time,end_time = load_test(test_url, num_users, requests_per_user, test_duration)

      # 2. Get Prometheus Data
      #start_time_f = time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime(time.time() - test_duration))
      #end_time_f = time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime())
      start_time_f = start_time.astimezone(pytz.utc).strftime("%Y-%m-%dT%H:%M:%SZ")
      end_time_f = end_time.astimezone(pytz.utc).strftime("%Y-%m-%dT%H:%M:%SZ")
      
      
      output_filename = f"./data/{instance}/{requests_per_user * num_users}_{start_time}.json"
      print(f'{start_time_f},{end_time_f}')
      prometheus_data = get_prometheus_data(prometheus_host, prometheus_port, metrics_to_fetch, start_time_f, end_time_f,instance+"-hello:8080")
      # 3. Save Data to Local FS
      save_data_to_file(prometheus_data, output_filename)

if __name__ == "__main__":
  main()