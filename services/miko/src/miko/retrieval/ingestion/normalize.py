from miko.retrieval.ingestion.load_documents import load_documents
from pathlib import Path
from unicodedata import normalize
import re
from typing import Any
import frontmatter

def normalize_documents(
    root: str | Path,
    ignored_files: set[str] | None = None,
    excluded_directories: set[str] | None = None,
) -> list[dict[str, Any]]:

    documents = load_documents(
        root=root,
        ignored_files=ignored_files,
        excluded_directories=excluded_directories
    )

    normalize_docs: list[dict[str, Any]] = []

    for doc in documents:
        raw_content = doc["content"]

        parsed = frontmatter.loads(raw_content)

        meta: dict[str, Any] = {
            "name": doc.get("name", ""),
            **parsed.metadata,
            "path": str(doc.get("path", "")),
        }

        cleaned_body = re.sub(
            r"\n{3,}",
            "\n\n",
            parsed.content.lstrip("\ufeff").replace("\r\n", "\n").replace("\r", "\n")
            .replace("\u00a0", " ").replace("\u200b", ""),
        ).strip()
        cleaned_body = normalize("NFC", cleaned_body)

        normalize_docs.append(
            {
            "metadata": meta,
            "content": cleaned_body
            }
        )

    return normalize_docs
