#!/usr/bin/env python3

from __future__ import annotations

import argparse
import shutil
import tarfile
from pathlib import Path


PRODUCT = "wuji"


def normalized_version(raw: str) -> str:
    return raw[1:] if raw.startswith("v") else raw


def copy_docs(stage_root: Path, repo_root: Path) -> None:
    readme = repo_root / "README.md"
    if readme.is_file():
        shutil.copy2(readme, stage_root / "README.md")


def create_archive(stage_dir: Path, output_dir: Path, archive_base: str) -> Path:
    output_dir.mkdir(parents=True, exist_ok=True)
    archive_path = output_dir / f"{archive_base}.tar.gz"
    with tarfile.open(archive_path, "w:gz") as archive:
        archive.add(stage_dir, arcname=stage_dir.name)
    return archive_path


def main() -> int:
    parser = argparse.ArgumentParser(description="Package wuji release binary")
    parser.add_argument("--version", required=True)
    parser.add_argument("--platform", required=True, choices=["linux", "macos"])
    parser.add_argument("--arch", required=True, choices=["x86_64", "aarch64"])
    parser.add_argument("--binary", required=True)
    parser.add_argument("--output-dir", required=True)
    args = parser.parse_args()

    version = normalized_version(args.version)
    binary_path = Path(args.binary).resolve()
    if not binary_path.is_file():
        raise FileNotFoundError(f"binary not found: {binary_path}")

    repo_root = Path(__file__).resolve().parents[2]
    output_dir = (repo_root / args.output_dir).resolve()
    archive_base = f"{PRODUCT}-{version}-{args.platform}-{args.arch}"
    stage_root = output_dir / archive_base
    if stage_root.exists():
        shutil.rmtree(stage_root)

    stage_bin_dir = stage_root / "bin"
    stage_bin_dir.mkdir(parents=True, exist_ok=True)
    shutil.copy2(binary_path, stage_bin_dir / PRODUCT)
    copy_docs(stage_root, repo_root)

    archive_path = create_archive(stage_root, output_dir, archive_base)
    print(archive_path)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
