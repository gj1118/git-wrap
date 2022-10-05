# git-wrap

Please make sure that the config file is present in the same folder where the binary/script is runnig from.

## Building

Please run the following command where you have cloned the repo
`go build index.go`

If you are on MacOS and are planning to build a Windows binary please set the appropriate environment vars and run the above command. For example as follows
`GOOS=windows GOARCH=amd64 go build -o bin/app-amd64.exe index.go`

There should be a binary present in the root folder. Windows binaries will be present in the `bin` folder.
