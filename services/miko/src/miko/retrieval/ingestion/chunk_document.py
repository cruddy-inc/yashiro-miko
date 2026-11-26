from miko.retrieval.ingestion.normalize import normalize_documents
from pathlib import Path
from langchain_text_splitters import (
    MarkdownHeaderTextSplitter,
    RecursiveCharacterTextSplitter,
)
from langchain_core.documents import Document
from miko.core.logging.logger import get_logger

okitsune = get_logger("ingestion.chunk_document")

headers_to_split_on = [
    ("#", "Header 1"),
    ("##", "Header 2"),
    ("###", "Header 3"),
]
chunk_size = 500
chunk_overlap = 60

pivotal_metadata_keys = {"id", "typology", "signals", "severity_range", "path"}
integral_metadata_keys = {
    "title",
    "sub_typology",
    "rails",
    "tools",
    "interventions",
}


def chunk_documents(
    root: str | Path,
    ignored_files: set[str] | None = None,
    excluded_directories: set[str] | None = None,
) -> list[Document]:

    norm_doc_list = normalize_documents(
        root,
        ignored_files,
        excluded_directories,
    )

    markdown_splitter = MarkdownHeaderTextSplitter(
        headers_to_split_on=headers_to_split_on,
        strip_headers=False,
    )

    text_splitter = RecursiveCharacterTextSplitter(
        chunk_size=chunk_size,
        chunk_overlap=chunk_overlap,
        separators=["\n\n", "\n", " ", ""],
    )

    all_chunks: list[Document] = []

    for norm_doc in norm_doc_list:
        body_text = norm_doc["content"]
        file_metadata = norm_doc["metadata"]
        file_metadata_keys = file_metadata.keys()

        for pivotal in pivotal_metadata_keys:
            if pivotal in file_metadata_keys:
                continue
            okitsune.warning(
                "Missing pivotal metadata in the document: | \nkeys=[%s] \nmissing=[%s]",
                file_metadata_keys,
                pivotal,
            )

        for integral in integral_metadata_keys:
            if integral in file_metadata_keys:
                continue
            okitsune.warning(
                "Document missing integral metadata | \nkeys=[%s] | \nmissing=[%s]",
                file_metadata_keys,
                integral,
            )

        md_header_splits = markdown_splitter.split_text(body_text)

        for header_split in md_header_splits:
            header_split.metadata.update(file_metadata)

        docs_chunks = text_splitter.split_documents(md_header_splits)

        for chunk in docs_chunks:
            if chunk.page_content.strip():
                all_chunks.append(chunk)


    return all_chunks