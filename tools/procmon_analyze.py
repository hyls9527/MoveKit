"""
MoveKit ProcMon CSV Analyzer
用法：把 ProcMon 导出的 CSV 保存到 trace.csv，然后运行 python procmon_analyze.py trace.csv
"""
import csv
import sys
import json
from collections import Counter, defaultdict

def analyze(csv_path):
    results = {
        'total_operations': 0,
        'by_type': Counter(),
        'by_result': Counter(),
        'registry_writes': [],
        'file_operations': [],
        'junction_operations': [],
        'failures': [],
        'errors': [],
        'paths_accessed': set(),
    }

    with open(csv_path, 'r', encoding='utf-8', errors='ignore') as f:
        reader = csv.DictReader(f)
        for row in reader:
            op = row.get('Operation', '')
            path = row.get('Path', '')
            result = row.get('Result', '')
            detail = row.get('Detail', '')

            results['total_operations'] += 1
            results['by_type'][op] += 1
            results['by_result'][result] += 1

            # Registry operations
            if 'Reg' in op and 'User Shell Folders' in path:
                results['registry_writes'].append({
                    'op': op, 'path': path, 'result': result, 'detail': detail
                })

            # File operations on MoveKit paths
            if any(x in path for x in ['MoveKit', 'Desktop', 'Downloads', 'Music', 'Documents', 'Pictures']):
                results['file_operations'].append({
                    'op': op, 'path': path, 'result': result
                })

            # Junction operations
            if 'FSCTL' in op or 'REPARSE' in path.upper():
                results['junction_operations'].append({
                    'op': op, 'path': path, 'result': result, 'detail': detail
                })

            # Failures
            if result not in ('SUCCESS', 'REPARSE', 'BUFFER OVERFLOW'):
                results['failures'].append({
                    'op': op, 'path': path, 'result': result, 'detail': detail
                })

            results['paths_accessed'].add(path)

    return results

def print_report(results):
    print("=" * 60)
    print("MoveKit ProcMon Analysis Report")
    print("=" * 60)

    print(f"\nTotal operations: {results['total_operations']}")

    print(f"\n--- Operation Types ---")
    for op, count in results['by_type'].most_common(15):
        print(f"  {op}: {count}")

    print(f"\n--- Results ---")
    for res, count in results['by_result'].most_common():
        marker = "❌" if res not in ('SUCCESS', 'REPARSE', 'BUFFER OVERFLOW') else "✅"
        print(f"  {marker} {res}: {count}")

    if results['registry_writes']:
        print(f"\n--- Registry Operations on User Shell Folders ---")
        for op in results['registry_writes']:
            print(f"  {op['op']} {op['path']}")
            if op['detail']:
                print(f"    Detail: {op['detail']}")

    if results['junction_operations']:
        print(f"\n--- Junction Operations ---")
        for op in results['junction_operations']:
            print(f"  {op['op']} {op['path']} → {op['result']}")

    if results['failures']:
        print(f"\n--- ❌ Failures ({len(results['failures'])}) ---")
        for f in results['failures'][:20]:
            print(f"  {f['op']} {f['path'][:80]} → {f['result']}")
            if f['detail']:
                print(f"    {f['detail']}")

    print(f"\n--- Unique paths accessed: {len(results['paths_accessed'])} ---")

if __name__ == '__main__':
    if len(sys.argv) < 2:
        print("Usage: python procmon_analyze.py <trace.csv>")
        sys.exit(1)
    results = analyze(sys.argv[1])
    print_report(results)
