"""A small throwaway agents directory so unit tests never depend on the real roster."""

import tempfile
from pathlib import Path

ALPHA = """---
name: ALPHA
description: Cryptography and key management
codename: ALPHA
tier: 1
id: 01
category: Foundational
---

# @ALPHA - Cryptography

## Core Capabilities

- Encryption protocols and signatures
- Reviews software designs
"""

# CRLF line endings and a quoted value, as files checked out on Windows have.
BETA = (
    "---\r\nname: beta\r\ndescription: \"Testing: unit and integration\"\r\n"
    "tier: 2\r\ncategory: Specialists\r\n---\r\n\r\n# @BETA - Testing\r\n\r\n"
    "Writes unit tests and reviews software designs. Café-grade coverage.\r\n"
)

# No front matter at all: the name falls back to the file name.
GAMMA = "# Gamma\n\nKubernetes deployment pipelines.\n"


def make_agents_dir() -> tempfile.TemporaryDirectory:
    tmp = tempfile.TemporaryDirectory()
    root = Path(tmp.name)
    (root / "ALPHA.agent.md").write_text(ALPHA, encoding="utf-8", newline="")
    (root / "BETA.agent.md").write_bytes(BETA.encode("utf-8"))
    (root / "gamma.agent.md").write_text(GAMMA, encoding="utf-8", newline="")
    (root / "README.md").write_text("not an agent", encoding="utf-8")
    return tmp
