import os
from pathlib import Path
from miko.core.logging.errors import (
    PathIsNotADirectoryError,
    ResourceFileNotFoundError,
)

def discover_documents(
    root: str | Path,
    ignored_files: set[str] | None = None,
    excluded_directories: set[str] | None = None,
) -> list[Path]:

    root_path = Path.home() / root
    kb_path = root_path / "data" / "kb"
    ignored_f = ignored_files or set()
    excluded_dir = excluded_directories or set()

    if not root_path.is_dir():
        if root_path.exists():
            raise PathIsNotADirectoryError(
                f"Expected a directory: {root_path}"
            )
        raise ResourceFileNotFoundError(
            f"Project directory not found: {root_path}"
        )

    if not kb_path.is_dir():
        if kb_path.exists():
            raise PathIsNotADirectoryError(
                f"Expected a directory: {kb_path}"
            )
        raise ResourceFileNotFoundError(
            f"Knowledge base directory not found: {kb_path}"
        )

    discovered_paths: list[Path] = []

    def handle_walk_error(error: OSError) -> None:
        raise ResourceFileNotFoundError(
            f"Failed to access directory during walk: {error.filename}",
            cause=error,
        )

    for current_root, dirs, files in os.walk(
        kb_path,
        topdown=True,
        onerror=handle_walk_error,
    ):
        dirs[:] = sorted(
            directory
            for directory in dirs
            if directory not in excluded_dir
        )

        valid_files = sorted(
            filename
            for filename in files
            if filename not in ignored_f
            and filename.lower().endswith(".md")
        )

        for filename in valid_files:
            discovered_paths.append(Path(current_root) / filename)

    return sorted(discovered_paths)