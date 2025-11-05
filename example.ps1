function GoGetItExample {
    # use in script
    $output = ./dist/getit.exe count a
    return $output
}

function GoGetItExample2 {
    # or if you add to PATH
    $output = getit count a
    return $output
}


$data = GoGetItExample
Write-Host "Result: $data"