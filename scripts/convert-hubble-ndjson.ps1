# Convert Hubble NDJSON output to PolicyPilot format
# Hubble outputs: {"flow":{...},"node_name":"...","time":"..."}
# PolicyPilot expects: {"schema":"cpp.flows.v1","flows":[...]}

param(
    [Parameter(Mandatory=$true)]
    [string]$InputFile,
    
    [Parameter(Mandatory=$false)]
    [string]$OutputFile = "converted-flows.json"
)

Write-Host "Converting Hubble NDJSON to PolicyPilot format..." -ForegroundColor Cyan
Write-Host "Input: $InputFile" -ForegroundColor Yellow
Write-Host "Output: $OutputFile" -ForegroundColor Yellow

$flows = @()
$lineCount = 0

# Read file line by line
Get-Content $InputFile | ForEach-Object {
    $line = $_.Trim()
    if ($line -eq "") { return }
    
    try {
        $obj = $line | ConvertFrom-Json
        
        # Extract the flow object
        if ($obj.flow) {
            $flows += $obj.flow
            $lineCount++
        }
    } catch {
        Write-Warning "Failed to parse line: $_"
    }
}

Write-Host "Parsed $lineCount flows" -ForegroundColor Green

# Create FlowCollection structure
$collection = @{
    schema = "cpp.flows.v1"
    flows = $flows
}

# Convert to JSON and normalize field names
$json = $collection | ConvertTo-Json -Depth 100

# Normalize field names: "IP" -> "ip", "ipVersion" string -> int
$json = $json -replace '"IP":', '"ip":'
$json = $json -replace '"ipVersion":\s*"IPv4"', '"ipVersion": 4'
$json = $json -replace '"ipVersion":\s*"IPv6"', '"ipVersion": 6'

# Write without BOM
[System.IO.File]::WriteAllText((Resolve-Path .).Path + "\" + $OutputFile, $json, [System.Text.UTF8Encoding]::new($false))

Write-Host "Converted flows saved to: $OutputFile" -ForegroundColor Green
Write-Host "Total flows: $lineCount" -ForegroundColor Green

