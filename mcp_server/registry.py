"""Agent registry: loads the ``*.agent.md`` definitions and searches them.

Every fact served comes from the definition files themselves. There is no
hand-kept keyword map that could drift from the roster.
"""

from __future__ import annotations

import math
import os
import re
from dataclasses import dataclass, field
from pathlib import Path

DEFAULT_AGENTS_DIR = Path(__file__).resolve().parent.parent / ".github" / "agents"
AGENT_SUFFIX = ".agent.md"

_FRONT_MATTER = re.compile(r"\A---\n(.*?)\n---\n?(.*)\Z", re.S)
_WORD = re.compile(r"[a-z0-9][a-z0-9+#]*")
_STOPWORDS = frozenset(
    """a about after all also an and any are as at be been best but by can do does for from get
    give has have help how i if in into is it its me my need needs of on one or our out please
    should so some than that the their them then there these this to up use using want was we
    what when which who why will with would you your""".split()
)

# Field weights for find(): where a task word was found matters.
_W_DESCRIPTION = 3.0
_W_HEADINGS = 2.0
_W_BODY = 1.0
_W_CODENAME = 10.0


class RegistryError(Exception):
    """The agents directory is missing, unreadable or has no definitions."""


def _stem(word: str) -> str:
    """Crude suffix stripping so 'testing' meets 'tests'. Not a real stemmer."""
    for suffix in ("ing", "ed", "es", "s"):
        if word.endswith(suffix) and len(word) - len(suffix) >= 4:
            return word[: -len(suffix)]
    return word


def tokenize(text: str) -> list[str]:
    return [
        _stem(w)
        for w in _WORD.findall(text.lower())
        if len(w) >= 2 and w not in _STOPWORDS
    ]


def parse_front_matter(raw: str) -> tuple[dict[str, str], str]:
    """Split a definition into (front-matter keys, body). Flat ``key: value`` only."""
    raw = raw.replace("\r\n", "\n").lstrip("﻿")
    match = _FRONT_MATTER.match(raw)
    if not match:
        return {}, raw
    meta: dict[str, str] = {}
    for line in match.group(1).split("\n"):
        key, sep, value = line.partition(":")
        if sep and key.strip():
            meta[key.strip().lower()] = value.strip().strip("\"'")
    return meta, match.group(2)


@dataclass(frozen=True)
class Agent:
    name: str
    description: str
    tier: int | None
    category: str
    agent_id: str
    source_file: str
    content: str  # the whole file, line endings normalised to \n
    prompt: str  # the body after the front matter
    _terms: dict[str, float] = field(default_factory=dict, repr=False, compare=False)

    def summary(self) -> dict:
        return {
            "name": self.name,
            "tier": self.tier,
            "category": self.category,
            "description": self.description,
        }


def _load_agent(path: Path) -> Agent:
    raw = path.read_text(encoding="utf-8")
    meta, body = parse_front_matter(raw)
    name = (meta.get("codename") or meta.get("name") or path.name[: -len(AGENT_SUFFIX)]).upper()
    try:
        tier: int | None = int(meta["tier"])
    except (KeyError, ValueError):
        tier = None
    body = body.strip()
    headings = " ".join(re.findall(r"^#+\s+(.*)$", body, re.M))
    terms: dict[str, float] = {}
    for text, weight in (
        (body, _W_BODY),
        (headings, _W_HEADINGS),
        (meta.get("description", ""), _W_DESCRIPTION),
    ):
        for token in set(tokenize(text)):
            terms[token] = max(terms.get(token, 0.0), weight)
    return Agent(
        name=name,
        description=meta.get("description", ""),
        tier=tier,
        category=meta.get("category", "Uncategorized"),
        agent_id=meta.get("id", ""),
        source_file=path.name,
        content=raw.replace("\r\n", "\n"),
        prompt=body,
        _terms=terms,
    )


class AgentRegistry:
    def __init__(self, agents_dir: str | os.PathLike | None = None):
        self.agents_dir = Path(
            agents_dir or os.environ.get("AGENTS_DIR") or DEFAULT_AGENTS_DIR
        )
        self._agents: dict[str, Agent] = {}
        self._idf: dict[str, float] = {}

    def load(self) -> "AgentRegistry":
        if not self.agents_dir.is_dir():
            raise RegistryError(f"agents directory not found: {self.agents_dir}")
        agents: dict[str, Agent] = {}
        for path in sorted(self.agents_dir.glob("*" + AGENT_SUFFIX)):
            agent = _load_agent(path)
            if agent.name in agents:
                raise RegistryError(
                    f"duplicate agent name {agent.name}: "
                    f"{agents[agent.name].source_file} and {agent.source_file}"
                )
            agents[agent.name] = agent
        if not agents:
            raise RegistryError(f"no *{AGENT_SUFFIX} files in {self.agents_dir}")
        self._agents = agents
        # Inverse document frequency: a word every agent uses says little.
        count = len(agents)
        frequency: dict[str, int] = {}
        for agent in agents.values():
            for token in agent._terms:
                frequency[token] = frequency.get(token, 0) + 1
        self._idf = {t: math.log(1 + count / n) for t, n in frequency.items()}
        return self

    def __len__(self) -> int:
        return len(self._agents)

    def list(self, tier: int | None = None, category: str | None = None) -> list[Agent]:
        agents = list(self._agents.values())
        if tier is not None:
            agents = [a for a in agents if a.tier == tier]
        if category:
            wanted = category.lower()
            agents = [a for a in agents if wanted in a.category.lower()]
        return sorted(agents, key=lambda a: (a.tier if a.tier is not None else 99, a.name))

    def get(self, name: str) -> Agent | None:
        return self._agents.get(name.strip().lstrip("@").upper())

    def find(self, task: str, limit: int = 3) -> list[dict]:
        """Rank agents by word overlap between the task and each definition.

        Lexical only (weighted by field and by rarity across the roster); it
        does not understand meaning. Returns [] when nothing overlaps.
        """
        tokens = set(tokenize(task))
        # A codename counts as named only when written @name or in capitals,
        # so "stream the data" does not summon STREAM.
        mentioned = {w.upper() for w in re.findall(r"@([A-Za-z_][\w-]*)", task)}
        mentioned |= set(re.findall(r"(?<![A-Za-z0-9_])[A-Z][A-Z0-9_-]+(?![A-Za-z0-9_])", task))
        results = []
        for agent in self._agents.values():
            matched = sorted(t for t in tokens if t in agent._terms)
            score = sum(agent._terms[t] * self._idf.get(t, 0.0) for t in matched)
            if agent.name in mentioned:
                score += _W_CODENAME
                matched.insert(0, "@" + agent.name)
            if score > 0:
                results.append(
                    {**agent.summary(), "score": round(score, 3), "matched": matched}
                )
        results.sort(key=lambda r: (-r["score"], r["name"]))
        return results[: max(1, limit)]

    def roster_markdown(self, agents: list[Agent] | None = None) -> str:
        agents = self.list() if agents is None else agents
        lines = [
            f"## Elite Agent Collective ({len(agents)} agents)",
            "",
            "| Agent | Tier | Category | Role |",
            "|-------|------|----------|------|",
        ]
        for a in agents:
            tier = "-" if a.tier is None else a.tier
            lines.append(f"| @{a.name} | {tier} | {a.category} | {a.description} |")
        return "\n".join(lines)
