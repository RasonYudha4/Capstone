"""
config.py — central settings with environment variable binding.

Priority order (highest to lowest):
  1. Actual environment variables (set in shell or Docker)
  2. Values in .env file
  3. Default values defined here
"""
from __future__ import annotations

from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):

    model_config = SettingsConfigDict(
        env_file=".env",
        env_file_encoding="utf-8",
        case_sensitive=False,
        extra="ignore",
    )

    # ── Embedding model (OpenVINO) ────────────────────────────────────────
    # Local path to pre-exported OpenVINO IR directory for the embed model.
    # Export with:
    #   optimum-cli export openvino `
    #     --model Qwen/Qwen3-0.6B `
    #     --task feature-extraction `
    #     --weight-format int8 `
    #     ./models/qwen3-0.6b-ov-int8
    embed_model_path:  str = "./models/qwen3-0.6b-ov-int8"
    embed_device:      str = "AUTO"
    embed_batch_size:  int = 128

    # ── Chat / generation model (OpenVINO) ────────────────────────────────
    # Local path to pre-exported + INT4-quantized OV IR directory.
    # One-time export command:
    #   optimum-cli export openvino 
    #       --model LiquidAI/LFM2.5-1.2B-Instruct \
    #       --task text-generation-with-past \
    #       --weight-format int4 \
    #       --group-size 128 \
    #       --ratio 1.0 \
    #       ./models/lfm2.5-1.2b-ov-int4
    #
    # --task text-generation-with-past  → enables KV cache reuse between tokens
    # --weight-format int4              → ~650MB, fastest decode
    # --group-size 128 --ratio 1.0      → quality-preserving INT4 (≈ Q4_K_M)
    chat_model_path:   str = "./models/qwen3-4b-instruct-2507-ov-int4"

    # OpenVINO device: "CPU" | "GPU" | "NPU" | "AUTO"
    chat_device:       str = "AUTO"

    # Generation parameters
    chat_max_new_tokens:      int   = 512
    chat_min_new_tokens:      int   = 32
    chat_temperature:         float = 0.7
    chat_top_k:               int   = 20
    chat_top_p:               float = 0.8
    chat_repetition_penalty:  float = 1.05

    # ── Chunking ─────────────────────────────────────────────────────────
    chunk_size:    int = 512
    chunk_overlap: int = 64
    top_k:         int = 5

    # ── Vector store ─────────────────────────────────────────────────────
    vector_path: str = "./data/vector_db"
    collection:  str = "rag_docs"


settings = Settings()