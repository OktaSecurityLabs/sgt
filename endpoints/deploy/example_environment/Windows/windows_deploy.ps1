Stop-Service -Name "osqueryd"

$secret_filename = "c:\ProgramData\osquery\osquery.secret"
$secret_content = "example-secret"

if (Test-Path -Path $secret_filename) {
    Remove-Item $secret_filename
    Write-Host "Removed Secrets file"
}

[IO.File]::WriteAllLines($secret_filename, $secret_content)

$default_flagpath = "C:\ProgramData\osquery\osquery.flags.default"

if (Test-Path -Path $default_flagpath) {
    Remove-item -Path $default_flagpath
    Write-Host "Removed default flags file"
}

$content = "--config_plugin=tls
--enroll_secret_path=C:\Programdata\osquery\osquery.secret
--enroll_tls_endpoint=/node/enroll
--config_tls_endpoint=/node/configure
--tls_hostname=example.domain.endpoint.com
--config_refresh=300
--config_tls_accelerated_refresh=300
--config_tls_max_attempts=9999"
[IO.File]::WriteAllLines($default_flagpath, $content)

$flagpath = "c:\ProgramData\osquery\osquery.flags"

if (Test-Path -Path $flagpath) {
    Remove-Item -Path $flagpath
    Write-Host "Removed flags file"
}

New-Item -Path C:\ProgramData\osquery\osquery.flags -ItemType SymbolicLink -Value C:\ProgramData\osquery\osquery.flags.default

Start-Service -Name "osqueryd"
