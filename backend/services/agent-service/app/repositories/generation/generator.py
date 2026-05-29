"""
generation/generator.py — OpenVINO-based text generation using qwen3-4b-instruct-2507-ov-int4.

"""
from __future__ import annotations

import time
import threading
from functools import lru_cache
from typing import Generator

from optimum.intel import OVModelForCausalLM
from transformers import AutoTokenizer, TextIteratorStreamer

from app.core.config import settings
from app.core.logger import get_logger, timer

log = get_logger("generator")


# ---------------------------------------------------------------------------
# Model singleton — loaded once, reused forever
# ---------------------------------------------------------------------------

@lru_cache(maxsize=1)
def _get_pipeline() -> tuple[OVModelForCausalLM, AutoTokenizer]:
    """
    Load the OV model and tokenizer once. lru_cache guarantees a single load
    even if called concurrently (Python's GIL protects the first call).
    """
    log.info(
        "loading OV chat model from %s on device=%s",
        settings.chat_model_path,
        settings.chat_device,
    )
    t0 = time.perf_counter()

    tokenizer = AutoTokenizer.from_pretrained(
        settings.chat_model_path,
        trust_remote_code=True,
    )

    model = OVModelForCausalLM.from_pretrained(
        settings.chat_model_path,
        device=settings.chat_device,
        ov_config={
            # Compress KV-cache to uint8 — cuts memory pressure during long
            # RAG prompts with large retrieved contexts, ~10-15% faster decode.
            "KV_CACHE_PRECISION": "u8",
            # Let OV scheduler saturate all available compute cores.
            "PERFORMANCE_HINT": "LATENCY",
        },
        trust_remote_code=True,
    )

    log.info("chat model loaded in %.1fs", time.perf_counter() - t0)
    return model, tokenizer


# ---------------------------------------------------------------------------
# Public API
# ---------------------------------------------------------------------------

def generate(prompt: str) -> str:
    """
    Single-shot generation. Returns the full response string.

    `prompt` is the raw RAG/intent prompt string built by prompt_builder.py.
    We wrap it in the ChatML template here so prompt_builder stays format-agnostic.
    """
    model, tokenizer = _get_pipeline()

    input_ids = _apply_template(tokenizer, prompt)

    log.info(
        "generating (device=%s, input_tokens=%d, prompt_chars=%d)",
        settings.chat_device, input_ids.shape[-1], len(prompt),
    )

    with timer(log, "ov generate"):
        output_ids = model.generate(
            input_ids,
            min_new_tokens=settings.chat_min_new_tokens,
            max_new_tokens=settings.chat_max_new_tokens,
            do_sample=True,
            temperature=settings.chat_temperature,
            top_k=settings.chat_top_k,
            top_p=settings.chat_top_p,
            repetition_penalty=settings.chat_repetition_penalty,
        )

    # Slice off the input tokens — output_ids contains prompt + response
    new_tokens = output_ids[0, input_ids.shape[-1]:]
    response = tokenizer.decode(new_tokens, skip_special_tokens=True).strip()

    log.info("response_len=%d chars, new_tokens=%d", len(response), len(new_tokens))
    return response


def generate_stream(prompt: str) -> Generator[str, None, None]:
    """
    Streaming generation — yields decoded text chunks as they are produced.

    OVModelForCausalLM.generate() is synchronous, so we run it in a background
    thread and surface tokens via TextIteratorStreamer.
    """
    model, tokenizer = _get_pipeline()

    input_ids = _apply_template(tokenizer, prompt)

    streamer = TextIteratorStreamer(
        tokenizer,
        skip_prompt=True,           # don't re-yield the input tokens
        skip_special_tokens=True,
    )

    generate_kwargs = dict(
        input_ids=input_ids,
        streamer=streamer,
        max_new_tokens=settings.chat_max_new_tokens,
        do_sample=True,
        temperature=settings.chat_temperature,
        top_k=settings.chat_top_k,
        repetition_penalty=settings.chat_repetition_penalty,
    )

    log.info(
        "streaming (device=%s, input_tokens=%d)",
        settings.chat_device, input_ids.shape[-1],
    )
    t_start = time.perf_counter()
    token_count = 0

    # Run generate() in a thread so we can yield from the streamer in this one
    thread = threading.Thread(target=model.generate, kwargs=generate_kwargs)
    thread.start()

    for chunk in streamer:
        if chunk:
            token_count += 1
            yield chunk

    thread.join()

    elapsed = (time.perf_counter() - t_start) * 1000
    log.info(
        "stream done — %d tokens in %.1fms (%.1f tok/s)",
        token_count, elapsed, token_count / (elapsed / 1000),
    )


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

def _apply_template(tokenizer: AutoTokenizer, prompt: str):
    """
    Wrap the raw prompt string in LFM2.5's ChatML template and tokenize.

    LFM2.5 template (from HF model card):
        <|startoftext|><|im_start|>system
        {system}<|im_end|>
        <|im_start|>user
        {user}<|im_end|>
        <|im_start|>assistant

    We pass the prompt as the user turn. The system turn establishes the
    hospital accreditation assistant persona, consistent with prompt_builder.py.
    """
    messages = [
        {
            "role": "system",
            "content": "Kamu adalah asisten sistem akreditasi rumah sakit yang membantu menjawab pertanyaan berdasarkan dokumen standar dan bukti yang tersedia.",
        },
        {
            "role": "user",
            "content": prompt,
        },
    ]

    # add_generation_prompt=True appends the <|im_start|>assistant token
    # so the model knows to start generating, not continue the user turn.
    input_ids = tokenizer.apply_chat_template(
        messages,
        add_generation_prompt=True,
        return_tensors="pt",
        tokenize=True,
    )

    return input_ids