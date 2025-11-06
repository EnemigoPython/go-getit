function GoGetItExample {
    # use in script
    $output = ./dist/getit.exe count
    return $output
}

function GoGetItExample2 {
    # or if you add to PATH
    $output = getit count
    return $output
}


$data = GoGetItExample
Write-Host "Result: $data"

$data = GoGetItExample2
Write-Host "Result: $data"