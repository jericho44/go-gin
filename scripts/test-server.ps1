# PowerShell script to test the server
Write-Host "Starting server test..." -ForegroundColor Green

# Start the server in background
$serverProcess = Start-Process -FilePath ".\bin\server.exe" -PassThru -WindowStyle Hidden

# Wait a moment for server to start
Start-Sleep -Seconds 2

try {
    # Test health endpoint
    Write-Host "Testing health endpoint..." -ForegroundColor Yellow
    $response = Invoke-RestMethod -Uri "http://localhost:8080/health" -Method GET
    
    Write-Host "Health check response:" -ForegroundColor Green
    $response | ConvertTo-Json -Depth 3
    
    # Test API v1 health endpoint
    Write-Host "Testing API v1 health endpoint..." -ForegroundColor Yellow
    $response2 = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/health" -Method GET
    
    Write-Host "API v1 health check response:" -ForegroundColor Green
    $response2 | ConvertTo-Json -Depth 3
    
    Write-Host "Server test completed successfully!" -ForegroundColor Green
}
catch {
    Write-Host "Error testing server: $($_.Exception.Message)" -ForegroundColor Red
}
finally {
    # Stop the server
    if ($serverProcess -and !$serverProcess.HasExited) {
        Write-Host "Stopping server..." -ForegroundColor Yellow
        Stop-Process -Id $serverProcess.Id -Force
    }
}