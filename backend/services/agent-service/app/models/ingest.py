from __future__ import annotations

from dataclasses import dataclass, field


@dataclass
class IngestResult:
    docs_found:      int = 0
    docs_parsed:     int = 0
    docs_failed:     int = 0
    docs_duplicate:  int = 0
    chunks_total:    int = 0
    chunks_upserted: int = 0
    parse_errors:    list[str] = field(default_factory=list)
    elapsed_s:       float = 0.0

    @property
    def success(self) -> bool:
        return self.docs_failed == 0 and self.chunks_upserted > 0

    def log_summary(self, log) -> None:
        log.info(
            "ingest summary | docs: %d parsed / %d failed | "
            "chunks: %d total / %d upserted | elapsed: %.1fs",
            self.docs_parsed, self.docs_failed,
            self.chunks_total, self.chunks_upserted,
            self.elapsed_s,
        )
        if self.parse_errors:
            log.warning("%d file(s) had errors:", len(self.parse_errors))
            for err in self.parse_errors:
                log.warning("  %s", err)