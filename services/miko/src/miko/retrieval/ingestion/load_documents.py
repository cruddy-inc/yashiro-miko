from miko.retrieval.ingestion.discovery import discover_documents
from pathlib import Path
from miko.core.logging.logger import get_logger

okitsune = get_logger("ingest.load_document")

def load_documents(
    root: str | Path,
    ignored_files: set[str] | None = None,
    excluded_directories: set[str] | None = None,
) -> list[dict[str, str]]:
    root_path = Path.home() / root
    files = discover_documents(
        root=root,
        ignored_files=ignored_files or set(),
        excluded_directories=excluded_directories or set(),
    )

    documents: list[dict[str, str]] = []
    for filename in files:
        doc_file = Path(filename)
        relative_path = str(doc_file.relative_to(root_path))

        try:
            documents.append({
                "path": relative_path,
                "name": doc_file.name,
                "content": doc_file.read_text(encoding="utf-8"),
            })
        except (OSError, UnicodeDecodeError) as e:
            okitsune.warning(
                f"Skipping unreadable document: {relative_path}",
                exc_info=e,
                extra={"file": relative_path},
            )

    return documents
