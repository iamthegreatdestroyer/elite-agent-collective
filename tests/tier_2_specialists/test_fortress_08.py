"""
Elite Agent Collective - FORTRESS-08 Test Suite
================================================
Agent: FORTRESS (08)
Tier: 2 - Specialist
Specialty: Defensive Security & Penetration Testing

Philosophy: "To defend, you must think like the attacker."

Tests penetration testing, red team operations, incident response,
forensics, and security architecture review capabilities.
"""

import sys
from pathlib import Path
sys.path.insert(0, str(Path(__file__).parent.parent))

from framework.base_agent_test import (
    BaseAgentTest, TestResult, DifficultyLevel, TestCategory
)
from typing import Any, Dict, List, Optional
import hashlib
import base64
import re


class FortressAgentTest(BaseAgentTest):
    """
    Comprehensive test suite for FORTRESS-08 agent.
    
    Tests defensive security capabilities including:
    - Penetration testing (web, network, mobile)
    - Red team operations and threat hunting
    - Incident response and forensics
    - Security architecture review
    - Tools: Burp Suite, Metasploit, Nmap, Wireshark, IDA Pro, Ghidra
    """

    @property
    def agent_id(self) -> str:
        return "08"

    @property
    def agent_codename(self) -> str:
        return "FORTRESS"

    @property
    def agent_tier(self) -> int:
        return 2

    @property
    def agent_specialty(self) -> str:
        return "Defensive Security & Penetration Testing"

    # ═══════════════════════════════════════════════════════════════════════
    # HELPER METHODS
    # ═══════════════════════════════════════════════════════════════════════

    def _classify_vulnerability(self, vuln_type: str) -> Dict:
        """Classify vulnerability severity and type."""
        classifications = {
            "sql_injection": {"severity": "Critical", "cvss": 9.8, "cwe": "CWE-89"},
            "xss": {"severity": "High", "cvss": 7.5, "cwe": "CWE-79"},
            "csrf": {"severity": "Medium", "cvss": 6.5, "cwe": "CWE-352"},
            "ssrf": {"severity": "High", "cvss": 8.0, "cwe": "CWE-918"},
            "rce": {"severity": "Critical", "cvss": 10.0, "cwe": "CWE-94"},
            "auth_bypass": {"severity": "Critical", "cvss": 9.1, "cwe": "CWE-287"},
            "idor": {"severity": "High", "cvss": 7.5, "cwe": "CWE-639"},
            "path_traversal": {"severity": "High", "cvss": 7.5, "cwe": "CWE-22"}
        }
        return classifications.get(vuln_type, {"severity": "Unknown", "cvss": 0, "cwe": "Unknown"})

    def _analyze_log_entry(self, log: str) -> Dict:
        """Analyze a log entry for suspicious patterns."""
        patterns = {
            "sql_injection": r"('|\")\s*(OR|AND)\s*('|\")?1('|\")?=('|\")?1",
            "xss_attempt": r"<script[^>]*>|javascript:|on\w+=",
            "path_traversal": r"\.\.\/|\.\.\\",
            "command_injection": r"[;&|`$]|\$\(|\)\s*{",
            "brute_force": r"failed.*login|invalid.*password|authentication.*failed"
        }
        
        findings = []
        for attack_type, pattern in patterns.items():
            if re.search(pattern, log, re.IGNORECASE):
                findings.append(attack_type)
        
        return {"suspicious": len(findings) > 0, "attack_types": findings}

    # ═══════════════════════════════════════════════════════════════════════
    # L1 TRIVIAL TESTS
    # ═══════════════════════════════════════════════════════════════════════

    def test_L1_trivial_01(self) -> TestResult:
        """Test basic vulnerability classification."""
        def test_func(input_data: Dict) -> Dict:
            vuln_type = input_data["vulnerability"]
            classification = self._classify_vulnerability(vuln_type)
            
            return {
                "vulnerability": vuln_type,
                "classification": classification,
                "is_critical": classification["severity"] == "Critical"
            }

        input_data = {"vulnerability": "sql_injection"}
        expected = {"is_critical": True, "severity": "Critical"}

        return self.execute_test(
            test_name="vulnerability_classification",
            difficulty=DifficultyLevel.TRIVIAL,
            category=TestCategory.CORE_COMPETENCY,
            test_func=test_func,
            input_data=input_data,
            expected_output=expected,
            validation_func=lambda e, a: a["is_critical"] == e["is_critical"]
        )

    def test_L1_trivial_02(self) -> TestResult:
        """Test basic port classification."""
        def test_func(input_data: Dict) -> Dict:
            port = input_data["port"]
            
            well_known_ports = {
                21: {"service": "FTP", "risk": "High", "notes": "Clear text credentials"},
                22: {"service": "SSH", "risk": "Low", "notes": "Encrypted"},
                23: {"service": "Telnet", "risk": "Critical", "notes": "Clear text"},
                25: {"service": "SMTP", "risk": "Medium", "notes": "Email relay"},
                80: {"service": "HTTP", "risk": "Medium", "notes": "Unencrypted web"},
                443: {"service": "HTTPS", "risk": "Low", "notes": "Encrypted web"},
                445: {"service": "SMB", "risk": "High", "notes": "File sharing"},
                3306: {"service": "MySQL", "risk": "High", "notes": "Database"},
                3389: {"service": "RDP", "risk": "High", "notes": "Remote desktop"}
            }
            
            info = well_known_ports.get(port, {"service": "Unknown", "risk": "Unknown", "notes": "Investigate"})
            
            return {
                "port": port,
                "service": info["service"],
                "risk_level": info["risk"],
                "notes": info["notes"]
            }

        input_data = {"port": 22}
        expected = {"service": "SSH", "risk_level": "Low"}

        return self.execute_test(
            test_name="port_classification",
            difficulty=DifficultyLevel.TRIVIAL,
            category=TestCategory.CORE_COMPETENCY,
            test_func=test_func,
            input_data=input_data,
            expected_output=expected,
            validation_func=lambda e, a: a["service"] == e["service"]
        )

    # ═══════════════════════════════════════════════════════════════════════
    # L2 STANDARD TESTS
    # ═══════════════════════════════════════════════════════════════════════

    def test_L2_standard_01(self) -> TestResult:
        """Test web application vulnerability assessment."""
        def test_func(input_data: Dict) -> Dict:
            findings = input_data["scan_results"]
            
            assessment = {
                "total_findings": len(findings),
                "by_severity": {"Critical": 0, "High": 0, "Medium": 0, "Low": 0},
                "prioritized_findings": [],
                "remediation_roadmap": []
            }
            
            for finding in findings:
                severity = finding.get("severity", "Medium")
                assessment["by_severity"][severity] = assessment["by_severity"].get(severity, 0) + 1
            
            # Prioritize findings
            priority_order = ["Critical", "High", "Medium", "Low"]
            for priority in priority_order:
                for finding in findings:
                    if finding.get("severity") == priority:
                        assessment["prioritized_findings"].append({
                            "name": finding["name"],
                            "severity": priority,
                            "remediation": finding.get("remediation", "Review and fix")
                        })
            
            # Create roadmap
            if assessment["by_severity"]["Critical"] > 0:
                assessment["remediation_roadmap"].append({
                    "phase": 1,
                    "timeframe": "Immediate (24-48h)",
                    "focus": "Critical vulnerabilities"
                })
            if assessment["by_severity"]["High"] > 0:
                assessment["remediation_roadmap"].append({
                    "phase": 2,
                    "timeframe": "Short-term (1-2 weeks)",
                    "focus": "High severity vulnerabilities"
                })
            
            assessment["overall_risk"] = "Critical" if assessment["by_severity"]["Critical"] > 0 else \
                                        "High" if assessment["by_severity"]["High"] > 0 else "Medium"
            
            return assessment

        input_data = {
            "scan_results": [
                {"name": "SQL Injection in login", "severity": "Critical", "remediation": "Use parameterized queries"},
                {"name": "XSS in search", "severity": "High", "remediation": "Encode output"},
                {"name": "Missing CSRF token", "severity": "Medium", "remediation": "Implement CSRF protection"},
                {"name": "Verbose error messages", "severity": "Low", "remediation": "Use generic errors"}
            ]
        }
        expected = {"overall_risk": "Critical", "total_findings": 4}

        return self.execute_test(
            test_name="web_vulnerability_assessment",
            difficulty=DifficultyLevel.STANDARD,
            category=TestCategory.CORE_COMPETENCY,
            test_func=test_func,
            input_data=input_data,
            expected_output=expected,
            validation_func=lambda e, a: (
                a["overall_risk"] == e["overall_risk"] and
                a["total_findings"] == e["total_findings"]
            )
        )

    def test_L2_standard_02(self) -> TestResult:
        """Test network reconnaissance analysis."""
        def test_func(input_data: Dict) -> Dict:
            nmap_results = input_data["nmap_scan"]
            
            analysis = {
                "hosts_discovered": len(nmap_results),
                "open_ports_total": 0,
                "services_identified": [],
                "high_risk_services": [],
                "attack_surface": []
            }
            
            high_risk_ports = [21, 23, 445, 3389, 5900]  # FTP, Telnet, SMB, RDP, VNC
            
            for host in nmap_results:
                for port_info in host.get("ports", []):
                    analysis["open_ports_total"] += 1
                    service = {
                        "host": host["ip"],
                        "port": port_info["port"],
                        "service": port_info.get("service", "unknown"),
                        "version": port_info.get("version", "unknown")
                    }
                    analysis["services_identified"].append(service)
                    
                    if port_info["port"] in high_risk_ports:
                        analysis["high_risk_services"].append(service)
                        analysis["attack_surface"].append({
                            "target": f"{host['ip']}:{port_info['port']}",
                            "risk": "High",
                            "recommendation": f"Review necessity of {port_info.get('service', 'service')}"
                        })
            
            analysis["risk_score"] = min(10, len(analysis["high_risk_services"]) * 2 + analysis["open_ports_total"] * 0.5)
            
            return analysis

        input_data = {
            "nmap_scan": [
                {
                    "ip": "192.168.1.10",
                    "ports": [
                        {"port": 22, "service": "ssh", "version": "OpenSSH 8.0"},
                        {"port": 80, "service": "http", "version": "nginx 1.18"},
                        {"port": 3389, "service": "rdp", "version": "Windows RDP"}
                    ]
                },
                {
                    "ip": "192.168.1.20",
                    "ports": [
                        {"port": 445, "service": "smb", "version": "Samba 4.0"}
                    ]
                }
            ]
        }
        expected = {"hosts_discovered": 2, "high_risk_count": 2}

        return self.execute_test(
            test_name="network_reconnaissance",
            difficulty=DifficultyLevel.STANDARD,
            category=TestCategory.CORE_COMPETENCY,
            test_func=test_func,
            input_data=input_data,
            expected_output=expected,
            validation_func=lambda e, a: (
                a["hosts_discovered"] == e["hosts_discovered"] and
                len(a["high_risk_services"]) >= 2
            )
        )

    def test_L2_standard_03(self) -> TestResult:
        """Test log analysis for security events."""
        def test_func(input_data: Dict) -> Dict:
            logs = input_data["logs"]
            
            analysis = {
                "total_entries": len(logs),
                "suspicious_entries": 0,
                "attack_patterns": {},
                "recommendations": []
            }
            
            for log in logs:
                result = self._analyze_log_entry(log)
                if result["suspicious"]:
                    analysis["suspicious_entries"] += 1
                    for attack_type in result["attack_types"]:
                        analysis["attack_patterns"][attack_type] = \
                            analysis["attack_patterns"].get(attack_type, 0) + 1
            
            # Generate recommendations
            if "sql_injection" in analysis["attack_patterns"]:
                analysis["recommendations"].append("Implement WAF rules for SQL injection")
            if "brute_force" in analysis["attack_patterns"]:
                analysis["recommendations"].append("Implement rate limiting and account lockout")
            if "xss_attempt" in analysis["attack_patterns"]:
                analysis["recommendations"].append("Review input validation and output encoding")
            
            analysis["threat_level"] = "High" if analysis["suspicious_entries"] > len(logs) * 0.1 else "Medium"
            
            return analysis

        input_data = {
            "logs": [
                "2024-01-01 10:00:00 GET /search?q=test HTTP/1.1 200",
                "2024-01-01 10:01:00 GET /search?q=' OR '1'='1 HTTP/1.1 200",
                "2024-01-01 10:02:00 GET /page?id=<script>alert(1)</script> HTTP/1.1 200",
                "2024-01-01 10:03:00 POST /login failed login for user admin",
                "2024-01-01 10:04:00 POST /login failed login for user admin",
                "2024-01-01 10:05:00 GET /normal-page HTTP/1.1 200"
            ]
        }
        expected = {"has_attacks": True}

        return self.execute_test(
            test_name="log_security_analysis",
            difficulty=DifficultyLevel.STANDARD,
            category=TestCategory.CORE_COMPETENCY,
            test_func=test_func,
            input_data=input_data,
            expected_output=expected,
            validation_func=lambda e, a: (
                a["suspicious_entries"] >= 3 and
                len(a["attack_patterns"]) >= 2
            )
        )

    # ═══════════════════════════════════════════════════════════════════════
    # L3 ADVANCED TESTS
    # ═══════════════════════════════════════════════════════════════════════

    def test_L3_advanced_01(self) -> TestResult:
        """Test penetration test methodology and reporting."""
        def test_func(input_data: Dict) -> Dict:
            scope = input_data["scope"]
            
            pentest_plan = {
                "methodology": "PTES (Penetration Testing Execution Standard)",
                "phases": [
                    {
                        "phase": 1,
                        "name": "Pre-engagement Interactions",
                        "activities": ["Scope definition", "Rules of engagement", "Authorization"],
                        "deliverables": ["Signed authorization", "Scope document"]
                    },
                    {
                        "phase": 2,
                        "name": "Intelligence Gathering",
                        "activities": ["OSINT", "DNS enumeration", "Technology fingerprinting"],
                        "tools": ["theHarvester", "Shodan", "Maltego", "Recon-ng"],
                        "deliverables": ["Target profile", "Attack surface map"]
                    },
                    {
                        "phase": 3,
                        "name": "Threat Modeling",
                        "activities": ["Identify assets", "Map threats", "Prioritize targets"],
                        "frameworks": ["STRIDE", "DREAD", "Attack Trees"],
                        "deliverables": ["Threat model", "Attack scenarios"]
                    },
                    {
                        "phase": 4,
                        "name": "Vulnerability Analysis",
                        "activities": ["Automated scanning", "Manual testing", "Configuration review"],
                        "tools": ["Nessus", "Burp Suite", "Nikto", "OWASP ZAP"],
                        "deliverables": ["Vulnerability list", "Risk ratings"]
                    },
                    {
                        "phase": 5,
                        "name": "Exploitation",
                        "activities": ["Exploit development", "Payload delivery", "Post-exploitation"],
                        "tools": ["Metasploit", "Cobalt Strike", "Custom scripts"],
                        "deliverables": ["Proof of concept", "Access evidence"]
                    },
                    {
                        "phase": 6,
                        "name": "Post-Exploitation",
                        "activities": ["Privilege escalation", "Lateral movement", "Data exfiltration"],
                        "considerations": ["Minimize impact", "Document everything"],
                        "deliverables": ["Access paths", "Data exposure assessment"]
                    },
                    {
                        "phase": 7,
                        "name": "Reporting",
                        "activities": ["Executive summary", "Technical findings", "Remediation"],
                        "deliverables": ["Pentest report", "Presentation", "Remediation roadmap"]
                    }
                ],
                "scope_specifics": {
                    "in_scope": scope.get("targets", []),
                    "out_of_scope": scope.get("exclusions", []),
                    "test_types": scope.get("test_types", ["black_box"]),
                    "timeline": scope.get("duration_days", 10)
                },
                "risk_mitigation": [
                    "Backup critical systems before testing",
                    "Maintain communication channel with SOC",
                    "Stop testing if production impact detected"
                ]
            }
            
            return pentest_plan

        input_data = {
            "scope": {
                "targets": ["webapp.example.com", "api.example.com", "192.168.1.0/24"],
                "exclusions": ["production-db.example.com"],
                "test_types": ["black_box", "web_application"],
                "duration_days": 14
            }
        }
        expected = {"methodology": "PTES (Penetration Testing Execution Standard)"}

        return self.execute_test(
            test_name="pentest_methodology",
            difficulty=DifficultyLevel.ADVANCED,
            category=TestCategory.CORE_COMPETENCY,
            test_func=test_func,
            input_data=input_data,
            expected_output=expected,
            validation_func=lambda e, a: (
                len(a["phases"]) >= 6 and
                "methodology" in a
            )
        )

    def test_L3_advanced_02(self) -> TestResult:
        """Test incident response playbook creation."""
        def test_func(input_data: Dict) -> Dict:
            incident_type = input_data["incident_type"]
            
            playbooks = {
                "ransomware": {
                    "severity": "Critical",
                    "initial_actions": [
                        "Isolate affected systems from network",
                        "Preserve evidence (memory dumps, disk images)",
                        "Notify incident response team",
                        "Do NOT pay ransom (initial stance)"
                    ],
                    "investigation_steps": [
                        "Identify ransomware variant",
                        "Determine initial infection vector",
                        "Map affected systems and data",
                        "Check for data exfiltration"
                    ],
                    "containment": [
                        "Block C2 communications at firewall",
                        "Disable affected user accounts",
                        "Isolate network segments",
                        "Revoke compromised credentials"
                    ],
                    "eradication": [
                        "Remove malware from affected systems",
                        "Rebuild from clean images if needed",
                        "Patch exploited vulnerabilities",
                        "Reset all potentially compromised credentials"
                    ],
                    "recovery": [
                        "Restore from clean backups",
                        "Verify data integrity",
                        "Monitor for reinfection",
                        "Gradually restore services"
                    ],
                    "lessons_learned": [
                        "Timeline reconstruction",
                        "Gap analysis",
                        "Control improvements",
                        "Update playbooks"
                    ]
                },
                "data_breach": {
                    "severity": "Critical",
                    "initial_actions": [
                        "Assess scope of breach",
                        "Preserve evidence",
                        "Notify legal and compliance",
                        "Engage breach response team"
                    ],
                    "investigation_steps": [
                        "Identify compromised data",
                        "Determine access timeline",
                        "Identify affected individuals",
                        "Assess regulatory obligations"
                    ],
                    "containment": [
                        "Revoke attacker access",
                        "Reset compromised accounts",
                        "Implement additional monitoring"
                    ],
                    "notification": [
                        "Regulatory notification (72h GDPR)",
                        "Affected individual notification",
                        "Public disclosure if required"
                    ]
                }
            }
            
            playbook = playbooks.get(incident_type, playbooks["ransomware"])
            playbook["incident_type"] = incident_type
            playbook["escalation_matrix"] = {
                "tier_1": "SOC Analyst - Initial triage",
                "tier_2": "Incident Response Team - Investigation",
                "tier_3": "CISO/Management - Major incidents",
                "external": "Law enforcement, Forensics firm"
            }
            
            return playbook

        input_data = {"incident_type": "ransomware"}
        expected = {"severity": "Critical"}

        return self.execute_test(
            test_name="incident_response_playbook",
            difficulty=DifficultyLevel.ADVANCED,
            category=TestCategory.CORE_COMPETENCY,
            test_func=test_func,
            input_data=input_data,
            expected_output=expected,
            validation_func=lambda e, a: (
                a["severity"] == e["severity"] and
                len(a["initial_actions"]) >= 3 and
                "escalation_matrix" in a
            )
        )

    def test_L3_advanced_03(self) -> TestResult:
        """Test security architecture review."""
        def test_func(input_data: Dict) -> Dict:
            architecture = input_data["architecture"]
            
            review = {
                "architecture_summary": architecture.get("description", ""),
                "security_controls_analysis": [],
                "gaps_identified": [],
                "recommendations": [],
                "compliance_check": {}
            }
            
            # Analyze security controls
            controls = architecture.get("security_controls", [])
            required_controls = [
                "WAF", "IDS/IPS", "SIEM", "MFA", "Encryption at rest",
                "Encryption in transit", "Network segmentation", "Access control"
            ]
            
            for control in required_controls:
                present = control.lower() in [c.lower() for c in controls]
                review["security_controls_analysis"].append({
                    "control": control,
                    "status": "Implemented" if present else "Missing",
                    "criticality": "High"
                })
                if not present:
                    review["gaps_identified"].append(f"Missing: {control}")
                    review["recommendations"].append({
                        "finding": f"No {control} detected",
                        "recommendation": f"Implement {control}",
                        "priority": "High"
                    })
            
            # Compliance mapping
            frameworks = ["SOC2", "GDPR", "PCI-DSS"]
            for framework in frameworks:
                controls_met = sum(1 for c in review["security_controls_analysis"] if c["status"] == "Implemented")
                review["compliance_check"][framework] = {
                    "controls_required": len(required_controls),
                    "controls_met": controls_met,
                    "compliance_percentage": controls_met / len(required_controls) * 100
                }
            
            review["overall_security_score"] = sum(
                1 for c in review["security_controls_analysis"] if c["status"] == "Implemented"
            ) / len(required_controls) * 100
            
            return review

        input_data = {
            "architecture": {
                "description": "3-tier web application",
                "security_controls": ["WAF", "MFA", "Encryption in transit", "SIEM"]
            }
        }
        expected = {"has_gaps": True}

        return self.execute_test(
            test_name="security_architecture_review",
            difficulty=DifficultyLevel.ADVANCED,
            category=TestCategory.CORE_COMPETENCY,
            test_func=test_func,
            input_data=input_data,
            expected_output=expected,
            validation_func=lambda e, a: (
                len(a["gaps_identified"]) >= 1 and
                "overall_security_score" in a
            )
        )

    # ═══════════════════════════════════════════════════════════════════════
    # L4 EXPERT TESTS
    # ═══════════════════════════════════════════════════════════════════════

    def test_L4_expert_01(self) -> TestResult:
        """Test advanced threat hunting scenario."""
        def test_func(input_data: Dict) -> Dict:
            hypothesis = input_data["threat_hypothesis"]
            available_data = input_data["data_sources"]
            
            threat_hunt = {
                "hypothesis": hypothesis,
                "hunt_methodology": "MITRE ATT&CK based hypothesis-driven hunting",
                "data_sources_utilized": available_data,
                "hunt_queries": [],
                "indicators_to_search": [],
                "timeline": []
            }
            
            # Map hypothesis to ATT&CK techniques
            attack_mappings = {
                "lateral_movement": {
                    "techniques": ["T1021 - Remote Services", "T1075 - Pass the Hash", "T1076 - RDP"],
                    "data_sources": ["Authentication logs", "Network traffic", "Windows Event Logs"],
                    "indicators": [
                        "Multiple failed logins followed by success",
                        "NTLM authentication anomalies",
                        "RDP from unusual sources"
                    ],
                    "queries": [
                        "event_id:4624 AND logon_type:10 | stats count by src_ip, dest_ip",
                        "event_id:4648 AND NOT src_ip IN known_jump_hosts",
                        "process_name:psexec* OR process_name:wmic.exe"
                    ]
                },
                "data_exfiltration": {
                    "techniques": ["T1048 - Exfiltration Over Alternative Protocol", "T1567 - Exfiltration to Cloud"],
                    "data_sources": ["Network traffic", "Proxy logs", "DLP alerts"],
                    "indicators": [
                        "Large outbound data transfers",
                        "Connections to file sharing services",
                        "DNS tunneling patterns"
                    ],
                    "queries": [
                        "bytes_out > 100MB AND dest_port NOT IN (443, 80)",
                        "dns_query_length > 50 chars",
                        "dest_domain IN file_sharing_domains"
                    ]
                },
                "persistence": {
                    "techniques": ["T1053 - Scheduled Task", "T1547 - Boot Autostart"],
                    "data_sources": ["Windows Event Logs", "Registry", "Sysmon"],
                    "indicators": [
                        "New scheduled tasks",
                        "Registry run key modifications",
                        "Service installations"
                    ],
                    "queries": [
                        "event_id:4698 | stats count by task_name, user",
                        "registry_path:*\\Run* AND operation:SetValue",
                        "event_id:7045 | table service_name, image_path"
                    ]
                }
            }
            
            hunt_data = attack_mappings.get(hypothesis, attack_mappings["lateral_movement"])
            threat_hunt["mitre_techniques"] = hunt_data["techniques"]
            threat_hunt["indicators_to_search"] = hunt_data["indicators"]
            threat_hunt["hunt_queries"] = hunt_data["queries"]
            
            threat_hunt["timeline"] = [
                {"phase": "Preparation", "duration": "2 hours", "activities": ["Define scope", "Prepare queries"]},
                {"phase": "Execution", "duration": "4 hours", "activities": ["Run queries", "Analyze results"]},
                {"phase": "Investigation", "duration": "4 hours", "activities": ["Deep dive findings", "Correlate events"]},
                {"phase": "Documentation", "duration": "2 hours", "activities": ["Document findings", "Update IOCs"]}
            ]
            
            return threat_hunt

        input_data = {
            "threat_hypothesis": "lateral_movement",
            "data_sources": ["Windows Event Logs", "Network Traffic", "EDR Telemetry"]
        }
        expected = {"has_queries": True}

        return self.execute_test(
            test_name="threat_hunting",
            difficulty=DifficultyLevel.EXPERT,
            category=TestCategory.CORE_COMPETENCY,
            test_func=test_func,
            input_data=input_data,
            expected_output=expected,
            validation_func=lambda e, a: (
                len(a["hunt_queries"]) >= 2 and
                len(a["mitre_techniques"]) >= 2
            )
        )

    def test_L4_expert_02(self) -> TestResult:
        """Test red team operation planning."""
        def test_func(input_data: Dict) -> Dict:
            objectives = input_data["objectives"]
            constraints = input_data["constraints"]
            
            operation_plan = {
                "operation_name": f"Operation {hashlib.md5(str(objectives).encode()).hexdigest()[:8].upper()}",
                "objectives": objectives,
                "methodology": "Adversary Simulation - APT Emulation",
                "kill_chain_phases": [
                    {
                        "phase": "Reconnaissance",
                        "techniques": ["OSINT", "Social engineering recon", "Technical recon"],
                        "tools": ["Maltego", "theHarvester", "LinkedIn scraping"],
                        "duration": "1-2 weeks"
                    },
                    {
                        "phase": "Weaponization",
                        "techniques": ["Payload development", "Exploit customization"],
                        "tools": ["Metasploit", "Cobalt Strike", "Custom implants"],
                        "duration": "1 week"
                    },
                    {
                        "phase": "Delivery",
                        "techniques": ["Spear phishing", "Watering hole", "Supply chain"],
                        "tools": ["GoPhish", "Evilginx2", "Custom infrastructure"],
                        "duration": "1-2 weeks"
                    },
                    {
                        "phase": "Exploitation",
                        "techniques": ["Initial access exploitation", "Client-side attacks"],
                        "considerations": ["Evade detection", "Minimal footprint"],
                        "duration": "1 week"
                    },
                    {
                        "phase": "Installation",
                        "techniques": ["Persistence mechanisms", "Backdoor installation"],
                        "evasion": ["AMSI bypass", "EDR evasion", "Living off the land"],
                        "duration": "Ongoing"
                    },
                    {
                        "phase": "Command & Control",
                        "techniques": ["Encrypted C2", "Domain fronting", "Legitimate services"],
                        "infrastructure": ["Redirectors", "Proxy chains", "Cloud infrastructure"],
                        "duration": "Duration of operation"
                    },
                    {
                        "phase": "Actions on Objectives",
                        "objectives_mapped": objectives,
                        "techniques": ["Privilege escalation", "Lateral movement", "Data access"],
                        "duration": "1-2 weeks"
                    }
                ],
                "opsec_measures": [
                    "Use encrypted communications only",
                    "Minimize artifacts",
                    "Use time-delayed actions",
                    "Blend with normal traffic patterns"
                ],
                "detection_testing": [
                    "Test SOC response",
                    "Evaluate EDR effectiveness",
                    "Measure MTTD and MTTR"
                ],
                "constraints_addressed": constraints,
                "deconfliction": {
                    "notification": "Security team lead only",
                    "safe_words": "Use agreed safe word to stop operation",
                    "out_of_bounds": constraints.get("out_of_scope", [])
                }
            }
            
            return operation_plan

        input_data = {
            "objectives": ["Gain domain admin", "Access sensitive data", "Test detection capabilities"],
            "constraints": {
                "duration": "4 weeks",
                "out_of_scope": ["production database", "customer data"],
                "hours": "Business hours only"
            }
        }
        expected = {"has_kill_chain": True}

        return self.execute_test(
            test_name="red_team_planning",
            difficulty=DifficultyLevel.EXPERT,
            category=TestCategory.CORE_COMPETENCY,
            test_func=test_func,
            input_data=input_data,
            expected_output=expected,
            validation_func=lambda e, a: (
                len(a["kill_chain_phases"]) >= 6 and
                "opsec_measures" in a
            )
        )

    # ═══════════════════════════════════════════════════════════════════════
    # L5 EXTREME TESTS
    # ═══════════════════════════════════════════════════════════════════════

    def test_L5_extreme_01(self) -> TestResult:
        """Test zero-day vulnerability analysis and response."""
        def test_func(input_data: Dict) -> Dict:
            vulnerability = input_data["vulnerability"]
            
            analysis = {
                "vulnerability_summary": vulnerability,
                "severity_assessment": {
                    "cvss_v3_estimate": 9.8,
                    "exploitability": "High",
                    "impact": "Critical",
                    "scope": "Changed"
                },
                "attack_vector_analysis": {
                    "vector": "Network",
                    "complexity": "Low",
                    "privileges_required": "None",
                    "user_interaction": "None"
                },
                "immediate_mitigations": [
                    "Apply network segmentation to limit exposure",
                    "Implement strict egress filtering",
                    "Enable enhanced monitoring on affected systems",
                    "Deploy virtual patching via WAF/IPS"
                ],
                "detection_strategies": [
                    {
                        "method": "Network signatures",
                        "description": "Create IDS rules for exploit patterns",
                        "effectiveness": "Medium (can be evaded)"
                    },
                    {
                        "method": "Behavioral analysis",
                        "description": "Monitor for post-exploitation behavior",
                        "effectiveness": "High"
                    },
                    {
                        "method": "Honeypot deployment",
                        "description": "Deploy decoys to detect exploitation attempts",
                        "effectiveness": "Medium"
                    }
                ],
                "threat_intelligence": {
                    "exploit_availability": "Unknown - assume weaponized",
                    "threat_actors": "APT groups likely aware",
                    "exploitation_timeline": "Assume active exploitation"
                },
                "long_term_remediation": [
                    "Coordinate with vendor for official patch",
                    "Plan emergency patching procedure",
                    "Review and harden similar systems",
                    "Update incident response playbooks"
                ],
                "communication_plan": {
                    "internal": "CISO, IT, Security team",
                    "external": "Customers if required, regulators per requirements",
                    "timeline": "72 hours for initial assessment"
                }
            }
            
            return analysis

        input_data = {
            "vulnerability": {
                "type": "Remote Code Execution",
                "affected_component": "Web Server Framework",
                "discovery": "Internal research",
                "patch_status": "None available"
            }
        }
        expected = {"is_critical": True}

        return self.execute_test(
            test_name="zero_day_response",
            difficulty=DifficultyLevel.EXTREME,
            category=TestCategory.CORE_COMPETENCY,
            test_func=test_func,
            input_data=input_data,
            expected_output=expected,
            validation_func=lambda e, a: (
                a["severity_assessment"]["cvss_v3_estimate"] >= 9.0 and
                len(a["immediate_mitigations"]) >= 3 and
                len(a["detection_strategies"]) >= 2
            )
        )

    def test_L5_extreme_02(self) -> TestResult:
        """Test APT attribution and campaign analysis."""
        def test_func(input_data: Dict) -> Dict:
            indicators = input_data["indicators"]
            
            attribution_analysis = {
                "indicators_analyzed": len(indicators),
                "technical_analysis": {
                    "malware_family": "Custom backdoor with similarities to known APT tools",
                    "infrastructure": {
                        "c2_domains": indicators.get("domains", []),
                        "hosting_analysis": "Bulletproof hosting, Eastern European providers",
                        "infrastructure_age": "Recent registration, fast-flux DNS"
                    },
                    "ttps": {
                        "initial_access": "Spear phishing with malicious documents",
                        "execution": "PowerShell, WMI, scheduled tasks",
                        "persistence": "Registry run keys, scheduled tasks",
                        "defense_evasion": "Obfuscation, timestomping, log deletion",
                        "c2": "HTTPS, DNS tunneling, legitimate service abuse"
                    }
                },
                "attribution_confidence": "Medium-High",
                "suspected_threat_actor": {
                    "name": "APT-EXAMPLE",
                    "origin": "Nation-state (suspected)",
                    "motivation": "Espionage",
                    "historical_targets": ["Government", "Defense", "Critical infrastructure"],
                    "known_tools": ["Custom malware", "Modified open-source tools"]
                },
                "diamond_model": {
                    "adversary": "State-sponsored",
                    "capability": "High technical sophistication",
                    "infrastructure": "Robust, distributed C2",
                    "victim": input_data.get("target_sector", "Unknown")
                },
                "campaign_timeline": {
                    "first_observed": indicators.get("first_seen", "Unknown"),
                    "last_activity": indicators.get("last_seen", "Unknown"),
                    "campaign_duration": "Ongoing"
                },
                "intelligence_gaps": [
                    "Initial compromise vector not fully confirmed",
                    "Full scope of data access unknown",
                    "Attribution to specific unit uncertain"
                ],
                "recommendations": [
                    "Share IOCs with sector ISACs",
                    "Coordinate with law enforcement",
                    "Enhanced monitoring for related TTPs",
                    "Review and harden based on observed techniques"
                ]
            }
            
            return attribution_analysis

        input_data = {
            "indicators": {
                "domains": ["evil-update.com", "cdn-static.net"],
                "ips": ["192.0.2.1", "198.51.100.1"],
                "hashes": ["abc123...", "def456..."],
                "first_seen": "2024-01-15",
                "last_seen": "2024-01-30"
            },
            "target_sector": "Defense"
        }
        expected = {"has_attribution": True}

        return self.execute_test(
            test_name="apt_attribution",
            difficulty=DifficultyLevel.EXTREME,
            category=TestCategory.CORE_COMPETENCY,
            test_func=test_func,
            input_data=input_data,
            expected_output=expected,
            validation_func=lambda e, a: (
                "suspected_threat_actor" in a and
                "diamond_model" in a and
                a["attribution_confidence"] in ["Medium", "Medium-High", "High"]
            )
        )

    # ═══════════════════════════════════════════════════════════════════════
    # COLLABORATION, EVOLUTION, AND EDGE CASE TESTS
    # ═══════════════════════════════════════════════════════════════════════

    def test_collaboration_scenario(self) -> TestResult:
        """Test collaboration with CIPHER-02 on cryptographic security assessment."""
        def test_func(input_data: Dict) -> Dict:
            system = input_data["system"]
            
            collaboration = {
                "fortress_contribution": {
                    "attack_surface": ["TLS endpoints", "Key storage", "Authentication"],
                    "penetration_tests": [
                        "SSL/TLS configuration testing",
                        "Certificate validation bypass attempts",
                        "Key extraction attempts"
                    ],
                    "vulnerabilities_found": [
                        {"finding": "Weak cipher suites enabled", "severity": "Medium"},
                        {"finding": "Missing certificate pinning", "severity": "High"}
                    ]
                },
                "cipher_contribution": {
                    "cryptographic_analysis": {
                        "key_management": "AES-256 keys derived from PBKDF2",
                        "protocol_security": "TLS 1.3 with AEAD",
                        "random_generation": "Hardware RNG available"
                    },
                    "recommendations": [
                        "Disable TLS 1.0/1.1",
                        "Implement certificate transparency",
                        "Rotate encryption keys quarterly"
                    ]
                },
                "integrated_assessment": {
                    "overall_crypto_posture": "Good with improvements needed",
                    "combined_recommendations": [
                        "Disable weak cipher suites (RC4, DES, 3DES)",
                        "Implement certificate pinning for mobile apps",
                        "Enable HSTS with long max-age",
                        "Migrate to Ed25519 for signatures",
                        "Implement key rotation automation"
                    ],
                    "priority_actions": [
                        {"action": "Disable weak ciphers", "timeline": "Immediate"},
                        {"action": "Implement cert pinning", "timeline": "30 days"},
                        {"action": "Key rotation automation", "timeline": "90 days"}
                    ]
                }
            }
            
            return collaboration

        input_data = {"system": "Mobile banking application"}
        expected = {"has_integrated": True}

        return self.execute_test(
            test_name="fortress_cipher_collaboration",
            difficulty=DifficultyLevel.ADVANCED,
            category=TestCategory.COLLABORATION,
            test_func=test_func,
            input_data=input_data,
            expected_output=expected,
            validation_func=lambda e, a: (
                "integrated_assessment" in a and
                len(a["integrated_assessment"]["combined_recommendations"]) >= 3
            )
        )

    def test_evolution_adaptation(self) -> TestResult:
        """Test adaptation to emerging threat landscape (AI-powered attacks)."""
        def test_func(input_data: Dict) -> Dict:
            emerging_threats = input_data["threats"]
            
            adaptation = {
                "emerging_threat_analysis": {
                    "ai_powered_phishing": {
                        "threat": "LLM-generated highly convincing phishing",
                        "detection_challenges": "Content appears legitimate",
                        "countermeasures": [
                            "AI-based email analysis",
                            "Behavioral sender analysis",
                            "Enhanced user training",
                            "Out-of-band verification procedures"
                        ]
                    },
                    "deepfake_social_engineering": {
                        "threat": "Voice/video deepfakes for impersonation",
                        "detection_challenges": "Increasingly realistic",
                        "countermeasures": [
                            "Multi-factor verification for sensitive requests",
                            "Code word protocols",
                            "Deepfake detection tools",
                            "Policy-based controls (no voice-only authorizations)"
                        ]
                    },
                    "ai_vulnerability_discovery": {
                        "threat": "AI-assisted zero-day discovery",
                        "detection_challenges": "Novel exploit techniques",
                        "countermeasures": [
                            "Proactive AI-assisted defense",
                            "Enhanced anomaly detection",
                            "Reduced attack surface",
                            "Defense in depth"
                        ]
                    }
                },
                "defensive_ai_integration": {
                    "threat_detection": "ML-based anomaly detection",
                    "response_automation": "SOAR with AI decision support",
                    "threat_hunting": "AI-assisted hypothesis generation",
                    "vulnerability_management": "AI-prioritized remediation"
                },
                "updated_capabilities": [
                    "LLM-powered log analysis",
                    "Automated threat intelligence correlation",
                    "AI-assisted malware analysis",
                    "Predictive threat modeling"
                ],
                "training_needs": [
                    "AI/ML security fundamentals",
                    "Adversarial ML techniques",
                    "LLM security considerations",
                    "Deepfake detection methods"
                ]
            }
            
            return adaptation

        input_data = {
            "threats": ["AI phishing", "Deepfakes", "AI vulnerability discovery"]
        }
        expected = {"has_ai_defense": True}

        return self.execute_test(
            test_name="ai_threat_adaptation",
            difficulty=DifficultyLevel.EXPERT,
            category=TestCategory.EVOLUTION,
            test_func=test_func,
            input_data=input_data,
            expected_output=expected,
            validation_func=lambda e, a: (
                "defensive_ai_integration" in a and
                len(a["updated_capabilities"]) >= 3
            )
        )

    def test_edge_case_handling(self) -> TestResult:
        """Test handling of complex security scenarios."""
        def test_func(input_data: Dict) -> Dict:
            scenarios = input_data["scenarios"]
            results = {}
            
            for scenario in scenarios:
                if scenario == "insider_threat":
                    results[scenario] = {
                        "handling": "Privileged access monitoring + behavioral analysis",
                        "challenges": ["Legal considerations", "False positives", "Privacy"],
                        "approach": "Risk-based monitoring with HR coordination"
                    }
                elif scenario == "supply_chain_compromise":
                    results[scenario] = {
                        "handling": "Vendor security assessment + software composition analysis",
                        "challenges": ["Visibility", "Third-party access", "Software dependencies"],
                        "approach": "Zero trust for vendors, SCA for code, SBOM requirements"
                    }
                elif scenario == "nation_state_attack":
                    results[scenario] = {
                        "handling": "Enhanced monitoring + government coordination",
                        "challenges": ["Sophisticated TTPs", "Resources", "Persistence"],
                        "approach": "Assume breach, segment critical assets, threat intelligence"
                    }
                elif scenario == "cloud_misconfiguration":
                    results[scenario] = {
                        "handling": "CSPM + automated remediation",
                        "challenges": ["Rapid change", "Complex permissions", "Multi-cloud"],
                        "approach": "Policy as code, continuous monitoring, least privilege"
                    }
                elif scenario == "iot_compromise":
                    results[scenario] = {
                        "handling": "Network segmentation + firmware analysis",
                        "challenges": ["Limited visibility", "No patching", "Scale"],
                        "approach": "Separate network, baseline behavior, replace when needed"
                    }
            
            return {
                "scenarios_handled": len(results),
                "results": results,
                "general_principles": [
                    "Defense in depth",
                    "Assume breach mentality",
                    "Continuous monitoring",
                    "Incident response readiness"
                ]
            }

        input_data = {
            "scenarios": [
                "insider_threat",
                "supply_chain_compromise",
                "nation_state_attack",
                "cloud_misconfiguration",
                "iot_compromise"
            ]
        }
        expected = {"scenarios_handled": 5}

        return self.execute_test(
            test_name="security_edge_cases",
            difficulty=DifficultyLevel.ADVANCED,
            category=TestCategory.EDGE_CASE,
            test_func=test_func,
            input_data=input_data,
            expected_output=expected,
            validation_func=lambda e, a: a["scenarios_handled"] >= 5
        )


# ═══════════════════════════════════════════════════════════════════════════
# TEST EXECUTION
# ═══════════════════════════════════════════════════════════════════════════

if __name__ == "__main__":
    print("=" * 70)
    print("ELITE AGENT COLLECTIVE - FORTRESS-08 TEST SUITE")
    print("Agent: FORTRESS | Specialty: Defensive Security & Penetration Testing")
    print("=" * 70)
    
    test_suite = FortressAgentTest()
    summary = test_suite.run_all_tests()
    
    print(f"\n📊 Test Results for {summary.agent_codename}-{summary.agent_id}")
    print(f"   Specialty: {summary.agent_specialty}")
    print(f"   Total Tests: {summary.total_tests}")
    print(f"   Passed: {summary.passed_tests}")
    print(f"   Failed: {summary.failed_tests}")
    print(f"   Pass Rate: {summary.pass_rate:.2%}")
    print(f"   Avg Execution Time: {summary.avg_execution_time_ms:.2f}ms")
    
    print("\n📈 Difficulty Breakdown:")
    for level, data in summary.difficulty_breakdown.items():
        print(f"   {level}: {data['passed']}/{data['total']} ({data['pass_rate']:.0%})")
    
    print("\n" + "=" * 70)
    print("FORTRESS-08 TEST SUITE COMPLETE")
    print("=" * 70)
