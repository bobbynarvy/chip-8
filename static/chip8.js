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
	const clearDisplay = () => ctx.clearRect(0, 0, display.width, display.height);
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
			if (byte) {
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

	return {
		clearDisplay,
		// The CHIP-8 display is a 64x32-pixel display. In this implementation,
		// each pixel is scaled by 10 pixels, meaning that the entire display is
		// 640x320 pixels large.
		draw: (pixels) => {
			clearDisplay();
			pixels.forEach((row, y) => {
				row.forEach((column, x) => {
					if (pixels[y][x] === 1) {
						ctx.beginPath();
						ctx.rect(x * 10, y * 10, 10, 10);
						ctx.fillStyle = "black";
						ctx.fill();
						ctx.lineWidth = 1;
						ctx.strokeStyle = "white";
						ctx.stroke();
					}
				});
			});
		},
		waitForKeyPress: () => {
			const keyWaitingDiv = elem("debug-key-waiting");
			keyWaitingDiv.style.display = "block";
			return new Promise((resolve) => {
				const listener = (event) => {
					const name = event.key;
					const code = event.code;
					// Alert the key name and key code on keydown
					console.log(`Key pressed ${name} \r\n Key code value: ${code}`);
					if (keyByteMap[name]) {
						document.removeEventListener("keydown", listener);
						resolve(keyByteMap[name]);
					}
					keyWaitingDiv.style.display = "none";
				};
				document.addEventListener("keydown", listener);
			});
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
