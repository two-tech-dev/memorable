<#
.SYNOPSIS
    Memorable — Auto-install MCP for AI agents (Windows)
.DESCRIPTION
    Detects installed AI agents and registers Memorable as an MCP server.
.PARAMETER Agent
    Target agent: cursor, claude, copilot, windsurf, all (default: all)
.PARAMETER Config
    Optional path to memorable.yaml
.EXAMPLE
    .\install.ps1
    .\install.ps1 -Agent cursor
    .\install.ps1 -Agent claude -Config "C:\Users\me\memorable.yaml"
#>

param(
    [ValidateSet("cursor", "claude", "copilot", "windsurf", "all")]
    [string]$Agent = "all",
    [string]$Config = ""
)

$ErrorActionPreference = "Stop"

function Write-Info($msg)  { Write-Host "[memorable] " -ForegroundColor Cyan -NoNewline; Write-Host $msg }
function Write-Ok($msg)    { Write-Host "[✓] " -ForegroundColor Green -NoNewline; Write-Host $msg }
function Write-Warn($msg)  { Write-Host "[!] " -ForegroundColor Yellow -NoNewline; Write-Host $msg }
function Write-Fail($msg)  { Write-Host "[✗] " -ForegroundColor Red -NoNewline; Write-Host $msg }

# ── Locate memorable binary ──────────────────────────────
function Find-MemorableBin {
    $candidates = @(
        (Get-Command "memorable" -ErrorAction SilentlyContinue | Select-Object -ExpandProperty Source),
        (Join-Path $PSScriptRoot "..\bin\memorable.exe"),
        (Join-Path $env:GOPATH "bin\memorable.exe"),
        (Join-Path $env:USERPROFILE "go\bin\memorable.exe")
    ) | Where-Object { $_ -and (Test-Path $_) }

    if ($candidates.Count -eq 0) {
        Write-Fail "memorable.exe not found. Run 'make build' or 'go install github.com/two-tech-dev/memorable/cmd/memorable@latest' first."
        exit 1
    }

    $bin = $candidates[0]
    Write-Ok "Found memorable at: $bin"
    return $bin
}

# ── Build args array ─────────────────────────────────────
function Get-MemorableArgs {
    if ($Config) {
        return @("-config", $Config)
    }
    return @()
}

# ── Write/merge MCP config ───────────────────────────────
function Set-McpConfig {
    param(
        [string]$FilePath,
        [string]$ServersKey,  # "mcpServers" or "servers"
        [string]$BinPath,
        [string[]]$Args
    )

    $dir = Split-Path $FilePath -Parent
    if (-not (Test-Path $dir)) {
        New-Item -ItemType Directory -Path $dir -Force | Out-Null
    }

    $data = @{}
    if (Test-Path $FilePath) {
        try {
            $data = Get-Content $FilePath -Raw | ConvertFrom-Json -AsHashtable
        } catch {
            $data = @{}
        }
    }

    if (-not $data.ContainsKey($ServersKey)) {
        $data[$ServersKey] = @{}
    }

    $entry = [ordered]@{
        command = $BinPath
        args    = $Args
    }
    if ($ServersKey -eq "servers") {
        $entry["type"] = "stdio"
    }

    $data[$ServersKey]["memorable"] = $entry

    $json = $data | ConvertTo-Json -Depth 10
    Set-Content -Path $FilePath -Value $json -Encoding UTF8
}

# ── Agent installers ─────────────────────────────────────
function Install-Cursor($bin, $args) {
    $configFile = Join-Path $env:USERPROFILE ".cursor\mcp.json"
    Write-Info "Configuring Cursor..."
    Set-McpConfig -FilePath $configFile -ServersKey "mcpServers" -BinPath $bin -Args $args
    Write-Ok "Cursor: $configFile"
}

function Install-Claude($bin, $args) {
    $configFile = Join-Path $env:APPDATA "Claude\claude_desktop_config.json"
    Write-Info "Configuring Claude Desktop / Claude Code..."
    Set-McpConfig -FilePath $configFile -ServersKey "mcpServers" -BinPath $bin -Args $args
    Write-Ok "Claude: $configFile"
}

function Install-Copilot($bin, $args) {
    $configFile = Join-Path (Get-Location) ".vscode\mcp.json"
    Write-Info "Configuring VS Code (GitHub Copilot)..."
    Set-McpConfig -FilePath $configFile -ServersKey "servers" -BinPath $bin -Args $args
    Write-Ok "VS Code Copilot: $configFile"
}

function Install-Windsurf($bin, $args) {
    $configFile = Join-Path $env:USERPROFILE ".codeium\windsurf\mcp_config.json"
    Write-Info "Configuring Windsurf..."
    Set-McpConfig -FilePath $configFile -ServersKey "mcpServers" -BinPath $bin -Args $args
    Write-Ok "Windsurf: $configFile"
}

# ── Main ─────────────────────────────────────────────────
Write-Host ""
Write-Host "  Memorable MCP Installer" -ForegroundColor White
Write-Host ""

$bin = Find-MemorableBin
$args = Get-MemorableArgs

switch ($Agent) {
    "cursor"   { Install-Cursor $bin $args }
    "claude"   { Install-Claude $bin $args }
    "copilot"  { Install-Copilot $bin $args }
    "windsurf" { Install-Windsurf $bin $args }
    "all" {
        Install-Cursor $bin $args
        Install-Claude $bin $args
        Install-Copilot $bin $args
        Install-Windsurf $bin $args
    }
}

Write-Host ""
Write-Ok "Done! Restart your agent to activate Memorable."
Write-Host ""
