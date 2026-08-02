"""
eval/run_intent_eval.py

Evaluation harness for IntentClassifier (generation/intent_extractor.py).

Runs classify() + classify_multi() against a labeled dataset and reports:
  - overall accuracy
  - per-class precision / recall / F1
  - confusion matrix (expected vs predicted)
  - nav_depth accuracy (for ui_navigation rows that specify one)
  - score-margin (top1 - top2) distribution, split by correct/incorrect —
    use this to pick a real threshold for classify_multi() instead of a guess

Usage:
    python -m eval.run_intent_eval
    python -m eval.run_intent_eval --dataset eval/intent_eval_dataset.jsonl --csv eval/results.csv

Add more rows to intent_eval_dataset.jsonl as you collect real queries —
the seed set here is a starting point, not ground truth.
"""
from __future__ import annotations

import argparse
import csv
import json
from collections import Counter, defaultdict
from pathlib import Path
from statistics import mean, median
from typing import Any


# ---------------------------------------------------------------------------
# Dataset loading
# ---------------------------------------------------------------------------

def load_dataset(path: Path) -> list[dict]:
    rows = []
    with path.open(encoding="utf-8") as f:
        for lineno, line in enumerate(f, 1):
            line = line.strip()
            if not line or line.startswith("//") or line.startswith("#"):
                continue
            try:
                row = json.loads(line)
            except json.JSONDecodeError as e:
                raise ValueError(f"{path}:{lineno} invalid JSON: {e}") from e
            if "query" not in row or "intent" not in row:
                raise ValueError(f"{path}:{lineno} row missing required 'query'/'intent' field")
            rows.append(row)
    return rows


# ---------------------------------------------------------------------------
# Running the classifier
# ---------------------------------------------------------------------------

def full_scores(classifier, embed_query_fn, query: str) -> dict[str, float]:
    """Cosine similarity against every centroid, primary intents AND nav-depth
    labels alike. Mirrors classify_multi()'s internal computation but without
    the _PRIMARY_INTENTS filter, so we get full visibility for diagnostics."""
    import numpy as np
    q_vec = np.array(embed_query_fn(query, classifier._embedder))
    return {label: float(np.dot(q_vec, centroid)) for label, centroid in classifier._centroids.items()}


def evaluate(classifier, embed_query_fn, infer_nav_depth_fn, dataset: list[dict]) -> list[dict]:
    """Runs each row through the REAL classify()/classify_multi() methods
    (so eval results always reflect actual production behavior), and
    separately computes full_scores() purely for diagnostics."""
    results = []
    for row in dataset:
        query = row["query"]
        expected_intent = row["intent"]
        expected_depth = row.get("nav_depth")

        predicted_intent = classifier.classify(query)
        active_intents = classifier.classify_multi(query)
        # question must be passed through — _infer_nav_depth only runs its
        # lexical-anchor resolution when a question string is supplied,
        # otherwise it silently falls back to rank order only.
        predicted_depth = infer_nav_depth_fn(active_intents, query)

        scores = full_scores(classifier, embed_query_fn, query)
        ranked = sorted(scores.items(), key=lambda kv: -kv[1])
        top1_label, top1_score = ranked[0]
        top2_score = ranked[1][1] if len(ranked) > 1 else float("nan")

        correct = predicted_intent == expected_intent
        depth_correct = (predicted_depth == expected_depth) if expected_depth is not None else None

        results.append({
            "query": query,
            "expected_intent": expected_intent,
            "predicted_intent": predicted_intent,
            "correct": correct,
            "expected_nav_depth": expected_depth,
            "predicted_nav_depth": predicted_depth if expected_depth is not None else None,
            "depth_correct": depth_correct,
            "active_intents": active_intents,
            "margin": top1_score - top2_score,
            "top1_label": top1_label,
            "top1_score": top1_score,
            "scores": scores,
            "note": row.get("note", ""),
        })
    return results


# ---------------------------------------------------------------------------
# Metrics
# ---------------------------------------------------------------------------

def compute_metrics(results: list[dict]) -> dict[str, Any]:
    total = len(results)
    correct = sum(r["correct"] for r in results)
    accuracy = correct / total if total else 0.0

    labels = sorted({r["expected_intent"] for r in results} | {r["predicted_intent"] for r in results})
    confusion: dict[str, Counter] = defaultdict(Counter)
    for r in results:
        confusion[r["expected_intent"]][r["predicted_intent"]] += 1

    per_class = {}
    for label in labels:
        tp = confusion[label][label]
        fp = sum(confusion[other][label] for other in labels if other != label)
        fn = sum(confusion[label][other] for other in labels if other != label)
        support = sum(confusion[label].values())
        precision = tp / (tp + fp) if (tp + fp) else 0.0
        recall = tp / (tp + fn) if (tp + fn) else 0.0
        f1 = 2 * precision * recall / (precision + recall) if (precision + recall) else 0.0
        per_class[label] = {"precision": precision, "recall": recall, "f1": f1, "support": support}

    depth_rows = [r for r in results if r["expected_nav_depth"] is not None]
    depth_correct = sum(1 for r in depth_rows if r["depth_correct"])
    depth_accuracy = depth_correct / len(depth_rows) if depth_rows else None

    return {
        "accuracy": accuracy,
        "total": total,
        "correct": correct,
        "labels": labels,
        "confusion": confusion,
        "per_class": per_class,
        "depth_accuracy": depth_accuracy,
        "depth_total": len(depth_rows),
    }


def margin_stats(results: list[dict]) -> dict[str, dict | None]:
    def stats(xs: list[float]) -> dict | None:
        if not xs:
            return None
        return {"n": len(xs), "mean": mean(xs), "median": median(xs), "min": min(xs), "max": max(xs)}

    return {
        "correct": stats([r["margin"] for r in results if r["correct"]]),
        "incorrect": stats([r["margin"] for r in results if not r["correct"]]),
    }


# ---------------------------------------------------------------------------
# Reporting
# ---------------------------------------------------------------------------

def print_report(metrics: dict, margins: dict, results: list[dict]) -> None:
    print("=" * 78)
    print(f"INTENT CLASSIFIER EVAL — {metrics['correct']}/{metrics['total']} correct "
          f"({metrics['accuracy']:.1%})")
    print("=" * 78)

    print(f"\n{'label':<20}{'precision':>10}{'recall':>10}{'f1':>10}{'support':>10}")
    for label, m in sorted(metrics["per_class"].items()):
        print(f"{label:<20}{m['precision']:>10.2f}{m['recall']:>10.2f}{m['f1']:>10.2f}{m['support']:>10}")

    labels = metrics["labels"]
    print("\nconfusion matrix (rows=expected, cols=predicted)")
    print(" " * 22 + "".join(f"{l[:10]:>12}" for l in labels))
    for expected in labels:
        row = f"{expected:<22}" + "".join(
            f"{metrics['confusion'][expected][pred]:>12}" for pred in labels
        )
        print(row)

    if metrics["depth_accuracy"] is not None:
        print(f"\nnav_depth accuracy: {metrics['depth_accuracy']:.1%} "
              f"({metrics['depth_total']} labeled rows)")

    print("\nscore margin (top1 - top2), correct vs incorrect predictions")
    print("  -> if these overlap heavily, a similarity threshold alone won't")
    print("     reliably separate right from wrong answers")
    for bucket, s in margins.items():
        if s:
            print(f"  {bucket:<10} n={s['n']:<4} mean={s['mean']:.4f} "
                  f"median={s['median']:.4f} min={s['min']:.4f} max={s['max']:.4f}")
        else:
            print(f"  {bucket:<10} (no rows)")

    misses = [r for r in results if not r["correct"]]
    if misses:
        print(f"\nMISCLASSIFIED ({len(misses)}):")
        for r in sorted(misses, key=lambda r: -r["margin"]):
            print(f"  '{r['query']}'")
            print(f"      expected={r['expected_intent']}  predicted={r['predicted_intent']}  "
                  f"margin={r['margin']:.4f}")
            top3 = sorted(r["scores"].items(), key=lambda kv: -kv[1])[:3]
            print("      top-3: " + ", ".join(f"{l}={s:.3f}" for l, s in top3))
            if r["note"]:
                print(f"      note: {r['note']}")

    depth_misses = [r for r in results if r["depth_correct"] is False]
    if depth_misses:
        print(f"\nNAV_DEPTH MISCLASSIFIED ({len(depth_misses)}):")
        for r in depth_misses:
            print(f"  '{r['query']}'  expected={r['expected_nav_depth']}  "
                  f"predicted={r['predicted_nav_depth']}  active={r['active_intents']}")

    print("=" * 78)


def write_csv(results: list[dict], path: Path) -> None:
    fieldnames = [
        "query", "expected_intent", "predicted_intent", "correct",
        "expected_nav_depth", "predicted_nav_depth", "depth_correct",
        "margin", "top1_label", "top1_score", "active_intents", "note", "scores_json",
    ]
    with path.open("w", newline="", encoding="utf-8") as f:
        writer = csv.DictWriter(f, fieldnames=fieldnames)
        writer.writeheader()
        for r in results:
            writer.writerow({
                "query": r["query"],
                "expected_intent": r["expected_intent"],
                "predicted_intent": r["predicted_intent"],
                "correct": r["correct"],
                "expected_nav_depth": r["expected_nav_depth"] or "",
                "predicted_nav_depth": r["predicted_nav_depth"] or "",
                "depth_correct": r["depth_correct"] if r["depth_correct"] is not None else "",
                "margin": f"{r['margin']:.4f}",
                "top1_label": r["top1_label"],
                "top1_score": f"{r['top1_score']:.4f}",
                "active_intents": "|".join(r["active_intents"]),
                "note": r["note"],
                "scores_json": json.dumps(r["scores"], ensure_ascii=False),
            })


# ---------------------------------------------------------------------------
# Entry point
# ---------------------------------------------------------------------------

def main() -> None:
    parser = argparse.ArgumentParser(description="Evaluate IntentClassifier against a labeled dataset")
    parser.add_argument("--dataset", type=Path, default=Path(__file__).parent / "intent_eval_dataset.jsonl")
    parser.add_argument("--csv", type=Path, default=Path(__file__).parent / "results.csv",
                         help="path to write full per-query results as CSV")
    args = parser.parse_args()

    # Import here (not at module top) so this file can still be syntax-checked
    # / dry-run tested without the real app package / OpenVINO model on disk.
    from app.repositories.ingestion.embedder import embed_query
    from app.repositories.generation.intent_extractor import classifier as default_classifier, _infer_nav_depth

    dataset = load_dataset(args.dataset)
    print(f"loaded {len(dataset)} rows from {args.dataset}")

    results = evaluate(default_classifier, embed_query, _infer_nav_depth, dataset)
    metrics = compute_metrics(results)
    margins = margin_stats(results)

    print_report(metrics, margins, results)

    write_csv(results, args.csv)
    print(f"\nfull per-query results written to {args.csv}")


if __name__ == "__main__":
    main()