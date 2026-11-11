package main

// The benchmark harness writes a new value into this file in a commit of its own before each warm run, so the
// build step gets new input and no provider can skip it. Do not edit it by hand.
const benchTouch = "initial"
