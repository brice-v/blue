# blue - a fun programming language

## Building

- Install deps for fyne
- make sure no errors with `go build`
    - [had this error](https://stackoverflow.com/questions/65387167/glfw-pkg-config-error-when-building-a-fyne-app)
    - windows has no requirements
    - still havent tested on macos
    - `fyne-cross` giving issues due to go1.19
- Install `upx` to make the binary super small
    - small exe cmd = `go build -ldflags="-s -w -extldflags='-static'" && strip blue && upx --ultra-brute blue`

## Notes

- bundler still not working perfectly
    - probably wont work with ui unless its setup
    - needs the file to be built in the same dir as this project

## Bugs

- Currently have a race condition/concurrency issue happening
