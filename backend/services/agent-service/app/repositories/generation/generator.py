"""
generation/generator.py — OpenVINO-based text generation using qwen3-4b-instruct-2507-ov-int4.

Uses OVModelForCausalLM (optimum.intel) — necessary because INT4 decoder exports
carry stateful KV-cache nodes that require Optimum's generate() loop to manage
beam/batch state correctly.
"""
from __future__ import annotations

import time
import threading
from dataclasses import dataclass, field
from pathlib import Path
from typing import Generator

from optimum.intel import OVModelForCausalLM
from transformers import AutoTokenizer, TextIteratorStreamer

from app.core.config import settings
from app.core.logger import get_logger, timer

log = get_logger("generator")


# ---------------------------------------------------------------------------
# Model wrapper
# ---------------------------------------------------------------------------

@dataclass
class GeneratorModel:
    model_name_or_path: str
    device: str = "GPU"

    _tokenizer: AutoTokenizer      = field(init=False, repr=False)
    _model:     OVModelForCausalLM = field(init=False, repr=False)

    def __post_init__(self) -> None:
        log.info(
            "loading OV generator model from %s on device=%s",
            self.model_name_or_path, self.device,
        )
        t0 = time.perf_counter()

        self._tokenizer = AutoTokenizer.from_pretrained(
            self.model_name_or_path,
            trust_remote_code=True,
            fix_mistral_regex=True
        )
        self._model = OVModelForCausalLM.from_pretrained(
            self.model_name_or_path,
            device=self.device,
            ov_config={
                "KV_CACHE_PRECISION": "u8",
                "PERFORMANCE_HINT":   "LATENCY",
                "NUM_STREAMS":           "1",
            },
            trust_remote_code=True,
        )

        log.info("generator model loaded in %.1fs", time.perf_counter() - t0)

    @classmethod
    def from_pretrained(
        cls,
        model_name_or_path: str | Path,
        device: str = "GPU",
    ) -> "GeneratorModel":
        return cls(model_name_or_path=str(model_name_or_path), device=device)
    
generator = GeneratorModel.from_pretrained(
    settings.chat_model_path,
    device=settings.chat_device,
)


# ---------------------------------------------------------------------------
# Public API
# ---------------------------------------------------------------------------

def generate(prompt: str, gen: GeneratorModel) -> str:
    input_ids, attention_mask = _apply_template(gen._tokenizer, prompt)

    log.info(
        "generating (device=%s, input_tokens=%d, prompt_chars=%d)",
        gen.device, input_ids.shape[-1], len(prompt),
    )

    with timer(log, "ov generate"):
        output_ids = gen._model.generate(
            input_ids,
            attention_mask=attention_mask,
            min_new_tokens=settings.chat_min_new_tokens,
            max_new_tokens=settings.chat_max_new_tokens,
            do_sample=True,
            temperature=settings.chat_temperature,
            top_k=settings.chat_top_k,
            top_p=settings.chat_top_p,
            repetition_penalty=settings.chat_repetition_penalty,
        )

    new_tokens = output_ids[0, input_ids.shape[-1]:]
    response = gen._tokenizer.decode(new_tokens, skip_special_tokens=True).strip()

    log.info("response_len=%d chars, new_tokens=%d", len(response), len(new_tokens))
    return response


def generate_stream(prompt: str, gen: GeneratorModel) -> Generator[str, None, None]:
    input_ids, attention_mask = _apply_template(gen._tokenizer, prompt)
    input_token_count = input_ids.shape[-1]

    streamer = TextIteratorStreamer(
        gen._tokenizer,
        skip_prompt=True,
        skip_special_tokens=True,
    )

    log.info(
        "streaming (device=%s, input_tokens=%d)",
        gen.device, input_token_count,
    )

    t_start = time.perf_counter()
    first_token_time: float | None = None
    token_count = 0

    thread = threading.Thread(
        target=gen._model.generate,
        kwargs=dict(
            input_ids=input_ids,
            attention_mask=attention_mask,
            streamer=streamer,
            max_new_tokens=settings.chat_max_new_tokens,
            do_sample=True,
            temperature=settings.chat_temperature,
            top_k=settings.chat_top_k,
            repetition_penalty=settings.chat_repetition_penalty,
        ),
    )
    thread.start()

    for chunk in streamer:
        if chunk:
            if first_token_time is None:
                first_token_time = time.perf_counter()
                ttft_ms = (first_token_time - t_start) * 1000
                log.info("TTFT=%.1fms (prefill %d tokens)", ttft_ms, input_token_count)
            token_count += 1
            yield chunk

    thread.join()

    t_end = time.perf_counter()
    total_ms = (t_end - t_start) * 1000
    generation_ms = (t_end - first_token_time) * 1000 if first_token_time else 0
    tps = token_count / max(generation_ms / 1000, 1e-9)

    log.info(
        "stream done — input=%d chunks | output=%d chunks | "
        "TTFT=%.1fms | gen=%.1fms | TPS=%.1f | total=%.1fms",
        input_token_count, token_count,
        ttft_ms, generation_ms, tps, total_ms,
    )


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

def _apply_template(tokenizer: AutoTokenizer, prompt: str) -> tuple:
    messages = [
        {
            "role": "system",
            "content": (
                "Kamu adalah asisten sistem akreditasi rumah sakit yang membantu "
                "menjawab pertanyaan berdasarkan dokumen standar dan bukti yang tersedia."
            ),
        },
        {"role": "user", "content": prompt},
    ]

    encoded = tokenizer.apply_chat_template(
        messages,
        add_generation_prompt=True,
        return_tensors="pt",
        tokenize=True,
        return_dict=True,
    )
    return encoded["input_ids"], encoded["attention_mask"]
