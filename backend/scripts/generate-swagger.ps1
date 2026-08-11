$dirs = @('cmd/api')

Get-ChildItem internal/feature -Directory | ForEach-Object {
    foreach ($leaf in @('domain', 'delivery')) {
        $candidate = Join-Path $_.FullName $leaf
        if (Test-Path $candidate) {
            $dirs += (Resolve-Path -Relative $candidate)
        }
    }
}

swag init -g main.go -d ($dirs -join ',') --parseInternal --parseDependency --parseDependencyLevel 3 --useStructName -o docs