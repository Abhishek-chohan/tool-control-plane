#!/usr/bin/env python3
"""
Description: Semantic search for relevant code or documentation using text similarity.

This tool performs semantic search across code and documentation files using
text similarity algorithms to find relevant content based on natural language queries.

Parameters:
  query (string, required): The search query in natural language.
  max_results (integer, optional): Maximum number of results to return (default: 10).
  file_types (array, optional): File types to search in (default: common code files).
  similarity_threshold (float, optional): Minimum similarity score to include (default: 0.1).
  context_size (integer, optional): Number of lines of context around matches (default: 3).
  search_path (string, optional): Directory to search in (default: current directory).
"""

import argparse
import math
import os
import re
import sys
from collections import Counter, defaultdict
from pathlib import Path
from typing import Any, Dict, List, Set


def semantic_search(
    query: str,
    max_results: int = 10,
    file_types: List[str] = None,
    similarity_threshold: float = 0.1,
    context_size: int = 3,
    search_path: str = None,
) -> List[Dict[str, Any]]:
    """
    Perform semantic search for relevant code or documentation.

    Args:
        query: The search query in natural language
        max_results: Maximum number of results to return
        file_types: File types to search in
        similarity_threshold: Minimum similarity score to include
        context_size: Number of lines of context around matches
        search_path: Directory to search in

    Returns:
        List of search results with similarity scores
    """
    if search_path is None:
        search_path = os.getcwd()

    if file_types is None:
        file_types = [
            ".py",
            ".js",
            ".ts",
            ".jsx",
            ".tsx",
            ".java",
            ".cpp",
            ".c",
            ".h",
            ".cs",
            ".php",
            ".rb",
            ".go",
            ".rs",
            ".swift",
            ".kt",
            ".scala",
            ".html",
            ".css",
            ".scss",
            ".md",
            ".rst",
            ".txt",
            ".json",
            ".yaml",
            ".yml",
            ".xml",
            ".sql",
            ".sh",
            ".bash",
            ".ps1",
            ".bat",
        ]

    # Get all files to search
    files_to_search = get_searchable_files(search_path, file_types)

    # Process query
    query_terms = preprocess_text(query)

    # Search each file
    results = []

    for file_path in files_to_search:
        try:
            file_results = search_file(file_path, query_terms, context_size)

            # Calculate similarity scores
            for result in file_results:
                score = calculate_similarity(query_terms, result["content_tokens"])

                if score >= similarity_threshold:
                    result["similarity_score"] = score
                    result["query"] = query
                    results.append(result)

        except Exception as e:
            # Skip files that can't be processed
            continue

    # Sort by similarity score
    results.sort(key=lambda x: x["similarity_score"], reverse=True)

    # Return top results
    return results[:max_results]


def get_searchable_files(search_path: str, file_types: List[str]) -> List[str]:
    """Get all files that match the specified file types."""
    files = []

    for root, dirs, filenames in os.walk(search_path):
        # Skip hidden directories
        dirs[:] = [d for d in dirs if not d.startswith(".")]

        for filename in filenames:
            if filename.startswith("."):
                continue

            file_path = Path(root) / filename

            # Check if file type is in our list
            if any(filename.lower().endswith(ext) for ext in file_types):
                files.append(str(file_path.resolve()))

    return files


def search_file(
    file_path: str, query_terms: Set[str], context_size: int
) -> List[Dict[str, Any]]:
    """Search for relevant content in a single file."""
    try:
        with open(file_path, "r", encoding="utf-8", errors="ignore") as f:
            lines = f.readlines()
    except Exception:
        return []

    results = []

    # Search for matches in each line
    for line_num, line in enumerate(lines):
        line_tokens = preprocess_text(line.strip())

        # Check if line contains any query terms
        if query_terms.intersection(line_tokens):
            # Get context around the match
            start_line = max(0, line_num - context_size)
            end_line = min(len(lines), line_num + context_size + 1)

            context_lines = []
            for i in range(start_line, end_line):
                context_lines.append(
                    {
                        "line_number": i + 1,
                        "content": lines[i].rstrip("\n\r"),
                        "is_match": i == line_num,
                    }
                )

            # Get all tokens from the context
            context_text = " ".join(cl["content"] for cl in context_lines)
            content_tokens = preprocess_text(context_text)

            result = {
                "file_path": file_path,
                "line_number": line_num + 1,
                "matched_line": line.strip(),
                "context": context_lines,
                "content_tokens": content_tokens,
                "file_type": Path(file_path).suffix.lower(),
            }

            results.append(result)

    # Also search in function/class definitions and comments
    function_results = extract_functions_and_classes(
        file_path, lines, query_terms, context_size
    )
    results.extend(function_results)

    # Remove duplicates based on line number
    seen_lines = set()
    unique_results = []
    for result in results:
        if result["line_number"] not in seen_lines:
            seen_lines.add(result["line_number"])
            unique_results.append(result)

    return unique_results


def extract_functions_and_classes(
    file_path: str, lines: List[str], query_terms: Set[str], context_size: int
) -> List[Dict[str, Any]]:
    """Extract functions and classes that might be relevant."""
    results = []
    file_ext = Path(file_path).suffix.lower()

    # Define patterns for different file types
    patterns = {
        ".py": [r"^\s*def\s+(\w+)", r"^\s*class\s+(\w+)", r"^\s*async\s+def\s+(\w+)"],
        ".js": [
            r"^\s*function\s+(\w+)",
            r"^\s*const\s+(\w+)\s*=\s*\(",
            r"^\s*let\s+(\w+)\s*=\s*\(",
            r"^\s*var\s+(\w+)\s*=\s*\(",
            r"^\s*(\w+)\s*:\s*function",
            r"^\s*class\s+(\w+)",
        ],
        ".ts": [
            r"^\s*function\s+(\w+)",
            r"^\s*const\s+(\w+)\s*=\s*\(",
            r"^\s*let\s+(\w+)\s*=\s*\(",
            r"^\s*class\s+(\w+)",
            r"^\s*interface\s+(\w+)",
            r"^\s*type\s+(\w+)",
        ],
        ".java": [
            r"^\s*public\s+.*\s+(\w+)\s*\(",
            r"^\s*private\s+.*\s+(\w+)\s*\(",
            r"^\s*protected\s+.*\s+(\w+)\s*\(",
            r"^\s*class\s+(\w+)",
            r"^\s*interface\s+(\w+)",
        ],
        ".cpp": [r"^\s*.*\s+(\w+)\s*\(", r"^\s*class\s+(\w+)", r"^\s*struct\s+(\w+)"],
        ".c": [r"^\s*.*\s+(\w+)\s*\(", r"^\s*struct\s+(\w+)"],
    }

    if file_ext not in patterns:
        return results

    for line_num, line in enumerate(lines):
        for pattern in patterns[file_ext]:
            match = re.search(pattern, line)
            if match:
                function_name = match.group(1)

                # Check if function name or line content matches query
                name_tokens = preprocess_text(function_name)
                line_tokens = preprocess_text(line.strip())

                if query_terms.intersection(name_tokens) or query_terms.intersection(
                    line_tokens
                ):
                    # Get context around the function
                    start_line = max(0, line_num - context_size)
                    end_line = min(len(lines), line_num + context_size + 1)

                    context_lines = []
                    for i in range(start_line, end_line):
                        context_lines.append(
                            {
                                "line_number": i + 1,
                                "content": lines[i].rstrip("\n\r"),
                                "is_match": i == line_num,
                            }
                        )

                    context_text = " ".join(cl["content"] for cl in context_lines)
                    content_tokens = preprocess_text(context_text)

                    result = {
                        "file_path": file_path,
                        "line_number": line_num + 1,
                        "matched_line": line.strip(),
                        "context": context_lines,
                        "content_tokens": content_tokens,
                        "file_type": file_ext,
                        "match_type": "function_definition",
                        "function_name": function_name,
                    }

                    results.append(result)

    return results


def preprocess_text(text: str) -> Set[str]:
    """Preprocess text for similarity comparison."""
    # Convert to lowercase
    text = text.lower()

    # Remove special characters and split into words
    words = re.findall(r"\b\w+\b", text)

    # Remove common stop words
    stop_words = {
        "the",
        "a",
        "an",
        "and",
        "or",
        "but",
        "in",
        "on",
        "at",
        "to",
        "for",
        "of",
        "with",
        "by",
        "from",
        "is",
        "are",
        "was",
        "were",
        "be",
        "been",
        "have",
        "has",
        "had",
        "do",
        "does",
        "did",
        "will",
        "would",
        "could",
        "should",
        "may",
        "might",
        "can",
        "this",
        "that",
        "these",
        "those",
        "i",
        "you",
        "he",
        "she",
        "it",
        "we",
        "they",
        "me",
        "him",
        "her",
        "us",
        "them",
        "my",
        "your",
        "his",
        "her",
        "its",
        "our",
        "their",
    }

    # Filter out stop words and short words
    filtered_words = [
        word for word in words if word not in stop_words and len(word) > 2
    ]

    return set(filtered_words)


def calculate_similarity(query_terms: Set[str], content_tokens: Set[str]) -> float:
    """Calculate similarity between query terms and content tokens."""
    if not query_terms or not content_tokens:
        return 0.0

    # Jaccard similarity: intersection / union
    intersection = len(query_terms.intersection(content_tokens))
    union = len(query_terms.union(content_tokens))

    if union == 0:
        return 0.0

    jaccard_score = intersection / union

    # Boost score if many query terms are present
    query_coverage = intersection / len(query_terms)

    # Combined score
    similarity_score = (jaccard_score * 0.6) + (query_coverage * 0.4)

    return similarity_score


def main():
    parser = argparse.ArgumentParser(
        description="Semantic search for relevant code or documentation using text similarity."
    )
    parser.add_argument("query", type=str, help="The search query in natural language")
    parser.add_argument(
        "--max_results",
        type=int,
        default=10,
        help="Maximum number of results to return (default: 10)",
    )
    parser.add_argument(
        "--file_types", nargs="+", help="File types to search in (e.g., .py .js .md)"
    )
    parser.add_argument(
        "--similarity_threshold",
        type=float,
        default=0.1,
        help="Minimum similarity score to include (default: 0.1)",
    )
    parser.add_argument(
        "--context_size",
        type=int,
        default=3,
        help="Number of lines of context around matches (default: 3)",
    )
    parser.add_argument(
        "--search_path",
        type=str,
        help="Directory to search in (default: current directory)",
    )
    parser.add_argument(
        "--output_format",
        choices=["default", "json", "compact"],
        default="default",
        help="Output format (default: default)",
    )

    args = parser.parse_args()

    results = semantic_search(
        args.query,
        max_results=args.max_results,
        file_types=args.file_types,
        similarity_threshold=args.similarity_threshold,
        context_size=args.context_size,
        search_path=args.search_path,
    )

    if not results:
        print(f"No results found for query: {args.query}")
        sys.exit(0)

    if args.output_format == "json":
        import json

        print(json.dumps(results, indent=2))

    elif args.output_format == "compact":
        print(f"Found {len(results)} results for: {args.query}")
        print("-" * 60)

        for i, result in enumerate(results, 1):
            print(
                f"{i}. {result['file_path']}:{result['line_number']} (score: {result['similarity_score']:.3f})"
            )
            print(f"   {result['matched_line'][:100]}...")
            print()

    else:  # default format
        print(f"Semantic search results for: {args.query}")
        print(f"Found {len(results)} results")
        print("=" * 80)

        for i, result in enumerate(results, 1):
            print(f"\nResult {i}: {result['file_path']}:{result['line_number']}")
            print(f"Similarity Score: {result['similarity_score']:.3f}")
            print(f"File Type: {result['file_type']}")

            if result.get("match_type") == "function_definition":
                print(f"Function: {result.get('function_name', 'unknown')}")

            print("Context:")
            for context_line in result["context"]:
                marker = ">>>" if context_line["is_match"] else "   "
                print(
                    f"{marker} {context_line['line_number']:4d}: {context_line['content']}"
                )

            print("-" * 80)


if __name__ == "__main__":
    main()
