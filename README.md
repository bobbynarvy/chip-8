# chip-8
A CHIP-8 Emulator

Try it out [here](https://bobbynarvy.github.io/chip-8/).

The purpose of this project is to learn about emulation, compiling Go code to WebAssembly and using it in the browser.

## Building

```
make
```

This produces a `main.wasm` binary in the `static` directory.

## Local development

Build and execute the package in the `server` directory. This will launch an HTTP server that listens to port `3000` and
serves all the assets required to run the emulator.

## Requirement

- Go 1.20

## Useful links

- [Cowgod's Chip-8 Technical Reference v1.0](http://devernay.free.fr/hacks/chip8/C8TECH10.HTM)
- [CHIP-8 ROMs](https://github.com/kripod/chip8-roms/tree/master/games)
