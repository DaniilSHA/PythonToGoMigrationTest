package main

import "PythonToGoMigrationTest/internal/calculator"

// --host 127.0.0.1 --port 8080 --c-lib ./source/libcalculator.so --rust-lib ./source/libcalculator_rust.so --interval 5
func main() {
	calculator.Start()
}
