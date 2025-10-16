# credmgr CLI Tool

Simple command-line tool for testing Windows Credential Manager functionality.

## Building

From WSL or Linux:
```bash
# Cross-compile for Windows
GOOS=windows GOARCH=amd64 go build -o credmgr.exe ./cmd/credmgr
```

From Windows:
```cmd
go build -o credmgr.exe ./cmd/credmgr
```

## Usage

```cmd
credmgr get <name>          # Retrieve credential
credmgr set <name> <data>   # Store credential  
credmgr del <name>          # Delete credential
```

## Examples

```cmd
# Store a credential
credmgr set myapp-token "secret123"

# Retrieve a credential
credmgr get myapp-token

# Delete a credential
credmgr del myapp-token

# Store data with spaces
credmgr set database-connection "Server=localhost;Database=mydb;User=admin;Password=secret"
```

## Testing on Windows

1. Copy `credmgr.exe` to your Windows machine
2. Run the test script: `test-credmgr.bat`
3. Or test manually:
   ```cmd
   credmgr.exe set test "hello world"
   credmgr.exe get test
   credmgr.exe del test
   ```

## Platform Behavior

- **Windows**: Uses Windows Credential Manager API
- **Linux/macOS**: Returns "not supported" error

The credential data is stored securely in Windows Credential Manager and can be viewed/managed through:
- Windows Credential Manager GUI (`control keymgr.dll`)
- PowerShell (`Get-StoredCredential`, etc.)
- Other applications that use the same API