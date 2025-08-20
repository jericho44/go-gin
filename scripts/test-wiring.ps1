# Test script to verify component wiring
Write-Host "Testing Gin Golang Structure Component Wiring..." -ForegroundColor Green

# Build the application
Write-Host "`nBuilding application..." -ForegroundColor Yellow
go build -o bin/server cmd/server/main.go

if ($LASTEXITCODE -eq 0) {
    Write-Host "Build successful" -ForegroundColor Green
} else {
    Write-Host "Build failed" -ForegroundColor Red
    exit 1
}

# Run tests
Write-Host "`nRunning component wiring tests..." -ForegroundColor Yellow
go test ./cmd/server -v

if ($LASTEXITCODE -eq 0) {
    Write-Host "All tests passed" -ForegroundColor Green
} else {
    Write-Host "Some tests failed" -ForegroundColor Red
    exit 1
}

Write-Host "`nComponent wiring verification complete!" -ForegroundColor Green