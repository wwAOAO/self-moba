$ErrorActionPreference = "Stop"

Write-Host "== gofmt =="
gofmt -w cmd internal

Write-Host "== go test =="
go test ./...

Write-Host "== frontend syntax =="
$files = Get-ChildItem -Path web/js -Recurse -Filter *.js | Sort-Object FullName
foreach ($file in $files) {
  node --check $file.FullName
}

Write-Host "ok"
