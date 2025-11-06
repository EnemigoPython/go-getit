# shutdown active server before continuing
# (assumes that the executable is in PATH on windows)
getit exit
# Client executable (command line application)
go build -o dist/getit.exe ./src  
# Server executable (built as GUI to avoid console window on startup)
go build -ldflags -H=windowsgui -o dist/getitserver.exe ./src
# restart server
getitserver --runtime=server
