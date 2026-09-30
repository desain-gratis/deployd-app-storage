#!/usr/bin/env python3
"""
Analyze k6 raw output (CSV or JSON) and plot a latency histogram + boxplot.

Requires:
    pip install pandas matplotlib

Usage:
    python3 analyze_latency.py raw_results.csv
    python3 analyze_latency.py raw_results.json
    python3 analyze_latency.py raw_results.csv --metric http_req_duration --bins 60 --out latency_report.png
"""

import argparse
import json
import sys

import pandas as pd
import matplotlib.pyplot as plt


def load_csv(path, metric):
    df = pd.read_csv(path)
    if "metric_name" not in df.columns or "metric_value" not in df.columns:
        sys.exit(
            "Unexpected CSV format - expected 'metric_name' and 'metric_value' columns. "
            "Make sure this file came from `k6 run --out csv=...`."
        )
    sub = df[df["metric_name"] == metric]
    return sub["metric_value"].astype(float)


def load_json(path, metric):
    """k6 --out json=... writes one JSON object per line ('Point' records)."""
    values = []
    with open(path, "r") as f:
        for line in f:
            line = line.strip()
            if not line:
                continue
            try:
                rec = json.loads(line)
            except json.JSONDecodeError:
                continue
            if rec.get("type") != "Point":
                continue
            if rec.get("metric") != metric:
                continue
            val = rec.get("data", {}).get("value")
            if val is not None:
                values.append(float(val))
    return pd.Series(values)


def main():
    parser = argparse.ArgumentParser(description="Plot histogram + boxplot from k6 raw output")
    parser.add_argument("input", help="Path to k6 raw output file (.csv or .json)")
    parser.add_argument(
        "--metric",
        default="http_req_duration",
        help="Metric to analyze, e.g. http_req_duration, http_req_waiting (default: http_req_duration)",
    )
    parser.add_argument("--bins", type=int, default=50, help="Number of histogram bins (default: 50)")
    parser.add_argument(
        "--out", default="latency_report.png", help="Output image file path (default: latency_report.png)"
    )
    args = parser.parse_args()

    if args.input.lower().endswith(".json"):
        data = load_json(args.input, args.metric)
    else:
        data = load_csv(args.input, args.metric)

    if data.empty:
        sys.exit(f"No data points found for metric '{args.metric}' in {args.input}")

    print(f"Loaded {len(data)} samples for metric '{args.metric}'")

    stats = {
        "count": len(data),
        "min": data.min(),
        "avg": data.mean(),
        "med (p50)": data.median(),
        "p90": data.quantile(0.90),
        "p95": data.quantile(0.95),
        "p99": data.quantile(0.99),
        "p99.9": data.quantile(0.999),
        "p99.99": data.quantile(0.9999),
        "max": data.max(),
    }
    print("\nLatency stats (ms):")
    for k, v in stats.items():
        print(f"  {k:10s}: {v}" if k == "count" else f"  {k:10s}: {v:.2f}")

    if len(data) < 10000:
        print(
            f"\nNote: only {len(data)} samples - p99.99 needs ~10,000+ requests "
            "to be statistically meaningful. Treat it as a rough estimate here."
        )

    fig, (ax1, ax2) = plt.subplots(1, 2, figsize=(14, 6))

    # --- Histogram ---
    ax1.hist(data, bins=args.bins, color="#4C72B0", edgecolor="black", alpha=0.85)
    for label, q in [("p50", 0.50), ("p90", 0.90), ("p95", 0.95), ("p99", 0.99), ("p99.99", 0.9999)]:
        val = data.quantile(q)
        ax1.axvline(val, linestyle="--", linewidth=1, color="firebrick")
        ax1.text(val, ax1.get_ylim()[1] * 0.95, label, rotation=90, va="top", ha="right", fontsize=8)
    ax1.set_title(f"{args.metric} — Histogram (n={len(data)})")
    ax1.set_xlabel("Latency (ms)")
    ax1.set_ylabel("Request count")

    # --- Boxplot ---
    ax2.boxplot(data, vert=True, showfliers=True)
    ax2.set_title(f"{args.metric} — Boxplot")
    ax2.set_ylabel("Latency (ms)")
    ax2.set_xticklabels([args.metric])

    plt.tight_layout()
    plt.savefig(args.out, dpi=150)
    print(f"\nSaved chart to {args.out}")


if __name__ == "__main__":
    main()
