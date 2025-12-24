# Elite Agent Collective - Test Suite Runner
# PowerShell script for executing comprehensive test suite on Windows

param(
    [ValidateSet("all", "tier1", "tier2", "tier3", "tier4", "integration", "comprehensive", "performance", "memory")]
    [string]$TestType = "all",
    
    [switch]$Verbose = $false,
    [switch]$GenerateReport = $true,
    [string]$ReportDir = "test-reports"
)

# Configuration
$ScriptRoot = Split-Path -Parent $MyInvocation.MyCommand.Path
$TestsDir = Join-Path $ScriptRoot "tests"
$PythonExe = "python"
$TestRunner = Join-Path $TestsDir "run_all_tests.py"

# Color codes for output
$Colors = @{
    Header  = "Cyan"
    Success = "Green"
    Warning = "Yellow"
    Error   = "Red"
    Info    = "White"
}

# Output functions
function Write-Header {
    param([string]$Message)
    Write-Host "`n" -NoNewline
    Write-Host "=" * 80 -ForegroundColor $Colors.Header
    Write-Host $Message -ForegroundColor $Colors.Header
    Write-Host "=" * 80 -ForegroundColor $Colors.Header
}

function Write-Success {
    param([string]$Message)
    Write-Host "✓ $Message" -ForegroundColor $Colors.Success
}

function Write-Error-Custom {
    param([string]$Message)
    Write-Host "✗ $Message" -ForegroundColor $Colors.Error
}

function Write-Info {
    param([string]$Message)
    Write-Host "ℹ $Message" -ForegroundColor $Colors.Info
}

function Write-Warning-Custom {
    param([string]$Message)
    Write-Host "⚠ $Message" -ForegroundColor $Colors.Warning
}

# Check Python installation
function Test-Python {
    try {
        $output = & $PythonExe --version 2>&1
        Write-Success "Python found: $output"
        return $true
    }
    catch {
        Write-Error-Custom "Python not found. Please install Python 3.8 or later."
        return $false
    }
}

# Check test files exist
function Test-TestFiles {
    if (-not (Test-Path $TestRunner)) {
        Write-Error-Custom "Test runner not found: $TestRunner"
        return $false
    }
    Write-Success "Test runner found"
    return $true
}

# Create report directory
function Initialize-ReportDir {
    if ($GenerateReport) {
        $fullPath = Join-Path $ScriptRoot $ReportDir
        if (-not (Test-Path $fullPath)) {
            New-Item -ItemType Directory -Path $fullPath -Force | Out-Null
            Write-Success "Report directory created: $fullPath"
        }
        else {
            Write-Info "Report directory: $fullPath"
        }
        return $fullPath
    }
    return $null
}

# Run tests
function Run-Tests {
    param(
        [string]$Type,
        [string]$ReportPath
    )
    
    Write-Header "RUNNING $Type TESTS"
    
    $startTime = Get-Date
    Write-Info "Start time: $startTime"
    Write-Info "Test type: $Type"
    Write-Info "Verbose: $Verbose"
    
    # Build command
    $cmd = @($TestRunner)
    if ($Verbose) {
        $cmd += "--verbose"
    }
    
    # Add test type filter
    if ($Type -ne "all") {
        $cmd += "--test-type", $Type
    }
    
    # Add report output
    if ($ReportPath) {
        $timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
        $reportFile = Join-Path $ReportPath "test-results-$Type-$timestamp.json"
        $cmd += "--report", $reportFile
        Write-Info "Report will be saved to: $reportFile"
    }
    
    try {
        # Run Python test suite
        Write-Host "`nExecuting: python $($cmd -join ' ')`n" -ForegroundColor Gray
        
        if ($Verbose) {
            & $PythonExe @cmd
        }
        else {
            & $PythonExe @cmd 2>&1 | Tee-Object -Variable testOutput | Out-Host
        }
        
        $exitCode = $LASTEXITCODE
        
        if ($exitCode -eq 0) {
            Write-Success "Tests completed successfully"
        }
        else {
            Write-Warning-Custom "Tests completed with status code: $exitCode"
        }
        
        $endTime = Get-Date
        $duration = ($endTime - $startTime).TotalSeconds
        Write-Info "Duration: $duration seconds"
        
        return $exitCode -eq 0
    }
    catch {
        Write-Error-Custom "Test execution failed: $_"
        return $false
    }
}

# Print test type information
function Show-TestTypeInfo {
    Write-Host "`nAvailable test types:" -ForegroundColor $Colors.Info
    Write-Host "  all              - Run all tests (default)" -ForegroundColor Gray
    Write-Host "  tier1            - Foundational agents (APEX, CIPHER, ARCHITECT, AXIOM, VELOCITY)" -ForegroundColor Gray
    Write-Host "  tier2            - Specialist agents (12 agents)" -ForegroundColor Gray
    Write-Host "  tier3            - Innovator agents (NEXUS, GENESIS)" -ForegroundColor Gray
    Write-Host "  tier4            - Meta agents (OMNISCIENT)" -ForegroundColor Gray
    Write-Host "  integration      - Multi-agent collaboration tests" -ForegroundColor Gray
    Write-Host "  comprehensive    - Full integration test suite (6 test modules)" -ForegroundColor Gray
    Write-Host "  performance      - Benchmark and load tests" -ForegroundColor Gray
    Write-Host "  memory           - MNEMONIC memory system tests" -ForegroundColor Gray
}

# Main execution
function Main {
    Write-Header "ELITE AGENT COLLECTIVE - TEST SUITE RUNNER"
    
    Write-Info "PowerShell Test Executor"
    Write-Info "Python-based comprehensive test framework"
    
    # Check prerequisites
    Write-Host "`n[PREREQUISITES]" -ForegroundColor Cyan
    if (-not (Test-Python)) {
        exit 1
    }
    
    if (-not (Test-TestFiles)) {
        exit 1
    }
    
    # Initialize report directory
    $reportPath = Initialize-ReportDir
    
    # Show available tests
    Show-TestTypeInfo
    
    # Run tests
    Write-Host "`n[TEST EXECUTION]" -ForegroundColor Cyan
    $success = Run-Tests -Type $TestType -ReportPath $reportPath
    
    # Final summary
    Write-Header "TEST EXECUTION COMPLETE"
    
    if ($success) {
        Write-Success "All tests passed successfully"
        Write-Info "For detailed results, check the test-reports directory"
        exit 0
    }
    else {
        Write-Error-Custom "Some tests failed"
        Write-Info "Review the output above for error details"
        exit 1
    }
}

# Run main function
Main
