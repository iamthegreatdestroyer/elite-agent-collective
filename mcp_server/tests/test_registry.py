import os
import tempfile
import unittest
from pathlib import Path
from unittest import mock

from mcp_server.registry import (
    DEFAULT_AGENTS_DIR,
    AgentRegistry,
    RegistryError,
    parse_front_matter,
    tokenize,
)
from mcp_server.tests.fixtures import ALPHA, make_agents_dir


class ParseTests(unittest.TestCase):
    def test_front_matter_split(self):
        meta, body = parse_front_matter(ALPHA)
        self.assertEqual(meta["name"], "ALPHA")
        self.assertEqual(meta["tier"], "1")
        self.assertTrue(body.lstrip().startswith("# @ALPHA"))

    def test_value_may_contain_colon_and_quotes(self):
        meta, _ = parse_front_matter('---\ndescription: "Testing: unit"\n---\nbody')
        self.assertEqual(meta["description"], "Testing: unit")

    def test_crlf_and_bom(self):
        meta, body = parse_front_matter("﻿---\r\nname: X\r\n---\r\nhello\r\n")
        self.assertEqual(meta, {"name": "X"})
        self.assertEqual(body, "hello\n")

    def test_no_front_matter(self):
        meta, body = parse_front_matter("# just text\n")
        self.assertEqual(meta, {})
        self.assertEqual(body, "# just text\n")

    def test_tokenize_drops_stopwords_and_stems(self):
        self.assertEqual(tokenize("I need the testing of tests"), ["test", "test"])


class RegistryTests(unittest.TestCase):
    def setUp(self):
        self.tmp = make_agents_dir()
        self.addCleanup(self.tmp.cleanup)
        self.registry = AgentRegistry(self.tmp.name).load()

    def test_loads_only_agent_files(self):
        self.assertEqual(len(self.registry), 3)
        self.assertEqual([a.name for a in self.registry.list()], ["ALPHA", "BETA", "GAMMA"])

    def test_fields(self):
        alpha = self.registry.get("ALPHA")
        self.assertEqual((alpha.tier, alpha.category, alpha.agent_id), (1, "Foundational", "01"))
        self.assertEqual(alpha.description, "Cryptography and key management")
        self.assertTrue(alpha.prompt.startswith("# @ALPHA"))
        self.assertNotIn("---", alpha.prompt.splitlines()[0])
        self.assertIn("codename: ALPHA", alpha.content)

    def test_crlf_file_is_normalised_and_utf8_kept(self):
        beta = self.registry.get("BETA")
        self.assertEqual(beta.description, "Testing: unit and integration")
        self.assertNotIn("\r", beta.content)
        self.assertIn("Café", beta.prompt)

    def test_missing_front_matter_falls_back_to_file_name(self):
        gamma = self.registry.get("GAMMA")
        self.assertIsNone(gamma.tier)
        self.assertEqual(gamma.category, "Uncategorized")
        self.assertEqual(self.registry.list()[-1].name, "GAMMA")  # no tier sorts last

    def test_get_is_case_insensitive_and_accepts_at_sign(self):
        self.assertIs(self.registry.get(" @alpha "), self.registry.get("ALPHA"))
        self.assertIsNone(self.registry.get("NOPE"))

    def test_list_filters(self):
        self.assertEqual([a.name for a in self.registry.list(tier=2)], ["BETA"])
        self.assertEqual([a.name for a in self.registry.list(category="found")], ["ALPHA"])
        self.assertEqual(self.registry.list(tier=7), [])

    def test_find_ranks_by_rarity_and_field(self):
        matches = self.registry.find("write unit tests for the encryption module")
        self.assertEqual(matches[0]["name"], "BETA")  # 'unit' and 'test' hit BETA's description
        self.assertIn("ALPHA", [m["name"] for m in matches])
        self.assertIn("unit", matches[0]["matched"])

    def test_find_word_shared_by_all_ranks_below_distinctive_word(self):
        # 'software' is in ALPHA and BETA; 'kubernetes' only in GAMMA.
        matches = self.registry.find("software on kubernetes")
        self.assertEqual(matches[0]["name"], "GAMMA")

    def test_find_codename_mention_wins(self):
        matches = self.registry.find("ask @gamma about unit tests")
        self.assertEqual(matches[0]["name"], "GAMMA")
        self.assertEqual(matches[0]["matched"][0], "@GAMMA")

    def test_find_codename_needs_at_sign_or_capitals(self):
        self.assertEqual(self.registry.find("GAMMA please")[0]["matched"][0], "@GAMMA")
        # the ordinary lower-case word is a plain word match, not a mention
        self.assertNotIn("@GAMMA", self.registry.find("gamma rays")[0]["matched"])
        self.assertNotIn("@ALPHA", self.registry.find("alpha software")[0]["matched"])

    def test_find_no_overlap_returns_empty(self):
        self.assertEqual(self.registry.find("zebra xylophone"), [])

    def test_find_limit(self):
        self.assertEqual(len(self.registry.find("software designs", limit=1)), 1)

    def test_roster_markdown(self):
        text = self.registry.roster_markdown()
        self.assertIn("(3 agents)", text)
        self.assertIn("| @ALPHA | 1 | Foundational | Cryptography and key management |", text)


class RegistryErrorTests(unittest.TestCase):
    def test_missing_directory(self):
        with self.assertRaises(RegistryError):
            AgentRegistry(os.path.join(tempfile.gettempdir(), "no-such-agents-dir-xyz")).load()

    def test_empty_directory(self):
        with tempfile.TemporaryDirectory() as tmp:
            with self.assertRaises(RegistryError):
                AgentRegistry(tmp).load()

    def test_duplicate_names(self):
        with tempfile.TemporaryDirectory() as tmp:
            for stem in ("one", "two"):
                Path(tmp, stem + ".agent.md").write_text("---\nname: SAME\n---\nx\n", encoding="utf-8")
            with self.assertRaises(RegistryError):
                AgentRegistry(tmp).load()

    def test_agents_dir_env_and_default(self):
        with mock.patch.dict(os.environ, {"AGENTS_DIR": "some-dir"}):
            self.assertEqual(AgentRegistry().agents_dir, Path("some-dir"))
        with mock.patch.dict(os.environ):
            os.environ.pop("AGENTS_DIR", None)
            self.assertEqual(AgentRegistry().agents_dir, DEFAULT_AGENTS_DIR)


class RealRosterTests(unittest.TestCase):
    """The definitions shipped in this repository."""

    @classmethod
    def setUpClass(cls):
        cls.registry = AgentRegistry(DEFAULT_AGENTS_DIR).load()

    def test_forty_agents_with_complete_front_matter(self):
        self.assertEqual(len(self.registry), 40)
        for agent in self.registry.list():
            self.assertTrue(agent.description, agent.name)
            self.assertIn(agent.tier, range(1, 9), agent.name)
            self.assertNotEqual(agent.category, "Uncategorized", agent.name)
            self.assertEqual(agent.source_file, agent.name + ".agent.md")
            self.assertTrue(agent.prompt, agent.name)

    def test_find_on_real_roster(self):
        for task, expected in (
            ("design a post-quantum cryptography key exchange", "CIPHER"),
            ("set up kubernetes deployment and ci/cd", "FLUX"),
            ("write unit tests and property-based testing", "ECLIPSE"),
        ):
            names = [m["name"] for m in self.registry.find(task, limit=3)]
            self.assertIn(expected, names, task)


if __name__ == "__main__":
    unittest.main()
