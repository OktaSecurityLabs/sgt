### Windows installation and configuration


* via chocolatey
  * @"%SystemRoot%\System32\WindowsPowerShell\v1.0\powershell.exe" -NoProfile -ExecutionPolicy Bypass -Command "iex ((New-Object System.Net.WebClient).DownloadString('https://chocolatey.org/install.ps1'))" && SET "PATH=%PATH%;%ALLUSERSPROFILE%\chocolatey\bin"
  * choco install osquery --params='/InstallService'
  
  * create osquery.flags file in c:\programdata\osquery
  * create osquery.secret in c:\programdata\osquery
  * ```Start-Service osqueryd```


#flagsfile

```bash
--config_plugin=tls
--enroll_secret_path=C:\ProgramData\osquery\osquery.secret
--enroll_tls_endpoint=/node/enroll
--config_tls_endpoint=/node/configure
--tls_hostname=<your tls hostname here>
--config_refresh=300
--config_tls_accelerated_refresh=300
--config_tls_max_attempts=9999
```


### Resources and further reading
https://brewfault.io/blog/2017/9/24/local-configuration-for-osquery-on-windows
