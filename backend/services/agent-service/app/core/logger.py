import logging 
import time 
from contextlib import contextmanager

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s | %(levelname)s | %(name)s | %(message)s",
    datefmt="%H:%M:%S"
)

def get_logger(name: str) -> logging.Logger:
    return logging.getLogger(name)

@contextmanager
def timer(logger: logging.Logger, label: str):
    start = time.perf_counter()
    logger.info(f"{label} - started")
    try: 
        yield
    finally: 
        elapsed = (time.perf_counter() - start) * 1000
        logger.info(f"{label} - done in {elapsed:.1f}ms")