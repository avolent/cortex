#!/usr/bin/env python3
"""
Sync Obsidian notes to Astro pages directory.

This script copies markdown files and images from an Obsidian vault
to the Astro project, processing links and filtering by publish status.
"""

import argparse
import os
import re
import shutil
import sys

# Default paths - can be overridden by environment variables or CLI args
DEFAULT_OBSIDIAN_PATH = os.getenv("OBSIDIAN_PATH", os.path.expanduser("~/Documents/Notes/Cortex"))
TMP_PAGES_PATH = "./tmp_pages"
PAGES_PATH = "./src/pages"
TMP_IMAGES_PATH = "./tmp_images"
IMAGES_PATH = "./public/images"
IGNORED_FILES = [".DS_Store"]


class ObsidianSyncError(Exception):
    """Custom exception for sync errors."""

    pass


def copy_folder(source: str, destination: str, dry_run: bool = False) -> None:
    """
    Copy a folder from source to destination.

    Args:
        source: Source directory path
        destination: Destination directory path
        dry_run: If True, only print what would be done without actually doing it

    Raises:
        ObsidianSyncError: If source doesn't exist or copy fails
    """
    if not os.path.exists(source):
        raise ObsidianSyncError(f"Source directory does not exist: {source}")

    if not os.path.isdir(source):
        raise ObsidianSyncError(f"Source is not a directory: {source}")

    try:
        # Delete destination folder if it exists
        if os.path.exists(destination):
            if dry_run:
                print(f"[DRY RUN] Would remove: {destination}")
            else:
                shutil.rmtree(destination)
                print(f"Removed existing: {destination}")

        if dry_run:
            print(f"[DRY RUN] Would create: {destination}")
        else:
            os.makedirs(destination)

        # Loop through all files and subdirectories in the source folder
        for root, _dirs, files in os.walk(source):
            # Construct the corresponding destination path
            relative_path = os.path.relpath(root, source)
            dest_dir = os.path.join(destination, relative_path)

            # Ensure the destination subdirectory exists
            if not os.path.exists(dest_dir):
                if dry_run:
                    print(f"[DRY RUN] Would create directory: {dest_dir}")
                else:
                    os.makedirs(dest_dir)

            # Copy each file
            for file in files:
                if file in IGNORED_FILES:
                    print(f"Ignoring: {file}")
                    continue

                src_file = os.path.join(root, file)
                dest_file = os.path.join(dest_dir, file)

                if dry_run:
                    print(f"[DRY RUN] Would copy: {src_file} -> {dest_file}")
                else:
                    shutil.copy2(src_file, dest_file)
                    print(f"Copied: {src_file} -> {dest_file}")

    except PermissionError as e:
        raise ObsidianSyncError(f"Permission denied: {e}") from e
    except OSError as e:
        raise ObsidianSyncError(f"OS error occurred: {e}") from e


def update_header_links(text: str) -> str:
    """
    Update header links to use lowercase with hyphens.

    Converts [Text](#Link) to [Text](#link) with proper formatting.
    """
    pattern = re.compile(r"\[([^\]]+)\]\(#([^\)]+)\)")

    def replace_func(match):
        text = match.group(1)
        link = match.group(2).replace("%20", " ").lower().replace(" ", "-")
        return f"[{text}](#{link})"

    return pattern.sub(replace_func, text)


def update_image_links(text: str) -> str:
    """
    Update image links by removing 'cortex/' prefix.

    Converts ![AltText](cortex/path) to ![AltText](/path)
    """
    pattern = re.compile(r"!\[([^\]]+)\]\((cortex/[^\)]+)\)")

    def replace_func(match):
        alt_text = match.group(1)
        link = match.group(2).replace("cortex/", "/")
        return f"![{alt_text}]({link})"

    return pattern.sub(replace_func, text)


def update_page_links(text: str) -> str:
    """
    Update page links by removing 'cortex/pages' prefix and .md extension.

    Converts [Text](cortex/pages/path/to/file.md) to [Text](/path/to/file)
    """
    pattern = re.compile(r"\[([^\]]+)\]\((cortex/pages/[^\)]+)\.md\)")

    def replace_func(match):
        text = match.group(1)
        link = re.sub(r"cortex/pages", "", match.group(2))
        return f"[{text}]({link})"

    return pattern.sub(replace_func, text)


def check_publish(text: str) -> bool:
    """
    Check if the file's frontmatter has publish: "true".

    Args:
        text: File content to check

    Returns:
        True if publish is "true", False otherwise
    """
    match = re.match(r"---\n(.*?)\n---", text, re.DOTALL)
    if not match:
        return False

    front_matter = match.group(1)
    publish_match = re.search(r'^publish:\s*"(true|false)"', front_matter, re.MULTILINE)

    if publish_match:
        return publish_match.group(1).strip().lower() == "true"

    return False


def process_files_recursively(folder_path: str, dry_run: bool = False) -> None:
    """
    Process markdown files: filter by publish status and update links.

    Args:
        folder_path: Path to folder containing markdown files
        dry_run: If True, only print what would be done

    Raises:
        ObsidianSyncError: If file processing fails
    """
    if not os.path.exists(folder_path):
        raise ObsidianSyncError(f"Folder does not exist: {folder_path}")

    for root, _directories, files in os.walk(folder_path):
        for filename in files:
            if not filename.endswith(".md"):
                continue

            file_path = os.path.join(root, filename)

            try:
                # Read input file
                with open(file_path, encoding="utf-8") as file:
                    content = file.read()
            except UnicodeDecodeError as e:
                print(f"Warning: Could not read {filename} due to encoding error: {e}")
                continue
            except OSError as e:
                print(f"Warning: Could not read {filename}: {e}")
                continue

            # Check if file should be published
            if not check_publish(content):
                if dry_run:
                    print(f"[DRY RUN] Would remove: {file_path} (not marked for publish)")
                else:
                    os.remove(file_path)
                    print(f"Removed: {filename} (not marked for publish)")

                # Also remove associated image folder if it exists
                image_folder = file_path.replace(TMP_PAGES_PATH, TMP_IMAGES_PATH).replace(".md", "")
                if os.path.exists(image_folder):
                    if dry_run:
                        print(f"[DRY RUN] Would remove: {image_folder}")
                    else:
                        shutil.rmtree(image_folder)
                        print(f"Removed: {image_folder} (associated with unpublished page)")
                continue

            # Process the file
            print(f"Processing: {file_path}")
            updated_content = update_header_links(content)
            updated_content = update_image_links(updated_content)
            updated_content = update_page_links(updated_content)

            if dry_run:
                print(f"[DRY RUN] Would update links in: {file_path}")
            else:
                try:
                    with open(file_path, "w", encoding="utf-8") as file:
                        file.write(updated_content)
                except OSError as e:
                    print(f"Warning: Could not write {filename}: {e}")


def cleanup_tmp_folders(dry_run: bool = False) -> None:
    """Clean up temporary folders."""
    for folder in [TMP_PAGES_PATH, TMP_IMAGES_PATH]:
        if os.path.exists(folder):
            if dry_run:
                print(f"[DRY RUN] Would remove: {folder}")
            else:
                shutil.rmtree(folder)
                print(f"Cleaned up: {folder}")


def main():
    """Main sync function."""
    parser = argparse.ArgumentParser(
        description="Sync Obsidian notes to Astro project",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Environment Variables:
  OBSIDIAN_PATH    Path to Obsidian vault (default: ~/Documents/Notes/Cortex)

Examples:
  # Use default path
  python obsidian_sync.py

  # Specify custom path
  python obsidian_sync.py --obsidian-path /path/to/vault

  # Dry run to see what would happen
  python obsidian_sync.py --dry-run
        """,
    )

    parser.add_argument(
        "--obsidian-path",
        default=DEFAULT_OBSIDIAN_PATH,
        help=f"Path to Obsidian vault (default: {DEFAULT_OBSIDIAN_PATH})",
    )

    parser.add_argument(
        "--dry-run", action="store_true", help="Show what would be done without actually doing it"
    )

    args = parser.parse_args()

    obsidian_path = args.obsidian_path
    dry_run = args.dry_run

    # Validate Obsidian path
    if not os.path.exists(obsidian_path):
        print(f"Error: Obsidian path does not exist: {obsidian_path}", file=sys.stderr)
        print("\nSet the correct path using:", file=sys.stderr)
        print("  export OBSIDIAN_PATH=/path/to/vault", file=sys.stderr)
        print("  or use --obsidian-path argument", file=sys.stderr)
        sys.exit(1)

    if dry_run:
        print("=" * 60)
        print("DRY RUN MODE - No changes will be made")
        print("=" * 60)

    try:
        print(f"\nObsidian vault: {obsidian_path}")
        print("\n1. Copying pages from Obsidian...")
        copy_folder(obsidian_path + "/pages", TMP_PAGES_PATH, dry_run)

        print("\n2. Copying images from Obsidian...")
        copy_folder(obsidian_path + "/images", TMP_IMAGES_PATH, dry_run)

        print("\n3. Processing files (filtering and updating links)...")
        process_files_recursively(TMP_PAGES_PATH, dry_run)

        print("\n4. Moving processed files to destination...")
        copy_folder(TMP_PAGES_PATH, PAGES_PATH, dry_run)
        copy_folder(TMP_IMAGES_PATH, IMAGES_PATH, dry_run)

        print("\n5. Cleaning up temporary folders...")
        cleanup_tmp_folders(dry_run)

        if dry_run:
            print("\n" + "=" * 60)
            print("DRY RUN COMPLETE - No actual changes were made")
            print("=" * 60)
        else:
            print("\n✓ Sync complete!")

    except ObsidianSyncError as e:
        print(f"\nError: {e}", file=sys.stderr)
        cleanup_tmp_folders(dry_run)
        sys.exit(1)
    except KeyboardInterrupt:
        print("\n\nSync interrupted by user", file=sys.stderr)
        cleanup_tmp_folders(dry_run)
        sys.exit(130)
    except Exception as e:
        print(f"\nUnexpected error: {e}", file=sys.stderr)
        cleanup_tmp_folders(dry_run)
        sys.exit(1)


if __name__ == "__main__":
    main()
