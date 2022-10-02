(function() {
    var wasm;
    const __exports = {};
    /**
    * @param {number} arg0
    * @param {number} arg1
    * @returns {void}
    */
    __exports.gen_maze_rust_silent = function(arg0, arg1) {
        return wasm.gen_maze_rust_silent(arg0, arg1);
    };

    const __wbg_random_8cdd17579946bb97_target = Math.random.bind(Math) || function() {
        throw new Error(`wasm-bindgen: Math.random.bind(Math) does not exist`);
    };

    __exports.__wbg_random_8cdd17579946bb97 = function() {
        return __wbg_random_8cdd17579946bb97_target();
    };

    function init(wasm_path) {
        const fetchPromise = fetch(wasm_path);
        let resultPromise;
        if (typeof WebAssembly.instantiateStreaming === 'function') {
            resultPromise = WebAssembly.instantiateStreaming(fetchPromise, { './rust': __exports });
        } else {
            resultPromise = fetchPromise
            .then(response => response.arrayBuffer())
            .then(buffer => WebAssembly.instantiate(buffer, { './rust': __exports }));
        }
        return resultPromise.then(({instance}) => {
            wasm = init.wasm = instance.exports;
            return;
        });
    };
    self.wasm_bindgen = Object.assign(init, __exports);
})();
