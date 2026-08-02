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
    embed_model_path:  str = "./data/models/qwen3-0.6b-ov-int8"
    # Intel GPU: "GPU" | AMD / no Intel GPU: set to "CPU"
    embed_device:      str = "GPU"
    embed_batch_size:  int = 128

    # ── Chat / generation model (OpenVINO) ────────────────────────────────
    # Local path to pre-exported + INT4-quantized OV IR directory.

    chat_model_path:   str = "./data/models/qwen3-4b-instruct-2507-ov-int4"

    # OpenVINO device: "CPU" | "GPU" | "NPU" | "AUTO"
    # Intel keep "GPU"; on AMD change this to "CPU".
    chat_device:       str = "GPU"

    # Generation parameters
    chat_max_new_tokens:      int   = 512
    chat_min_new_tokens:      int   = 32
    chat_temperature:         float = 0.7
    chat_top_k:               int   = 5
    chat_top_p:               float = 0.8
    chat_repetition_penalty:  float = 1.05

    # ── Chunking ─────────────────────────────────────────────────────────
    chunk_size:    int = 512
    chunk_overlap: int = 64
    top_k:         int = 5

    # ── Vector store ─────────────────────────────────────────────────────
    vector_path: str = "./data/vector_db"
    collection:  str = "rag_docs"

    rewrite_cycles_to_keep: int = 3

    @property
    def rewrite_turns_to_keep(self) -> int:
        return self.rewrite_cycles_to_keep * 2


settings = Settings()