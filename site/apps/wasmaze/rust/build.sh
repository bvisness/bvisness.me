#!/bin/bash

mkdir -p dist
cargo build --target wasm32-unknown-unknown
wasm-bindgen target/wasm32-unknown-unknown/debug/rust.wasm --browser --no-modules --out-dir dist

