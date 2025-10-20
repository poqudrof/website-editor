#!/usr/bin/env python3
"""
Example AI Agent Script for Testing

This script simulates an AI agent that can be run via the AI Agent API.
It demonstrates streaming output and graceful interruption handling.

Usage:
    python -u example-agent.py [--task TASK] [--delay DELAY]

Options:
    --task TASK      Task to perform: count, analyze, process (default: count)
    --delay DELAY    Delay between outputs in seconds (default: 0.5)
"""

import sys
import time
import argparse
import signal

# Global flag for handling interrupts
interrupted = False

def signal_handler(sig, frame):
    """Handle interrupt signals gracefully"""
    global interrupted
    print("\n[INFO] Received interrupt signal, shutting down gracefully...")
    interrupted = True

# Register signal handler
signal.signal(signal.SIGINT, signal_handler)
signal.signal(signal.SIGTERM, signal_handler)

def task_count(delay=0.5):
    """Simple counting task"""
    print("[INFO] Starting counting task...")
    for i in range(1, 21):
        if interrupted:
            break
        print(f"Count: {i}/20")
        time.sleep(delay)
    print("[SUCCESS] Counting task completed!")

def task_analyze(delay=0.5):
    """Simulated analysis task"""
    steps = [
        "Loading data...",
        "Preprocessing data...",
        "Analyzing patterns...",
        "Computing statistics...",
        "Generating insights...",
        "Validating results...",
        "Preparing report...",
        "Finalizing analysis..."
    ]

    print("[INFO] Starting analysis task...")
    for i, step in enumerate(steps, 1):
        if interrupted:
            break
        print(f"[{i}/{len(steps)}] {step}")
        time.sleep(delay)

        # Simulate progress
        if i == 3:
            print("  → Found 42 patterns")
        elif i == 4:
            print("  → Mean: 123.45, Median: 118.20")
        elif i == 5:
            print("  → Key insight: Trend is increasing by 15%")

    if not interrupted:
        print("[SUCCESS] Analysis completed successfully!")
        print("[RESULT] Overall score: 87.5/100")

def task_process(delay=0.5):
    """Simulated data processing task"""
    print("[INFO] Starting data processing task...")

    files = [
        "data_001.csv",
        "data_002.csv",
        "data_003.csv",
        "data_004.csv",
        "data_005.csv"
    ]

    for i, filename in enumerate(files, 1):
        if interrupted:
            break

        print(f"[{i}/{len(files)}] Processing {filename}...")
        time.sleep(delay * 0.5)

        # Simulate processing steps
        steps = ["Reading", "Parsing", "Transforming", "Validating", "Writing"]
        for step in steps:
            if interrupted:
                break
            print(f"  • {step}... OK")
            time.sleep(delay * 0.2)

        print(f"  ✓ {filename} processed successfully")

    if not interrupted:
        print("[SUCCESS] All files processed!")
        print(f"[RESULT] Processed {len(files)} files, 0 errors")

def main():
    parser = argparse.ArgumentParser(
        description='Example AI Agent for testing the AI Agent API'
    )
    parser.add_argument(
        '--task',
        choices=['count', 'analyze', 'process'],
        default='count',
        help='Task to perform (default: count)'
    )
    parser.add_argument(
        '--delay',
        type=float,
        default=0.5,
        help='Delay between outputs in seconds (default: 0.5)'
    )

    args = parser.parse_args()

    print("=" * 50)
    print("AI Agent Example Script")
    print("=" * 50)
    print(f"Task: {args.task}")
    print(f"Delay: {args.delay}s")
    print("=" * 50)
    print()

    try:
        if args.task == 'count':
            task_count(args.delay)
        elif args.task == 'analyze':
            task_analyze(args.delay)
        elif args.task == 'process':
            task_process(args.delay)

        if interrupted:
            print("\n[WARNING] Task was interrupted!")
            sys.exit(1)
        else:
            print("\n[INFO] Agent finished successfully")
            sys.exit(0)

    except KeyboardInterrupt:
        print("\n[WARNING] Keyboard interrupt received")
        sys.exit(1)
    except Exception as e:
        print(f"\n[ERROR] Unexpected error: {e}", file=sys.stderr)
        sys.exit(1)

if __name__ == '__main__':
    main()
