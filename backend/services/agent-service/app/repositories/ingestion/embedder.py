"""
embedder.py — Stage 3: embed Chunks via OpenVINO native API (Qwen3-0.6b).

Dependencies: openvino, transformers (tokenizer only), numpy
No torch required — mean pooling and L2 norm are pure numpy.
"""
from __future__ import annotations

import time
from dataclasses import dataclass, field
from pathlib import Path

import numpy as np
import openvino as ov
from transformers import AutoTokenizer

from app.models import Chunk
from app.core.config import settings
from app.core.logger import get_logger

log = get_logger("embedder")

# ---------------------------------------------------------------------------
# Model wrapper
# ---------------------------------------------------------------------------

@dataclass
class EmbedderModel:
    model_name_or_path: str
    device: str = "GPU"

    _tokenizer:      AutoTokenizer   = field(init=False, repr=False)
    _compiled_model: ov.CompiledModel = field(init=False, repr=False)
    _infer_req:      ov.InferRequest  = field(init=False, repr=False)

    def __post_init__(self) -> None:
        log.info("loading OVModel from %s on device=%s",
                 self.model_name_or_path, self.device)
        t0 = time.perf_counter()

        self._tokenizer = AutoTokenizer.from_pretrained(
            self.model_name_or_path,
            trust_remote_code=True,
            fix_mistral_regex=True,
        )
        core = ov.Core()
        model_xml = Path(self.model_name_or_path) / "openvino_model.xml"
        self._compiled_model = core.compile_model(
            core.read_model(str(model_xml)),
            self.device,
            config={
                "PERFORMANCE_HINT":         "THROUGHPUT",
                "INFERENCE_PRECISION_HINT": "f16",
            },
        )
        self._infer_req = self._compiled_model.create_infer_request()
        log.info("model loaded in %.1fs", time.perf_counter() - t0)

    def infer(self, texts: list[str]) -> list[list[float]]:
        return self._infer_impl(texts)

    def _infer_impl(self, texts: list[str]) -> list[list[float]]:
        encoded = self._tokenizer(
            texts,
            padding=True,          # dynamic padding — pad to longest in batch
            truncation=True,
            max_length=512,
            return_tensors="np",
        )

        inputs = {
            "input_ids":      encoded["input_ids"],
            "attention_mask": encoded["attention_mask"],
        }
        if "token_type_ids" in encoded:
            inputs["token_type_ids"] = encoded["token_type_ids"]

        self._infer_req.infer(inputs)

        # last_hidden_state: (batch, seq_len, hidden_dim)
        hidden = self._infer_req.get_output_tensor(0).data.copy()

        return _mean_pool_and_normalize(hidden, encoded["attention_mask"]).tolist()

    @classmethod
    def from_pretrained(
        cls,
        model_name_or_path: str | Path,
        device: str = "GPU",
    ) -> "EmbedderModel":
        return cls(model_name_or_path=str(model_name_or_path), device=device)
    
embedder = EmbedderModel.from_pretrained(
    settings.embed_model_path,
    device=settings.embed_device,
)


# ---------------------------------------------------------------------------
# Public API
# ---------------------------------------------------------------------------

def embed_chunks(
    chunks: list[Chunk],
    embedder: EmbedderModel,
    batch_size: int = 64,
) -> list[Chunk]:
    if not chunks:
        return chunks

    log.info(
        "embedding %d chunk(s) with model=%s batch_size=%d device=%s",
        len(chunks), embedder.model_name_or_path, batch_size, embedder.device,
    )
    t0 = time.perf_counter()

    for batch_start in range(0, len(chunks), batch_size):
        batch   = chunks[batch_start : batch_start + batch_size]
        vectors = embedder.infer([c.text for c in batch])
        for chunk, vec in zip(batch, vectors):
            chunk.vector = vec

    dims = {len(c.vector) for c in chunks if c.vector}
    if len(dims) > 1:
        raise EmbeddingError(
            f"inconsistent embedding dimensions across batch: {dims}"
        )

    failed = [c for c in chunks if not c.vector]
    if failed:
        raise EmbeddingError(
            f"{len(failed)} chunk(s) failed to embed: "
            + ", ".join(str(c.chunk_index) for c in failed[:5])
        )

    elapsed = time.perf_counter() - t0
    dim     = len(chunks[0].vector)
    log.info(
        "embedded %d chunk(s) in %.1fs (%.0f chunks/s) dim=%d",
        len(chunks), elapsed, len(chunks) / elapsed, dim,
    )
    return chunks


def embed_query(text: str, embedder: EmbedderModel) -> list[float]:
    """Embed a single query string"""
    return embedder.infer([text])[0]


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

def _mean_pool_and_normalize(
    hidden:  np.ndarray,   # (batch, seq, dim)  float32
    mask:    np.ndarray,   # (batch, seq)        int64
) -> np.ndarray:           # (batch, dim)        float32
    """
    Mean-pool token embeddings weighted by attention mask, then L2-normalise.
    Pure numpy — no torch dependency.
    """
    mask_f = mask[..., np.newaxis].astype(np.float32)     # (batch, seq, 1)
    summed = (hidden * mask_f).sum(axis=1)                 # (batch, dim)
    counts = mask_f.sum(axis=1).clip(min=1e-9)             # (batch, 1)
    embeddings = summed / counts                           # (batch, dim)

    norms = np.linalg.norm(embeddings, axis=1, keepdims=True).clip(min=1e-9)
    return (embeddings / norms).astype(np.float32)


# ---------------------------------------------------------------------------
# Exceptions
# ---------------------------------------------------------------------------

class EmbeddingError(RuntimeError):
    pass
