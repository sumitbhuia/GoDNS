import re
import pandas as pd
import matplotlib.pyplot as plt

def parse_benchmark_results(filepath):
    """Parses a Go benchmark results file for multiple benchmarks."""
    results = []
    with open(filepath, 'r') as f:
        for line in f:
            # regex to capture benchmark names and their ns/op value
            match = re.search(r'^(Benchmark\w+)(-\d+)?\s+\d+\s+([\d\.]*)\s+ns/op', line)
            if match:
                results.append({
                    'name': match.group(1),
                    'latency_ns': float(match.group(3))
                })
    return pd.DataFrame(results)

def create_latency_chart(df, output_path='benchmark_latency.png'):
    """Creates a box plot for latency benchmarks."""
    latency_df = df[df['name'] == 'BenchmarkServerE2E']
    if latency_df.empty:
        print("No data found for 'BenchmarkServerE2E'. Skipping latency chart.")
        return

    plt.style.use('seaborn-v0_8-whitegrid')
    fig, ax = plt.subplots(figsize=(8, 6))
    latency_df.boxplot(column='latency_ns', by='name', ax=ax, grid=False)
    
    ax.set_title('End-to-End DNS Query Latency', fontsize=16, pad=20)
    ax.set_ylabel('Latency (nanoseconds per op)', fontsize=12)
    ax.set_xlabel('')
    fig.suptitle('')
    plt.xticks(rotation=0)
    plt.tight_layout()
    
    print(f"Saving latency chart to {output_path}...")
    plt.savefig(output_path, dpi=150)
    plt.close(fig)

def create_throughput_chart(df, output_path='benchmark_throughput.png'):
    """Creates a bar chart for throughput benchmarks."""
    throughput_df = df[df['name'] == 'BenchmarkServerThroughput']
    if throughput_df.empty:
        print("No data found for 'BenchmarkServerThroughput'. Skipping throughput chart.")
        return

    # Calculating Queries Per Second
    throughput_df['qps'] = 1_000_000_000 / throughput_df['latency_ns']
    
    # Calculating the average QPS to display on the bar
    avg_qps = throughput_df['qps'].mean()

    plt.style.use('seaborn-v0_8-whitegrid')
    fig, ax = plt.subplots(figsize=(8, 6))

    bars = ax.bar(['Server Throughput'], [avg_qps], color='skyblue', width=0.5)
    
    ax.set_title('Server Throughput', fontsize=16, pad=20)
    ax.set_ylabel('Queries Per Second (QPS)', fontsize=12)
    ax.set_ylim(bottom=0) # Make sure y-axis starts at 0
    
    # Adding the QPS value on top of the bar for clarity
    ax.bar_label(bars, fmt='{:,.0f} QPS'.format, fontsize=12, padding=5)

    print(f"Saving throughput chart to {output_path}...")
    plt.savefig(output_path, dpi=150)
    plt.close(fig)

if __name__ == '__main__':
    df = parse_benchmark_results('bench_results.txt')
    if not df.empty:
        create_latency_chart(df)
        create_throughput_chart(df)
    else:
        print("No benchmark data was found in bench_results.txt.")