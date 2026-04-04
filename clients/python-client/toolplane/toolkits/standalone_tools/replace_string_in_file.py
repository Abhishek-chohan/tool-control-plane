#!/usr/bin/env python3
"""
Description: Replace strings in files with enhanced features and safety checks.

This tool performs string replacements in files with validation, backup options,
and various safety features to prevent accidental data loss.

Parameters:
  file_path (string, required): The absolute path to the file to edit.
  old_string (string, required): The string to be replaced (must match exactly).
  new_string (string, required): The replacement string.
  create_backup (boolean, optional): Create a backup before editing (default: True).
  dry_run (boolean, optional): Show what would be changed without making changes (default: False).
  whole_word (boolean, optional): Only replace whole words (default: False).
  ignore_case (boolean, optional): Perform case-insensitive matching (default: False).
  max_replacements (integer, optional): Maximum number of replacements to make (default: unlimited).
"""

import argparse
import os
import re
import shutil
import sys
import tempfile
import time
from pathlib import Path
from typing import Any, Dict, Optional


def replace_string_in_file(
    file_path: str,
    old_string: str,
    new_string: str,
    create_backup: bool = True,
    dry_run: bool = False,
    whole_word: bool = False,
    ignore_case: bool = False,
    max_replacements: Optional[int] = None,
) -> Dict[str, Any]:
    """
    Replace strings in a file with enhanced features and safety checks.

    Args:
        file_path: The absolute path to the file to edit
        old_string: The string to be replaced
        new_string: The replacement string
        create_backup: Create a backup before editing
        dry_run: Show what would be changed without making changes
        whole_word: Only replace whole words
        ignore_case: Perform case-insensitive matching
        max_replacements: Maximum number of replacements to make

    Returns:
        Dictionary containing operation results and statistics
    """
    try:
        path = Path(file_path)

        # Validate input
        if not path.exists():
            return {
                "success": False,
                "error": f"File does not exist: {file_path}",
                "replacements_made": 0,
            }

        if not path.is_file():
            return {
                "success": False,
                "error": f"Path is not a file: {file_path}",
                "replacements_made": 0,
            }

        # Read the file
        try:
            with open(path, "r", encoding="utf-8", errors="replace") as f:
                original_content = f.read()
        except UnicodeDecodeError:
            # Try with different encodings
            for encoding in ["latin-1", "cp1252"]:
                try:
                    with open(path, "r", encoding=encoding) as f:
                        original_content = f.read()
                    break
                except UnicodeDecodeError:
                    continue
            else:
                return {
                    "success": False,
                    "error": f"Could not read file with any supported encoding: {file_path}",
                    "replacements_made": 0,
                }

        # Prepare search pattern
        if whole_word:
            if ignore_case:
                pattern = re.compile(
                    r"\b" + re.escape(old_string) + r"\b", re.IGNORECASE
                )
            else:
                pattern = re.compile(r"\b" + re.escape(old_string) + r"\b")
        else:
            if ignore_case:
                pattern = re.compile(re.escape(old_string), re.IGNORECASE)
            else:
                pattern = re.compile(re.escape(old_string))

        # Find all matches
        matches = list(pattern.finditer(original_content))

        if not matches:
            return {
                "success": True,
                "message": f"No matches found for '{old_string}' in {file_path}",
                "replacements_made": 0,
                "dry_run": dry_run,
            }

        # Apply max_replacements limit
        if max_replacements is not None and len(matches) > max_replacements:
            matches = matches[:max_replacements]

        # Perform replacements
        replacements_made = len(matches)

        if dry_run:
            # Show what would be changed
            changes = []
            for i, match in enumerate(matches[:5]):  # Show first 5 matches
                start = match.start()
                end = match.end()

                # Get context around the match
                context_start = max(0, start - 50)
                context_end = min(len(original_content), end + 50)

                before = original_content[context_start:context_end]
                after = before.replace(match.group(), new_string)

                changes.append(
                    {
                        "match_number": i + 1,
                        "position": start,
                        "before": before,
                        "after": after,
                        "line_number": original_content[:start].count("\n") + 1,
                    }
                )

            return {
                "success": True,
                "message": f"DRY RUN: Would replace {replacements_made} occurrences of '{old_string}' with '{new_string}' in {file_path}",
                "replacements_made": replacements_made,
                "dry_run": True,
                "changes_preview": changes,
            }

        # Create backup if requested
        backup_path = None
        if create_backup:
            backup_path = create_file_backup(path)

        # Apply replacements
        new_content = pattern.sub(
            new_string, original_content, count=max_replacements or 0
        )

        # Write the modified content
        try:
            with open(path, "w", encoding="utf-8") as f:
                f.write(new_content)
        except Exception as e:
            # Restore backup if write fails
            if backup_path and backup_path.exists():
                shutil.copy2(backup_path, path)
            raise e

        result = {
            "success": True,
            "message": f"Successfully replaced {replacements_made} occurrences of '{old_string}' with '{new_string}' in {file_path}",
            "replacements_made": replacements_made,
            "dry_run": False,
            "backup_created": backup_path is not None,
            "backup_path": str(backup_path) if backup_path else None,
            "original_size": len(original_content),
            "new_size": len(new_content),
            "size_change": len(new_content) - len(original_content),
        }

        return result

    except PermissionError:
        return {
            "success": False,
            "error": f"Permission denied: {file_path}",
            "replacements_made": 0,
        }
    except Exception as e:
        return {
            "success": False,
            "error": f"Error processing file: {e}",
            "replacements_made": 0,
        }


def create_file_backup(file_path: Path) -> Optional[Path]:
    """Create a backup of the file with timestamp."""
    try:
        timestamp = time.strftime("%Y%m%d_%H%M%S")
        backup_path = file_path.with_suffix(f".backup_{timestamp}{file_path.suffix}")

        shutil.copy2(file_path, backup_path)
        return backup_path

    except Exception:
        return None


def validate_old_string_uniqueness(file_path: str, old_string: str) -> Dict[str, Any]:
    """Validate that the old_string appears uniquely in the file."""
    try:
        path = Path(file_path)

        with open(path, "r", encoding="utf-8", errors="replace") as f:
            content = f.read()

        occurrences = content.count(old_string)

        if occurrences == 0:
            return {
                "unique": False,
                "count": 0,
                "message": f"String '{old_string}' not found in file",
            }
        elif occurrences == 1:
            return {
                "unique": True,
                "count": 1,
                "message": f"String '{old_string}' found once (unique)",
            }
        else:
            # Find all occurrences and their contexts
            contexts = []
            start = 0
            for i in range(min(occurrences, 5)):  # Show first 5 occurrences
                pos = content.find(old_string, start)
                if pos == -1:
                    break

                context_start = max(0, pos - 50)
                context_end = min(len(content), pos + len(old_string) + 50)
                context = content[context_start:context_end]
                line_number = content[:pos].count("\n") + 1

                contexts.append(
                    {"position": pos, "line_number": line_number, "context": context}
                )

                start = pos + 1

            return {
                "unique": False,
                "count": occurrences,
                "message": f"String '{old_string}' found {occurrences} times (not unique)",
                "contexts": contexts,
            }

    except Exception as e:
        return {
            "unique": False,
            "count": 0,
            "message": f"Error checking uniqueness: {e}",
        }


def main():
    parser = argparse.ArgumentParser(
        description="Replace strings in files with enhanced features and safety checks."
    )
    parser.add_argument(
        "file_path", type=str, help="The absolute path to the file to edit"
    )
    parser.add_argument(
        "old_string", type=str, help="The string to be replaced (must match exactly)"
    )
    parser.add_argument("new_string", type=str, help="The replacement string")
    parser.add_argument(
        "--create_backup",
        action="store_true",
        default=True,
        help="Create a backup before editing (default: True)",
    )
    parser.add_argument(
        "--no_backup",
        action="store_true",
        default=False,
        help="Don't create a backup before editing",
    )
    parser.add_argument(
        "--dry_run",
        action="store_true",
        default=False,
        help="Show what would be changed without making changes (default: False)",
    )
    parser.add_argument(
        "--whole_word",
        action="store_true",
        default=False,
        help="Only replace whole words (default: False)",
    )
    parser.add_argument(
        "--ignore_case",
        action="store_true",
        default=False,
        help="Perform case-insensitive matching (default: False)",
    )
    parser.add_argument(
        "--max_replacements",
        type=int,
        help="Maximum number of replacements to make (default: unlimited)",
    )
    parser.add_argument(
        "--check_uniqueness",
        action="store_true",
        default=False,
        help="Check if the old_string is unique before replacing",
    )
    parser.add_argument(
        "--output_format",
        choices=["default", "json"],
        default="default",
        help="Output format (default: default)",
    )

    args = parser.parse_args()

    # Handle backup flag
    create_backup = args.create_backup and not args.no_backup

    # Check uniqueness if requested
    if args.check_uniqueness:
        uniqueness_result = validate_old_string_uniqueness(
            args.file_path, args.old_string
        )

        if args.output_format == "json":
            import json

            print(json.dumps(uniqueness_result, indent=2))
        else:
            print(uniqueness_result["message"])

            if not uniqueness_result["unique"] and uniqueness_result["count"] > 1:
                print(f"\nOccurrences found:")
                for i, ctx in enumerate(uniqueness_result.get("contexts", []), 1):
                    print(
                        f"{i}. Line {ctx['line_number']}, Position {ctx['position']}:"
                    )
                    print(f"   Context: {ctx['context'][:100]}...")
                    print()

        if not uniqueness_result["unique"] and uniqueness_result["count"] > 1:
            print(
                "String is not unique. Use --dry_run to see all changes before applying."
            )

    # Perform replacement
    result = replace_string_in_file(
        args.file_path,
        args.old_string,
        args.new_string,
        create_backup=create_backup,
        dry_run=args.dry_run,
        whole_word=args.whole_word,
        ignore_case=args.ignore_case,
        max_replacements=args.max_replacements,
    )

    if args.output_format == "json":
        import json

        print(json.dumps(result, indent=2))
    else:
        if result["success"]:
            print(result["message"])

            if args.dry_run and "changes_preview" in result:
                print("\nChanges preview:")
                for change in result["changes_preview"]:
                    print(
                        f"\nMatch {change['match_number']} at line {change['line_number']}:"
                    )
                    print(f"Before: {change['before']}")
                    print(f"After:  {change['after']}")

            if result.get("backup_created"):
                print(f"Backup created: {result['backup_path']}")

        else:
            print(f"Error: {result['error']}", file=sys.stderr)

    # Exit with error code if operation failed
    sys.exit(0 if result["success"] else 1)


if __name__ == "__main__":
    main()
