$ErrorActionPreference = 'Stop'

Push-Location -LiteralPath $PSScriptRoot
try {
    Write-Host '== Building C library =='
    & gcc -shared -O2 '-Wl,--export-all-symbols' -o libcalculator.dll c_lib/calculator.c
    if ($LASTEXITCODE -ne 0) {
        throw 'C library build failed.'
    }
    Write-Host '  -> libcalculator.dll'

    Write-Host '== Building Rust library =='
    $cargoCommand = Get-Command cargo -ErrorAction SilentlyContinue
    if ($cargoCommand) {
        $cargoPath = $cargoCommand.Source
    } else {
        $cargoPath = Join-Path $env:USERPROFILE '.cargo/bin/cargo.exe'
        if (-not (Test-Path -LiteralPath $cargoPath)) {
            throw 'Cargo was not found. Install the x86_64-pc-windows-gnu Rust toolchain.'
        }
    }

    $targetDir = Join-Path $PSScriptRoot 'rust_lib/target'
    & $cargoPath rustc --release --target x86_64-pc-windows-gnu --manifest-path rust_lib/Cargo.toml --target-dir $targetDir -- -C link-self-contained=yes
    if ($LASTEXITCODE -ne 0) {
        throw 'Rust library build failed.'
    }
    Copy-Item -LiteralPath (Join-Path $targetDir 'x86_64-pc-windows-gnu/release/calculator_rust.dll') -Destination (Join-Path $PSScriptRoot 'libcalculator_rust.dll')
    Write-Host '  -> libcalculator_rust.dll'

    Write-Host 'Build complete.'
} finally {
    Pop-Location
}
