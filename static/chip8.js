const elem = document.getElementById.bind(document)
const go = new Go()
const onWasmLoad = (result) => {
  go.run(result.instance)
  const romInput = elem('rom')

  romInput.addEventListener('change', () => {
    const file = romInput.files[0]
    const reader = new FileReader()

    reader.addEventListener('load', (event) => {
      const buffer = event.target.result
      const byteArray = new Uint8Array(buffer)

      createNewVm(...byteArray)
    })
    reader.readAsArrayBuffer(file)
  })

  // DOM element interactions
  elem('debug').addEventListener('change', toggleDebug)
  elem('next-inst').addEventListener('click', nextInst)
}

window.Chip8 = (() => {
  const display = elem('chip8-display')
  const ctx = display.getContext('2d')
  const createPixelGrid = () =>
    new Array(32).fill(0).map(() => new Array(64).fill(0))
  let pixels = createPixelGrid()
  const assembly = []

  // Draws on the specified pixel position. Returns result of xor 1
  // and the value of the pixel position before drawing. False means
  // that the pixel is being erased.
  const drawPixel = (x, y) => {
    // wrap the positions when they exceed the limit
    const xPos = x >= 64 ? x - 64 : x
    const yPos = y >= 32 ? y - 32 : y
    const result = 1 ^ pixels[yPos][xPos]
    pixels[yPos][xPos] = result
    return result
  }

  // Draw the binary representation of the desired picture given a byte.
  // E.g. the byte 0xF0 (0b11110000 in binary) draws 4 pixels.
  // x is the starting offset. Returns whether drawing the bytes
  // erases a pixel somewhere.
  const drawByte = (byte, x, y) => {
    let hasErased = false
    for (let shift = 7; shift >= 0; shift--) {
      if (byte & (2 ** shift)) {
        if (!drawPixel(x + 7 - shift, y)) {
          hasErased = true
        }
      }
    }
    return hasErased
  }

  // The CHIP-8 display is a 64x32-pixel display. In this implementation,
  // each pixel is scaled by 10 pixels, meaning that the entire display is
  // 640x320 pixels large.
  const drawPixels = () =>
    pixels.forEach((row, y) => {
      row.forEach((column, x) => {
        if (pixels[y][x] === 1) {
          ctx.beginPath()
          ctx.rect(x * 10, y * 10, 10, 10)
          ctx.fillStyle = 'black'
          ctx.fill()
          ctx.lineWidth = 1
          ctx.strokeStyle = 'white'
          ctx.stroke()
        }
      })
    })

  const runStateChangeHandler = state => {
    // if (state.romLoaded && state.inDebug) {
    //   elem('next-inst').disabled = false
    // } else {
    //   elem('next-inst').disabled = true
    // }
    // elem('debug-container').style.display = state.inDebug ? 'block' : 'none'
  }

  return {
    clearDisplay: () => {
      pixels = createPixelGrid()
      ctx.clearRect(0, 0, display.width, display.height)
    },
    draw: (x, y, bytes) => {
      let hasErased = false
      bytes.forEach((byte, vOffset) => {
        const erased = drawByte(byte, x, y + vOffset)
        if (erased) {
          hasErased = true
        }
      })

      drawPixels()
      return hasErased
    },
    waitForKeyPress: () => {
      const keyWaitingDiv = elem('debug-key-waiting')
      keyWaitingDiv.style.display = 'block'
      return new Promise((resolve) => {
        const listener = event => {
          const name = event.key
          const code = event.code
          // Alert the key name and key code on keydown
          console.log(`Key pressed ${name} \r\n Key code value: ${code}`)
          const keyByteMap = {
            1: 0x1,
            2: 0x2,
            3: 0x3,
            q: 0x4,
            w: 0x5,
            e: 0x6,
            a: 0x7,
            s: 0x8,
            z: 0xa,
            x: 0x0,
            c: 0xb,
            4: 0xc,
            r: 0xd,
            f: 0xe,
            v: 0xf
          }
          if (keyByteMap[name]) {
            document.removeEventListener('keydown', listener)
            resolve(keyByteMap[name])
          }
          keyWaitingDiv.style.display = 'none'
        }
        document.addEventListener('keydown', listener)
      })
    },
    onRunStateInit: runStateChangeHandler,
    onRunStateUpdate: runStateChangeHandler,
    onVmUpdate: (state) => {
      // show the register data
      ['Pc', 'Sp', 'I'].forEach(reg => {
        elem(reg).textContent = state[reg].toString(16)
      })
      const regTrs = elem('registers').children
      state.Regs.forEach((reg, i) => {
        const td = regTrs.item(i % 8).children.item(i < 8 ? 1 : 3)
        td.textContent = reg.toString(16)
      })

      // show the stack
      const stackTrs = elem('stack').children
      state.Stack.forEach((item, i) => {
        const tds = stackTrs.item(i).children
        tds.item(1).textContent = item.toString(16)
        Array.from(tds).forEach(td => {
          if (i <= state.Sp) {
            td.classList.remove('stack-inactive')
          } else {
            td.classList.add('stack-inactive')
          }
        })
      })
      assembly[state.Pc] = state.Assembly

      // show the instructions
      const list = elem('debug-instructions-list')
      const lis = list.getElementsByTagName('li')
      if (lis.length <= state.Pc) {
        for (let i = lis.length; i <= state.Pc; i++) {
          const li = document.createElement('li')
          list.appendChild(li)
        }
      }
      const previousActive = list.querySelector('.active-inst')
      if (previousActive) {
        previousActive.classList.remove('active-inst')
      }
      const newActive = lis.item(state.Pc)
      newActive.textContent = state.Assembly
      newActive.classList.add('active-inst')
      newActive.scrollIntoView({ block: 'center' })
    }
  }
})()

WebAssembly.instantiateStreaming(fetch('main.wasm'), go.importObject).then(
  onWasmLoad
)
