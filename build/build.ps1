Set-StrictMode -Version Latest

function IsWindows {
    ($PSVersionTable.PSEdition -ne 'Core') -or $IsWindows
}

function ExePath([string] $dir, [string]$name) {
    $path = Join-Path $dir $name
    if (IsWindows) {
        $path += '.exe'
    }
    $path
}

function FileExists([string] $file) {
    Test-Path -PathType Leaf -Path $file
}
function DirExists([string] $dir) {
    Test-Path -PathType Container -Path $dir
}
function CreateDir([string] $dir) {
    if (-not (DirExists $dir)) {
        New-Item -Path $dir -ItemType Directory > $null
    }
}

$initialPwd = $PWD
Write-Host -ForegroundColor Green "Running build in $initialPwd"

try {
    if ($env:GOPATH) {
        $mageExe = ExePath (Join-Path $env:GOPATH 'bin') 'mage'
        

        if (-not (FileExists $mageExe)) {
            Write-Host -ForegroundColor Green "mage tool not found at $mageExe ...build it from local source in vendor/mage"

            $vendorDir = Join-Path $PSScriptRoot '../vendor'
            $mageVendorDir = Join-Path $vendorDir 'mage'
            $mageBootstrap = Join-Path $mageVendorDir 'bootstrap.go'

            if (-not (FileExists $mageBootstrap)) {
                Write-Host -ForegroundColor Green 'Need to download mage source from https://github.com/magefile/mage'

                CreateDir $vendorDir
                Set-Location $vendorDir
                git clone https://github.com/magefile/mage
            }

            if (FileExists $mageBootstrap) {
                Set-Location $mageVendorDir
                go run bootstrap.go
            }
            else {
                Write-Error "mage's bootstrap.go not found"
                exit 1
            }
        }
        else {
            Write-Host -ForegroundColor Green "mage tool found at $mageExe"
        }

        & $mageExe -v
    }
    else {
        Write-Error "GOPATH env var isn't set. Please install golang first."
        exit 1
    }

}
finally {
    Set-Location $initialPwd
}
