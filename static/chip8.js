const elem = document.getElementById.bind(document);
const go = new Go();
const onWasmLoad = (result) => {
	go.run(result.instance);
	const romInput = elem("rom");

	romInput.addEventListener("change", () => {
		const file = romInput.files[0];
		const reader = new FileReader();

		reader.addEventListener("load", (event) => {
			const buffer = event.target.result;
			const byteArray = new Uint8Array(buffer);

			createNewVm(...byteArray);
		});
		reader.readAsArrayBuffer(file);
	});

	// DOM element interactions
	elem("debug").addEventListener("change", toggleDebug);
	elem("next-inst").addEventListener("click", nextInst);
};

window.Chip8 = (() => {
	const display = elem("chip8-display");
	const ctx = display.getContext("2d");
	const assembly = [];
	let pixels = [];
	const keyByteMap = {
		1: 0x1,
		2: 0x2,
		3: 0x3,
		q: 0x4,
		w: 0x5,
		e: 0x6,
		a: 0x7,
		s: 0x8,
		d: 0x9,
		z: 0xa,
		x: 0x0,
		c: 0xb,
		4: 0xc,
		r: 0xd,
		f: 0xe,
		v: 0xf,
	};

	const runStateChangeHandler = (state) => {
		elem("debug").disabled = !state.romLoaded;
		elem("next-inst").disabled = !(state.romLoaded && state.inDebug);
		elem("debug-container").style.display = state.inDebug ? "block" : "none";
	};

	// Tell the VM which keys are being pressed
	["keydown", "keyup"].forEach((keyEvent) => {
		document.body.addEventListener(keyEvent, (event) => {
			const byte = keyByteMap[event.key];
			// explicit test for 0; it will be ignored otherwise
			if (byte || byte === 0) {
				setKey(byte, keyEvent === "keydown");
			}
		});
	});

	const help = elem("help");
	const helpToggle = elem("help-toggle");
	helpToggle.onclick = () => {
		if (help.style.display == "none" || help.style.display == "") {
			help.style.display = "block";
		} else {
			help.style.display = "none";
		}
	};

	// contains arrays of pixels that have been erased; first one is oldest
	let previousPixels = [];
	(() => {
		const alphaDelta = 0.2;
		const framesToShowErased = 3;
		// delete erased pixels that have already been shown for a
		// certain number of frames
		const deletePreviousPixels = () => {
			if (previousPixels[0]?.framesShown >= framesToShowErased) {
				previousPixels.shift();
				deletePreviousPixels();
			}
		};
		const draw = (alpha) => ([x, y]) => {
			ctx.globalAlpha = alpha;
			ctx.beginPath();
			ctx.rect(x * 10, y * 10, 10, 10); // scale each CHIP-8 pixel by 10 pixels
			ctx.fillStyle = "black";
			ctx.fill();
			ctx.lineWidth = 1;
			ctx.strokeStyle = "white";
			ctx.stroke();
		};
		const drawPixels = () => {
			ctx.clearRect(0, 0, display.width, display.height);
			deletePreviousPixels();
			// draw the erased pixels for a certain number of frames in order to reduce
			// flickering; for each frame the erased pixel fades out a little
			previousPixels.forEach((pixelSet) => {
				pixelSet.pixels.forEach(draw(pixelSet.alpha));
				pixelSet.framesShown++;
				pixelSet.alpha -= alphaDelta;
			});
			// draw the newest set of pixels
			pixels.forEach(draw(1));
			window.requestAnimationFrame(drawPixels);
		};
		window.requestAnimationFrame(drawPixels);
	})();

	const findErasedPixels = (curPixels, newPixels) => {
		const pixelMap = newPixels.reduce((acc, [x, y]) => {
			if (!acc[x]) {
				acc[x] = {};
			}
			acc[x][y] = 1;
			return acc;
		}, {});

		return curPixels.reduce((acc, pixels) => {
			const [x, y] = pixels;
			if (!pixelMap[x]?.[y]) {
				acc.push(pixels);
			}
			return acc;
		}, []);
	};

	return {
		clearDisplay: () => {
			pixels = [];
			ctx.clearRect(0, 0, display.width, display.height);
		},
		draw: (vmPixels) => {
			previousPixels.push({
				pixels: findErasedPixels(pixels, vmPixels),
				framesShown: 0,
				alpha: 0.95,
			});
			pixels = vmPixels;
		},
		waitForKeyPress: () => {
			const keyWaitingDiv = elem("debug-key-waiting");
			keyWaitingDiv.style.display = "block";
			return {
				onRelease: (keyReleased) => {
					console.log(`Key released: ${keyReleased}`);
					keyWaitingDiv.style.display = "none";
				},
			};
		},
		onRunStateInit: runStateChangeHandler,
		onRunStateUpdate: runStateChangeHandler,
		onVmUpdate: (state) => {
			// show the register data
			["Pc", "Sp", "I", "DT"].forEach((reg) => {
				elem(reg).textContent = state[reg].toString(16);
			});
			const regTrs = elem("registers").children;
			state.Regs.forEach((reg, i) => {
				const td = regTrs.item(i % 8).children.item(i < 8 ? 1 : 3);
				td.textContent = reg.toString(16);
			});

			// show the stack
			const stackTrs = elem("stack").children;
			state.Stack.forEach((item, i) => {
				const tds = stackTrs.item(i).children;
				tds.item(1).textContent = item.toString(16);
				Array.from(tds).forEach((td) => {
					if (i <= state.Sp) {
						td.classList.remove("stack-inactive");
					} else {
						td.classList.add("stack-inactive");
					}
				});
			});
			assembly[state.Pc] = state.Assembly;

			// show the instructions
			const list = elem("debug-instructions-list");
			const lis = list.getElementsByTagName("li");
			if (lis.length <= state.Pc) {
				for (let i = lis.length; i <= state.Pc; i++) {
					const li = document.createElement("li");
					list.appendChild(li);
				}
			}
			const previousActive = list.querySelector(".active-inst");
			if (previousActive) {
				previousActive.classList.remove("active-inst");
			}
			const newActive = lis.item(state.Pc);
			newActive.textContent = state.Assembly;
			newActive.classList.add("active-inst");
			newActive.scrollIntoView({ block: "center" });
		},
	};
})();

WebAssembly.instantiateStreaming(fetch("main.wasm"), go.importObject).then(
	onWasmLoad,
);
