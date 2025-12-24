#!/bin/bash
# Elite Agent Collective - Test Suite Runner
# Bash script for executing comprehensive test suite on Unix/Linux/macOS

set -o pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TESTS_DIR="${SCRIPT_DIR}/tests"
TEST_RUNNER="${TESTS_DIR}/run_all_tests.py"
PYTHON_EXE="${PYTHON_EXE:-python3}"

# Default parameters
TEST_TYPE="${1:-all}"
VERBOSE="${VERBOSE:-false}"
GENERATE_REPORT="${GENERATE_REPORT:-true}"
REPORT_DIR="${REPORT_DIR:-test-reports}"

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
WHITE='\033[0;37m'
NC='\033[0m' # No Color

# Output functions
print_header() {
    echo -e "\n${CYAN}$(printf '=%.0s' {1..80})${NC}"
    echo -e "${CYAN}$1${NC}"
    echo -e "${CYAN}$(printf '=%.0s' {1..80})${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_info() {
    echo -e "${WHITE}ℹ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

# Check Python installation
check_python() {
    if ! command -v $PYTHON_EXE &> /dev/null; then
        print_error "Python not found. Please install Python 3.8 or later."
        return 1
    fi
    
    VERSION=$($PYTHON_EXE --version 2>&1)
    print_success "Python found: $VERSION"
    return 0
}

# Check test files
check_test_files() {
    if [ ! -f "$TEST_RUNNER" ]; then
        print_error "Test runner not found: $TEST_RUNNER"
        return 1
    fi
    print_success "Test runner found"
    return 0
}

# Initialize report directory
initialize_report_dir() {
    if [ "$GENERATE_REPORT" = "true" ]; then
        FULL_PATH="${SCRIPT_DIR}/${REPORT_DIR}"
        mkdir -p "$FULL_PATH"
        print_success "Report directory: $FULL_PATH"
    fi
}

# Show help
show_help() {
    cat << EOF
Elite Agent Collective - Comprehensive Test Suite

Usage: $0 [TEST_TYPE] [OPTIONS]

Test Types:
  all              - Run all tests (default)
  tier1            - Foundational agents (APEX, CIPHER, ARCHITECT, AXIOM, VELOCITY)
  tier2            - Specialist agents (12 agents)
  tier3            - Innovator agents (NEXUS, GENESIS)
  tier4            - Meta agents (OMNISCIENT)
  integration      - Multi-agent collaboration tests
  comprehensive    - Full integration test suite (6 test modules)
  performance      - Benchmark and load tests
  memory           - MNEMONIC memory system tests

Options:
  --verbose        - Enable verbose output
  --no-report      - Disable report generation
  --report-dir DIR - Specify report directory (default: test-reports)
  --help           - Show this help message

Environment Variables:
  PYTHON_EXE       - Python executable (default: python3)
  VERBOSE          - Enable verbose mode (default: false)
  GENERATE_REPORT  - Generate test report (default: true)
  REPORT_DIR       - Report directory (default: test-reports)

Examples:
  # Run all tests
  $0

  # Run tier 1 tests with verbose output
  $0 tier1 --verbose

  # Run comprehensive integration tests
  $0 comprehensive

  # Run performance tests and save to custom directory
  $0 performance --report-dir ./results
EOF
}

# Parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --verbose)
                VERBOSE="true"
                shift
                ;;
            --no-report)
                GENERATE_REPORT="false"
                shift
                ;;
            --report-dir)
                REPORT_DIR="$2"
                shift 2
                ;;
            --help)
                show_help
                exit 0
                ;;
            *)
                TEST_TYPE="$1"
                shift
                ;;
        esac
    done
}

# Run tests
run_tests() {
    print_header "RUNNING $TEST_TYPE TESTS"
    
    local START_TIME=$(date +%s)
    print_info "Start time: $(date)"
    print_info "Test type: $TEST_TYPE"
    print_info "Verbose: $VERBOSE"
    print_info "Report generation: $GENERATE_REPORT"
    
    # Build command
    local CMD=("$PYTHON_EXE" "$TEST_RUNNER")
    
    if [ "$VERBOSE" = "true" ]; then
        CMD+=("--verbose")
    fi
    
    if [ "$TEST_TYPE" != "all" ]; then
        CMD+=("--test-type" "$TEST_TYPE")
    fi
    
    if [ "$GENERATE_REPORT" = "true" ]; then
        local TIMESTAMP=$(date +%Y%m%d-%H%M%S)
        local REPORT_FILE="${SCRIPT_DIR}/${REPORT_DIR}/test-results-${TEST_TYPE}-${TIMESTAMP}.json"
        CMD+=("--report" "$REPORT_FILE")
        print_info "Report will be saved to: $REPORT_FILE"
    fi
    
    # Execute tests
    print_info "Executing: ${CMD[@]}"
    echo ""
    
    if "${CMD[@]}"; then
        print_success "Tests completed successfully"
        local EXIT_CODE=0
    else
        EXIT_CODE=$?
        print_warning "Tests completed with exit code: $EXIT_CODE"
    fi
    
    local END_TIME=$(date +%s)
    local DURATION=$((END_TIME - START_TIME))
    print_info "Duration: ${DURATION}s"
    
    return $EXIT_CODE
}

# Main execution
main() {
    print_header "ELITE AGENT COLLECTIVE - TEST SUITE RUNNER"
    
    print_info "Bash Test Executor"
    print_info "Python-based comprehensive test framework"
    
    # Parse arguments
    parse_args "$@"
    
    # Check prerequisites
    echo ""
    print_info "[PREREQUISITES]"
    
    if ! check_python; then
        exit 1
    fi
    
    if ! check_test_files; then
        exit 1
    fi
    
    initialize_report_dir
    
    # Show available tests
    echo ""
    echo -e "${WHITE}Available test types:${NC}"
    echo -e "  ${WHITE}all${NC}              - Run all tests (default)"
    echo -e "  ${WHITE}tier1${NC}            - Foundational agents"
    echo -e "  ${WHITE}tier2${NC}            - Specialist agents"
    echo -e "  ${WHITE}tier3${NC}            - Innovator agents"
    echo -e "  ${WHITE}tier4${NC}            - Meta agents"
    echo -e "  ${WHITE}integration${NC}      - Multi-agent collaboration tests"
    echo -e "  ${WHITE}comprehensive${NC}    - Full integration test suite"
    echo -e "  ${WHITE}performance${NC}      - Benchmark and load tests"
    echo -e "  ${WHITE}memory${NC}           - MNEMONIC memory system tests"
    
    # Run tests
    echo ""
    print_info "[TEST EXECUTION]"
    
    if run_tests; then
        print_header "TEST EXECUTION COMPLETE"
        print_success "All tests passed successfully"
        print_info "For detailed results, check the test-reports directory"
        exit 0
    else
        print_header "TEST EXECUTION COMPLETE"
        print_error "Some tests failed"
        print_info "Review the output above for error details"
        exit 1
    fi
}

# Run main function
main "$@"
