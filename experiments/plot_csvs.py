#!/usr/bin/env python3

import argparse
import csv
import os
from pathlib import Path

import matplotlib
matplotlib.use("Agg")
import matplotlib.pyplot as plt


def read_csv(path: Path):
    with path.open(newline="", encoding="utf-8") as f:
        return list(csv.DictReader(f))


def ensure_dir(path: Path):
    path.mkdir(parents=True, exist_ok=True)


def to_float(x: str, default=None):
    try:
        return float(x)
    except Exception:
        return default


def to_int(x: str, default=None):
    try:
        return int(float(x))
    except Exception:
        return default


def plot_bully(rows, outdir: Path):
    # run,nodes,failover_ms,new_master_priority,new_master_port
    xs = [to_int(r.get("run")) for r in rows]
    ys = [to_int(r.get("failover_ms")) for r in rows]

    plt.figure(figsize=(7, 4))
    plt.plot(xs, ys, marker="o")
    plt.title("Bully failover time")
    plt.xlabel("run")
    plt.ylabel("failover (ms)")
    plt.grid(True, alpha=0.3)
    plt.tight_layout()
    plt.savefig(outdir / "bully_failover.png", dpi=200)
    plt.close()

    plt.figure(figsize=(7, 4))
    plt.hist([y for y in ys if y is not None and y >= 0], bins=10)
    plt.title("Bully failover time (hist)")
    plt.xlabel("failover (ms)")
    plt.ylabel("count")
    plt.grid(True, alpha=0.3)
    plt.tight_layout()
    plt.savefig(outdir / "bully_failover_hist.png", dpi=200)
    plt.close()


def plot_deadlock(rows, outdir: Path):
    # run,resource_a,resource_b,resolve_ms,victim_priority
    xs = [to_int(r.get("run")) for r in rows]
    ys = [to_int(r.get("resolve_ms")) for r in rows]

    plt.figure(figsize=(7, 4))
    plt.plot(xs, ys, marker="o")
    plt.title("Deadlock resolve time")
    plt.xlabel("run")
    plt.ylabel("resolve (ms)")
    plt.grid(True, alpha=0.3)
    plt.tight_layout()
    plt.savefig(outdir / "deadlock_resolve.png", dpi=200)
    plt.close()

    plt.figure(figsize=(7, 4))
    plt.hist([y for y in ys if y is not None and y >= 0], bins=10)
    plt.title("Deadlock resolve time (hist)")
    plt.xlabel("resolve (ms)")
    plt.ylabel("count")
    plt.grid(True, alpha=0.3)
    plt.tight_layout()
    plt.savefig(outdir / "deadlock_resolve_hist.png", dpi=200)
    plt.close()


def plot_lock(rows, outdir: Path):
    # run,node_idx,priority,wait_ms,order
    orders = [to_int(r.get("order")) for r in rows]
    waits = [to_int(r.get("wait_ms")) for r in rows]

    # boxplot by acquisition order
    by_order = {}
    for r in rows:
        o = to_int(r.get("order"))
        w = to_int(r.get("wait_ms"))
        if o is None or w is None:
            continue
        by_order.setdefault(o, []).append(w)

    keys = sorted(by_order.keys())
    data = [by_order[k] for k in keys]

    plt.figure(figsize=(8, 4))
    if data:
        plt.boxplot(data, labels=[str(k) for k in keys], showfliers=False)
    plt.title("Lock wait time by acquisition order")
    plt.xlabel("order")
    plt.ylabel("wait (ms)")
    plt.grid(True, axis="y", alpha=0.3)
    plt.tight_layout()
    plt.savefig(outdir / "lock_wait_boxplot_by_order.png", dpi=200)
    plt.close()

    # scatter order vs wait
    plt.figure(figsize=(7, 4))
    plt.scatter(orders, waits, s=18)
    plt.title("Lock wait scatter")
    plt.xlabel("order")
    plt.ylabel("wait (ms)")
    plt.grid(True, alpha=0.3)
    plt.tight_layout()
    plt.savefig(outdir / "lock_wait_scatter.png", dpi=200)
    plt.close()


def plot_streaming(rows, outdir: Path):
    # run,size_bytes,upload_ms,upload_mbps,download_ms,download_mbps
    xs = [to_int(r.get("run")) for r in rows]
    up = [to_float(r.get("upload_mbps"), default=None) for r in rows]
    down = [to_float(r.get("download_mbps"), default=None) for r in rows]

    plt.figure(figsize=(7, 4))
    plt.plot(xs, up, marker="o", label="upload")
    plt.plot(xs, down, marker="o", label="download")
    plt.title("Streaming throughput")
    plt.xlabel("run")
    plt.ylabel("Mbps")
    plt.grid(True, alpha=0.3)
    plt.legend()
    plt.tight_layout()
    plt.savefig(outdir / "streaming_throughput.png", dpi=200)
    plt.close()


def main():
    p = argparse.ArgumentParser()
    p.add_argument(
        "--csv-dir",
        default=str(Path(__file__).resolve().parents[1] / "backend" / "experiments"),
        help="directory containing CSVs (default: backend/experiments)",
    )
    p.add_argument(
        "--out-dir",
        default=str(Path(__file__).resolve().parent / "plots"),
        help="output directory for PNGs (default: experiments/plots)",
    )
    args = p.parse_args()

    csv_dir = Path(args.csv_dir)
    out_dir = Path(args.out_dir)
    ensure_dir(out_dir)

    handlers = {
        "bully.csv": plot_bully,
        "deadlock.csv": plot_deadlock,
        "lock.csv": plot_lock,
        "streaming.csv": plot_streaming,
    }

    produced = 0
    for name, fn in handlers.items():
        path = csv_dir / name
        if not path.exists():
            print(f"skip: {path} (missing)")
            continue
        rows = read_csv(path)
        if not rows:
            print(f"skip: {path} (empty)")
            continue
        fn(rows, out_dir)
        produced += 1
        print(f"ok: plotted {name}")

    if produced == 0:
        raise SystemExit("no plots generated (no CSVs found)")

    print(f"plots in: {out_dir}")


if __name__ == "__main__":
    main()
