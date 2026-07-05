$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

$natsConfig = Join-Path $root ".nats/nats.conf"
if (Test-Path $natsConfig) {
  Write-Host "Starting NATS with $natsConfig"
  Start-Process -WindowStyle Hidden -FilePath "nats-server" -ArgumentList "-c", $natsConfig
} else {
  Write-Host "Starting NATS on 4222"
  Start-Process -WindowStyle Hidden -FilePath "nats-server" -ArgumentList "-js", "-p", "4222", "-sd", ".nats/jetstream"
}

Write-Host "Starting battle server on :6969"
go run ./cmd/server
