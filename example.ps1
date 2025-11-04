function GoGetItSingleLine {
    $output = ./go-getit.exe count a
    return $output
}

function GoGetItStream {
    $results = @()

    & "./go-getit.exe" "load" "a" | ForEach-Object {
        $results += $_
    }

    return $results
}

$data = Get-GoData
Write-Host "Result: $data"